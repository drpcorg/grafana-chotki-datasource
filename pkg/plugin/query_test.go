package plugin

import (
	"encoding/base64"
	"encoding/json"
	"testing"

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
