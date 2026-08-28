package plugin

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"testing"

	"github.com/drpcorg/grafana-chotki-datasource/pkg/api"
	"github.com/google/uuid"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"google.golang.org/grpc"
)

// fakeAggregator implements just enough of AggregatorService to exercise the
// datasource end-to-end over a real gRPC connection.
type fakeAggregator struct {
	api.UnimplementedAggregatorServiceServer

	blockHit bool
}

func (f *fakeAggregator) GetAuthSnapshot(ctx context.Context, req *api.GetAuthSnapshotRequest) (*api.GetAuthSnapshotResponse, error) {
	return &api.GetAuthSnapshotResponse{
		Snapshot: &api.AuthSnapshot{
			Owner:            &api.Owner{OwnerID: req.OwnerId, Discounts: map[string]int32{"default": 10}, BillingVersion: 1},
			Key:              &api.Key{KeyId: req.KeyId},
			OwnerTotalCU:     100,
			OwnerFreeBalance: 25,
			KeyDailyUsedCU:   5,
			PackagePools: []*api.PackagePoolSnapshot{
				{Tag: "pro", Credited: 500, Spent: 200},
			},
		},
	}, nil
}

func (f *fakeAggregator) GetAuthSnapshots(ctx context.Context, req *api.GetAuthSnapshotsRequest) (*api.GetAuthSnapshotsResponse, error) {
	results := make([]*api.AuthSnapshotResult, 0, len(req.Refs))
	for i := range req.Refs {
		if i == 0 {
			results = append(results, &api.AuthSnapshotResult{
				Snapshot: &api.AuthSnapshot{OwnerTotalCU: 7, OwnerFreeBalance: 3, KeyDailyUsedCU: 2},
			})
			continue
		}
		results = append(results, &api.AuthSnapshotResult{Error: "not found"})
	}
	return &api.GetAuthSnapshotsResponse{Results: results}, nil
}

func (f *fakeAggregator) GetPackagePools(ctx context.Context, req *api.GetPackagePoolsRequest) (*api.GetPackagePoolsResponse, error) {
	return &api.GetPackagePoolsResponse{
		Pools: []*api.PackagePoolSnapshot{
			{Tag: "pro", Credited: 500, Spent: 200, PeriodEnd: "2026-12-31", AddonId: 1, SubscriptionId: 9},
		},
	}, nil
}

func (f *fakeAggregator) BlockHeight(ctx context.Context, req *api.BlockHeighRequest) (*api.BlockHeightResponse, error) {
	f.blockHit = true
	return &api.BlockHeightResponse{Found: true, Height: 123}, nil
}

func startFakeAggregator(t *testing.T) (*fakeAggregator, string) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := grpc.NewServer()
	fake := &fakeAggregator{}
	api.RegisterAggregatorServiceServer(server, fake)
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)

	return fake, listener.Addr().String()
}

func newTestDatasource(t *testing.T, addr string) *Datasource {
	t.Helper()

	settings := backend.DataSourceInstanceSettings{
		JSONData: json.RawMessage(`{"grpcAddress":"` + addr + `","insecure":true,"timeoutMs":4000}`),
	}
	instance, err := NewDatasource(context.Background(), settings)
	if err != nil {
		t.Fatalf("NewDatasource: %v", err)
	}
	ds, ok := instance.(*Datasource)
	if !ok {
		t.Fatalf("unexpected instance type %T", instance)
	}
	t.Cleanup(ds.Dispose)
	return ds
}

func runQuery(t *testing.T, ds *Datasource, query string) backend.DataResponse {
	t.Helper()

	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: json.RawMessage(query)}},
	}
	resp, err := ds.QueryData(context.Background(), req)
	if err != nil {
		t.Fatalf("QueryData: %v", err)
	}
	return resp.Responses["A"]
}

func TestQueryData_GetAuthSnapshot(t *testing.T) {
	_, addr := startFakeAggregator(t)
	ds := newTestDatasource(t, addr)
	ownerID, keyID := uuid.New().String(), uuid.New().String()

	query := `{"mode":"rpc","method":"GetAuthSnapshot","params":{"ownerId":"` + ownerID + `","keyId":"` + keyID + `"}}`
	response := runQuery(t, ds, query)
	if response.Error != nil {
		t.Fatalf("query error: %v", response.Error)
	}
	frame := response.Frames[0]
	if frame.Name != "get_auth_snapshot" || frame.Rows() != 1 {
		t.Fatalf("unexpected frame: %s rows=%d", frame.Name, frame.Rows())
	}
	if got, _ := frame.Fields[0].ConcreteAt(0); got.(string) != ownerID {
		t.Fatalf("unexpected owner_id: %v", got)
	}
	if got, _ := frame.Fields[4].ConcreteAt(0); got.(int64) != 5 {
		t.Fatalf("unexpected key_daily_used_cu: %v", got)
	}
}

func TestQueryData_GetAuthSnapshots(t *testing.T) {
	_, addr := startFakeAggregator(t)
	ds := newTestDatasource(t, addr)
	ownerID, keyID := uuid.New().String(), uuid.New().String()

	query := `{"mode":"rpc","method":"GetAuthSnapshots","params":{"refs":"` + ownerID + `:` + keyID + `,` + ownerID + `:` + keyID + `"}}`
	response := runQuery(t, ds, query)
	if response.Error != nil {
		t.Fatalf("query error: %v", response.Error)
	}
	frame := response.Frames[0]
	if frame.Name != "get_auth_snapshots" || frame.Rows() != 2 {
		t.Fatalf("unexpected frame: %s rows=%d", frame.Name, frame.Rows())
	}
	if got, _ := frame.Fields[2].ConcreteAt(0); got.(string) != "" {
		t.Fatalf("expected empty error on row 0, got %v", got)
	}
	if got, _ := frame.Fields[2].ConcreteAt(1); got.(string) != "not found" {
		t.Fatalf("expected 'not found' on row 1, got %v", got)
	}
}

func TestQueryData_GetPackagePools(t *testing.T) {
	_, addr := startFakeAggregator(t)
	ds := newTestDatasource(t, addr)
	ownerID := uuid.New().String()

	query := `{"mode":"rpc","method":"GetPackagePools","params":{"ownerId":"` + ownerID + `"}}`
	response := runQuery(t, ds, query)
	if response.Error != nil {
		t.Fatalf("query error: %v", response.Error)
	}
	frame := response.Frames[0]
	if frame.Name != "get_package_pools" || frame.Rows() != 1 {
		t.Fatalf("unexpected frame: %s rows=%d", frame.Name, frame.Rows())
	}
	if got, _ := frame.Fields[4].ConcreteAt(0); got.(int64) != 300 {
		t.Fatalf("unexpected remaining: %v", got)
	}
}

func TestCheckHealth_UsesBlockHeight(t *testing.T) {
	fake, addr := startFakeAggregator(t)
	ds := newTestDatasource(t, addr)

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth: %v", err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Fatalf("unexpected status: %s (%s)", result.Status, result.Message)
	}
	if !fake.blockHit {
		t.Fatalf("expected BlockHeight to be called for health check")
	}
}

func TestQueryData_RejectsWriteMethods(t *testing.T) {
	_, addr := startFakeAggregator(t)
	ds := newTestDatasource(t, addr)

	response := runQuery(t, ds, `{"mode":"rpc","method":"BillOwnerBalance","params":{}}`)
	if response.Error == nil {
		t.Fatalf("expected error for write method")
	}
}

func TestQueryData_ListNodeCoreKeys_ExecutesRPC(t *testing.T) {
	_, addr := startFakeAggregator(t)
	ds := newTestDatasource(t, addr)
	ownerID := uuid.New().String()

	// The fake server does not implement ListNodeCoreKeys, so the RPC returns
	// Unimplemented — but the query must get past parsing and reach the RPC
	// (guards against the handler case silently disappearing again).
	response := runQuery(t, ds, `{"mode":"rpc","method":"ListNodeCoreKeys","params":{"ownerId":"`+ownerID+`"}}`)
	if response.Error == nil {
		t.Fatalf("expected RPC-level error from unimplemented method")
	}
	if strings.Contains(response.Error.Error(), "not allowed") || strings.Contains(response.Error.Error(), "unsupported method") {
		t.Fatalf("method should be allowed, got: %v", response.Error)
	}
}
