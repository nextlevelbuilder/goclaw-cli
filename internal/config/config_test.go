package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
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

func testLoadCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "agents"}
	flags := cmd.Flags()
	flags.String("server", "", "")
	flags.String("token", "", "")
	flags.String("output", "", "")
	flags.String("profile", "", "")
	flags.String("tenant-id", "", "")
	flags.Bool("insecure", false, "")
	flags.Bool("verbose", false, "")
	flags.Bool("yes", false, "")
	return cmd
}

func setTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestLoadUsesGOCLAWProfile(t *testing.T) {
	setTestHome(t)
	t.Setenv("GOCLAW_PROFILE", "staging")
	if err := Save(Profile{Name: "default", Server: "https://default.example.com"}, true); err != nil {
		t.Fatalf("save default: %v", err)
	}
	if err := Save(Profile{Name: "staging", Server: "https://staging.example.com"}, false); err != nil {
		t.Fatalf("save staging: %v", err)
	}

	cfg, err := Load(testLoadCommand())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Profile != "staging" || cfg.Server != "https://staging.example.com" {
		t.Fatalf("expected staging profile, got profile=%q server=%q", cfg.Profile, cfg.Server)
	}
}

func TestLoadTracksProfileOutputDefault(t *testing.T) {
	setTestHome(t)
	if err := Save(Profile{Name: "default", Server: "https://default.example.com", OutputFormat: "yaml"}, true); err != nil {
		t.Fatalf("save: %v", err)
	}

	cfg, err := Load(testLoadCommand())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.ProfileOutputFormat != "yaml" {
		t.Fatalf("expected profile output yaml, got %q", cfg.ProfileOutputFormat)
	}
}

func TestLoadMissingExplicitProfileErrors(t *testing.T) {
	setTestHome(t)
	if err := Save(Profile{Name: "default", Server: "https://default.example.com"}, true); err != nil {
		t.Fatalf("save: %v", err)
	}
	cmd := testLoadCommand()
	if err := cmd.Flags().Set("profile", "missing"); err != nil {
		t.Fatalf("set profile: %v", err)
	}

	_, err := Load(cmd)
	if err == nil {
		t.Fatal("expected missing profile error")
	}
	if _, ok := err.(*ProfileNotFoundError); !ok {
		t.Fatalf("expected ProfileNotFoundError, got %T %v", err, err)
	}
}

func TestLegacyConfigMigratesToProfileAndCredential(t *testing.T) {
	home := setTestHome(t)
	dir := filepath.Join(home, ".goclaw")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	legacy := []byte("server: https://legacy.example.com\ntoken: secret-token\noutput: json\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), legacy, 0600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	cfg, err := Load(testLoadCommand())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Profile != DefaultProfileName || cfg.Server != "https://legacy.example.com" {
		t.Fatalf("unexpected migrated cfg: profile=%q server=%q", cfg.Profile, cfg.Server)
	}
	if cfg.Token != "secret-token" {
		t.Fatalf("expected token from migrated credential, got %q", cfg.Token)
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("read migrated: %v", err)
	}
	if containsString(string(data), "secret-token") {
		t.Fatal("migrated config must not contain token")
	}
}
