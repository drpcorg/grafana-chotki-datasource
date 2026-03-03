package plugin

import (
	"context"
	"testing"

	"github.com/drpcorg/grafana-chotki-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestEnsureReady(t *testing.T) {
	ds := &Datasource{settings: &models.PluginSettings{GRPCAddress: ""}}
	if err := ds.ensureReady(); err == nil {
		t.Fatal("expected error when grpcAddress is missing")
	}
}

func TestGrpcErrorToResponse(t *testing.T) {
	resp := grpcErrorToResponse(context.DeadlineExceeded)
	if resp.Error == nil {
		t.Fatalf("expected timeout data response error")
	}
	if resp.Status != backend.StatusTimeout {
		t.Fatalf("expected StatusTimeout, got %d", resp.Status)
	}
}
