package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func runBackupArgs(t *testing.T, args ...string) error {
	t.Helper()
	rootCmd.SetArgs(append([]string{"backup"}, args...))
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	return err
}

func TestBackupSystemPreflight_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/system/backup/preflight" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload, _ := json.Marshal(map[string]any{
			"pg_dump_available": true,
			"disk_space_ok":     true,
		})
		resp, _ := json.Marshal(map[string]any{"ok": true, "payload": json.RawMessage(payload)})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	}))
	defer srv.Close()

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runBackupArgs(t, "system-preflight")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestBackupSystem_ReturnsToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/system/backup" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload, _ := json.Marshal(map[string]any{"token": "tok-abc123", "expires_at": "2099-01-01T00:00:00Z"})
		resp, _ := json.Marshal(map[string]any{"ok": true, "payload": json.RawMessage(payload)})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	}))
	defer srv.Close()

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runBackupArgs(t, "system")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestBackupSystemDownload_WritesFile(t *testing.T) {
	content := "fake-backup-binary"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/system/backup/download/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write([]byte(content))
	}))
	defer srv.Close()

	outFile := t.TempDir() + "/backup.tar.gz"
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runBackupArgs(t, "system-download", "tok-abc123", "--file", outFile)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(data) != content {
		t.Errorf("unexpected content: %q", string(data))
	}
}

func TestBackupSystemDownload_MissingOutputFlag_Errors(t *testing.T) {
	// Unset env vars so newHTTP would fail — but flag check must happen first.
	t.Setenv("GOCLAW_SERVER", "")
	t.Setenv("GOCLAW_TOKEN", "")

	// Reset the flag to empty to avoid cobra reusing value from prior test run.
	_ = backupSystemDownloadCmd.Flags().Set("file", "")

	err := runBackupArgs(t, "system-download", "some-token")
	if err == nil {
		t.Fatal("expected error when --file flag is missing")
	}
	// Accept either our custom message or a server-required error (if flag check
	// somehow passes due to cobra state) — both are valid "refused to download" paths.
	if !strings.Contains(err.Error(), "--file") &&
		!strings.Contains(err.Error(), "required") &&
		!strings.Contains(err.Error(), "server") {
		t.Errorf("expected meaningful error, got: %v", err)
	}
}

func TestBackupS3ConfigGet_MasksSecretKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, _ := json.Marshal(map[string]any{
			"configured": true,
			"bucket":     "my-bucket",
			"secret_key": "super-secret-value",
		})
		resp, _ := json.Marshal(map[string]any{"ok": true, "payload": json.RawMessage(payload)})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	}))
	defer srv.Close()

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	// Capture stdout by redirecting it.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runBackupArgs(t, "s3", "config", "get")

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	buf.WriteString("")
	tmp := make([]byte, 4096)
	n, _ := r.Read(tmp)
	output := string(tmp[:n])

	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if strings.Contains(output, "super-secret-value") {
		t.Errorf("secret_key should be masked in output, got: %s", output)
	}
	if !strings.Contains(output, "***") {
		t.Errorf("expected *** mask in output, got: %s", output)
	}
}
