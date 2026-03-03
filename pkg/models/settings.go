package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	DefaultTimeoutMs = 4000
	DefaultLimit     = 200
	DefaultHardLimit = 1000
)

type PluginSettings struct {
	GRPCAddress      string `json:"grpcAddress"`
	Insecure         bool   `json:"insecure"`
	TimeoutMs        int    `json:"timeoutMs"`
	DefaultLimit     int    `json:"defaultLimit"`
	HardLimit        int    `json:"hardLimit"`
	DecodeIDs        bool   `json:"decodeIds"`
	DecodeEnums      bool   `json:"decodeEnums"`
	DecodeTimestamps bool   `json:"decodeTimestamps"`

	Secrets SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	AuthToken     string `json:"authToken"`
	TLSCACert     string `json:"tlsCACert"`
	TLSClientCert string `json:"tlsClientCert"`
	TLSClientKey  string `json:"tlsClientKey"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := &PluginSettings{
		Insecure:         true,
		TimeoutMs:        DefaultTimeoutMs,
		DefaultLimit:     DefaultLimit,
		HardLimit:        DefaultHardLimit,
		DecodeIDs:        true,
		DecodeEnums:      true,
		DecodeTimestamps: true,
	}

	if len(source.JSONData) > 0 {
		if err := json.Unmarshal(source.JSONData, settings); err != nil {
			return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
		}
	}

	settings.GRPCAddress = strings.TrimSpace(settings.GRPCAddress)
	if settings.TimeoutMs <= 0 {
		settings.TimeoutMs = DefaultTimeoutMs
	}
	if settings.HardLimit <= 0 {
		settings.HardLimit = DefaultHardLimit
	}
	if settings.DefaultLimit <= 0 {
		settings.DefaultLimit = DefaultLimit
	}
	if settings.DefaultLimit > settings.HardLimit {
		settings.DefaultLimit = settings.HardLimit
	}

	settings.Secrets = SecretPluginSettings{
		AuthToken:     source.DecryptedSecureJSONData["authToken"],
		TLSCACert:     source.DecryptedSecureJSONData["tlsCACert"],
		TLSClientCert: source.DecryptedSecureJSONData["tlsClientCert"],
		TLSClientKey:  source.DecryptedSecureJSONData["tlsClientKey"],
	}

	return settings, nil
}

func (s *PluginSettings) RequestTimeout() time.Duration {
	return time.Duration(s.TimeoutMs) * time.Millisecond
}

func (s *PluginSettings) ClampLimit(limit int64) int64 {
	if limit <= 0 {
		return int64(s.DefaultLimit)
	}
	if limit > int64(s.HardLimit) {
		return int64(s.HardLimit)
	}
	return limit
}
