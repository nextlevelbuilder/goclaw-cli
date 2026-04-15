package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

func TestEdition_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/edition" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// Edition endpoint returns raw JSON (not envelope) from server,
		// but HTTPClient handles non-envelope responses transparently.
		json.NewEncoder(w).Encode(map[string]any{
			"edition": "community",
			"version": "1.0.0",
		})
	}))
	defer srv.Close()

	cfg = &config.Config{Server: srv.URL, Token: "", OutputFormat: "json"}
	printer = output.NewPrinter("json")

	if err := editionCmd.RunE(editionCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEdition_NoServer(t *testing.T) {
	cfg = &config.Config{Server: "", Token: "", OutputFormat: "json"}
	printer = output.NewPrinter("json")

	err := editionCmd.RunE(editionCmd, nil)
	if err == nil {
		t.Fatal("expected error when server not configured")
	}
}

func TestEdition_NoTokenAllowed(t *testing.T) {
	// Edition endpoint should work without a token.
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		// Server should receive request even without Authorization header.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"edition": "pro"})
	}))
	defer srv.Close()

	cfg = &config.Config{Server: srv.URL, Token: "", OutputFormat: "json"}
	printer = output.NewPrinter("json")

	if err := editionCmd.RunE(editionCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("server was never called")
	}
}
