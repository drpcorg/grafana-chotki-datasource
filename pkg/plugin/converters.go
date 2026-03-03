package plugin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	api "github.com/drpcorg/grafana-chotki-datasource/pkg/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func formatID(raw []byte, decode bool) string {
	if len(raw) == 0 {
		return ""
	}
	if decode {
		if parsed, err := uuid.FromBytes(raw); err == nil {
			return parsed.String()
		}
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func formatMaybeUUIDString(raw string, decode bool) string {
	if decode {
		return raw
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return raw
	}
	return base64.StdEncoding.EncodeToString(parsed[:])
}

func timestampFields(ts *timestamppb.Timestamp, includeRFC bool) (time.Time, string, int64) {
	if ts == nil {
		return time.Time{}, "", 0
	}
	timeValue := ts.AsTime().UTC()
	rfc := ""
	if includeRFC {
		rfc = timeValue.Format(time.RFC3339Nano)
	}
	return timeValue, rfc, timeValue.Unix()
}

func tierLabel(raw int64, decode bool) string {
	if !decode {
		return ""
	}
	switch raw {
	case 0:
		return "free"
	case 1:
		return "paid"
	default:
		return "unknown"
	}
}

func mevModeLabel(raw int32, decode bool) string {
	if !decode {
		return ""
	}
	switch raw {
	case 0:
		return "unset"
	case 1:
		return "enabled"
	case 2:
		return "disabled"
	default:
		return "unknown"
	}
}

func stringSliceDual(values []string) (jsonValue string, csvValue string) {
	jsonValue = toJSON(values)
	csvValue = strings.Join(values, ",")
	return jsonValue, csvValue
}

func int32SliceDual(values []int32) (jsonValue string, csvValue string) {
	jsonValue = toJSON(values)
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(int(value)))
	}
	csvValue = strings.Join(parts, ",")
	return jsonValue, csvValue
}

func publicKeysDual(values []*api.PublicKey) (jsonValue string, csvValue string) {
	if len(values) == 0 {
		return "[]", ""
	}

	names := make([]string, 0, len(values))
	payload := make([]map[string]string, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		names = append(names, value.Name)
		payload = append(payload, map[string]string{
			"name":      value.Name,
			"bytes_b64": base64.StdEncoding.EncodeToString(value.Bytes),
		})
	}
	return toJSON(payload), strings.Join(names, ",")
}

func normalizeClientSpec(raw string) (value string, valid bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	if !json.Valid([]byte(trimmed)) {
		return raw, false
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(trimmed)); err != nil {
		return raw, false
	}
	return compact.String(), true
}

func toJSON(value any) string {
	if value == nil {
		return "null"
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("\"marshal_error:%v\"", err)
	}
	return string(encoded)
}

func sortStrings(values []string) []string {
	cloned := append([]string(nil), values...)
	sort.Strings(cloned)
	return cloned
}
