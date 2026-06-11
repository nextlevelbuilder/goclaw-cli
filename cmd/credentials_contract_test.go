package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCredentialsListAndPresetsReadServerEnvelopes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cli-credentials":
			okJSON(t, w, map[string]any{"items": []map[string]any{{"id": "cred-1", "binary_name": "git", "created_at": "2026-06-12T00:00:00Z"}}})
		case "/v1/cli-credentials/presets":
			okJSON(t, w, map[string]any{"presets": map[string]any{"git": map[string]any{"binary_name": "git"}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := runCmdCaptureStdout(t, "credentials", "list")
	if err != nil {
		t.Fatalf("credentials list: %v", err)
	}
	if !strings.Contains(out, "cred-1") || !strings.Contains(out, "git") {
		t.Fatalf("credentials list output missing item:\n%s", out)
	}
	out, err = runCmdCaptureStdout(t, "credentials", "presets")
	if err != nil {
		t.Fatalf("credentials presets: %v", err)
	}
	if !strings.Contains(out, "git") {
		t.Fatalf("credentials presets output missing preset:\n%s", out)
	}
}

func TestCredentialsCreateCanSendServerPayload(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cli-credentials" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		okJSON(t, w, map[string]any{"id": "cred-1", "binary_name": "git"})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	err := runCmd(t, "credentials", "create", "--body", `{"preset":"git"}`)
	if err != nil {
		t.Fatalf("credentials create: %v", err)
	}
	if got["preset"] != "git" {
		t.Fatalf("create body = %#v, want preset git", got)
	}
}

func TestCredentialNestedListsReadServerEnvelopes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cli-credentials/cred-1/agent-grants":
			okJSON(t, w, map[string]any{"grants": []map[string]any{{"id": "grant-1", "agent_id": "agent-1"}}})
		case "/v1/cli-credentials/cred-1/user-credentials":
			okJSON(t, w, map[string]any{"user_credentials": []map[string]any{{"user_id": "user-1", "env_keys": []string{"GITHUB_TOKEN"}}}})
		case "/v1/cli-credentials/cred-1/agent-credentials":
			okJSON(t, w, map[string]any{"agent_credentials": []map[string]any{{"agent_id": "agent-1", "credential_type": "pat"}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"agent grants", []string{"credentials", "agent-grants", "list", "cred-1"}, "grant-1"},
		{"user credentials", []string{"credentials", "user-credentials", "list", "cred-1"}, "user-1"},
		{"agent credentials", []string{"credentials", "agent-credentials", "list", "cred-1"}, "agent-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCmdCaptureStdout(t, tc.args...)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !strings.Contains(out, tc.want) {
				t.Fatalf("%s output missing %q:\n%s", tc.name, tc.want, out)
			}
		})
	}
}

func TestCredentialAgentCredentialsRoutes(t *testing.T) {
	var setBody map[string]any
	var deleted bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/cli-credentials/cred-1/agent-credentials/agent-1":
			okJSON(t, w, map[string]any{"agent_id": "agent-1", "credential_type": "pat"})
		case r.Method == http.MethodPut && r.URL.Path == "/v1/cli-credentials/cred-1/agent-credentials/agent-1":
			_ = json.NewDecoder(r.Body).Decode(&setBody)
			okJSON(t, w, map[string]any{"ok": true})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/cli-credentials/cred-1/agent-credentials/agent-1":
			deleted = true
			okJSON(t, w, map[string]any{"ok": true})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	if err := runCmd(t, "credentials", "agent-credentials", "get", "cred-1", "agent-1"); err != nil {
		t.Fatalf("agent credential get: %v", err)
	}
	if err := runCmd(t, "credentials", "agent-credentials", "set", "cred-1", "agent-1", "--body", `{"credential_type":"pat"}`); err != nil {
		t.Fatalf("agent credential set: %v", err)
	}
	if err := runCmd(t, "credentials", "agent-credentials", "delete", "cred-1", "agent-1", "--yes"); err != nil {
		t.Fatalf("agent credential delete: %v", err)
	}
	if setBody["credential_type"] != "pat" {
		t.Fatalf("set body = %#v", setBody)
	}
	if !deleted {
		t.Fatalf("delete route was not called")
	}
}
