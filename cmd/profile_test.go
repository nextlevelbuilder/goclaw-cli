package cmd

import (
	"errors"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

func setupProfileCommandTest(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	cfg = &config.Config{Profile: config.DefaultProfileName, OutputFormat: "json", Yes: true}
	printer = output.NewPrinter("json")
}

func TestProfileCreateUseAndDelete(t *testing.T) {
	setupProfileCommandTest(t)
	cfg.Server = "https://staging.example.com"

	if err := profileCreateCmd.RunE(profileCreateCmd, []string{"staging"}); err != nil {
		t.Fatalf("profile create: %v", err)
	}
	profiles, _, err := config.ListProfiles()
	if err != nil {
		t.Fatalf("list profiles: %v", err)
	}
	if len(profiles) != 1 || profiles[0].Name != "staging" || profiles[0].Server != cfg.Server {
		t.Fatalf("unexpected profiles after create: %#v", profiles)
	}

	if err := profileUseCmd.RunE(profileUseCmd, []string{"staging"}); err != nil {
		t.Fatalf("profile use: %v", err)
	}
	_, active, err := config.ListProfiles()
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if active != "staging" {
		t.Fatalf("expected active staging, got %q", active)
	}

	cfg.Profile = "default"
	if err := profileDeleteCmd.RunE(profileDeleteCmd, []string{"staging"}); err != nil {
		t.Fatalf("profile delete: %v", err)
	}
	profiles, _, err = config.ListProfiles()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected no profiles after delete, got %#v", profiles)
	}
}

func TestProfileRejectsUnsafeName(t *testing.T) {
	setupProfileCommandTest(t)
	err := profileCreateCmd.RunE(profileCreateCmd, []string{"../prod"})
	if err == nil {
		t.Fatal("expected unsafe profile name error")
	}
}

func TestProfileCreateCopyFromMissingErrors(t *testing.T) {
	setupProfileCommandTest(t)
	_ = profileCreateCmd.Flags().Set("copy-from", "missing")
	t.Cleanup(func() { _ = profileCreateCmd.Flags().Set("copy-from", "") })

	err := profileCreateCmd.RunE(profileCreateCmd, []string{"staging"})
	var notFound *config.ProfileNotFoundError
	if !errors.As(err, &notFound) || notFound.Name != "missing" {
		t.Fatalf("expected missing copy-from profile error, got %T %v", err, err)
	}
}
