package plugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/drpcorg/grafana-chotki-datasource/pkg/api"
	"github.com/drpcorg/grafana-chotki-datasource/pkg/models"
	"github.com/google/uuid"
)

const (
	methodGetOwner             = "GetOwner"
	methodGetFullOwner         = "GetFullOwner"
	methodGetOwnerHits         = "GetOwnerHits"
	methodGetOwnerMetadata     = "GetOwnerMetadata"
	methodGetKey               = "GetKey"
	methodGetKeyHits           = "GetKeyHits"
	methodListKeys             = "ListKeys"
	methodGetOwnersWithBalance = "GetOwnersWithBalance"
	methodGetAllOwnerIDs       = "GetAllOwnerIds"
	methodGetNodeCoreKey       = "GetNodeCoreKey"
	methodListNodeCoreKeys     = "ListNodeCoreKeys"
	methodGetAuthSnapshot      = "GetAuthSnapshot"
	methodGetAuthSnapshots     = "GetAuthSnapshots"
	methodGetPackagePools      = "GetPackagePools"
)

var allowedMethods = map[string]struct{}{
	methodGetOwner:             {},
	methodGetFullOwner:         {},
	methodGetOwnerHits:         {},
	methodGetOwnerMetadata:     {},
	methodGetKey:               {},
	methodGetKeyHits:           {},
	methodListKeys:             {},
	methodGetOwnersWithBalance: {},
	methodGetAllOwnerIDs:       {},
	methodGetNodeCoreKey:       {},
	methodListNodeCoreKeys:     {},
	methodGetAuthSnapshot:      {},
	methodGetAuthSnapshots:     {},
	methodGetPackagePools:      {},
}

type queryModel struct {
	RefID      string                 `json:"refId"`
	Mode       string                 `json:"mode"`
	Method     string                 `json:"method"`
	Params     map[string]any         `json:"params"`
	Options    *queryOptions          `json:"options,omitempty"`
	EditorMode string                 `json:"editorMode,omitempty"`
	RawQuery   string                 `json:"rawQuery,omitempty"`
	Raw        map[string]interface{} `json:"-"`
}

type queryOptions struct {
	Format           string `json:"format,omitempty"`
	DecodeIDs        *bool  `json:"decodeIds,omitempty"`
	DecodeEnums      *bool  `json:"decodeEnums,omitempty"`
	DecodeTimestamps *bool  `json:"decodeTimestamps,omitempty"`
	Limit            *int64 `json:"limit,omitempty"`
}

type queryExecOptions struct {
	Format           string
	DecodeIDs        bool
	DecodeEnums      bool
	DecodeTimestamps bool
	Limit            int64
}

func parseQueryModel(raw []byte, refID string, settings *models.PluginSettings) (*queryModel, queryExecOptions, error) {
	if settings == nil {
		return nil, queryExecOptions{}, fmt.Errorf("settings are required")
	}

	var qm queryModel
	if err := json.Unmarshal(raw, &qm); err != nil {
		return nil, queryExecOptions{}, fmt.Errorf("json unmarshal: %w", err)
	}

	if qm.RefID == "" {
		qm.RefID = refID
	}
	if err := qm.mergeRawQuery(); err != nil {
		return nil, queryExecOptions{}, err
	}

	qm.Mode = strings.TrimSpace(qm.Mode)
	if qm.Mode == "" {
		qm.Mode = "rpc"
	}
	if qm.Mode != "rpc" {
		return nil, queryExecOptions{}, fmt.Errorf("unsupported mode %q", qm.Mode)
	}

	qm.Method = strings.TrimSpace(qm.Method)
	if qm.Method == "" {
		return nil, queryExecOptions{}, fmt.Errorf("method is required")
	}
	if _, ok := allowedMethods[qm.Method]; !ok {
		return nil, queryExecOptions{}, fmt.Errorf("method %q is not allowed", qm.Method)
	}

	if qm.Params == nil {
		qm.Params = map[string]any{}
	}

	opts, err := resolveOptions(qm.Options, settings)
	if err != nil {
		return nil, queryExecOptions{}, err
	}

	return &qm, opts, nil
}

func (qm *queryModel) mergeRawQuery() error {
	raw := strings.TrimSpace(qm.RawQuery)
	if raw == "" {
		return nil
	}

	if qm.Method != "" && qm.EditorMode != "raw" {
		return nil
	}

	var rawObj map[string]any
	if err := json.Unmarshal([]byte(raw), &rawObj); err != nil {
		return fmt.Errorf("rawQuery parse error: %w", err)
	}
	qm.Raw = rawObj

	if v, ok := rawObj["mode"].(string); ok && strings.TrimSpace(v) != "" {
		qm.Mode = v
	}
	if v, ok := rawObj["method"].(string); ok && strings.TrimSpace(v) != "" {
		qm.Method = v
	}
	if v, ok := rawObj["params"].(map[string]any); ok && len(v) > 0 {
		qm.Params = v
	}

	if rawOptions, ok := rawObj["options"]; ok {
		marshalled, err := json.Marshal(rawOptions)
		if err != nil {
			return fmt.Errorf("rawQuery options marshal: %w", err)
		}

		var options queryOptions
		if err := json.Unmarshal(marshalled, &options); err != nil {
			return fmt.Errorf("rawQuery options parse: %w", err)
		}
		qm.Options = &options
	}

	return nil
}

func resolveOptions(raw *queryOptions, settings *models.PluginSettings) (queryExecOptions, error) {
	opts := queryExecOptions{
		Format:           "table",
		DecodeIDs:        settings.DecodeIDs,
		DecodeEnums:      settings.DecodeEnums,
		DecodeTimestamps: settings.DecodeTimestamps,
		Limit:            int64(settings.DefaultLimit),
	}

	if raw == nil {
		opts.Limit = settings.ClampLimit(opts.Limit)
		return opts, nil
	}

	if raw.Format != "" {
		opts.Format = strings.ToLower(strings.TrimSpace(raw.Format))
	}
	if opts.Format != "table" && opts.Format != "stat" {
		return queryExecOptions{}, fmt.Errorf("options.format must be table or stat")
	}

	if raw.DecodeIDs != nil {
		opts.DecodeIDs = *raw.DecodeIDs
	}
	if raw.DecodeEnums != nil {
		opts.DecodeEnums = *raw.DecodeEnums
	}
	if raw.DecodeTimestamps != nil {
		opts.DecodeTimestamps = *raw.DecodeTimestamps
	}
	if raw.Limit != nil {
		opts.Limit = *raw.Limit
	}

	opts.Limit = settings.ClampLimit(opts.Limit)
	return opts, nil
}

func getParam(params map[string]any, aliases ...string) (any, bool) {
	for _, alias := range aliases {
		if v, ok := params[alias]; ok {
			return v, true
		}
	}
	return nil, false
}

func getRequiredUUIDParam(params map[string]any, aliases ...string) ([]byte, error) {
	value, ok := getParam(params, aliases...)
	if !ok {
		return nil, fmt.Errorf("missing required parameter %q", aliases[0])
	}
	return parseUUIDLike(value, aliases[0])
}

func getOptionalUUIDParam(params map[string]any, aliases ...string) ([]byte, bool, error) {
	value, ok := getParam(params, aliases...)
	if !ok {
		return nil, false, nil
	}
	parsed, err := parseUUIDLike(value, aliases[0])
	if err != nil {
		return nil, false, err
	}
	return parsed, true, nil
}

func parseUUIDLike(value any, paramName string) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return parseUUIDStringOrBase64(v, paramName)
	case []byte:
		if len(v) == 0 {
			return nil, fmt.Errorf("parameter %q cannot be empty", paramName)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("parameter %q must be a UUID/base64 string", paramName)
	}
}

func parseUUIDStringOrBase64(value string, paramName string) ([]byte, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, fmt.Errorf("parameter %q cannot be empty", paramName)
	}

	if parsedUUID, err := uuid.Parse(trimmed); err == nil {
		bytesValue := make([]byte, len(parsedUUID))
		copy(bytesValue, parsedUUID[:])
		return bytesValue, nil
	}

	if raw, err := base64.StdEncoding.DecodeString(trimmed); err == nil && len(raw) > 0 {
		return raw, nil
	}
	if raw, err := base64.RawStdEncoding.DecodeString(trimmed); err == nil && len(raw) > 0 {
		return raw, nil
	}

	return nil, fmt.Errorf("parameter %q must be UUID or base64 bytes", paramName)
}

func getBoolParam(params map[string]any, defaultValue bool, aliases ...string) (bool, error) {
	value, ok := getParam(params, aliases...)
	if !ok {
		return defaultValue, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return defaultValue, fmt.Errorf("parameter %q must be boolean", aliases[0])
		}
		return parsed, nil
	default:
		return defaultValue, fmt.Errorf("parameter %q must be boolean", aliases[0])
	}
}

func getOptionalInt64Param(params map[string]any, aliases ...string) (int64, bool, error) {
	value, ok := getParam(params, aliases...)
	if !ok {
		return 0, false, nil
	}

	switch v := value.(type) {
	case int:
		return int64(v), true, nil
	case int32:
		return int64(v), true, nil
	case int64:
		return v, true, nil
	case float64:
		return int64(v), true, nil
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, false, fmt.Errorf("parameter %q must be int64", aliases[0])
		}
		return parsed, true, nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0, false, fmt.Errorf("parameter %q must be int64", aliases[0])
		}
		return parsed, true, nil
	default:
		return 0, false, fmt.Errorf("parameter %q must be int64", aliases[0])
	}
}

// parseAuthRefs builds AggregatorService.AuthRef pairs for GetAuthSnapshots.
// Accepted param shapes:
//   - refs: JSON array string, e.g. [{"ownerId":"...","keyId":"..."}]
//   - refs: comma-separated "ownerId:keyId" pairs
//   - ownerIds + keyIds: parallel lists (arrays or comma-separated strings)
func parseAuthRefs(params map[string]any) ([]*api.AuthRef, error) {
	if raw, ok := getParam(params, "refs"); ok && raw != nil {
		return parseAuthRefsValue(raw)
	}

	ownerRaw, hasOwners := getParam(params, "ownerIds", "owner_ids")
	keyRaw, hasKeys := getParam(params, "keyIds", "key_ids")
	if !hasOwners || !hasKeys {
		return nil, fmt.Errorf("refs (JSON array) or ownerIds+keyIds params are required")
	}

	ownerIDs, err := parseUUIDListParam(ownerRaw, "ownerIds")
	if err != nil {
		return nil, err
	}
	keyIDs, err := parseUUIDListParam(keyRaw, "keyIds")
	if err != nil {
		return nil, err
	}
	if len(ownerIDs) != len(keyIDs) {
		return nil, fmt.Errorf("ownerIds (%d) and keyIds (%d) must have the same length", len(ownerIDs), len(keyIDs))
	}

	refs := make([]*api.AuthRef, 0, len(ownerIDs))
	for i := range ownerIDs {
		refs = append(refs, &api.AuthRef{OwnerId: ownerIDs[i], KeyId: keyIDs[i]})
	}
	return refs, nil
}

func parseAuthRefsValue(raw any) ([]*api.AuthRef, error) {
	switch value := raw.(type) {
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, fmt.Errorf("refs must not be empty")
		}
		if strings.HasPrefix(trimmed, "[") {
			var entries []map[string]any
			if err := json.Unmarshal([]byte(trimmed), &entries); err != nil {
				return nil, fmt.Errorf("refs json: %w", err)
			}
			refs := make([]*api.AuthRef, 0, len(entries))
			for i, entry := range entries {
				ownerID, err := getRequiredUUIDParam(entry, "ownerId", "owner_id", "ownerID")
				if err != nil {
					return nil, fmt.Errorf("refs[%d]: %w", i, err)
				}
				keyID, err := getRequiredUUIDParam(entry, "keyId", "key_id", "keyID")
				if err != nil {
					return nil, fmt.Errorf("refs[%d]: %w", i, err)
				}
				refs = append(refs, &api.AuthRef{OwnerId: ownerID, KeyId: keyID})
			}
			if len(refs) == 0 {
				return nil, fmt.Errorf("refs must contain at least one entry")
			}
			return refs, nil
		}
		parts := strings.Split(trimmed, ",")
		refs := make([]*api.AuthRef, 0, len(parts))
		for i, part := range parts {
			pair := strings.SplitN(strings.TrimSpace(part), ":", 2)
			if len(pair) != 2 {
				return nil, fmt.Errorf("refs[%d]: expected ownerId:keyId", i)
			}
			ownerID, err := parseUUIDLike(pair[0], "ownerId")
			if err != nil {
				return nil, fmt.Errorf("refs[%d]: %w", i, err)
			}
			keyID, err := parseUUIDLike(pair[1], "keyId")
			if err != nil {
				return nil, fmt.Errorf("refs[%d]: %w", i, err)
			}
			refs = append(refs, &api.AuthRef{OwnerId: ownerID, KeyId: keyID})
		}
		return refs, nil
	case []any:
		refs := make([]*api.AuthRef, 0, len(value))
		for i, item := range value {
			itemRefs, err := parseAuthRefsValue(item)
			if err != nil {
				return nil, fmt.Errorf("refs[%d]: %w", i, err)
			}
			refs = append(refs, itemRefs...)
		}
		if len(refs) == 0 {
			return nil, fmt.Errorf("refs must contain at least one entry")
		}
		return refs, nil
	default:
		return nil, fmt.Errorf("refs must be a JSON array string")
	}
}

func parseUUIDListParam(raw any, paramName string) ([][]byte, error) {
	switch value := raw.(type) {
	case []any:
		ids := make([][]byte, 0, len(value))
		for i, item := range value {
			id, err := parseUUIDLike(item, fmt.Sprintf("%s[%d]", paramName, i))
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, nil
	case []string:
		ids := make([][]byte, 0, len(value))
		for i, item := range value {
			id, err := parseUUIDStringOrBase64(item, fmt.Sprintf("%s[%d]", paramName, i))
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, nil
	case string:
		parts := strings.Split(value, ",")
		ids := make([][]byte, 0, len(parts))
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := parseUUIDStringOrBase64(part, fmt.Sprintf("%s[%d]", paramName, i))
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, nil
	default:
		return nil, fmt.Errorf("%s must be a list of UUIDs", paramName)
	}
}
