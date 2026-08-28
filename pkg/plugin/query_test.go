package plugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/drpcorg/grafana-chotki-datasource/pkg/api"
	"github.com/drpcorg/grafana-chotki-datasource/pkg/models"
	"github.com/google/uuid"
)

func testSettings() *models.PluginSettings {
	return &models.PluginSettings{
		Insecure:         true,
		TimeoutMs:        4000,
		DefaultLimit:     200,
		HardLimit:        1000,
		DecodeIDs:        true,
		DecodeEnums:      true,
		DecodeTimestamps: true,
	}
}

func TestParseQueryModel_Defaults(t *testing.T) {
	raw := []byte(`{"mode":"rpc","method":"GetOwnerHits","params":{"ownerId":"b6b1f765-0fe6-4ea0-8f45-3486f149f299"}}`)

	qm, opts, err := parseQueryModel(raw, "A", testSettings())
	if err != nil {
		t.Fatalf("parseQueryModel() error = %v", err)
	}
	if qm.RefID != "A" {
		t.Fatalf("unexpected refID: %s", qm.RefID)
	}
	if qm.Method != methodGetOwnerHits {
		t.Fatalf("unexpected method: %s", qm.Method)
	}
	if opts.Format != "table" {
		t.Fatalf("unexpected default format: %s", opts.Format)
	}
	if opts.Limit != 200 {
		t.Fatalf("unexpected default limit: %d", opts.Limit)
	}
}

func TestParseQueryModel_RawQueryMerge(t *testing.T) {
	raw := []byte(`{
		"mode":"rpc",
		"editorMode":"raw",
		"rawQuery":"{\"method\":\"GetAllOwnerIds\",\"params\":{},\"options\":{\"format\":\"stat\",\"limit\":10}}"
	}`)

	qm, opts, err := parseQueryModel(raw, "A", testSettings())
	if err != nil {
		t.Fatalf("parseQueryModel() error = %v", err)
	}
	if qm.Method != methodGetAllOwnerIDs {
		t.Fatalf("unexpected method: %s", qm.Method)
	}
	if opts.Format != "stat" {
		t.Fatalf("unexpected format: %s", opts.Format)
	}
	if opts.Limit != 10 {
		t.Fatalf("unexpected limit: %d", opts.Limit)
	}
}

func TestParseUUIDStringOrBase64(t *testing.T) {
	id := uuid.New()

	parsedUUID, err := parseUUIDStringOrBase64(id.String(), "ownerId")
	if err != nil {
		t.Fatalf("parseUUIDStringOrBase64() uuid error = %v", err)
	}
	if string(parsedUUID) != string(id[:]) {
		t.Fatalf("parsed UUID mismatch")
	}

	b64 := base64.StdEncoding.EncodeToString(id[:])
	parsedB64, err := parseUUIDStringOrBase64(b64, "ownerId")
	if err != nil {
		t.Fatalf("parseUUIDStringOrBase64() base64 error = %v", err)
	}
	if string(parsedB64) != string(id[:]) {
		t.Fatalf("parsed base64 mismatch")
	}
}

func TestGetOptionalInt64Param(t *testing.T) {
	params := map[string]any{"limit": json.Number("150")}
	value, ok, err := getOptionalInt64Param(params, "limit")
	if err != nil {
		t.Fatalf("getOptionalInt64Param() error = %v", err)
	}
	if !ok || value != 150 {
		t.Fatalf("unexpected value=%d ok=%v", value, ok)
	}
}

func TestParseAuthRefs_JSONString(t *testing.T) {
	ownerID := uuid.New()
	keyID := uuid.New()
	params := map[string]any{
		"refs": fmt.Sprintf(`[{"ownerId":%q,"keyId":%q}]`, ownerID.String(), keyID.String()),
	}

	refs, err := parseAuthRefs(params)
	if err != nil {
		t.Fatalf("parseAuthRefs() error = %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("unexpected refs length: %d", len(refs))
	}
	if string(refs[0].GetOwnerId()) != string(ownerID[:]) || string(refs[0].GetKeyId()) != string(keyID[:]) {
		t.Fatalf("ref ids mismatch")
	}
}

func TestParseAuthRefs_Pairs(t *testing.T) {
	ownerID := uuid.New()
	keyID := uuid.New()
	params := map[string]any{"refs": ownerID.String() + ":" + keyID.String()}

	refs, err := parseAuthRefs(params)
	if err != nil {
		t.Fatalf("parseAuthRefs() error = %v", err)
	}
	if len(refs) != 1 || string(refs[0].GetOwnerId()) != string(ownerID[:]) || string(refs[0].GetKeyId()) != string(keyID[:]) {
		t.Fatalf("ref ids mismatch")
	}
}

func TestParseAuthRefs_ParallelLists(t *testing.T) {
	ownerA, ownerB := uuid.New(), uuid.New()
	keyA, keyB := uuid.New(), uuid.New()
	params := map[string]any{
		"ownerIds": []any{ownerA.String(), ownerB.String()},
		"keyIds":   []any{keyA.String(), keyB.String()},
	}

	refs, err := parseAuthRefs(params)
	if err != nil {
		t.Fatalf("parseAuthRefs() error = %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("unexpected refs length: %d", len(refs))
	}
	if string(refs[1].GetOwnerId()) != string(ownerB[:]) || string(refs[1].GetKeyId()) != string(keyB[:]) {
		t.Fatalf("ref[1] ids mismatch")
	}

	bad := map[string]any{"ownerIds": []any{ownerA.String()}, "keyIds": "not-a-uuid"}
	if _, err := parseAuthRefs(bad); err == nil {
		t.Fatalf("expected error for invalid keyIds")
	}
}

func TestParseAuthRefs_Missing(t *testing.T) {
	if _, err := parseAuthRefs(map[string]any{}); err == nil {
		t.Fatalf("expected error for missing refs")
	}
}

func TestParseQueryModel_NewMethods(t *testing.T) {
	for _, method := range []string{methodGetAuthSnapshot, methodGetAuthSnapshots, methodGetPackagePools} {
		raw := []byte(`{"mode":"rpc","method":"` + method + `","params":{}}`)
		qm, _, err := parseQueryModel(raw, "A", testSettings())
		if err != nil {
			t.Fatalf("method %s rejected: %v", method, err)
		}
		if qm.Method != method {
			t.Fatalf("unexpected method: %s", qm.Method)
		}
	}
}

func TestBuildGetOwnerFrame_BillingFields(t *testing.T) {
	ownerID := uuid.New()
	owner := &api.Owner{
		OwnerID:        ownerID[:],
		Discounts:      map[string]int32{"default": 15},
		BillingVersion: 2,
	}

	frame, _, err := buildGetOwnerFrame(owner, queryExecOptions{})
	if err != nil {
		t.Fatalf("buildGetOwnerFrame() error = %v", err)
	}
	names := make([]string, 0, len(frame.Fields))
	for _, field := range frame.Fields {
		names = append(names, field.Name)
	}
	if names[len(names)-2] != "discounts_json" || names[len(names)-1] != "billing_version" {
		t.Fatalf("unexpected trailing fields: %v", names)
	}
	if got, ok := frame.Fields[len(frame.Fields)-1].ConcreteAt(0); !ok || got.(int64) != 2 {
		t.Fatalf("unexpected billing_version: %v", got)
	}
}

func TestBuildGetPackagePoolsFrame(t *testing.T) {
	ownerID := uuid.New()
	pools := []*api.PackagePoolSnapshot{
		{Tag: "pro", Credited: 1000, Spent: 400, PeriodEnd: "2026-12-31"},
		{Tag: "scale", Credited: 500, Spent: 500},
	}

	frame, statValue := buildGetPackagePoolsFrame(ownerID[:], pools, queryExecOptions{DecodeIDs: true})
	if frame.Name != "get_package_pools" || frame.Rows() != 2 {
		t.Fatalf("unexpected frame: %s rows=%d", frame.Name, frame.Rows())
	}
	remainingField := frame.Fields[4]
	if remainingField.Name != "remaining" {
		t.Fatalf("unexpected field order: %s", remainingField.Name)
	}
	if got, _ := remainingField.ConcreteAt(0); got.(int64) != 600 {
		t.Fatalf("unexpected remaining[0]: %v", got)
	}
	if got, _ := remainingField.ConcreteAt(1); got.(int64) != 0 {
		t.Fatalf("unexpected remaining[1]: %v", got)
	}
	if statValue != 600 {
		t.Fatalf("unexpected stat value: %v", statValue)
	}
}

func TestBuildGetAuthSnapshotsFrame(t *testing.T) {
	ownerID := uuid.New()
	keyID := uuid.New()
	refs := []*api.AuthRef{{OwnerId: ownerID[:], KeyId: keyID[:]}}
	results := []*api.AuthSnapshotResult{
		{
			Snapshot: &api.AuthSnapshot{
				Owner:            &api.Owner{OwnerID: ownerID[:]},
				OwnerTotalCU:     42,
				OwnerFreeBalance: 7,
				KeyDailyUsedCU:   13,
			},
		},
	}

	frame := buildGetAuthSnapshotsFrame(refs, results, queryExecOptions{DecodeIDs: true})
	if frame.Name != "get_auth_snapshots" || frame.Rows() != 1 {
		t.Fatalf("unexpected frame: %s rows=%d", frame.Name, frame.Rows())
	}
	if got, _ := frame.Fields[0].ConcreteAt(0); got.(string) != ownerID.String() {
		t.Fatalf("unexpected requested_owner_id: %v", got)
	}
	if got, _ := frame.Fields[4].ConcreteAt(0); got.(int64) != 42 {
		t.Fatalf("unexpected owner_total_cu: %v", got)
	}

	errResults := []*api.AuthSnapshotResult{{Error: "not found"}}
	frame = buildGetAuthSnapshotsFrame(refs, errResults, queryExecOptions{DecodeIDs: true})
	if got, _ := frame.Fields[2].ConcreteAt(0); got.(string) != "not found" {
		t.Fatalf("unexpected error: %v", got)
	}
}
