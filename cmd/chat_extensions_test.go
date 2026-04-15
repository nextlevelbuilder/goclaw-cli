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

// mockChatServer creates a test WebSocket server responding to chat.* methods.
func mockChatServer(t *testing.T, responses map[string]any) *httptest.Server {
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
				Type   string          `json:"type"`
				ID     string          `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			if err := json.Unmarshal(msg, &req); err != nil {
				return
			}
			respData, ok := responses[req.Method]
			if !ok {
				respData = map[string]any{"ok": true}
			}
			payload, _ := json.Marshal(respData)
			resp := map[string]any{
				"type":    "res",
				"id":      req.ID,
				"ok":      true,
				"payload": json.RawMessage(payload),
			}
			if err := conn.WriteJSON(resp); err != nil {
				return
			}
		}
	}))
}

func setupChatTest(serverURL string) {
	cfg = &config.Config{
		Server:       strings.Replace(serverURL, "http://", "ws://", 1),
		Token:        "test-token",
		OutputFormat: "json",
	}
	printer = output.NewPrinter("json")
}

// --- chat history tests ---

func TestChatHistory_ReturnsMessageArray(t *testing.T) {
	messages := []any{
		map[string]any{"role": "user", "content": "Hello", "created_at": "2024-01-01T00:00:00Z"},
		map[string]any{"role": "assistant", "content": "Hi there!", "created_at": "2024-01-01T00:00:01Z"},
	}
	srv := mockChatServer(t, map[string]any{"chat.history": messages})
	defer srv.Close()
	setupChatTest(srv.URL)

	// Reset flags to defaults
	chatHistoryCmd.Flags().Set("limit", "50")
	chatHistoryCmd.Flags().Set("before", "")
	chatHistoryCmd.Flags().Set("session", "")

	if err := chatHistoryCmd.RunE(chatHistoryCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatHistory_WithLimitAndBefore(t *testing.T) {
	srv := mockChatServer(t, map[string]any{"chat.history": []any{}})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatHistoryCmd.Flags().Set("limit", "10")
	chatHistoryCmd.Flags().Set("before", "2024-06-01T00:00:00Z")
	chatHistoryCmd.Flags().Set("session", "")

	if err := chatHistoryCmd.RunE(chatHistoryCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatHistory_WithSession(t *testing.T) {
	srv := mockChatServer(t, map[string]any{"chat.history": []any{}})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatHistoryCmd.Flags().Set("limit", "50")
	chatHistoryCmd.Flags().Set("before", "")
	chatHistoryCmd.Flags().Set("session", "sess-abc")

	if err := chatHistoryCmd.RunE(chatHistoryCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- chat inject tests ---

func TestChatInject_UserRole(t *testing.T) {
	srv := mockChatServer(t, map[string]any{
		"chat.inject": map[string]any{"injected": true, "message_id": "msg-1"},
	})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatInjectCmd.Flags().Set("role", "user")
	chatInjectCmd.Flags().Set("content", "Test message")
	chatInjectCmd.Flags().Set("session", "")

	if err := chatInjectCmd.RunE(chatInjectCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatInject_SystemRole(t *testing.T) {
	srv := mockChatServer(t, map[string]any{
		"chat.inject": map[string]any{"injected": true, "message_id": "msg-2"},
	})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatInjectCmd.Flags().Set("role", "system")
	chatInjectCmd.Flags().Set("content", "You are a helpful assistant.")
	chatInjectCmd.Flags().Set("session", "")

	if err := chatInjectCmd.RunE(chatInjectCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatInject_InvalidRole(t *testing.T) {
	srv := mockChatServer(t, map[string]any{})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatInjectCmd.Flags().Set("role", "bot") // invalid
	chatInjectCmd.Flags().Set("content", "hello")
	chatInjectCmd.Flags().Set("session", "")

	err := chatInjectCmd.RunE(chatInjectCmd, []string{"my-agent"})
	if err == nil {
		t.Fatal("expected error for invalid role, got nil")
	}
	if !strings.Contains(err.Error(), "--role must be") {
		t.Fatalf("expected role validation error, got: %v", err)
	}
}

func TestChatInject_EmptyContent(t *testing.T) {
	srv := mockChatServer(t, map[string]any{})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatInjectCmd.Flags().Set("role", "user")
	chatInjectCmd.Flags().Set("content", "") // empty — invalid after readContent
	chatInjectCmd.Flags().Set("session", "")

	err := chatInjectCmd.RunE(chatInjectCmd, []string{"my-agent"})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestChatInject_WithSession(t *testing.T) {
	srv := mockChatServer(t, map[string]any{
		"chat.inject": map[string]any{"injected": true, "session_key": "sess-1"},
	})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatInjectCmd.Flags().Set("role", "assistant")
	chatInjectCmd.Flags().Set("content", "Prior response context")
	chatInjectCmd.Flags().Set("session", "sess-1")

	if err := chatInjectCmd.RunE(chatInjectCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- chat session-status tests ---

func TestChatSessionStatus_ReturnsState(t *testing.T) {
	srv := mockChatServer(t, map[string]any{
		"chat.session.status": map[string]any{
			"agent_key":   "my-agent",
			"state":       "idle",
			"turn_count":  5,
			"last_active": "2024-01-01T12:00:00Z",
		},
	})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatSessionStatusCmd.Flags().Set("session", "")

	if err := chatSessionStatusCmd.RunE(chatSessionStatusCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatSessionStatus_WithSession(t *testing.T) {
	srv := mockChatServer(t, map[string]any{
		"chat.session.status": map[string]any{
			"state": "running", "session_key": "sess-42",
		},
	})
	defer srv.Close()
	setupChatTest(srv.URL)

	chatSessionStatusCmd.Flags().Set("session", "sess-42")

	if err := chatSessionStatusCmd.RunE(chatSessionStatusCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
