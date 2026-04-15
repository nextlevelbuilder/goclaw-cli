package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

func setupSysConfigTest(serverURL string) {
	cfg = &config.Config{Server: serverURL, Token: "test-token", OutputFormat: "json"}
	printer = output.NewPrinter("json")
}

func TestSystemConfigsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/system-configs" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"agent.default_model": "gpt-4o"},
		})
	}))
	defer srv.Close()
	setupSysConfigTest(srv.URL)

	if err := systemConfigsListCmd.RunE(systemConfigsListCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSystemConfigsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/system-configs/agent.default_model" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"key": "agent.default_model", "value": "gpt-4o"},
		})
	}))
	defer srv.Close()
	setupSysConfigTest(srv.URL)

	if err := systemConfigsGetCmd.RunE(systemConfigsGetCmd, []string{"agent.default_model"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSystemConfigsSet_StringValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v1/system-configs/agent.default_model" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["value"] != "gpt-4o" {
			t.Errorf("unexpected value: %v", body["value"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"key": "agent.default_model", "value": "gpt-4o"},
		})
	}))
	defer srv.Close()
	setupSysConfigTest(srv.URL)

	// Reset json flag
	systemConfigsSetCmd.Flags().Set("json", "false")
	if err := systemConfigsSetCmd.RunE(systemConfigsSetCmd, []string{"agent.default_model", "gpt-4o"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSystemConfigsSet_JSONValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		// When --json is set the value field is parsed as JSON object.
		if _, ok := body["value"].(map[string]any); !ok {
			t.Errorf("expected JSON object value, got %T", body["value"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"key": "feature.flags"},
		})
	}))
	defer srv.Close()
	setupSysConfigTest(srv.URL)

	systemConfigsSetCmd.Flags().Set("json", "true")
	defer systemConfigsSetCmd.Flags().Set("json", "false")

	if err := systemConfigsSetCmd.RunE(systemConfigsSetCmd, []string{"feature.flags", `{"beta":true}`}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSystemConfigsSet_InvalidJSON(t *testing.T) {
	setupSysConfigTest("http://localhost")

	systemConfigsSetCmd.Flags().Set("json", "true")
	defer systemConfigsSetCmd.Flags().Set("json", "false")

	err := systemConfigsSetCmd.RunE(systemConfigsSetCmd, []string{"key", "not-json"})
	if err == nil || !strings.Contains(err.Error(), "not valid JSON") {
		t.Fatalf("expected JSON parse error, got: %v", err)
	}
}

func TestSystemConfigsDelete_WithYes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/system-configs/feature.beta" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	setupSysConfigTest(srv.URL)
	cfg.Yes = true

	if err := systemConfigsDeleteCmd.RunE(systemConfigsDeleteCmd, []string{"feature.beta"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
