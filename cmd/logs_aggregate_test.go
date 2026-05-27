package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func resetLogsAggregateFlags(t *testing.T) {
	t.Helper()
	resetTestFlag(logsAggregateCmd, "group-by", "level")
	for _, name := range []string{"level", "source", "from"} {
		resetTestFlag(logsAggregateCmd, name, "")
	}
}

func TestLogsAggregate_DefaultGroupBy(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })

	var path, rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		rawQuery = r.URL.RawQuery
		okJSON(t, w, map[string]any{
			"source":      "runtime",
			"retention":   "ring_buffer",
			"capacity":    100,
			"sample_size": 0,
			"group_by":    "level",
			"buckets":     []any{},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "logs", "aggregate"); err != nil {
		t.Fatalf("logs aggregate: %v", err)
	}
	if path != "/v1/logs/runtime/aggregate" {
		t.Fatalf("path = %q", path)
	}
	q, _ := url.ParseQuery(rawQuery)
	// Default group_by=level should appear in query.
	if q.Get("group_by") != "level" {
		t.Errorf("expected group_by=level in raw query, got: %q", rawQuery)
	}
}

func TestLogsAggregate_InvalidGroupBy(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "logs", "aggregate", "--group-by=foo")
	if err == nil {
		t.Fatal("expected error for invalid --group-by=foo")
	}
}

func TestLogsAggregate_QueryStringFilters(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })

	var rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		okJSON(t, w, map[string]any{
			"source": "runtime", "retention": "ring_buffer", "capacity": 100,
			"sample_size": 0, "group_by": "source", "buckets": []any{},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "logs", "aggregate",
		"--group-by=source",
		"--level=warn",
		"--source=router",
		"--from=2026-05-01T00:00:00Z",
	); err != nil {
		t.Fatalf("logs aggregate: %v", err)
	}
	q, _ := url.ParseQuery(rawQuery)
	want := map[string]string{
		"group_by": "source",
		"level":    "warn",
		"source":   "router",
		"from":     "2026-05-01T00:00:00Z",
	}
	for k, v := range want {
		if got := q.Get(k); got != v {
			t.Errorf("query[%s] = %q, want %q (raw: %s)", k, got, v, rawQuery)
		}
	}
}

func TestLogsAggregate_InvalidFromRFC3339(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "logs", "aggregate", "--from=not-a-date")
	if err == nil {
		t.Fatal("expected error for non-RFC3339 --from")
	}
}

func TestLogsAggregate_JSONPreservesFields(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"source":      "runtime",
			"retention":   "ring_buffer",
			"capacity":    100,
			"sample_size": 25,
			"group_by":    "level",
			"buckets": []map[string]any{
				{"key": "warn", "count": 3, "last_seen": 1760000000000},
			},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "logs", "aggregate")
	})
	if err != nil {
		t.Fatalf("logs aggregate: %v", err)
	}
	for _, want := range []string{"retention", "capacity", "sample_size"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q in: %s", want, out)
		}
	}
}

// Numeric-last_seen rendering regression: the runtime logs endpoint returns
// last_seen as epoch millis (a JSON number), which json.Unmarshal decodes as
// float64. Naively rendering with fmt.Sprintf("%v", ...) produces scientific
// notation (e.g. "1.76e+12") for large numbers — useless in a table. The
// formatLastSeen helper must type-switch and emit RFC3339.
func TestLogsAggregate_LastSeenRendersRFC3339(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"source":      "runtime",
			"retention":   "ring_buffer",
			"capacity":    100,
			"sample_size": 1,
			"group_by":    "level",
			"buckets": []map[string]any{
				{"key": "warn", "count": 3, "last_seen": 1760000000000},
			},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "logs", "aggregate")
	})
	if err != nil {
		t.Fatalf("logs aggregate: %v", err)
	}
	if strings.Contains(out, "e+12") || strings.Contains(out, "e+11") {
		t.Errorf("LAST_SEEN must not render in scientific notation:\n%s", out)
	}
	rfc3339 := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)
	if !rfc3339.MatchString(out) {
		t.Errorf("expected RFC3339 timestamp in LAST_SEEN cell:\n%s", out)
	}
}

func TestLogsAggregate_TableHeaders(t *testing.T) {
	t.Cleanup(func() { resetLogsAggregateFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"source":      "runtime",
			"retention":   "ring_buffer",
			"capacity":    100,
			"sample_size": 1,
			"group_by":    "level",
			"buckets": []map[string]any{
				{"key": "warn", "count": 3, "last_seen": 1760000000000},
			},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "logs", "aggregate")
	})
	if err != nil {
		t.Fatalf("logs aggregate: %v", err)
	}
	for _, want := range []string{"KEY", "COUNT", "LAST_SEEN"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing header %q in:\n%s", want, out)
		}
	}
}

func TestLogsAggregate_DistinctFromTail(t *testing.T) {
	if logsAggregateCmd == nil {
		t.Fatal("logsAggregateCmd not declared")
	}
	if !strings.HasPrefix(logsAggregateCmd.Use, "aggregate") {
		t.Fatalf("Use = %q", logsAggregateCmd.Use)
	}
	// Sanity: aggregate is NOT a watch loop — has no --follow flag.
	if f := logsAggregateCmd.Flags().Lookup("follow"); f != nil {
		t.Errorf("logs aggregate should not have --follow flag (that's logs tail)")
	}
}
