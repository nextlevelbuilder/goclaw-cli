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

// mockAgentsWSServer creates a WS server responding to any method with a canned response.
func mockAgentsWSServer(t *testing.T, responses map[string]any) *httptest.Server {
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

// mockAgentsHTTPServer creates an HTTP test server for agent HTTP endpoints.
func mockAgentsHTTPServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func setupAgentsTest(serverURL string) {
	cfg = &config.Config{
		Server:       serverURL,
		Token:        "test-token",
		OutputFormat: "json",
	}
	printer = output.NewPrinter("json")
}

func setupAgentsWSTest(serverURL string) {
	cfg = &config.Config{
		Server:       strings.Replace(serverURL, "http://", "ws://", 1),
		Token:        "test-token",
		OutputFormat: "json",
	}
	printer = output.NewPrinter("json")
}

// --- agents wake ---

func TestAgentsWake(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/wake") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"waking"}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsWakeCmd.RunE(agentsWakeCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents identity (AI-critical) ---

func TestAgentsIdentity_ReturnsIdentityJSON(t *testing.T) {
	srv := mockAgentsWSServer(t, map[string]any{
		"agent.identity.get": map[string]any{
			"agent_key":    "my-agent",
			"display_name": "My Agent",
			"persona":      "helpful assistant",
			"traits":       []string{"curious", "precise"},
		},
	})
	defer srv.Close()
	setupAgentsWSTest(srv.URL)

	if err := agentsIdentityCmd.RunE(agentsIdentityCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents wait (AI-critical: blocking + timeout) ---

func TestAgentsWait_SuccessfulStateMatch(t *testing.T) {
	srv := mockAgentsWSServer(t, map[string]any{
		"agent.wait": map[string]any{
			"agent_key":  "my-agent",
			"state":      "idle",
			"reached_at": "2024-01-01T00:00:00Z",
		},
	})
	defer srv.Close()
	setupAgentsWSTest(srv.URL)

	agentsWaitCmd.Flags().Set("state", "idle")
	agentsWaitCmd.Flags().Set("timeout", "10s")

	if err := agentsWaitCmd.RunE(agentsWaitCmd, []string{"my-agent"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsWait_InvalidTimeout(t *testing.T) {
	srv := mockAgentsWSServer(t, map[string]any{})
	defer srv.Close()
	setupAgentsWSTest(srv.URL)

	agentsWaitCmd.Flags().Set("state", "online")
	agentsWaitCmd.Flags().Set("timeout", "notaduration")

	err := agentsWaitCmd.RunE(agentsWaitCmd, []string{"my-agent"})
	if err == nil {
		t.Fatal("expected parse error for invalid timeout")
	}
	if !strings.Contains(err.Error(), "invalid --timeout") {
		t.Fatalf("expected timeout parse error, got: %v", err)
	}
}

func TestAgentsWait_DefaultTimeout(t *testing.T) {
	// Server responds immediately — no timeout hit.
	srv := mockAgentsWSServer(t, map[string]any{
		"agent.wait": map[string]any{"state": "online"},
	})
	defer srv.Close()
	setupAgentsWSTest(srv.URL)

	agentsWaitCmd.Flags().Set("state", "online")
	agentsWaitCmd.Flags().Set("timeout", "5m")

	if err := agentsWaitCmd.RunE(agentsWaitCmd, []string{"agent-x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents sync-workspace ---

func TestAgentsSyncWorkspace(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/agents/sync-workspace" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"synced":true}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsSyncWorkspaceCmd.RunE(agentsSyncWorkspaceCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents prompt-preview ---

func TestAgentsPromptPreview(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/system-prompt-preview") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"prompt":"You are a helpful assistant..."}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsPromptPreviewCmd.RunE(agentsPromptPreviewCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents evolution ---

func TestAgentsEvolutionMetrics(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/evolution/metrics") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"total_turns":100,"avg_latency_ms":250}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsEvolutionMetricsCmd.RunE(agentsEvolutionMetricsCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsEvolutionSuggestions(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"sugg-1","type":"prompt","description":"Improve brevity"}]`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsEvolutionSuggestionsCmd.RunE(agentsEvolutionSuggestionsCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsEvolutionUpdate_Accept(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	agentsEvolutionUpdateCmd.Flags().Set("action", "accept")

	if err := agentsEvolutionUpdateCmd.RunE(agentsEvolutionUpdateCmd, []string{"agent-1", "sugg-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsEvolutionUpdate_InvalidAction(t *testing.T) {
	agentsEvolutionUpdateCmd.Flags().Set("action", "ignore")
	err := agentsEvolutionUpdateCmd.RunE(agentsEvolutionUpdateCmd, []string{"agent-1", "sugg-1"})
	if err == nil {
		t.Fatal("expected validation error for invalid action")
	}
	if !strings.Contains(err.Error(), "--action must be") {
		t.Fatalf("expected action validation error, got: %v", err)
	}
}

// --- agents episodic ---

func TestAgentsEpisodicList(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/episodic") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"ep-1","type":"interaction","summary":"User asked about deployment"}]`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsEpisodicListCmd.RunE(agentsEpisodicListCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsEpisodicSearch(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"ep-2","score":0.92,"summary":"Deployment discussion"}]`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsEpisodicSearchCmd.RunE(agentsEpisodicSearchCmd, []string{"agent-1", "deployment"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents v3-flags ---

func TestAgentsV3FlagsGet(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"kg_auto_extract":true,"multi_session":false}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsV3FlagsGetCmd.RunE(agentsV3FlagsGetCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsV3FlagsToggle(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	agentsV3FlagsToggleCmd.Flags().Set("flag", "kg_auto_extract")

	if err := agentsV3FlagsToggleCmd.RunE(agentsV3FlagsToggleCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents orchestration + codex-pool-activity ---

func TestAgentsOrchestration(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"mode":"auto","max_delegates":5}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsOrchestrationCmd.RunE(agentsOrchestrationCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsCodexPoolActivity(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"active_contexts":3,"total_tokens":12000}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsCodexPoolActivityCmd.RunE(agentsCodexPoolActivityCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- agents instances ---

func TestAgentsInstancesList(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"user_id":"user-1","status":"active"}]`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	if err := agentsInstancesListCmd.RunE(agentsInstancesListCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsInstancesGetFile(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"content":"file content here"}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	agentsInstancesGetFileCmd.Flags().Set("user", "user-1")
	agentsInstancesGetFileCmd.Flags().Set("file", "context.md")

	if err := agentsInstancesGetFileCmd.RunE(agentsInstancesGetFileCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsInstancesUpdateMetadata_InvalidJSON(t *testing.T) {
	agentsInstancesUpdateMetadataCmd.Flags().Set("user", "user-1")
	agentsInstancesUpdateMetadataCmd.Flags().Set("metadata", "not-json")

	err := agentsInstancesUpdateMetadataCmd.RunE(agentsInstancesUpdateMetadataCmd, []string{"agent-1"})
	if err == nil {
		t.Fatal("expected JSON parse error")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("expected JSON error, got: %v", err)
	}
}

func TestAgentsInstancesUpdateMetadata_ValidJSON(t *testing.T) {
	srv := mockAgentsHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()
	setupAgentsTest(srv.URL)

	agentsInstancesUpdateMetadataCmd.Flags().Set("user", "user-1")
	agentsInstancesUpdateMetadataCmd.Flags().Set("metadata", `{"tier":"premium"}`)

	if err := agentsInstancesUpdateMetadataCmd.RunE(agentsInstancesUpdateMetadataCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
