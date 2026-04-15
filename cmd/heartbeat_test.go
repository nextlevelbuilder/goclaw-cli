package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

// wsPayload is the v3 protocol frame used by mockHeartbeatServer.
type wsPayload struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// mockHeartbeatServer creates a test WebSocket server that responds to connect
// and any heartbeat.* method with a canned OK response.
func mockHeartbeatServer(t *testing.T, method string, response map[string]any) *httptest.Server {
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
			var req wsPayload
			json.Unmarshal(msg, &req)

			payload, _ := json.Marshal(response)
			resp := wsPayload{Type: "res", ID: req.ID, OK: true, Payload: payload}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	}))
}

func setupHeartbeatTest(serverURL string) {
	cfg = &config.Config{Server: serverURL, Token: "test-token", OutputFormat: "json"}
	printer = output.NewPrinter("json")
}

func httpURLToWS(u string) string {
	return strings.Replace(u, "http://", "ws://", 1)
}

func TestHeartbeatGet(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.get", map[string]any{
		"heartbeat": map[string]any{"agentId": "agent-1", "enabled": true},
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatGetCmd.Flags().Set("agent", "agent-1")
	if err := heartbeatGetCmd.RunE(heartbeatGetCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatSet(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.set", map[string]any{
		"heartbeat": map[string]any{"agentId": "agent-1", "intervalSec": 3600},
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatSetCmd.Flags().Set("agent", "agent-1")
	heartbeatSetCmd.Flags().Set("interval", "3600")
	if err := heartbeatSetCmd.RunE(heartbeatSetCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatToggle(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.toggle", map[string]any{
		"agentId": "agent-1", "enabled": true,
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatToggleCmd.Flags().Set("agent", "agent-1")
	heartbeatToggleCmd.Flags().Set("enabled", "true")
	if err := heartbeatToggleCmd.RunE(heartbeatToggleCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatTest(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.test", map[string]any{"ok": true})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatTestCmd.Flags().Set("agent", "agent-1")
	if err := heartbeatTestCmd.RunE(heartbeatTestCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatTargets(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.targets", map[string]any{
		"targets": []any{
			map[string]any{"channel": "telegram", "chat_id": "12345", "enabled": true},
		},
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	if err := heartbeatTargetsCmd.RunE(heartbeatTargetsCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatLogs(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.logs", map[string]any{
		"logs":  []any{map[string]any{"id": "log-1", "status": "ok"}},
		"total": 1,
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatLogsCmd.Flags().Set("agent", "agent-1")
	heartbeatLogsCmd.Flags().Set("follow", "false")
	heartbeatLogsCmd.Flags().Set("tail", "10")
	if err := heartbeatLogsCmd.RunE(heartbeatLogsCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatChecklistGet(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.checklist.get", map[string]any{
		"content": "## Health Check\n- [ ] API responding",
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatChecklistGetCmd.Flags().Set("agent", "agent-1")
	if err := heartbeatChecklistGetCmd.RunE(heartbeatChecklistGetCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeatChecklistSet(t *testing.T) {
	srv := mockHeartbeatServer(t, "heartbeat.checklist.set", map[string]any{
		"ok": true, "length": 10,
	})
	defer srv.Close()
	setupHeartbeatTest(srv.URL)

	heartbeatChecklistSetCmd.Flags().Set("agent", "agent-1")
	heartbeatChecklistSetCmd.Flags().Set("content", "## Health Check")
	if err := heartbeatChecklistSetCmd.RunE(heartbeatChecklistSetCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
