package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

func setupP5HTTPTest(serverURL string) {
	cfg = &config.Config{Server: serverURL, Token: "test-token", OutputFormat: "json"}
	printer = output.NewPrinter("json")
}

func resetFlag(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("missing flag %s", name)
	}
	_ = flag.Value.Set(flag.DefValue)
	flag.Changed = false
}

func TestTeamsAttachmentsDownloadWritesOutputAndSendsAuth(t *testing.T) {
	body := []byte("attachment-data")
	outFile := filepath.Join(t.TempDir(), "nested", "artifact.bin")
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, teamsAttachmentsDownloadCmd, "output")
	resetFlag(t, teamsAttachmentsDownloadCmd, "force")
	_ = teamsAttachmentsDownloadCmd.Flags().Set("output", outFile)

	if err := teamsAttachmentsDownloadCmd.RunE(teamsAttachmentsDownloadCmd, []string{"team alpha", "att/42"}); err != nil {
		t.Fatalf("download: %v", err)
	}
	if gotPath != "/v1/teams/team%20alpha/attachments/att%2F42/download" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("authorization = %q", gotAuth)
	}
	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("output = %q", got)
	}
}

func TestTeamsAttachmentsDownloadRequiresOutputBeforeNetwork(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, teamsAttachmentsDownloadCmd, "output")
	resetFlag(t, teamsAttachmentsDownloadCmd, "force")

	err := teamsAttachmentsDownloadCmd.RunE(teamsAttachmentsDownloadCmd, []string{"team-1", "att-1"})
	if err == nil || !strings.Contains(err.Error(), "--output is required") {
		t.Fatalf("expected output validation error, got %v", err)
	}
	if called {
		t.Fatal("server was called before output validation")
	}
}

func TestTeamsAttachmentsDownloadRefusesExistingFileUnlessForce(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "artifact.bin")
	if err := os.WriteFile(outFile, []byte("old"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = w.Write([]byte("new"))
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, teamsAttachmentsDownloadCmd, "output")
	resetFlag(t, teamsAttachmentsDownloadCmd, "force")
	_ = teamsAttachmentsDownloadCmd.Flags().Set("output", outFile)

	err := teamsAttachmentsDownloadCmd.RunE(teamsAttachmentsDownloadCmd, []string{"team-1", "att-1"})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected overwrite guard, got %v", err)
	}
	if called {
		t.Fatal("server was called before overwrite guard")
	}

	_ = teamsAttachmentsDownloadCmd.Flags().Set("force", "true")
	if err := teamsAttachmentsDownloadCmd.RunE(teamsAttachmentsDownloadCmd, []string{"team-1", "att-1"}); err != nil {
		t.Fatalf("force download: %v", err)
	}
	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("output = %q", got)
	}
}

func TestTeamsAttachmentsDownloadLocalOutputDoesNotOverrideFormat(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "artifact.bin")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("attachment"))
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")
	resetFlag(t, teamsAttachmentsDownloadCmd, "output")
	resetFlag(t, teamsAttachmentsDownloadCmd, "force")

	if err := runCmd(t, "teams", "attachments", "download", "team-1", "att-1", "--output", outFile); err != nil {
		t.Fatalf("download: %v", err)
	}
	if cfg.OutputFormat != "json" {
		t.Fatalf("cfg.OutputFormat = %q, want json", cfg.OutputFormat)
	}
}

func TestAgentsEvolutionUpdateMapsActionToStatusAndEscapesRoute(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		okJSON(t, w, map[string]any{"status": "approved"})
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, agentsEvolutionUpdateCmd, "action")
	_ = agentsEvolutionUpdateCmd.Flags().Set("action", "accept")

	if err := agentsEvolutionUpdateCmd.RunE(agentsEvolutionUpdateCmd, []string{"agent/1", "sugg 1"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if gotPath != "/v1/agents/agent%2F1/evolution/suggestions/sugg%201" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["status"] != "approved" || gotBody["action"] != nil {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestAgentsEvolutionUpdateRejectMapsToRejected(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		okJSON(t, w, map[string]any{"status": "rejected"})
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, agentsEvolutionUpdateCmd, "action")
	_ = agentsEvolutionUpdateCmd.Flags().Set("action", "reject")

	if err := agentsEvolutionUpdateCmd.RunE(agentsEvolutionUpdateCmd, []string{"agent-1", "sugg-1"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if gotBody["status"] != "rejected" {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestAgentsEvolutionSkillApplySendsApprovedWithDraftFile(t *testing.T) {
	draftPath := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(draftPath, []byte("skill draft\n"), 0644); err != nil {
		t.Fatalf("write draft: %v", err)
	}
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.EscapedPath() != "/v1/agents/agent%2F1/evolution/suggestions" {
				t.Errorf("unexpected preflight path: %s", r.URL.EscapedPath())
			}
			okJSON(t, w, []map[string]any{{"id": "sugg 1", "suggestion_type": "skill_add"}})
			return
		case http.MethodPatch:
			gotPath = r.URL.EscapedPath()
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Errorf("decode body: %v", err)
			}
			okJSON(t, w, map[string]any{"applied": true})
			return
		default:
			t.Errorf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, agentsEvolutionSkillApplyCmd, "skill-draft")
	_ = agentsEvolutionSkillApplyCmd.Flags().Set("skill-draft", "@"+draftPath)

	if err := agentsEvolutionSkillApplyCmd.RunE(agentsEvolutionSkillApplyCmd, []string{"agent/1", "sugg 1"}); err != nil {
		t.Fatalf("skill apply: %v", err)
	}
	if gotPath != "/v1/agents/agent%2F1/evolution/suggestions/sugg%201" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody["status"] != "approved" || gotBody["skill_draft"] != "skill draft\n" {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestAgentsEvolutionSkillApplyWithoutDraftSendsApprovedOnly(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			okJSON(t, w, []map[string]any{{"id": "sugg-1", "suggestion_type": "skill_add"}})
			return
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		} else if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		okJSON(t, w, map[string]any{"applied": true})
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, agentsEvolutionSkillApplyCmd, "skill-draft")

	if err := agentsEvolutionSkillApplyCmd.RunE(agentsEvolutionSkillApplyCmd, []string{"agent-1", "sugg-1"}); err != nil {
		t.Fatalf("skill apply: %v", err)
	}
	if gotBody["status"] != "approved" || gotBody["skill_draft"] != nil {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestAgentsEvolutionSkillApplyRejectsNonSkillSuggestionBeforePatch(t *testing.T) {
	patchCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patchCalled = true
			t.Error("PATCH should not be called for non-skill suggestion")
			return
		}
		okJSON(t, w, []map[string]any{{"id": "sugg-1", "suggestion_type": "tool_order"}})
	}))
	defer srv.Close()
	setupP5HTTPTest(srv.URL)
	resetFlag(t, agentsEvolutionSkillApplyCmd, "skill-draft")

	err := agentsEvolutionSkillApplyCmd.RunE(agentsEvolutionSkillApplyCmd, []string{"agent-1", "sugg-1"})
	if err == nil || !strings.Contains(err.Error(), "not skill_add") {
		t.Fatalf("expected skill_add validation error, got %v", err)
	}
	if patchCalled {
		t.Fatal("PATCH was called")
	}
}
