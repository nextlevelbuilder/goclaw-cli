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

// mockTeamsWSServer creates a WS server for teams method testing.
func mockTeamsWSServer(t *testing.T, responses map[string]any) *httptest.Server {
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
				Type   string `json:"type"`
				ID     string `json:"id"`
				Method string `json:"method"`
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

func setupTeamsTest(serverURL string) {
	cfg = &config.Config{
		Server:       strings.Replace(serverURL, "http://", "ws://", 1),
		Token:        "test-token",
		OutputFormat: "json",
	}
	printer = output.NewPrinter("json")
}

// --- teams tasks list / get ---

func TestTeamsTasksList(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.tasks.list": []any{
			map[string]any{"id": "task-1", "title": "Build feature", "status": "open"},
		},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	teamsTasksListCmd.Flags().Set("status", "")
	if err := teamsTasksListCmd.RunE(teamsTasksListCmd, []string{"team-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamsTasksGet(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.tasks.get": map[string]any{"id": "task-1", "title": "Build feature", "description": "Full details"},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	if err := teamsTasksGetCmd.RunE(teamsTasksGetCmd, []string{"team-1", "task-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamsTasksGetLight(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.tasks.get-light": map[string]any{"id": "task-1", "title": "Build feature", "status": "open"},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	if err := teamsTasksGetLightCmd.RunE(teamsTasksGetLightCmd, []string{"team-1", "task-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- teams tasks create / assign ---

func TestTeamsTasksCreate(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.tasks.create": map[string]any{"id": "task-new", "title": "New task"},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	teamsTasksCreateCmd.Flags().Set("title", "New task")
	teamsTasksCreateCmd.Flags().Set("description", "")
	teamsTasksCreateCmd.Flags().Set("assignee", "")
	if err := teamsTasksCreateCmd.RunE(teamsTasksCreateCmd, []string{"team-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamsTasksAssign(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{"teams.tasks.assign": map[string]any{"ok": true}})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	teamsTasksAssignCmd.Flags().Set("agent", "agent-1")
	if err := teamsTasksAssignCmd.RunE(teamsTasksAssignCmd, []string{"team-1", "task-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- teams tasks delete ---

func TestTeamsTasksDelete_WithYes(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{"teams.tasks.delete": map[string]any{"ok": true}})
	defer srv.Close()
	cfg = &config.Config{
		Server:       strings.Replace(srv.URL, "http://", "ws://", 1),
		Token:        "test-token",
		OutputFormat: "json",
		Yes:          true,
	}
	printer = output.NewPrinter("json")

	if err := teamsTasksDeleteCmd.RunE(teamsTasksDeleteCmd, []string{"team-1", "task-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamsTasksDelete_ConfirmDeclined(t *testing.T) {
	// Yes=false → Confirm returns false in non-interactive → no-op
	cfg = &config.Config{
		Server:       "ws://localhost",
		Token:        "test-token",
		OutputFormat: "json",
		Yes:          false,
	}
	printer = output.NewPrinter("json")

	err := teamsTasksDeleteCmd.RunE(teamsTasksDeleteCmd, []string{"team-1", "task-1"})
	if err != nil {
		t.Fatalf("expected nil when confirm declined, got: %v", err)
	}
}

// --- teams tasks delete-bulk ---

func TestTeamsTasksDeleteBulk_WithYes(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{"teams.tasks.delete-bulk": map[string]any{"ok": true}})
	defer srv.Close()
	cfg = &config.Config{
		Server:       strings.Replace(srv.URL, "http://", "ws://", 1),
		Token:        "test-token",
		OutputFormat: "json",
		Yes:          true,
	}
	printer = output.NewPrinter("json")

	teamsTasksDeleteBulkCmd.Flags().Set("ids", "task-1,task-2,task-3")
	if err := teamsTasksDeleteBulkCmd.RunE(teamsTasksDeleteBulkCmd, []string{"team-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamsTasksDeleteBulk_MissingIDs(t *testing.T) {
	cfg = &config.Config{Server: "ws://localhost", Token: "tok", OutputFormat: "json", Yes: true}
	printer = output.NewPrinter("json")

	teamsTasksDeleteBulkCmd.Flags().Set("ids", "")
	err := teamsTasksDeleteBulkCmd.RunE(teamsTasksDeleteBulkCmd, []string{"team-1"})
	if err == nil {
		t.Fatal("expected error when --ids is empty")
	}
	if !strings.Contains(err.Error(), "--ids is required") {
		t.Fatalf("expected ids required error, got: %v", err)
	}
}

// --- teams tasks events (one-shot) ---

func TestTeamsTasksEvents_OneShot(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.tasks.events": []any{
			map[string]any{"type": "status_changed", "data": map[string]any{"status": "assigned"}},
		},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	teamsTasksEventsCmd.Flags().Set("follow", "false")
	if err := teamsTasksEventsCmd.RunE(teamsTasksEventsCmd, []string{"team-1", "task-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- teams tasks active ---

func TestTeamsTasksActive(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.tasks.active-by-session": []any{
			map[string]any{"id": "task-1", "status": "running"},
		},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	teamsTasksActiveCmd.Flags().Set("session", "sess-abc")
	if err := teamsTasksActiveCmd.RunE(teamsTasksActiveCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamsTasksActive_MissingSession(t *testing.T) {
	cfg = &config.Config{Server: "ws://localhost", Token: "tok", OutputFormat: "json"}
	printer = output.NewPrinter("json")

	teamsTasksActiveCmd.Flags().Set("session", "")
	err := teamsTasksActiveCmd.RunE(teamsTasksActiveCmd, nil)
	if err == nil {
		t.Fatal("expected error when --session is missing")
	}
	if !strings.Contains(err.Error(), "--session is required") {
		t.Fatalf("expected session required error, got: %v", err)
	}
}

// --- teams scopes ---

func TestTeamsScopes(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.scopes": map[string]any{"scopes": []string{"tasks.create", "tasks.approve"}},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	if err := teamsScopesCmd.RunE(teamsScopesCmd, []string{"team-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- teams events list (one-shot) ---

func TestTeamsEventsList_OneShot(t *testing.T) {
	srv := mockTeamsWSServer(t, map[string]any{
		"teams.events.list": []any{
			map[string]any{"id": "evt-1", "type": "task_created"},
		},
	})
	defer srv.Close()
	setupTeamsTest(srv.URL)

	teamsEventsListCmd.Flags().Set("follow", "false")
	if err := teamsEventsListCmd.RunE(teamsEventsListCmd, []string{"team-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
