package client

import (
	"fmt"
	"os"
	"path/filepath"
)

// CredentialStore manages token persistence.
// Uses a simple file at ~/.goclaw/credentials with 0600 permissions.
type CredentialStore struct {
	dir string
}

// NewCredentialStore creates a store using ~/.goclaw/ directory.
func NewCredentialStore() *CredentialStore {
	home, _ := os.UserHomeDir()
	return &CredentialStore{dir: filepath.Join(home, ".goclaw")}
}

// SaveToken stores a token for a profile.
func (cs *CredentialStore) SaveToken(profile, token string) error {
	if err := os.MkdirAll(cs.dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(cs.dir, "credentials_"+profile)
	return os.WriteFile(path, []byte(token), 0600)
}

// LoadToken retrieves a stored token for a profile.
func (cs *CredentialStore) LoadToken(profile string) (string, error) {
	path := filepath.Join(cs.dir, "credentials_"+profile)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("no stored credentials for profile %q", profile)
	}
	return string(data), nil
}

// DeleteToken removes stored credentials for a profile.
func (cs *CredentialStore) DeleteToken(profile string) error {
	path := filepath.Join(cs.dir, "credentials_"+profile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// SaveSenderID stores a device pairing sender_id.
func (cs *CredentialStore) SaveSenderID(profile, senderID string) error {
	if err := os.MkdirAll(cs.dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(cs.dir, "sender_"+profile)
	return os.WriteFile(path, []byte(senderID), 0600)
}

// LoadSenderID retrieves a stored sender_id.
func (cs *CredentialStore) LoadSenderID(profile string) (string, error) {
	path := filepath.Join(cs.dir, "sender_"+profile)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
