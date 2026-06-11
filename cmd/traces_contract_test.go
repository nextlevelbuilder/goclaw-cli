package cmd

import (
	"encoding/json"
	"net/http"
	"testing"
)

func rawJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("write raw json: %v", err)
	}
}

func resetTracesListFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"agent", "user", "session-key", "status", "channel", "since",
		"q", "agent-query", "channel-query", "from", "to", "tool-name", "has-tool-calls",
	} {
		resetTestFlag(tracesListCmd, name, "")
	}
	resetTestFlag(tracesListCmd, "root-only", "false")
	resetTestFlag(tracesListCmd, "limit", "20")
	resetTestFlag(tracesListCmd, "offset", "0")
	for _, name := range []string{
		"min-input-tokens", "max-input-tokens",
		"min-output-tokens", "max-output-tokens",
		"min-tool-calls", "max-tool-calls",
	} {
		resetTestFlag(tracesListCmd, name, "0")
	}
}

func assertNoTracesReplayCommand(t *testing.T) {
	t.Helper()
	for _, c := range tracesCmd.Commands() {
		if c.Name() == "replay" {
			t.Fatal("traces replay command must not exist until the server exposes POST /v1/traces/{id}/replay")
		}
	}
}
