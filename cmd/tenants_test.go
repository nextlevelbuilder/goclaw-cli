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

// setupTenantTest wires cfg and printer for a test server URL.
func setupTenantTest(serverURL string) {
	cfg = &config.Config{Server: serverURL, Token: "test-token", OutputFormat: "json"}
	printer = output.NewPrinter("json")
}

func TestTenantsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/tenants" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"payload": map[string]any{
				"tenants": []any{
					map[string]any{"id": "t1", "name": "Acme", "slug": "acme", "status": "active"},
				},
			},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)

	if err := tenantsListCmd.RunE(tenantsListCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTenantsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenants/t1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"id": "t1", "name": "Acme"},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)

	if err := tenantsGetCmd.RunE(tenantsGetCmd, []string{"t1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTenantsCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tenants" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Acme" || body["slug"] != "acme" {
			t.Errorf("unexpected body: %v", body)
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"id": "new-t1"},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)

	tenantsCreateCmd.Flags().Set("name", "Acme")
	tenantsCreateCmd.Flags().Set("slug", "acme")
	if err := tenantsCreateCmd.RunE(tenantsCreateCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTenantsUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/tenants/t1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"ok": "true"},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)

	tenantsUpdateCmd.Flags().Set("name", "NewName")
	if err := tenantsUpdateCmd.RunE(tenantsUpdateCmd, []string{"t1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTenantsUsersList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenants/t1/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"payload": map[string]any{
				"users": []any{
					map[string]any{"user_id": "u1", "role": "admin"},
				},
			},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)

	if err := tenantsUsersListCmd.RunE(tenantsUsersListCmd, []string{"t1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTenantsUsersAdd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/tenants/t1/users" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["user_id"] != "u1" {
			t.Errorf("unexpected user_id: %v", body["user_id"])
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"ok": "true"},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)

	tenantsUsersAddCmd.Flags().Set("user-id", "u1")
	tenantsUsersAddCmd.Flags().Set("role", "member")
	if err := tenantsUsersAddCmd.RunE(tenantsUsersAddCmd, []string{"t1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTenantsUsersRemove_ConfirmMismatch(t *testing.T) {
	setupTenantTest("http://localhost")
	cfg.Yes = false

	tenantsUsersRemoveCmd.Flags().Set("confirm", "wrong-id")
	err := tenantsUsersRemoveCmd.RunE(tenantsUsersRemoveCmd, []string{"t1", "u1"})
	if err == nil || !strings.Contains(err.Error(), "confirmation mismatch") {
		t.Fatalf("expected confirmation mismatch error, got: %v", err)
	}
}

func TestTenantsUsersRemove_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/tenants/t1/users/u1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"ok": "true"},
		})
	}))
	defer srv.Close()
	setupTenantTest(srv.URL)
	cfg.Yes = true // skip interactive prompt

	tenantsUsersRemoveCmd.Flags().Set("confirm", "u1")
	if err := tenantsUsersRemoveCmd.RunE(tenantsUsersRemoveCmd, []string{"t1", "u1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToList(t *testing.T) {
	input := []any{
		map[string]any{"id": "1"},
		map[string]any{"id": "2"},
	}
	out := toList(input)
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	if out[0]["id"] != "1" {
		t.Errorf("expected id=1, got %v", out[0]["id"])
	}
}

func TestToList_Nil(t *testing.T) {
	if toList(nil) != nil {
		t.Error("expected nil for nil input")
	}
}
