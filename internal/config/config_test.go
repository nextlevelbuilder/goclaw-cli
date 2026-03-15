package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileConfig_FindProfile(t *testing.T) {
	fc := &FileConfig{
		ActiveProfile: "prod",
		Profiles: []Profile{
			{Name: "prod", Server: "https://prod.example.com"},
			{Name: "staging", Server: "https://staging.example.com"},
		},
	}

	p := fc.FindProfile("prod")
	if p == nil {
		t.Fatal("expected to find prod profile")
	}
	if p.Server != "https://prod.example.com" {
		t.Errorf("expected prod server, got %s", p.Server)
	}

	p = fc.FindProfile("staging")
	if p == nil {
		t.Fatal("expected to find staging profile")
	}

	p = fc.FindProfile("nonexistent")
	if p != nil {
		t.Error("expected nil for nonexistent profile")
	}
}

func TestSaveAndLoadProfiles(t *testing.T) {
	// Use a temp dir to avoid touching real config
	tmpDir := t.TempDir()
	origDir := Dir
	// Override Dir function via file operations directly
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a config
	fc := &FileConfig{
		ActiveProfile: "test",
		Profiles: []Profile{
			{Name: "test", Server: "https://test.example.com"},
		},
	}
	data, err := marshalConfig(fc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read it back
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	fc2, err := parseConfig(readData)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if fc2.ActiveProfile != "test" {
		t.Errorf("expected active_profile=test, got %s", fc2.ActiveProfile)
	}
	if len(fc2.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(fc2.Profiles))
	}
	if fc2.Profiles[0].Server != "https://test.example.com" {
		t.Errorf("expected test server, got %s", fc2.Profiles[0].Server)
	}

	_ = origDir // suppress unused
}

func TestTokenNotInConfigYAML(t *testing.T) {
	// Profile has Token with yaml:"-", so it should not appear in marshaled YAML
	fc := &FileConfig{
		ActiveProfile: "test",
		Profiles: []Profile{
			{Name: "test", Server: "https://test.example.com", Token: "secret-token"},
		},
	}
	data, err := marshalConfig(fc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	yaml := string(data)
	if containsString(yaml, "secret-token") {
		t.Error("token should NOT appear in marshaled YAML (yaml:\"-\" tag)")
	}
}

func containsString(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
