package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
)

// resetTracesFollowFlags returns flags to default state between subtests.
func resetTracesFollowFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"session-key", "agent", "since", "status", "channel", "include-spans"} {
		resetTestFlag(tracesFollowCmd, name, "")
	}
	resetTestFlag(tracesFollowCmd, "limit", "0")
}

func TestTracesFollow_SessionKeyBuildsQuery(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })

	var calls int64
	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		if r.URL.Path != "/v1/traces/follow" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		query = r.URL.Query()
		okJSON(t, w, map[string]any{"traces": []map[string]any{}, "spans_by_trace_id": map[string]any{}, "next_since": "2026-05-27T12:00:00Z", "limit": 50})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "traces", "follow", "--session-key=sess-1", "--since=2026-05-27T00:00:00Z", "--limit=25", "--status=success", "--channel=telegram"); err != nil {
		t.Fatalf("traces follow: %v", err)
	}
	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected exactly 1 request, got %d (no watch loop allowed)", atomic.LoadInt64(&calls))
	}
	if query.Get("session_key") != "sess-1" {
		t.Errorf("session_key = %q", query.Get("session_key"))
	}
	if query.Get("since") != "2026-05-27T00:00:00Z" {
		t.Errorf("since = %q", query.Get("since"))
	}
	if query.Get("limit") != "25" {
		t.Errorf("limit = %q", query.Get("limit"))
	}
	if query.Get("status") != "success" {
		t.Errorf("status = %q", query.Get("status"))
	}
	if query.Get("channel") != "telegram" {
		t.Errorf("channel = %q", query.Get("channel"))
	}
}

func TestTracesFollow_AgentTargetBuildsQuery(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })

	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		okJSON(t, w, map[string]any{"traces": []map[string]any{}})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "traces", "follow", "--agent=agent-1", "--include-spans"); err != nil {
		t.Fatalf("traces follow: %v", err)
	}
	if query.Get("agent_id") != "agent-1" {
		t.Errorf("agent_id = %q", query.Get("agent_id"))
	}
	if query.Get("include_spans") != "true" {
		t.Errorf("include_spans = %q", query.Get("include_spans"))
	}
	if query.Has("session_key") {
		t.Errorf("session_key should not be set: %#v", query)
	}
}

func TestTracesFollow_RejectMissingTarget(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "traces", "follow")
	if err == nil {
		t.Fatal("expected validation error for missing target")
	}
	if !strings.Contains(err.Error(), "session-key") && !strings.Contains(err.Error(), "agent") {
		t.Errorf("error should mention target flags: %v", err)
	}
}

func TestTracesFollow_RejectBothTargets(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "traces", "follow", "--session-key=sess-1", "--agent=agent-1")
	if err == nil {
		t.Fatal("expected validation error when both target flags set")
	}
}

func TestTracesFollow_RejectInvalidSince(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "traces", "follow", "--session-key=sess-1", "--since=not-a-timestamp")
	if err == nil {
		t.Fatal("expected RFC3339 validation error")
	}
}

func TestTracesFollow_JSONPreservesEnvelope(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"traces":            []map[string]any{{"trace_id": "t1"}},
			"spans_by_trace_id": map[string]any{"t1": []any{}},
			"next_since":        "2026-05-27T13:00:00Z",
			"server_time":       "2026-05-27T12:30:00Z",
			"limit":             50,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "follow", "--session-key=sess-1")
	})
	if err != nil {
		t.Fatalf("traces follow: %v", err)
	}
	if !strings.Contains(out, "next_since") || !strings.Contains(out, "spans_by_trace_id") {
		t.Fatalf("stdout missing fields: %s", out)
	}
}

func TestTracesFollow_TableHeaders(t *testing.T) {
	t.Cleanup(func() { resetTracesFollowFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"traces": []map[string]any{
				{"trace_id": "t1", "agent_id": "agent-1", "status": "success", "duration_ms": 120, "input_tokens": 50, "output_tokens": 30, "cost": "0.001"},
			},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "follow", "--session-key=sess-1")
	})
	if err != nil {
		t.Fatalf("traces follow: %v", err)
	}
	headerRE := regexp.MustCompile(`TRACE_ID.*AGENT.*STATUS.*DURATION_MS.*INPUT_TOKENS.*OUTPUT_TOKENS.*COST`)
	if !headerRE.MatchString(out) {
		t.Fatalf("table headers missing in:\n%s", out)
	}
}

func TestTracesFollow_DoesNotImportFollowStream(t *testing.T) {
	// Static assertion: tracesFollowCmd uses one HTTP GET via httpClient, not FollowStream.
	// Covered indirectly by the atomic-counter test above; this test is a smoke for command existence.
	if tracesFollowCmd == nil {
		t.Fatal("tracesFollowCmd not declared")
	}
	if tracesFollowCmd.Use == "" || !strings.HasPrefix(tracesFollowCmd.Use, "follow") {
		t.Fatalf("tracesFollowCmd.Use = %q, expected to start with 'follow'", tracesFollowCmd.Use)
	}
}
