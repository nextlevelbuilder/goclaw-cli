package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func runCmdCaptureStdout(t *testing.T, args ...string) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	cmdErr := runCmd(t, args...)
	_ = w.Close()
	os.Stdout = old
	out, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("read stdout: %v", readErr)
	}
	return string(out), cmdErr
}

func TestPackagesListRendersGroupedServerPayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/packages" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		okJSON(t, w, map[string]any{
			"system": []map[string]any{{"name": "curl", "version": "8.9.1-r1"}},
			"pip":    []map[string]any{{"name": "pandas", "version": "2.0.0"}},
			"npm":    []map[string]any{{"name": "typescript", "version": "5.1.0"}},
			"github": []map[string]any{{"name": "gh", "repo": "cli/cli", "tag": "v2.72.0", "binaries": []string{"gh"}}},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := runCmdCaptureStdout(t, "packages", "list")
	if err != nil {
		t.Fatalf("packages list: %v", err)
	}
	for _, want := range []string{"system", "pip", "npm", "github", "curl", "pandas", "typescript", "gh"} {
		if !strings.Contains(out, want) {
			t.Fatalf("packages list output missing %q:\n%s", want, out)
		}
	}
}

func TestPackagesInstallAndUninstallSendPackageKey(t *testing.T) {
	var installPackage, uninstallPackage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch r.URL.Path {
		case "/v1/packages/install":
			installPackage, _ = body["package"].(string)
		case "/v1/packages/uninstall":
			uninstallPackage, _ = body["package"].(string)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		okJSON(t, w, map[string]any{"ok": true})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	if err := runCmd(t, "packages", "install", "pandas", "--runtime", "python"); err != nil {
		t.Fatalf("packages install: %v", err)
	}
	if err := runCmd(t, "packages", "uninstall", "typescript", "--runtime", "node", "--yes"); err != nil {
		t.Fatalf("packages uninstall: %v", err)
	}
	if installPackage != "pip:pandas" {
		t.Fatalf("install package = %q, want pip:pandas", installPackage)
	}
	if uninstallPackage != "npm:typescript" {
		t.Fatalf("uninstall package = %q, want npm:typescript", uninstallPackage)
	}
}

func TestPackagesRuntimesRendersStatusObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/packages/runtimes" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		okJSON(t, w, map[string]any{
			"ready":    false,
			"runtimes": []map[string]any{{"name": "python3", "available": true, "version": "Python 3.12.0"}},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := runCmdCaptureStdout(t, "packages", "runtimes")
	if err != nil {
		t.Fatalf("packages runtimes: %v", err)
	}
	for _, want := range []string{"python3", "true", "Python 3.12.0"} {
		if !strings.Contains(out, want) {
			t.Fatalf("runtimes output missing %q:\n%s", want, out)
		}
	}
}

func TestPackagesGitHubReleasesRequiresRepoAndSendsQuery(t *testing.T) {
	var gotRepo, gotLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/packages/github-releases" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		gotRepo = r.URL.Query().Get("repo")
		gotLimit = r.URL.Query().Get("limit")
		okJSON(t, w, map[string]any{"releases": []map[string]any{{"tag": "v2.72.0", "name": "GitHub CLI"}}})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	if _, err := runCmdCaptureStdout(t, "packages", "github-releases"); err == nil {
		t.Fatalf("github-releases without --repo should fail before HTTP")
	}
	out, err := runCmdCaptureStdout(t, "packages", "github-releases", "--repo", "cli/cli", "--limit", "10")
	if err != nil {
		t.Fatalf("github-releases: %v", err)
	}
	if gotRepo != "cli/cli" || gotLimit != "10" {
		t.Fatalf("query repo=%q limit=%q", gotRepo, gotLimit)
	}
	if !strings.Contains(out, "v2.72.0") {
		t.Fatalf("github releases output missing tag:\n%s", out)
	}
}
