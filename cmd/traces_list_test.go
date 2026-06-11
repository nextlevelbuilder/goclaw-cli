package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestTracesList_ServerEnvelope_TableRows(t *testing.T) {
	t.Cleanup(func() { resetTracesListFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/traces" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rawJSON(t, w, traceListFixture())
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "list", "--output", "table")
	})
	if err != nil {
		t.Fatalf("traces list: %v", err)
	}
	headerRE := regexp.MustCompile(`ID.*AGENT.*STATUS.*DURATION_MS.*TOTAL_INPUT_TOKENS.*TOTAL_OUTPUT_TOKENS.*TOTAL_COST`)
	if !headerRE.MatchString(out) {
		t.Fatalf("table headers missing in:\n%s", out)
	}
	for _, want := range []string{"trace-1", "agent-1", "completed", "0.01"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestTracesList_JSONPreservesEnvelope(t *testing.T) {
	t.Cleanup(func() { resetTracesListFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawJSON(t, w, traceListFixture())
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "list", "--output", "json")
	})
	if err != nil {
		t.Fatalf("traces list: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, out)
	}
	if got["total"] != float64(1) || got["limit"] != float64(20) || got["offset"] != float64(0) {
		t.Fatalf("envelope pagination fields not preserved: %#v", got)
	}
	if _, ok := got["traces"].([]any); !ok {
		t.Fatalf("traces array not preserved: %#v", got["traces"])
	}
}

func TestTracesList_ServerFilters(t *testing.T) {
	t.Cleanup(func() { resetTracesListFlags(t) })
	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		rawJSON(t, w, traceListFixture())
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	err := runCmd(t, "traces", "list",
		"--agent=agent-1",
		"--user=user-1",
		"--session-key=session-1",
		"--status=failed",
		"--channel=telegram",
		"--q=trace_%",
		"--agent-query=helper",
		"--channel-query=ops",
		"--from=2026-06-10T01:02:03Z",
		"--to=2026-06-11T04:05:06Z",
		"--min-input-tokens=10",
		"--max-input-tokens=20",
		"--min-output-tokens=30",
		"--max-output-tokens=40",
		"--min-tool-calls=1",
		"--max-tool-calls=3",
		"--tool-name=web_%",
		"--has-tool-calls=true",
		"--limit=5",
		"--offset=10",
	)
	if err != nil {
		t.Fatalf("traces list: %v", err)
	}
	want := map[string]string{
		"agent_id":          "agent-1",
		"user_id":           "user-1",
		"session_key":       "session-1",
		"status":            "failed",
		"channel":           "telegram",
		"q":                 "trace_%",
		"agent":             "helper",
		"channel_query":     "ops",
		"from":              "2026-06-10T01:02:03Z",
		"to":                "2026-06-11T04:05:06Z",
		"min_input_tokens":  "10",
		"max_input_tokens":  "20",
		"min_output_tokens": "30",
		"max_output_tokens": "40",
		"min_tool_calls":    "1",
		"max_tool_calls":    "3",
		"tool_name":         "web_%",
		"has_tool_calls":    "true",
		"limit":             "5",
		"offset":            "10",
	}
	for key, val := range want {
		if got := query.Get(key); got != val {
			t.Fatalf("%s = %q, want %q; query=%s", key, got, val, query.Encode())
		}
	}
	if query.Has("since") || query.Has("root_only") {
		t.Fatalf("unsupported list filters must not be sent: %s", query.Encode())
	}
}

func TestTracesList_ServerFiltersForwardExplicitFalseAndZero(t *testing.T) {
	t.Cleanup(func() { resetTracesListFlags(t) })
	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		rawJSON(t, w, traceListFixture())
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	err := runCmd(t, "traces", "list",
		"--min-input-tokens=0",
		"--max-tool-calls=0",
		"--has-tool-calls=false",
	)
	if err != nil {
		t.Fatalf("traces list: %v", err)
	}
	want := map[string]string{
		"min_input_tokens": "0",
		"max_tool_calls":   "0",
		"has_tool_calls":   "false",
	}
	for key, val := range want {
		if got := query.Get(key); got != val {
			t.Fatalf("%s = %q, want %q; query=%s", key, got, val, query.Encode())
		}
	}
}

func TestTracesReplayCommandAbsent(t *testing.T) {
	assertNoTracesReplayCommand(t)
}

func traceListFixture() map[string]any {
	return map[string]any{
		"traces": []map[string]any{{
			"id":                  "trace-1",
			"agent_id":            "agent-1",
			"session_key":         "session-1",
			"status":              "completed",
			"duration_ms":         2500,
			"total_input_tokens":  10,
			"total_output_tokens": 5,
			"total_cost":          0.01,
		}},
		"total":  1,
		"limit":  20,
		"offset": 0,
	}
}
