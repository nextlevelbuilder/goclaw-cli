package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func runAgentsArgs(t *testing.T, args ...string) error {
	t.Helper()
	rootCmd.SetArgs(append([]string{"agents"}, args...))
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	return err
}

func TestAgentsImport_PreviewByDefault(t *testing.T) {
	var calledPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"payload":{}}`))
	}))
	defer srv.Close()

	f, err := os.CreateTemp(t.TempDir(), "agent-*.tar.gz")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_, _ = f.WriteString("fake-archive")
	f.Close()

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err = runAgentsArgs(t, "import", f.Name())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	// Must hit the preview endpoint, not the real import.
	if !strings.Contains(calledPath, "preview") {
		t.Errorf("expected preview endpoint, got path: %s", calledPath)
	}
}

func TestAgentsImport_ApplyFlag_HitsImportEndpoint(t *testing.T) {
	var calledPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"payload":{}}`))
	}))
	defer srv.Close()

	f, err := os.CreateTemp(t.TempDir(), "agent-*.tar.gz")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_, _ = f.WriteString("fake-archive")
	f.Close()

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err = runAgentsArgs(t, "import", f.Name(), "--apply")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if strings.Contains(calledPath, "preview") {
		t.Errorf("expected non-preview endpoint with --apply, got path: %s", calledPath)
	}
	if !strings.Contains(calledPath, "/v1/agents/import") {
		t.Errorf("expected /v1/agents/import path, got: %s", calledPath)
	}
}

func TestAgentsExport_WritesToFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/export") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write([]byte("fake-gzip-content"))
	}))
	defer srv.Close()

	outFile := t.TempDir() + "/out.tar.gz"
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runAgentsArgs(t, "export", "agent-123", "--file", outFile)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(data) != "fake-gzip-content" {
		t.Errorf("unexpected file content: %q", string(data))
	}
}

func TestAgentsImportMerge_PreviewByDefault(t *testing.T) {
	var calledQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"payload":{}}`))
	}))
	defer srv.Close()

	f, err := os.CreateTemp(t.TempDir(), "agent-*.tar.gz")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_, _ = f.WriteString("fake-archive")
	f.Close()

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err = runAgentsArgs(t, "import-merge", "agent-123", f.Name())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !strings.Contains(calledQuery, "preview=true") {
		t.Errorf("expected preview=true in query, got: %s", calledQuery)
	}
}
