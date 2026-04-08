package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestAllCommandsRegistered verifies that all expected root-level commands are
// registered on rootCmd. It checks by name prefix since some Use strings include
// argument placeholders (e.g. "get <id>").
func TestAllCommandsRegistered(t *testing.T) {
	// Expected root-level command names (alphabetical).
	// Note: "completion" and "help" are injected by Cobra only after Execute()
	// is called, so they are not checked here.
	expected := []string{
		"agents",
		"api-docs",
		"api-keys",
		"approvals",
		"auth",
		"channels",
		"chat",
		"config",
		"contacts",
		"credentials",
		"cron",
		"delegations",
		"devices",
		"export",
		"heartbeat",
		"import",
		"knowledge-graph",
		"logs",
		"mcp",
		"media",
		"memory",
		"packages",
		"pending-messages",
		"providers",
		"sessions",
		"skills",
		"status",
		"storage",
		"system-config",
		"teams",
		"tenants",
		"tools",
		"traces",
		"tts",
		"usage",
		"version",
	}

	// Build a set of registered command names from Use field (first word only).
	registered := make(map[string]bool)
	for _, c := range rootCmd.Commands() {
		name := strings.SplitN(c.Use, " ", 2)[0]
		registered[name] = true
	}

	for _, name := range expected {
		if !registered[name] {
			t.Errorf("expected command %q to be registered on rootCmd, but it was not found", name)
		}
	}
}

// TestRootHelp verifies that running --help on rootCmd does not panic and
// returns without error.
func TestRootHelp(t *testing.T) {
	// Capture output to avoid polluting test output.
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with --help returned error: %v", err)
	}
}

// TestCommandUseFields verifies that key commands have the correct Use field
// (name matches expected value).
func TestCommandUseFields(t *testing.T) {
	cases := []struct {
		wantName string
	}{
		{"agents"},
		{"api-docs"},
		{"api-keys"},
		{"approvals"},
		{"auth"},
		{"channels"},
		{"chat"},
		{"config"},
		{"contacts"},
		{"credentials"},
		{"cron"},
		{"delegations"},
		{"devices"},
		{"export"},
		{"heartbeat"},
		{"import"},
		{"knowledge-graph"},
		{"logs"},
		{"mcp"},
		{"media"},
		{"memory"},
		{"packages"},
		{"pending-messages"},
		{"providers"},
		{"sessions"},
		{"skills"},
		{"status"},
		{"storage"},
		{"system-config"},
		{"teams"},
		{"tenants"},
		{"tools"},
		{"traces"},
		{"tts"},
		{"usage"},
		{"version"},
	}

	cmdMap := make(map[string]*cobra.Command)
	for _, c := range rootCmd.Commands() {
		name := strings.SplitN(c.Use, " ", 2)[0]
		cmdMap[name] = c
	}

	for _, tc := range cases {
		c, ok := cmdMap[tc.wantName]
		if !ok {
			t.Errorf("command %q not found", tc.wantName)
			continue
		}
		// Use field must start with the expected name.
		if !strings.HasPrefix(c.Use, tc.wantName) {
			t.Errorf("command %q: Use=%q does not start with expected name", tc.wantName, c.Use)
		}
	}
}
