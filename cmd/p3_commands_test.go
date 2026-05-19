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

func setupP3CommandTest(serverURL string) {
	cfg = &config.Config{Server: serverURL, Token: "test-token", OutputFormat: "json", Yes: true}
	printer = output.NewPrinter("json")
}

func mockRPCServer(t *testing.T, seen chan<- string) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("ws upgrade: %v", err)
		}
		defer conn.Close()
		for {
			var req struct {
				Type   string          `json:"type"`
				ID     string          `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			if err := conn.ReadJSON(&req); err != nil {
				return
			}
			if seen != nil {
				seen <- req.Method
			}
			payload, _ := json.Marshal(map[string]any{"ok": true, "method": req.Method})
			_ = conn.WriteJSON(map[string]any{
				"type":    "res",
				"id":      req.ID,
				"ok":      true,
				"payload": json.RawMessage(payload),
			})
		}
	}))
}

func TestSessionsCompactUsesWSMethod(t *testing.T) {
	seen := make(chan string, 2)
	srv := mockRPCServer(t, seen)
	defer srv.Close()
	setupP3CommandTest(srv.URL)

	if err := sessionsCompactCmd.RunE(sessionsCompactCmd, []string{"session-1"}); err != nil {
		t.Fatalf("sessions compact: %v", err)
	}
	<-seen // connect
	if method := <-seen; method != "sessions.compact" {
		t.Fatalf("expected sessions.compact, got %s", method)
	}
}

func TestHealthUsesWSMethodWhenAuthenticated(t *testing.T) {
	seen := make(chan string, 2)
	srv := mockRPCServer(t, seen)
	defer srv.Close()
	setupP3CommandTest(srv.URL)

	if err := healthCmd.RunE(healthCmd, nil); err != nil {
		t.Fatalf("health: %v", err)
	}
	<-seen // connect
	if method := <-seen; method != "health" {
		t.Fatalf("expected health, got %s", method)
	}
}

func TestTracesListAddsP3Filters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/traces" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		if q.Get("agent_id") != "agent-1" || q.Get("status") != "error" ||
			q.Get("since") != "1h" || q.Get("root_only") != "true" || q.Get("limit") != "5" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		okJSON(t, w, []map[string]any{{"trace_id": "trace-1"}})
	}))
	defer srv.Close()
	setupP3CommandTest(srv.URL)

	_ = tracesListCmd.Flags().Set("agent", "agent-1")
	_ = tracesListCmd.Flags().Set("status", "error")
	_ = tracesListCmd.Flags().Set("since", "1h")
	_ = tracesListCmd.Flags().Set("root-only", "true")
	_ = tracesListCmd.Flags().Set("limit", "5")
	t.Cleanup(func() {
		_ = tracesListCmd.Flags().Set("agent", "")
		_ = tracesListCmd.Flags().Set("status", "")
		_ = tracesListCmd.Flags().Set("since", "")
		_ = tracesListCmd.Flags().Set("root-only", "false")
		_ = tracesListCmd.Flags().Set("limit", "20")
	})

	if err := tracesListCmd.RunE(tracesListCmd, nil); err != nil {
		t.Fatalf("traces list: %v", err)
	}
}
