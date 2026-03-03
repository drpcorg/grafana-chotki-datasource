package plugin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/drpcorg/grafana-chotki-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	api "github.com/drpcorg/grafana-chotki-datasource/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

type methodStats struct {
	Calls         int64
	Errors        int64
	LastLatencyMs int64
}

// Datasource handles gRPC read RPC execution for AggregatorService.
type Datasource struct {
	settings *models.PluginSettings

	conn   *grpc.ClientConn
	client api.AggregatorServiceClient

	initErr error

	mu    sync.Mutex
	stats map[string]*methodStats
}

func NewDatasource(ctx context.Context, source backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := models.LoadPluginSettings(source)
	if err != nil {
		return nil, err
	}

	ds := &Datasource{
		settings: settings,
		stats:    map[string]*methodStats{},
	}

	if strings.TrimSpace(settings.GRPCAddress) == "" {
		return ds, nil
	}

	conn, err := dialGRPC(ctx, settings)
	if err != nil {
		ds.initErr = err
		return ds, nil
	}

	ds.conn = conn
	ds.client = api.NewAggregatorServiceClient(conn)
	return ds, nil
}

func dialGRPC(ctx context.Context, settings *models.PluginSettings) (*grpc.ClientConn, error) {
	if settings == nil {
		return nil, fmt.Errorf("settings are required")
	}
	if strings.TrimSpace(settings.GRPCAddress) == "" {
		return nil, fmt.Errorf("grpcAddress is required")
	}

	transportCreds, err := buildTransportCredentials(settings)
	if err != nil {
		return nil, err
	}

	dialCtx, cancel := context.WithTimeout(ctx, settings.RequestTimeout())
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32 << 20)),
		grpc.WithBlock(), //nolint:staticcheck // Keep blocking dial behavior to fail fast on invalid datasource config.
	}

	//nolint:staticcheck // We intentionally use DialContext for blocking init semantics in datasource construction.
	conn, err := grpc.DialContext(dialCtx, settings.GRPCAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("gRPC dial %s: %w", settings.GRPCAddress, err)
	}
	return conn, nil
}

func buildTransportCredentials(settings *models.PluginSettings) (credentials.TransportCredentials, error) {
	if settings.Insecure {
		return insecure.NewCredentials(), nil
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if strings.TrimSpace(settings.Secrets.TLSCACert) != "" {
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(settings.Secrets.TLSCACert)); !ok {
			return nil, fmt.Errorf("tlsCACert is not a valid PEM certificate")
		}
		tlsCfg.RootCAs = pool
	}

	clientCert := strings.TrimSpace(settings.Secrets.TLSClientCert)
	clientKey := strings.TrimSpace(settings.Secrets.TLSClientKey)
	if clientCert != "" || clientKey != "" {
		if clientCert == "" || clientKey == "" {
			return nil, fmt.Errorf("both tlsClientCert and tlsClientKey are required together")
		}
		pair, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
		if err != nil {
			return nil, fmt.Errorf("invalid client TLS key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{pair}
	}

	return credentials.NewTLS(tlsCfg), nil
}

func (d *Datasource) Dispose() {
	if d.conn != nil {
		_ = d.conn.Close()
	}
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, query := range req.Queries {
		response.Responses[query.RefID] = d.query(ctx, query)
	}
	return response, nil
}

func (d *Datasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	if err := d.ensureReady(); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	qm, opts, err := parseQueryModel(query.JSON, query.RefID, d.settings)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	requestCtx, cancel := context.WithTimeout(ctx, d.settings.RequestTimeout())
	defer cancel()

	frame, statValue, err := d.executeRPC(requestCtx, qm, opts)
	if err != nil {
		return grpcErrorToResponse(err)
	}

	if opts.Format == "stat" {
		return backend.DataResponse{Frames: []*data.Frame{buildStatFrame(qm.Method, statValue)}}
	}
	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) executeRPC(ctx context.Context, qm *queryModel, opts queryExecOptions) (*data.Frame, float64, error) {
	switch qm.Method {
	case methodGetOwner:
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetOwner(callCtx, &api.GetOwnerRequest{OwnerId: ownerID})
			if err != nil {
				return nil, 0, err
			}
			return buildGetOwnerFrame(resp.GetOwner(), opts)
		})

	case methodGetFullOwner:
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		loadBalance, err := getBoolParam(qm.Params, false, "loadBalance", "load_balance")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetFullOwner(callCtx, &api.GetFullOwnerRequest{OwnerId: ownerID, LoadBalance: loadBalance})
			if err != nil {
				return nil, 0, err
			}
			return buildGetFullOwnerFrame(resp.GetOwner(), opts)
		})

	case methodGetOwnerHits:
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetOwnerHits(callCtx, &api.GetOwnerHitsRequest{OwnerId: ownerID})
			if err != nil {
				return nil, 0, err
			}
			frame, statValue := buildGetOwnerHitsFrame(ownerID, resp.GetHits(), opts)
			return frame, statValue, nil
		})

	case methodGetOwnerMetadata:
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetOwnerMetadata(callCtx, &api.GetOwnerMetadataRequest{OwnerId: ownerID})
			if err != nil {
				return nil, 0, err
			}
			return buildGetOwnerMetadataFrame(resp.GetMetadata(), opts)
		})

	case methodGetKey:
		keyID, err := getRequiredUUIDParam(qm.Params, "keyId", "key_id", "keyID")
		if err != nil {
			return nil, 0, err
		}
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetKey(callCtx, &api.GetKeyRequest{KeyId: keyID, OwnerId: ownerID})
			if err != nil {
				return nil, 0, err
			}
			return buildGetKeyFrame(resp.GetKey(), opts)
		})

	case methodGetKeyHits:
		keyID, err := getRequiredUUIDParam(qm.Params, "keyId", "key_id", "keyID")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetKeyHits(callCtx, &api.GetKeyHitsRequest{KeyId: keyID})
			if err != nil {
				return nil, 0, err
			}
			frame, statValue := buildGetKeyHitsFrame(keyID, resp.GetHits(), opts)
			return frame, statValue, nil
		})

	case methodListKeys:
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		lastKeyID, hasLastKeyID, err := getOptionalUUIDParam(qm.Params, "lastKeyId", "last_key_id", "lastKeyID")
		if err != nil {
			return nil, 0, err
		}
		limit := opts.Limit
		if overrideLimit, ok, err := getOptionalInt64Param(qm.Params, "limit"); err != nil {
			return nil, 0, err
		} else if ok {
			limit = d.settings.ClampLimit(overrideLimit)
		}

		request := &api.ListKeysRequest{OwnerId: ownerID, Limit: limit}
		if hasLastKeyID {
			request.LastKeyId = lastKeyID
		}

		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.ListKeys(callCtx, request)
			if err != nil {
				return nil, 0, err
			}
			return buildListKeysFrame(resp.GetKeys(), opts)
		})

	case methodGetOwnersWithBalance:
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetOwnersWithBalance(callCtx, &api.GetOwnersWithBalanceRequest{})
			if err != nil {
				return nil, 0, err
			}
			frame, statValue := buildGetOwnersWithBalanceFrame(resp.GetBalances(), opts)
			return frame, statValue, nil
		})

	case methodGetAllOwnerIDs:
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetAllOwnerIds(callCtx, &api.GetAllOwnerIdsRequest{})
			if err != nil {
				return nil, 0, err
			}
			frame, statValue := buildGetAllOwnerIDsFrame(resp.GetOwnerIds(), opts)
			return frame, statValue, nil
		})

	case methodGetNodeCoreKey:
		keyID, err := getRequiredUUIDParam(qm.Params, "keyId", "key_id", "keyID")
		if err != nil {
			return nil, 0, err
		}
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.GetNodeCoreKey(callCtx, &api.GetNodeCoreKeyRequest{KeyId: keyID, OwnerId: ownerID})
			if err != nil {
				return nil, 0, err
			}
			return buildGetNodeCoreKeyFrame(resp.GetKey(), opts)
		})

	case methodListNodeCoreKeys:
		ownerID, err := getRequiredUUIDParam(qm.Params, "ownerId", "owner_id", "ownerID")
		if err != nil {
			return nil, 0, err
		}
		lastKeyID, hasLastKeyID, err := getOptionalUUIDParam(qm.Params, "lastKeyId", "last_key_id", "lastKeyID")
		if err != nil {
			return nil, 0, err
		}
		limit := opts.Limit
		if overrideLimit, ok, err := getOptionalInt64Param(qm.Params, "limit"); err != nil {
			return nil, 0, err
		} else if ok {
			limit = d.settings.ClampLimit(overrideLimit)
		}

		request := &api.ListNodeCoreKeysRequest{OwnerId: ownerID, Limit: limit}
		if hasLastKeyID {
			request.LastKeyId = lastKeyID
		}

		return d.callRead(ctx, qm.Method, func(callCtx context.Context) (*data.Frame, float64, error) {
			resp, err := d.client.ListNodeCoreKeys(callCtx, request)
			if err != nil {
				return nil, 0, err
			}
			return buildListNodeCoreKeysFrame(resp.GetKeys(), opts)
		})
	}

	return nil, 0, fmt.Errorf("unsupported method %q", qm.Method)
}

func (d *Datasource) callRead(ctx context.Context, method string, fn func(context.Context) (*data.Frame, float64, error)) (*data.Frame, float64, error) {
	const maxAttempts = 2

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		callCtx := d.withAuth(ctx)
		startedAt := time.Now()
		frame, statValue, err := fn(callCtx)
		d.recordMethodCall(method, time.Since(startedAt), err)
		if err == nil {
			return frame, statValue, nil
		}

		lastErr = err
		statusErr, ok := status.FromError(err)
		retryable := ok && (statusErr.Code() == codes.Unavailable || statusErr.Code() == codes.ResourceExhausted)
		if !retryable || attempt == maxAttempts {
			return nil, 0, err
		}

		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		case <-time.After(120 * time.Millisecond):
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("read request failed")
	}
	return nil, 0, lastErr
}

func (d *Datasource) withAuth(ctx context.Context) context.Context {
	token := strings.TrimSpace(d.settings.Secrets.AuthToken)
	if token == "" {
		return ctx
	}
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", token)
}

func (d *Datasource) recordMethodCall(method string, latency time.Duration, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	metric := d.stats[method]
	if metric == nil {
		metric = &methodStats{}
		d.stats[method] = metric
	}
	metric.Calls++
	metric.LastLatencyMs = latency.Milliseconds()
	if err != nil {
		metric.Errors++
	}

	if err != nil {
		backend.Logger.Warn("grafana chotki datasource query failed", "method", method, "latency_ms", metric.LastLatencyMs, "error", err.Error())
		return
	}
	backend.Logger.Debug("grafana chotki datasource query succeeded", "method", method, "latency_ms", metric.LastLatencyMs)
}

func (d *Datasource) ensureReady() error {
	if d == nil {
		return fmt.Errorf("datasource is not initialized")
	}
	if d.settings == nil {
		return fmt.Errorf("datasource settings are not loaded")
	}
	if strings.TrimSpace(d.settings.GRPCAddress) == "" {
		return fmt.Errorf("grpcAddress is required")
	}
	if d.initErr != nil {
		return fmt.Errorf("gRPC initialization failed: %w", d.initErr)
	}
	if d.client == nil {
		return fmt.Errorf("gRPC client is not ready")
	}
	return nil
}

func (d *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if err := d.ensureReady(); err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}

	healthTimeout := d.settings.RequestTimeout()
	if healthTimeout > 2*time.Second {
		healthTimeout = 2 * time.Second
	}
	healthCtx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()

	_, err := d.client.GetAllOwnerIds(d.withAuth(healthCtx), &api.GetAllOwnerIdsRequest{})
	if err != nil {
		return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: err.Error()}, nil
	}

	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: "connected to AggregatorService"}, nil
}

func grpcErrorToResponse(err error) backend.DataResponse {
	if err == nil {
		return backend.DataResponse{}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return backend.ErrDataResponse(backend.StatusTimeout, err.Error())
	}

	statusErr, ok := status.FromError(err)
	if !ok {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}

	switch statusErr.Code() {
	case codes.InvalidArgument, codes.FailedPrecondition:
		return backend.ErrDataResponse(backend.StatusBadRequest, statusErr.Message())
	case codes.NotFound:
		return backend.ErrDataResponse(backend.StatusNotFound, statusErr.Message())
	case codes.DeadlineExceeded:
		return backend.ErrDataResponse(backend.StatusTimeout, statusErr.Message())
	case codes.Unavailable:
		return backend.ErrDataResponse(backend.StatusBadGateway, statusErr.Message())
	case codes.Unauthenticated, codes.PermissionDenied:
		return backend.ErrDataResponse(backend.StatusUnauthorized, statusErr.Message())
	default:
		return backend.ErrDataResponse(backend.StatusInternal, statusErr.Message())
	}
}
