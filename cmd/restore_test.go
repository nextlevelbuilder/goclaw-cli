package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func runRestoreArgs(t *testing.T, args ...string) error {
	t.Helper()
	rootCmd.SetArgs(append([]string{"restore"}, args...))
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	return err
}

func TestRestoreSystem_NoYes_Refuses(t *testing.T) {
	err := runRestoreArgs(t, "system", "somefile.tar.gz", "--confirm=somefile.tar.gz")
	if err == nil {
		t.Fatal("expected error when --yes is missing")
	}
	if !strings.Contains(err.Error(), "DESTRUCTIVE") {
		t.Errorf("expected DESTRUCTIVE in error, got: %v", err)
	}
}

func TestRestoreSystem_WrongConfirm_Refuses(t *testing.T) {
	err := runRestoreArgs(t, "system", "backup.tar.gz", "--yes", "--confirm=wrong.tar.gz")
	if err == nil {
		t.Fatal("expected error on confirmation mismatch")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("expected mismatch in error, got: %v", err)
	}
}

func TestRestoreSystem_CorrectConfirm_Proceeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"payload":{}}`))
	}))
	defer srv.Close()

	// Create a minimal temp file to upload.
	f, err := os.CreateTemp(t.TempDir(), "backup-*.tar.gz")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	_, _ = f.WriteString("fake-archive")
	f.Close()
	archivePath := f.Name()
	baseName := archivePath[strings.LastIndex(archivePath, string(os.PathSeparator))+1:]

	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err = runRestoreArgs(t, "system", archivePath, "--yes", "--confirm="+baseName)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestRestoreTenant_NoYes_Refuses(t *testing.T) {
	err := runRestoreArgs(t, "tenant", "backup.tar.gz", "--tenant-id=abc", "--confirm=abc")
	if err == nil {
		t.Fatal("expected error when --yes is missing")
	}
	if !strings.Contains(err.Error(), "DESTRUCTIVE") {
		t.Errorf("expected DESTRUCTIVE in error, got: %v", err)
	}
}

func TestRestoreTenant_WrongConfirm_Refuses(t *testing.T) {
	err := runRestoreArgs(t, "tenant", "backup.tar.gz",
		"--yes", "--tenant-id=real-id", "--confirm=wrong-id")
	if err == nil {
		t.Fatal("expected error on confirmation mismatch")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("expected mismatch in error, got: %v", err)
	}
}

func TestRestoreTenant_MissingTenantID_Refuses(t *testing.T) {
	// --tenant-id is required flag; cobra itself returns the error.
	err := runRestoreArgs(t, "tenant", "backup.tar.gz", "--yes", "--confirm=abc")
	if err == nil {
		t.Fatal("expected error when --tenant-id is missing")
	}
}
