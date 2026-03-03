package plugin

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
)

func TestFormatID(t *testing.T) {
	id := uuid.New()
	decoded := formatID(id[:], true)
	if decoded != id.String() {
		t.Fatalf("decoded id mismatch: got %s want %s", decoded, id.String())
	}

	raw := formatID(id[:], false)
	if raw != base64.StdEncoding.EncodeToString(id[:]) {
		t.Fatalf("raw id mismatch")
	}
}

func TestNormalizeClientSpec(t *testing.T) {
	valid, ok := normalizeClientSpec(`[{"client_type":"geth","client_version":"1.13.0"}]`)
	if !ok {
		t.Fatalf("expected valid client spec")
	}
	if valid == "" {
		t.Fatalf("expected normalized value")
	}

	invalid, ok := normalizeClientSpec(`[{broken`) // invalid JSON
	if ok {
		t.Fatalf("expected invalid client spec")
	}
	if invalid != `[{broken` {
		t.Fatalf("expected original invalid value")
	}
}

func TestArrayDualRepresentation(t *testing.T) {
	jsonValue, csvValue := stringSliceDual([]string{"a", "b", "c"})
	if jsonValue != `["a","b","c"]` {
		t.Fatalf("unexpected json value: %s", jsonValue)
	}
	if csvValue != "a,b,c" {
		t.Fatalf("unexpected csv value: %s", csvValue)
	}

	intJSON, intCSV := int32SliceDual([]int32{1, 2, 3})
	if intJSON != `[1,2,3]` {
		t.Fatalf("unexpected int json value: %s", intJSON)
	}
	if intCSV != "1,2,3" {
		t.Fatalf("unexpected int csv value: %s", intCSV)
	}
}
