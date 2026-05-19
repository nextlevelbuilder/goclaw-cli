package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

func TestCodexPoolActivityRoutes(t *testing.T) {
	defer resetTestFlag(codexPoolActivityCmd, "agent", "")
	defer resetTestFlag(codexPoolActivityCmd, "provider", "")
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		okJSON(t, w, map[string]any{"ok": true})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "codex-pool", "activity", "--agent=agent-1"); err != nil {
		t.Fatalf("codex-pool agent: %v", err)
	}
	_ = codexPoolActivityCmd.Flags().Set("agent", "")
	if err := runCmd(t, "codex-pool", "activity", "--provider=provider-1"); err != nil {
		t.Fatalf("codex-pool provider: %v", err)
	}
	want := []string{
		"/v1/agents/agent-1/codex-pool-activity",
		"/v1/providers/provider-1/codex-pool-activity",
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("path[%d] = %q, want %q", i, paths[i], want[i])
		}
	}
}

func TestAPIKeysRotatePartialFailureMapsExitFive(t *testing.T) {
	defer resetTestFlag(apiKeysRotateCmd, "name", "")
	defer resetTestFlag(apiKeysRotateCmd, "scopes", "")
	defer resetTestFlag(apiKeysRotateCmd, "expires-in", "0")
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/v1/api-keys":
			okJSON(t, w, map[string]any{"id": "new-key", "key": "goclaw_new_raw", "name": "rotated"})
		case "/v1/api-keys/old-key/revoke":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"ok":false,"error":{"code":"INTERNAL","message":"boom"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	stdout, err := captureStdout(t, func() error {
		return runCmd(t, "api-keys", "rotate", "old-key", "--name=rotated", "--scopes=operator.read", "--yes")
	})
	if err == nil {
		t.Fatal("expected partial failure")
	}
	if strings.Count(stdout, "goclaw_new_raw") != 1 {
		t.Fatalf("raw key count in stdout = %d, stdout = %s", strings.Count(stdout, "goclaw_new_raw"), stdout)
	}
	if got := output.FromError(err); got != output.ExitServer {
		t.Fatalf("exit = %d, want %d", got, output.ExitServer)
	}
	var detail *output.ErrorDetail
	if !errors.As(err, &detail) {
		t.Fatalf("expected ErrorDetail, got %T", err)
	}
	details, ok := detail.Details.(map[string]any)
	if !ok {
		t.Fatalf("details = %#v", detail.Details)
	}
	if _, ok := details["new_key"]; ok {
		t.Fatalf("partial failure details must not repeat raw key: %#v", details)
	}
	if details["new_key_id"] != "new-key" || details["old_key_id"] != "old-key" ||
		details["old_revoke_status"] != "failed" {
		t.Fatalf("details = %#v", details)
	}
	if len(paths) != 2 || paths[0] != "/v1/api-keys" || paths[1] != "/v1/api-keys/old-key/revoke" {
		t.Fatalf("paths = %#v", paths)
	}
}

func TestAPIKeysCreateUsesSharedScopeParser(t *testing.T) {
	defer resetTestFlag(apiKeysCreateCmd, "name", "")
	defer resetTestFlag(apiKeysCreateCmd, "scopes", "")
	defer resetTestFlag(apiKeysCreateCmd, "expires-in", "0")
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/api-keys" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		okJSON(t, w, map[string]any{"id": "new-key", "key": "goclaw_new_raw"})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	err := runCmd(t, "api-keys", "create", "--name=ci", "--scopes= operator.read, , operator.write ")
	if err != nil {
		t.Fatalf("api-keys create: %v", err)
	}
	scopes := body["scopes"].([]any)
	if len(scopes) != 2 || scopes[0] != "operator.read" || scopes[1] != "operator.write" {
		t.Fatalf("scopes = %#v", scopes)
	}
}

func TestLegacyCodexPoolAliasesUseCobraDeprecationMetadata(t *testing.T) {
	if agentsCodexPoolActivityCmd.Deprecated == "" {
		t.Fatal("agents codex-pool-activity should be marked deprecated")
	}
	if providersCodexPoolActivityCmd.Deprecated == "" {
		t.Fatal("providers codex-pool-activity should be marked deprecated")
	}
}

func TestConfigDefaultsUsesWSMethod(t *testing.T) {
	seen := make(chan string, 2)
	srv := mockRPCServer(t, seen)
	defer srv.Close()
	setupP3CommandTest(srv.URL)

	if err := configDefaultsCmd.RunE(configDefaultsCmd, nil); err != nil {
		t.Fatalf("config defaults: %v", err)
	}
	<-seen // connect
	if method := <-seen; method != "config.defaults" {
		t.Fatalf("method = %q", method)
	}
}

func TestToolsInvokeArgsReadsFile(t *testing.T) {
	defer resetTestFlag(toolsInvokeCmd, "args", "")
	defer resetTestFlag(toolsInvokeCmd, "param", "")
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tools/invoke" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		okJSON(t, w, map[string]any{"ok": true})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	path := filepath.Join(t.TempDir(), "args.json")
	if err := os.WriteFile(path, []byte(`{"city":"Saigon"}`), 0o600); err != nil {
		t.Fatalf("write args: %v", err)
	}
	if err := runCmd(t, "tools", "invoke", "weather", "--args=@"+path, "--param=unit=c"); err != nil {
		t.Fatalf("tools invoke: %v", err)
	}
	params := body["parameters"].(map[string]any)
	if params["city"] != "Saigon" || params["unit"] != "c" {
		t.Fatalf("params = %#v", params)
	}
}

func TestChatReplayAndResumeUseExistingWSContracts(t *testing.T) {
	defer resetTestFlag(chatReplayCmd, "session", "")
	defer resetTestFlag(chatReplayCmd, "before", "")
	defer resetTestFlag(chatReplayCmd, "limit", "50")
	defer resetTestFlag(chatSessionsResumeCmd, "session", "")
	defer resetTestFlag(chatSessionsResumeCmd, "message", "")
	defer resetTestFlag(chatSessionsResumeCmd, "no-stream", "false")
	seen := make(chan wsCall, 4)
	srv := mockCaptureRPCServer(t, seen)
	defer srv.Close()
	setupP3CommandTest(srv.URL)

	_ = chatReplayCmd.Flags().Set("session", "sess-1")
	_ = chatReplayCmd.Flags().Set("before", "")
	_ = chatReplayCmd.Flags().Set("limit", "50")
	if err := chatReplayCmd.RunE(chatReplayCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("chat replay: %v", err)
	}
	<-seen // connect
	history := <-seen
	if history.Method != "chat.history" || history.Params["agent_key"] != "agent-1" ||
		history.Params["session_key"] != "sess-1" {
		t.Fatalf("history call = %#v", history)
	}

	_ = chatSessionsResumeCmd.Flags().Set("session", "sess-2")
	_ = chatSessionsResumeCmd.Flags().Set("message", "hi")
	_ = chatSessionsResumeCmd.Flags().Set("no-stream", "true")
	if err := chatSessionsResumeCmd.RunE(chatSessionsResumeCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("chat sessions resume: %v", err)
	}
	<-seen // connect
	send := <-seen
	if send.Method != "chat.send" || send.Params["agent_key"] != "agent-1" ||
		send.Params["session_key"] != "sess-2" || send.Params["message"] != "hi" {
		t.Fatalf("send call = %#v", send)
	}
}

func resetTestFlag(cmd *cobra.Command, name, value string) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return
	}
	_ = flag.Value.Set(value)
	flag.Changed = false
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fnErr := fn()
	_ = w.Close()
	os.Stdout = old
	data, readErr := io.ReadAll(r)
	_ = r.Close()
	if readErr != nil {
		t.Fatalf("read stdout: %v", readErr)
	}
	return string(data), fnErr
}

type wsCall struct {
	Method string
	Params map[string]any
}

func mockCaptureRPCServer(t *testing.T, seen chan<- wsCall) *httptest.Server {
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
				ID     string         `json:"id"`
				Method string         `json:"method"`
				Params map[string]any `json:"params"`
			}
			if err := conn.ReadJSON(&req); err != nil {
				return
			}
			seen <- wsCall{Method: req.Method, Params: req.Params}
			payload, _ := json.Marshal([]map[string]any{{"content": "ok"}})
			if req.Method == "chat.send" || req.Method == "connect" {
				payload, _ = json.Marshal(map[string]any{"content": "ok"})
			}
			_ = conn.WriteJSON(map[string]any{
				"type":    "res",
				"id":      req.ID,
				"ok":      true,
				"payload": json.RawMessage(payload),
			})
		}
	}))
}
