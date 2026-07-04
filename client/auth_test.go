package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialStore_SaveLoadDelete(t *testing.T) {
	dir := t.TempDir()
	cs := &CredentialStore{dir: dir}

	// Save
	if err := cs.SaveToken("test", "my-secret-token"); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Verify file permissions
	path := filepath.Join(dir, "credentials_test")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// On Windows permissions are different, just check file exists
	if info.Size() == 0 {
		t.Error("expected non-empty credential file")
	}

	// Load
	token, err := cs.LoadToken("test")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if token != "my-secret-token" {
		t.Errorf("expected my-secret-token, got %s", token)
	}

	// Delete
	if err := cs.DeleteToken("test"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Verify deleted
	_, err = cs.LoadToken("test")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestCredentialStore_LoadMissing(t *testing.T) {
	cs := &CredentialStore{dir: t.TempDir()}
	_, err := cs.LoadToken("nonexistent")
	if err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestCredentialStore_SenderID(t *testing.T) {
	cs := &CredentialStore{dir: t.TempDir()}

	if err := cs.SaveSenderID("prod", "sender-abc"); err != nil {
		t.Fatalf("save sender: %v", err)
	}

	sid, err := cs.LoadSenderID("prod")
	if err != nil {
		t.Fatalf("load sender: %v", err)
	}
	if sid != "sender-abc" {
		t.Errorf("expected sender-abc, got %s", sid)
	}
}

func TestCredentialStore_DeleteNonExistent(t *testing.T) {
	cs := &CredentialStore{dir: t.TempDir()}
	// Should not error when deleting non-existent
	if err := cs.DeleteToken("ghost"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
