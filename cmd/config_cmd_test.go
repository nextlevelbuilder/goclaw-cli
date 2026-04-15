package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

// mockConfigWSServer creates a WS test server that responds to any method with canned payload.
func mockConfigWSServer(t *testing.T, response map[string]any) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("ws upgrade: %v", err)
		}
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var req struct {
				ID string `json:"id"`
			}
			json.Unmarshal(msg, &req)
			payload, _ := json.Marshal(response)
			conn.WriteJSON(map[string]any{
				"type": "res", "id": req.ID, "ok": true, "payload": json.RawMessage(payload),
			})
		}
	}))
}

func setupConfigTest(serverURL string) {
	cfg = &config.Config{Server: serverURL, Token: "test-token", OutputFormat: "json"}
	printer = output.NewPrinter("json")
}

func TestConfigGet(t *testing.T) {
	srv := mockConfigWSServer(t, map[string]any{"key": "agent.model", "value": "gpt-4o"})
	defer srv.Close()
	setupConfigTest(srv.URL)

	configGetCmd.Flags().Set("key", "agent.model")
	if err := configGetCmd.RunE(configGetCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigSchema(t *testing.T) {
	srv := mockConfigWSServer(t, map[string]any{"schema": map[string]any{}})
	defer srv.Close()
	setupConfigTest(srv.URL)

	if err := configSchemaCmd.RunE(configSchemaCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigPatch(t *testing.T) {
	srv := mockConfigWSServer(t, map[string]any{"ok": true})
	defer srv.Close()
	setupConfigTest(srv.URL)

	configPatchCmd.Flags().Set("key", "agent.model")
	configPatchCmd.Flags().Set("value", "gpt-4o")
	if err := configPatchCmd.RunE(configPatchCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Config Permissions ---

func TestConfigPermissionsList(t *testing.T) {
	srv := mockConfigWSServer(t, map[string]any{
		"permissions": []any{
			map[string]any{"agentId": "a1", "userId": "u1", "permission": "read"},
		},
	})
	defer srv.Close()
	setupConfigTest(srv.URL)

	configPermissionsListCmd.Flags().Set("agent", "a1")
	if err := configPermissionsListCmd.RunE(configPermissionsListCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigPermissionsGrant(t *testing.T) {
	srv := mockConfigWSServer(t, map[string]any{"ok": true})
	defer srv.Close()
	setupConfigTest(srv.URL)

	configPermissionsGrantCmd.Flags().Set("agent", "a1")
	configPermissionsGrantCmd.Flags().Set("user", "u1")
	configPermissionsGrantCmd.Flags().Set("scope", "agent")
	configPermissionsGrantCmd.Flags().Set("config-type", "system")
	configPermissionsGrantCmd.Flags().Set("permission", "read")
	if err := configPermissionsGrantCmd.RunE(configPermissionsGrantCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigPermissionsRevoke(t *testing.T) {
	srv := mockConfigWSServer(t, map[string]any{"ok": true})
	defer srv.Close()
	setupConfigTest(srv.URL)
	cfg.Yes = true

	configPermissionsRevokeCmd.Flags().Set("agent", "a1")
	configPermissionsRevokeCmd.Flags().Set("user", "u1")
	configPermissionsRevokeCmd.Flags().Set("scope", "agent")
	configPermissionsRevokeCmd.Flags().Set("config-type", "system")
	if err := configPermissionsRevokeCmd.RunE(configPermissionsRevokeCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
