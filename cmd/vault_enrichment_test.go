package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVaultEnrichmentStatus_CallsEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/vault/enrichment/status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{
			"running":   true,
			"total":     10,
			"done":      4,
			"failed":    0,
			"percent":   40,
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "enrichment", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultEnrichmentStop_RequiresYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/vault/enrichment/stop" {
			called = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{"stopped": true}))
	}))
	defer srv.Close()

	// Without --yes, non-interactive mode must refuse.
	err := runVaultArgs(t, srv.URL, "enrichment", "stop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("enrichment stop should not be called without --yes in non-interactive mode")
	}
}

func TestVaultEnrichmentStop_WithYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/vault/enrichment/stop" {
			called = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{"stopped": true}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "enrichment", "stop", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected enrichment stop endpoint to be called with --yes")
	}
}
