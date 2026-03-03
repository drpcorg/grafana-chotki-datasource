package plugin

import (
	"fmt"
	"sort"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	api "github.com/p2p-org/dproxy/pkg/api"
)

type keyRow struct {
	KeyID                          string
	OwnerID                        string
	Name                           string
	Status                         bool
	UpdatedAt                      time.Time
	UpdatedAtRFC3339               string
	UpdatedAtUnix                  int64
	APIKey                         string
	RateLimit                      int64
	IpWhitelistJSON                string
	IpWhitelistCSV                 string
	ProvidersJSON                  string
	ProvidersCSV                   string
	CorsOriginsJSON                string
	CorsOriginsCSV                 string
	JWTEnabled                     bool
	JWTPublicKeysJSON              string
	JWTPublicKeysCSV               string
	FallbackEnabled                bool
	FallbackProvidersJSON          string
	FallbackProvidersCSV           string
	MEVEnabled                     bool
	MEVProvidersJSON               string
	MEVProvidersCSV                string
	MEVFallback                    bool
	ComputeUnitDayLimit            int64
	MethodBlacklistJSON            string
	MethodBlacklistCSV             string
	ContractWhitelistJSON          string
	ContractWhitelistCSV           string
	Archived                       bool
	Description                    string
	TagsJSON                       string
	TagsCSV                        string
	MethodWhitelistJSON            string
	MethodWhitelistCSV             string
	ChainsWhitelistJSON            string
	ChainsWhitelistCSV             string
	EnableTraceID                  bool
	BalanceNotificationOptionsJSON string
	BalanceNotificationOptionsCSV  string
	ClientSpec                     string
	ClientSpecValidJSON            bool
	MEVMode                        int32
	MEVModeLabel                   string
}

type nodeCoreKeyRow struct {
	KeyID                 string
	OwnerID               string
	Name                  string
	Status                bool
	UpdatedAt             time.Time
	UpdatedAtRFC3339      string
	UpdatedAtUnix         int64
	Description           string
	IpWhitelistJSON       string
	IpWhitelistCSV        string
	ContractWhitelistJSON string
	ContractWhitelistCSV  string
	MethodWhitelistJSON   string
	MethodWhitelistCSV    string
	MethodBlacklistJSON   string
	MethodBlacklistCSV    string
	TagsJSON              string
	TagsCSV               string
	CorsOriginsJSON       string
	CorsOriginsCSV        string
}

func buildStatFrame(method string, value float64) *data.Frame {
	now := time.Now().UTC()
	return data.NewFrame(
		"stat",
		data.NewField("time", nil, []time.Time{now}),
		data.NewField("method", nil, []string{method}),
		data.NewField("value", nil, []float64{value}),
	)
}

func buildGetOwnerFrame(owner *api.Owner, opts queryExecOptions) (*data.Frame, float64, error) {
	if owner == nil {
		return nil, 0, fmt.Errorf("owner is empty")
	}

	createdAt, createdAtRFC3339, createdAtUnix := timestampFields(owner.CreatedAt, opts.DecodeTimestamps)
	frame := data.NewFrame(
		"get_owner",
		data.NewField("owner_id", nil, []string{formatID(owner.OwnerID, opts.DecodeIDs)}),
		data.NewField("compute_unit_balance", nil, []int64{owner.ComputeUnitBalance}),
		data.NewField("tier", nil, []int64{owner.Tier}),
		data.NewField("tier_label", nil, []string{tierLabel(owner.Tier, opts.DecodeEnums)}),
		data.NewField("special", nil, []bool{owner.Special}),
		data.NewField("discount_percent", nil, []int32{owner.DiscountPercent}),
		data.NewField("overdraft_limit", nil, []int64{owner.OverdraftLimit}),
		data.NewField("free_compute_unit_balance", nil, []int64{owner.FreeComputeUnitBalance}),
		data.NewField("created_at", nil, []time.Time{createdAt}),
		data.NewField("created_at_rfc3339", nil, []string{createdAtRFC3339}),
		data.NewField("created_at_unix", nil, []int64{createdAtUnix}),
		data.NewField("node_core_api_token", nil, []string{owner.NodeCoreApiToken}),
		data.NewField("addons_json", nil, []string{toJSON(owner.Addons)}),
	)
	return frame, float64(owner.ComputeUnitBalance), nil
}

func buildGetFullOwnerFrame(owner *api.OwnerFull, opts queryExecOptions) (*data.Frame, float64, error) {
	if owner == nil {
		return nil, 0, fmt.Errorf("owner is empty")
	}

	createdAt, createdAtRFC3339, createdAtUnix := timestampFields(owner.CreatedAt, opts.DecodeTimestamps)
	frame := data.NewFrame(
		"get_full_owner",
		data.NewField("owner_id", nil, []string{formatID(owner.OwnerId, opts.DecodeIDs)}),
		data.NewField("balance", nil, []int64{owner.Balance}),
		data.NewField("balance_low_warn", nil, []int64{owner.BalanceLowWarn}),
		data.NewField("balance_low_crit", nil, []int64{owner.BalanceLowCrit}),
		data.NewField("discount_percent", nil, []int32{owner.DiscountPercent}),
		data.NewField("overdraft_limit", nil, []int64{owner.OverdraftLimit}),
		data.NewField("tier", nil, []int32{owner.Tier}),
		data.NewField("tier_label", nil, []string{tierLabel(int64(owner.Tier), opts.DecodeEnums)}),
		data.NewField("total_spent", nil, []int64{owner.TotalSpent}),
		data.NewField("free_compute_unit_balance", nil, []int64{owner.FreeComputeUnitBalance}),
		data.NewField("created_at", nil, []time.Time{createdAt}),
		data.NewField("created_at_rfc3339", nil, []string{createdAtRFC3339}),
		data.NewField("created_at_unix", nil, []int64{createdAtUnix}),
		data.NewField("node_core_api_token", nil, []string{owner.NodeCoreApiToken}),
		data.NewField("addons_json", nil, []string{toJSON(owner.Addons)}),
	)
	return frame, float64(owner.Balance), nil
}

func buildGetOwnerHitsFrame(ownerID []byte, hits int64, opts queryExecOptions) (*data.Frame, float64) {
	frame := data.NewFrame(
		"get_owner_hits",
		data.NewField("owner_id", nil, []string{formatID(ownerID, opts.DecodeIDs)}),
		data.NewField("hits", nil, []int64{hits}),
	)
	return frame, float64(hits)
}

func buildGetOwnerMetadataFrame(metadata *api.OwnerMetadata, opts queryExecOptions) (*data.Frame, float64, error) {
	if metadata == nil {
		return nil, 0, fmt.Errorf("metadata is empty")
	}

	frame := data.NewFrame(
		"get_owner_metadata",
		data.NewField("owner_id", nil, []string{formatID(metadata.OwnerId, opts.DecodeIDs)}),
		data.NewField("balance_low_warning", nil, []int64{metadata.BalanceLowWarning}),
		data.NewField("balance_low_critical", nil, []int64{metadata.BalanceLowCritical}),
		data.NewField("discount_percent", nil, []int32{metadata.DiscountPercent}),
		data.NewField("overdraft_limit", nil, []int64{metadata.OverdraftLimit}),
	)
	return frame, float64(metadata.BalanceLowWarning), nil
}

func flattenKeyRow(key *api.Key, opts queryExecOptions) (keyRow, error) {
	if key == nil {
		return keyRow{}, fmt.Errorf("key is nil")
	}

	updatedAt, updatedAtRFC3339, updatedAtUnix := timestampFields(key.UpdatedAt, opts.DecodeTimestamps)
	ipWhitelistJSON, ipWhitelistCSV := stringSliceDual(key.IpWhitelist)
	providersJSON, providersCSV := stringSliceDual(key.Providers)
	corsOriginsJSON, corsOriginsCSV := stringSliceDual(key.CorsOrigins)
	jwtPublicKeysJSON, jwtPublicKeysCSV := publicKeysDual(key.JwtPublicKeys)
	fallbackProvidersJSON, fallbackProvidersCSV := stringSliceDual(key.FallbackProviders)
	mevProvidersJSON, mevProvidersCSV := stringSliceDual(key.MevProviders)
	methodBlacklistJSON, methodBlacklistCSV := stringSliceDual(key.MethodBlacklist)
	contractWhitelistJSON, contractWhitelistCSV := stringSliceDual(key.ContractWhiteList)
	tagsJSON, tagsCSV := int32SliceDual(key.Tags)
	methodWhitelistJSON, methodWhitelistCSV := stringSliceDual(key.MethodWhitelist)
	chainsWhitelistJSON, chainsWhitelistCSV := int32SliceDual(key.ChainsWhitelist)
	balanceOptionsJSON, balanceOptionsCSV := int32SliceDual(key.BalanceNotificationOptions)
	clientSpecValue, clientSpecValid := normalizeClientSpec(key.ClientSpec)

	return keyRow{
		KeyID:                          formatID(key.KeyId, opts.DecodeIDs),
		OwnerID:                        formatID(key.OwnerId, opts.DecodeIDs),
		Name:                           key.Name,
		Status:                         key.Status,
		UpdatedAt:                      updatedAt,
		UpdatedAtRFC3339:               updatedAtRFC3339,
		UpdatedAtUnix:                  updatedAtUnix,
		APIKey:                         key.ApiKey,
		RateLimit:                      key.RateLimit,
		IpWhitelistJSON:                ipWhitelistJSON,
		IpWhitelistCSV:                 ipWhitelistCSV,
		ProvidersJSON:                  providersJSON,
		ProvidersCSV:                   providersCSV,
		CorsOriginsJSON:                corsOriginsJSON,
		CorsOriginsCSV:                 corsOriginsCSV,
		JWTEnabled:                     key.JwtEnabled,
		JWTPublicKeysJSON:              jwtPublicKeysJSON,
		JWTPublicKeysCSV:               jwtPublicKeysCSV,
		FallbackEnabled:                key.FallbackEnabled,
		FallbackProvidersJSON:          fallbackProvidersJSON,
		FallbackProvidersCSV:           fallbackProvidersCSV,
		MEVEnabled:                     key.MevEnabled,
		MEVProvidersJSON:               mevProvidersJSON,
		MEVProvidersCSV:                mevProvidersCSV,
		MEVFallback:                    key.MevFallback,
		ComputeUnitDayLimit:            key.ComputeUnitDayLimit,
		MethodBlacklistJSON:            methodBlacklistJSON,
		MethodBlacklistCSV:             methodBlacklistCSV,
		ContractWhitelistJSON:          contractWhitelistJSON,
		ContractWhitelistCSV:           contractWhitelistCSV,
		Archived:                       key.Archived,
		Description:                    key.Description,
		TagsJSON:                       tagsJSON,
		TagsCSV:                        tagsCSV,
		MethodWhitelistJSON:            methodWhitelistJSON,
		MethodWhitelistCSV:             methodWhitelistCSV,
		ChainsWhitelistJSON:            chainsWhitelistJSON,
		ChainsWhitelistCSV:             chainsWhitelistCSV,
		EnableTraceID:                  key.EnableTraceID,
		BalanceNotificationOptionsJSON: balanceOptionsJSON,
		BalanceNotificationOptionsCSV:  balanceOptionsCSV,
		ClientSpec:                     clientSpecValue,
		ClientSpecValidJSON:            clientSpecValid,
		MEVMode:                        key.MevMode,
		MEVModeLabel:                   mevModeLabel(key.MevMode, opts.DecodeEnums),
	}, nil
}

func buildKeyRowsFrame(frameName string, rows []keyRow) *data.Frame {
	keyIDs := make([]string, 0, len(rows))
	ownerIDs := make([]string, 0, len(rows))
	names := make([]string, 0, len(rows))
	statuses := make([]bool, 0, len(rows))
	updatedAts := make([]time.Time, 0, len(rows))
	updatedAtRFC3339 := make([]string, 0, len(rows))
	updatedAtUnix := make([]int64, 0, len(rows))
	apiKeys := make([]string, 0, len(rows))
	rateLimits := make([]int64, 0, len(rows))
	ipWhitelistJSON := make([]string, 0, len(rows))
	ipWhitelistCSV := make([]string, 0, len(rows))
	providersJSON := make([]string, 0, len(rows))
	providersCSV := make([]string, 0, len(rows))
	corsOriginsJSON := make([]string, 0, len(rows))
	corsOriginsCSV := make([]string, 0, len(rows))
	jwtEnabled := make([]bool, 0, len(rows))
	jwtPublicKeysJSON := make([]string, 0, len(rows))
	jwtPublicKeysCSV := make([]string, 0, len(rows))
	fallbackEnabled := make([]bool, 0, len(rows))
	fallbackProvidersJSON := make([]string, 0, len(rows))
	fallbackProvidersCSV := make([]string, 0, len(rows))
	mevEnabled := make([]bool, 0, len(rows))
	mevProvidersJSON := make([]string, 0, len(rows))
	mevProvidersCSV := make([]string, 0, len(rows))
	mevFallback := make([]bool, 0, len(rows))
	computeUnitDayLimit := make([]int64, 0, len(rows))
	methodBlacklistJSON := make([]string, 0, len(rows))
	methodBlacklistCSV := make([]string, 0, len(rows))
	contractWhitelistJSON := make([]string, 0, len(rows))
	contractWhitelistCSV := make([]string, 0, len(rows))
	archived := make([]bool, 0, len(rows))
	descriptions := make([]string, 0, len(rows))
	tagsJSON := make([]string, 0, len(rows))
	tagsCSV := make([]string, 0, len(rows))
	methodWhitelistJSON := make([]string, 0, len(rows))
	methodWhitelistCSV := make([]string, 0, len(rows))
	chainsWhitelistJSON := make([]string, 0, len(rows))
	chainsWhitelistCSV := make([]string, 0, len(rows))
	enableTraceID := make([]bool, 0, len(rows))
	balanceOptionsJSON := make([]string, 0, len(rows))
	balanceOptionsCSV := make([]string, 0, len(rows))
	clientSpec := make([]string, 0, len(rows))
	clientSpecValid := make([]bool, 0, len(rows))
	mevMode := make([]int32, 0, len(rows))
	mevModeLabel := make([]string, 0, len(rows))

	for _, row := range rows {
		keyIDs = append(keyIDs, row.KeyID)
		ownerIDs = append(ownerIDs, row.OwnerID)
		names = append(names, row.Name)
		statuses = append(statuses, row.Status)
		updatedAts = append(updatedAts, row.UpdatedAt)
		updatedAtRFC3339 = append(updatedAtRFC3339, row.UpdatedAtRFC3339)
		updatedAtUnix = append(updatedAtUnix, row.UpdatedAtUnix)
		apiKeys = append(apiKeys, row.APIKey)
		rateLimits = append(rateLimits, row.RateLimit)
		ipWhitelistJSON = append(ipWhitelistJSON, row.IpWhitelistJSON)
		ipWhitelistCSV = append(ipWhitelistCSV, row.IpWhitelistCSV)
		providersJSON = append(providersJSON, row.ProvidersJSON)
		providersCSV = append(providersCSV, row.ProvidersCSV)
		corsOriginsJSON = append(corsOriginsJSON, row.CorsOriginsJSON)
		corsOriginsCSV = append(corsOriginsCSV, row.CorsOriginsCSV)
		jwtEnabled = append(jwtEnabled, row.JWTEnabled)
		jwtPublicKeysJSON = append(jwtPublicKeysJSON, row.JWTPublicKeysJSON)
		jwtPublicKeysCSV = append(jwtPublicKeysCSV, row.JWTPublicKeysCSV)
		fallbackEnabled = append(fallbackEnabled, row.FallbackEnabled)
		fallbackProvidersJSON = append(fallbackProvidersJSON, row.FallbackProvidersJSON)
		fallbackProvidersCSV = append(fallbackProvidersCSV, row.FallbackProvidersCSV)
		mevEnabled = append(mevEnabled, row.MEVEnabled)
		mevProvidersJSON = append(mevProvidersJSON, row.MEVProvidersJSON)
		mevProvidersCSV = append(mevProvidersCSV, row.MEVProvidersCSV)
		mevFallback = append(mevFallback, row.MEVFallback)
		computeUnitDayLimit = append(computeUnitDayLimit, row.ComputeUnitDayLimit)
		methodBlacklistJSON = append(methodBlacklistJSON, row.MethodBlacklistJSON)
		methodBlacklistCSV = append(methodBlacklistCSV, row.MethodBlacklistCSV)
		contractWhitelistJSON = append(contractWhitelistJSON, row.ContractWhitelistJSON)
		contractWhitelistCSV = append(contractWhitelistCSV, row.ContractWhitelistCSV)
		archived = append(archived, row.Archived)
		descriptions = append(descriptions, row.Description)
		tagsJSON = append(tagsJSON, row.TagsJSON)
		tagsCSV = append(tagsCSV, row.TagsCSV)
		methodWhitelistJSON = append(methodWhitelistJSON, row.MethodWhitelistJSON)
		methodWhitelistCSV = append(methodWhitelistCSV, row.MethodWhitelistCSV)
		chainsWhitelistJSON = append(chainsWhitelistJSON, row.ChainsWhitelistJSON)
		chainsWhitelistCSV = append(chainsWhitelistCSV, row.ChainsWhitelistCSV)
		enableTraceID = append(enableTraceID, row.EnableTraceID)
		balanceOptionsJSON = append(balanceOptionsJSON, row.BalanceNotificationOptionsJSON)
		balanceOptionsCSV = append(balanceOptionsCSV, row.BalanceNotificationOptionsCSV)
		clientSpec = append(clientSpec, row.ClientSpec)
		clientSpecValid = append(clientSpecValid, row.ClientSpecValidJSON)
		mevMode = append(mevMode, row.MEVMode)
		mevModeLabel = append(mevModeLabel, row.MEVModeLabel)
	}

	return data.NewFrame(
		frameName,
		data.NewField("key_id", nil, keyIDs),
		data.NewField("owner_id", nil, ownerIDs),
		data.NewField("name", nil, names),
		data.NewField("status", nil, statuses),
		data.NewField("updated_at", nil, updatedAts),
		data.NewField("updated_at_rfc3339", nil, updatedAtRFC3339),
		data.NewField("updated_at_unix", nil, updatedAtUnix),
		data.NewField("api_key", nil, apiKeys),
		data.NewField("rate_limit", nil, rateLimits),
		data.NewField("ip_whitelist_json", nil, ipWhitelistJSON),
		data.NewField("ip_whitelist_csv", nil, ipWhitelistCSV),
		data.NewField("providers_json", nil, providersJSON),
		data.NewField("providers_csv", nil, providersCSV),
		data.NewField("cors_origins_json", nil, corsOriginsJSON),
		data.NewField("cors_origins_csv", nil, corsOriginsCSV),
		data.NewField("jwt_enabled", nil, jwtEnabled),
		data.NewField("jwt_public_keys_json", nil, jwtPublicKeysJSON),
		data.NewField("jwt_public_keys_csv", nil, jwtPublicKeysCSV),
		data.NewField("fallback_enabled", nil, fallbackEnabled),
		data.NewField("fallback_providers_json", nil, fallbackProvidersJSON),
		data.NewField("fallback_providers_csv", nil, fallbackProvidersCSV),
		data.NewField("mev_enabled", nil, mevEnabled),
		data.NewField("mev_providers_json", nil, mevProvidersJSON),
		data.NewField("mev_providers_csv", nil, mevProvidersCSV),
		data.NewField("mev_fallback", nil, mevFallback),
		data.NewField("compute_unit_day_limit", nil, computeUnitDayLimit),
		data.NewField("method_blacklist_json", nil, methodBlacklistJSON),
		data.NewField("method_blacklist_csv", nil, methodBlacklistCSV),
		data.NewField("contract_whitelist_json", nil, contractWhitelistJSON),
		data.NewField("contract_whitelist_csv", nil, contractWhitelistCSV),
		data.NewField("archived", nil, archived),
		data.NewField("description", nil, descriptions),
		data.NewField("tags_json", nil, tagsJSON),
		data.NewField("tags_csv", nil, tagsCSV),
		data.NewField("method_whitelist_json", nil, methodWhitelistJSON),
		data.NewField("method_whitelist_csv", nil, methodWhitelistCSV),
		data.NewField("chains_whitelist_json", nil, chainsWhitelistJSON),
		data.NewField("chains_whitelist_csv", nil, chainsWhitelistCSV),
		data.NewField("enable_trace_id", nil, enableTraceID),
		data.NewField("balance_notification_options_json", nil, balanceOptionsJSON),
		data.NewField("balance_notification_options_csv", nil, balanceOptionsCSV),
		data.NewField("client_spec", nil, clientSpec),
		data.NewField("client_spec_valid_json", nil, clientSpecValid),
		data.NewField("mev_mode", nil, mevMode),
		data.NewField("mev_mode_label", nil, mevModeLabel),
	)
}

func buildGetKeyFrame(key *api.Key, opts queryExecOptions) (*data.Frame, float64, error) {
	row, err := flattenKeyRow(key, opts)
	if err != nil {
		return nil, 0, err
	}
	frame := buildKeyRowsFrame("get_key", []keyRow{row})
	return frame, 1, nil
}

func buildListKeysFrame(keys []*api.Key, opts queryExecOptions) (*data.Frame, float64, error) {
	rows := make([]keyRow, 0, len(keys))
	for _, key := range keys {
		row, err := flattenKeyRow(key, opts)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
	}
	frame := buildKeyRowsFrame("list_keys", rows)
	return frame, float64(len(rows)), nil
}

func buildGetKeyHitsFrame(keyID []byte, hits int64, opts queryExecOptions) (*data.Frame, float64) {
	frame := data.NewFrame(
		"get_key_hits",
		data.NewField("key_id", nil, []string{formatID(keyID, opts.DecodeIDs)}),
		data.NewField("hits", nil, []int64{hits}),
	)
	return frame, float64(hits)
}

func buildGetOwnersWithBalanceFrame(balances map[string]int64, opts queryExecOptions) (*data.Frame, float64) {
	ownerIDs := make([]string, 0, len(balances))
	for ownerID := range balances {
		ownerIDs = append(ownerIDs, ownerID)
	}
	ownerIDs = sortStrings(ownerIDs)

	if opts.Limit > 0 && int64(len(ownerIDs)) > opts.Limit {
		ownerIDs = ownerIDs[:opts.Limit]
	}

	rowsOwnerIDs := make([]string, 0, len(ownerIDs))
	rowsBalances := make([]int64, 0, len(ownerIDs))
	var total int64
	for _, ownerID := range ownerIDs {
		rowsOwnerIDs = append(rowsOwnerIDs, formatMaybeUUIDString(ownerID, opts.DecodeIDs))
		rowsBalances = append(rowsBalances, balances[ownerID])
		total += balances[ownerID]
	}

	frame := data.NewFrame(
		"get_owners_with_balance",
		data.NewField("owner_id", nil, rowsOwnerIDs),
		data.NewField("balance", nil, rowsBalances),
	)
	return frame, float64(total)
}

func buildGetAllOwnerIDsFrame(ownerIDsRaw [][]byte, opts queryExecOptions) (*data.Frame, float64) {
	ownerIDs := ownerIDsRaw
	if opts.Limit > 0 && int64(len(ownerIDs)) > opts.Limit {
		ownerIDs = ownerIDs[:opts.Limit]
	}

	rows := make([]string, 0, len(ownerIDs))
	for _, ownerID := range ownerIDs {
		rows = append(rows, formatID(ownerID, opts.DecodeIDs))
	}

	frame := data.NewFrame(
		"get_all_owner_ids",
		data.NewField("owner_id", nil, rows),
	)
	return frame, float64(len(rows))
}

func flattenNodeCoreKeyRow(key *api.NodeCoreKey, opts queryExecOptions) (nodeCoreKeyRow, error) {
	if key == nil {
		return nodeCoreKeyRow{}, fmt.Errorf("node core key is nil")
	}

	updatedAt, updatedAtRFC3339, updatedAtUnix := timestampFields(key.UpdatedAt, opts.DecodeTimestamps)
	ipWhitelistJSON, ipWhitelistCSV := stringSliceDual(key.IpWhitelist)
	contractWhitelistJSON, contractWhitelistCSV := stringSliceDual(key.ContractWhitelist)
	methodWhitelistJSON, methodWhitelistCSV := stringSliceDual(key.MethodWhitelist)
	methodBlacklistJSON, methodBlacklistCSV := stringSliceDual(key.MethodBlacklist)
	tagsJSON, tagsCSV := int32SliceDual(key.Tags)
	corsOriginsJSON, corsOriginsCSV := stringSliceDual(key.CorsOrigins)

	return nodeCoreKeyRow{
		KeyID:                 formatID(key.KeyId, opts.DecodeIDs),
		OwnerID:               formatID(key.OwnerId, opts.DecodeIDs),
		Name:                  key.Name,
		Status:                key.Status,
		UpdatedAt:             updatedAt,
		UpdatedAtRFC3339:      updatedAtRFC3339,
		UpdatedAtUnix:         updatedAtUnix,
		Description:           key.Description,
		IpWhitelistJSON:       ipWhitelistJSON,
		IpWhitelistCSV:        ipWhitelistCSV,
		ContractWhitelistJSON: contractWhitelistJSON,
		ContractWhitelistCSV:  contractWhitelistCSV,
		MethodWhitelistJSON:   methodWhitelistJSON,
		MethodWhitelistCSV:    methodWhitelistCSV,
		MethodBlacklistJSON:   methodBlacklistJSON,
		MethodBlacklistCSV:    methodBlacklistCSV,
		TagsJSON:              tagsJSON,
		TagsCSV:               tagsCSV,
		CorsOriginsJSON:       corsOriginsJSON,
		CorsOriginsCSV:        corsOriginsCSV,
	}, nil
}

func buildNodeCoreKeyRowsFrame(frameName string, rows []nodeCoreKeyRow) *data.Frame {
	keyIDs := make([]string, 0, len(rows))
	ownerIDs := make([]string, 0, len(rows))
	names := make([]string, 0, len(rows))
	statuses := make([]bool, 0, len(rows))
	updatedAts := make([]time.Time, 0, len(rows))
	updatedAtRFC3339 := make([]string, 0, len(rows))
	updatedAtUnix := make([]int64, 0, len(rows))
	descriptions := make([]string, 0, len(rows))
	ipWhitelistJSON := make([]string, 0, len(rows))
	ipWhitelistCSV := make([]string, 0, len(rows))
	contractWhitelistJSON := make([]string, 0, len(rows))
	contractWhitelistCSV := make([]string, 0, len(rows))
	methodWhitelistJSON := make([]string, 0, len(rows))
	methodWhitelistCSV := make([]string, 0, len(rows))
	methodBlacklistJSON := make([]string, 0, len(rows))
	methodBlacklistCSV := make([]string, 0, len(rows))
	tagsJSON := make([]string, 0, len(rows))
	tagsCSV := make([]string, 0, len(rows))
	corsOriginsJSON := make([]string, 0, len(rows))
	corsOriginsCSV := make([]string, 0, len(rows))

	for _, row := range rows {
		keyIDs = append(keyIDs, row.KeyID)
		ownerIDs = append(ownerIDs, row.OwnerID)
		names = append(names, row.Name)
		statuses = append(statuses, row.Status)
		updatedAts = append(updatedAts, row.UpdatedAt)
		updatedAtRFC3339 = append(updatedAtRFC3339, row.UpdatedAtRFC3339)
		updatedAtUnix = append(updatedAtUnix, row.UpdatedAtUnix)
		descriptions = append(descriptions, row.Description)
		ipWhitelistJSON = append(ipWhitelistJSON, row.IpWhitelistJSON)
		ipWhitelistCSV = append(ipWhitelistCSV, row.IpWhitelistCSV)
		contractWhitelistJSON = append(contractWhitelistJSON, row.ContractWhitelistJSON)
		contractWhitelistCSV = append(contractWhitelistCSV, row.ContractWhitelistCSV)
		methodWhitelistJSON = append(methodWhitelistJSON, row.MethodWhitelistJSON)
		methodWhitelistCSV = append(methodWhitelistCSV, row.MethodWhitelistCSV)
		methodBlacklistJSON = append(methodBlacklistJSON, row.MethodBlacklistJSON)
		methodBlacklistCSV = append(methodBlacklistCSV, row.MethodBlacklistCSV)
		tagsJSON = append(tagsJSON, row.TagsJSON)
		tagsCSV = append(tagsCSV, row.TagsCSV)
		corsOriginsJSON = append(corsOriginsJSON, row.CorsOriginsJSON)
		corsOriginsCSV = append(corsOriginsCSV, row.CorsOriginsCSV)
	}

	return data.NewFrame(
		frameName,
		data.NewField("key_id", nil, keyIDs),
		data.NewField("owner_id", nil, ownerIDs),
		data.NewField("name", nil, names),
		data.NewField("status", nil, statuses),
		data.NewField("updated_at", nil, updatedAts),
		data.NewField("updated_at_rfc3339", nil, updatedAtRFC3339),
		data.NewField("updated_at_unix", nil, updatedAtUnix),
		data.NewField("description", nil, descriptions),
		data.NewField("ip_whitelist_json", nil, ipWhitelistJSON),
		data.NewField("ip_whitelist_csv", nil, ipWhitelistCSV),
		data.NewField("contract_whitelist_json", nil, contractWhitelistJSON),
		data.NewField("contract_whitelist_csv", nil, contractWhitelistCSV),
		data.NewField("method_whitelist_json", nil, methodWhitelistJSON),
		data.NewField("method_whitelist_csv", nil, methodWhitelistCSV),
		data.NewField("method_blacklist_json", nil, methodBlacklistJSON),
		data.NewField("method_blacklist_csv", nil, methodBlacklistCSV),
		data.NewField("tags_json", nil, tagsJSON),
		data.NewField("tags_csv", nil, tagsCSV),
		data.NewField("cors_origins_json", nil, corsOriginsJSON),
		data.NewField("cors_origins_csv", nil, corsOriginsCSV),
	)
}

func buildGetNodeCoreKeyFrame(key *api.NodeCoreKey, opts queryExecOptions) (*data.Frame, float64, error) {
	row, err := flattenNodeCoreKeyRow(key, opts)
	if err != nil {
		return nil, 0, err
	}
	frame := buildNodeCoreKeyRowsFrame("get_node_core_key", []nodeCoreKeyRow{row})
	return frame, 1, nil
}

func buildListNodeCoreKeysFrame(keys []*api.NodeCoreKey, opts queryExecOptions) (*data.Frame, float64, error) {
	rows := make([]nodeCoreKeyRow, 0, len(keys))
	for _, key := range keys {
		row, err := flattenNodeCoreKeyRow(key, opts)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, row)
	}
	frame := buildNodeCoreKeyRowsFrame("list_node_core_keys", rows)
	return frame, float64(len(rows)), nil
}

func sortedMapKeys(input map[string]int64) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
