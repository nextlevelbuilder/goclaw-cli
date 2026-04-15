package output

import (
	"os"
	"testing"
)

func TestIsTTY_Stdin(t *testing.T) {
	// In CI/test environments stdin is not a TTY — just verify no panic.
	_ = IsTTY(int(os.Stdin.Fd()))
}

func TestIsTTY_InvalidFD(t *testing.T) {
	// A clearly invalid fd should return false without panic.
	got := IsTTY(99999)
	if got {
		t.Error("expected false for invalid fd")
	}
}

func TestResolveFormat_FlagPriority(t *testing.T) {
	t.Setenv("GOCLAW_OUTPUT", "yaml")
	got := ResolveFormat("json")
	if got != "json" {
		t.Errorf("flag should win over env: got %q, want json", got)
	}
}

func TestResolveFormat_EnvOverTTY(t *testing.T) {
	t.Setenv("GOCLAW_OUTPUT", "yaml")
	// flagVal empty → env should win
	got := ResolveFormat("")
	if got != "yaml" {
		t.Errorf("env should win when flag empty: got %q, want yaml", got)
	}
}

func TestResolveFormat_DefaultNonTTY(t *testing.T) {
	t.Setenv("GOCLAW_OUTPUT", "")
	// In test runner stdout is not a TTY → should return "json"
	got := ResolveFormat("")
	// Accept both "json" and "table" depending on environment
	if got != "json" && got != "table" {
		t.Errorf("ResolveFormat() = %q, want json or table", got)
	}
}

func TestResolveFormat_ExplicitTable(t *testing.T) {
	t.Setenv("GOCLAW_OUTPUT", "")
	got := ResolveFormat("table")
	if got != "table" {
		t.Errorf("explicit table flag: got %q, want table", got)
	}
}

func TestResolveFormat_ExplicitYAML(t *testing.T) {
	got := ResolveFormat("yaml")
	if got != "yaml" {
		t.Errorf("explicit yaml flag: got %q, want yaml", got)
	}
}

func TestResolveFormat_InvalidEnvFallsThrough(t *testing.T) {
	// Invalid GOCLAW_OUTPUT must not be used — fall through to TTY detect.
	t.Setenv("GOCLAW_OUTPUT", "xml")
	got := ResolveFormat("")
	if got == "xml" {
		t.Error("invalid env value must not be returned; should fall through to TTY detect")
	}
	if got != "json" && got != "table" {
		t.Errorf("fallback should be json or table, got %q", got)
	}
}

func TestResolveFormat_InvalidFlagPassedThrough(t *testing.T) {
	// Invalid flag values pass through — cobra validation catches them.
	got := ResolveFormat("xml")
	if got != "xml" {
		t.Errorf("invalid flag should pass through for caller validation: got %q", got)
	}
}
