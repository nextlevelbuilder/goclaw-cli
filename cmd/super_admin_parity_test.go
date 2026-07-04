package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/nextlevelbuilder/goclaw-cli/client"
)

func TestAPIKeysRevokeUsesPostRevokeRoute(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if method != http.MethodPost || path != "/v1/api-keys/key-1/revoke" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		okJSON(t, w, map[string]any{"revoked": true})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "api-keys", "revoke", "key-1", "--yes"); err != nil {
		t.Fatalf("api-keys revoke: %v", err)
	}
	if method != http.MethodPost || path != "/v1/api-keys/key-1/revoke" {
		t.Fatalf("method/path = %s %s", method, path)
	}
}

func TestGatewayUpgradeStartSendsTriggerHeader(t *testing.T) {
	var gotHeader, gotTag string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/system/gateway/upgrade" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		gotHeader = r.Header.Get(gatewayUpgradeTokenHeader)
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotTag = body["tag"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"accepted":true}`))
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_UPGRADE_TRIGGER_TOKEN", "upgrade-secret")

	if err := runCmd(t, "system", "upgrade", "start", "--tag=v3.12.0"); err != nil {
		t.Fatalf("gateway upgrade start: %v", err)
	}
	if gotHeader != "upgrade-secret" {
		t.Fatalf("upgrade header = %q", gotHeader)
	}
	if gotTag != "v3.12.0" {
		t.Fatalf("tag = %q", gotTag)
	}
}

func TestPackagesApplyAllFailsOnPartialWithoutAllowPartial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/packages/updates/apply-all" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		okJSON(t, w, map[string]any{
			"succeeded": []any{},
			"failed":    []map[string]string{{"package": "github:gh", "reason": "boom"}},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "packages", "updates", "apply-all")
	if err == nil || !strings.Contains(err.Error(), "package updates failed") {
		t.Fatalf("expected partial failure error, got %v", err)
	}
	if _, ok := err.(*client.APIError); !ok {
		t.Fatalf("expected structured APIError, got %T", err)
	}
	if err := runCmd(t, "packages", "updates", "apply-all", "--allow-partial"); err != nil {
		t.Fatalf("allow-partial should succeed: %v", err)
	}
}

func TestWorkstationsLinkAgentUsesWSRPCContract(t *testing.T) {
	var gotMethod string
	var gotParams map[string]any
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		defer conn.Close()
		for i := 0; i < 2; i++ {
			var req struct {
				ID     string         `json:"id"`
				Method string         `json:"method"`
				Params map[string]any `json:"params"`
			}
			if err := conn.ReadJSON(&req); err != nil {
				t.Fatalf("read ws: %v", err)
			}
			if req.Method == "workstations.linkAgent" {
				gotMethod = req.Method
				gotParams = req.Params
			}
			_ = conn.WriteJSON(map[string]any{
				"type": "res", "id": req.ID, "ok": true,
				"payload": map[string]any{"ok": true},
			})
		}
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "workstations", "link-agent",
		"--agent=agent-1", "--workstation=ws-1", "--default")
	if err != nil {
		t.Fatalf("workstations link-agent: %v", err)
	}
	if gotMethod != "workstations.linkAgent" {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotParams["agentId"] != "agent-1" || gotParams["workstationId"] != "ws-1" || gotParams["isDefault"] != true {
		t.Fatalf("params = %#v", gotParams)
	}
}
