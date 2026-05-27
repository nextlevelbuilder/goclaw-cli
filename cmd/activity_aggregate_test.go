package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func resetActivityAggregateFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"group-by", "from", "to", "actor-type", "actor-id", "action", "entity-type", "entity-id"} {
		resetTestFlag(activityAggregateCmd, name, "")
	}
	resetTestFlag(activityAggregateCmd, "limit", "0")
}

func TestActivityAggregate_MissingGroupBy(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "activity", "aggregate")
	if err == nil {
		t.Fatal("expected error for missing --group-by")
	}
}

func TestActivityAggregate_InvalidGroupBy(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "activity", "aggregate", "--group-by=foo")
	if err == nil {
		t.Fatal("expected error for invalid --group-by=foo")
	}
}

func TestActivityAggregate_ValidGroupByValues(t *testing.T) {
	for _, gb := range []string{"action", "actor_type", "entity_type", "actor_id"} {
		gb := gb
		t.Run(gb, func(t *testing.T) {
			t.Cleanup(func() { resetActivityAggregateFlags(t) })
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				okJSON(t, w, map[string]any{"source": "activity", "group_by": gb, "total": 0, "buckets": []any{}})
			}))
			defer srv.Close()
			t.Setenv("GOCLAW_SERVER", srv.URL)
			t.Setenv("GOCLAW_TOKEN", "test-token")
			if err := runCmd(t, "activity", "aggregate", "--group-by="+gb); err != nil {
				t.Fatalf("group-by=%s: %v", gb, err)
			}
		})
	}
}

func TestActivityAggregate_InvalidFromRFC3339(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "activity", "aggregate", "--group-by=action", "--from=not-a-date")
	if err == nil {
		t.Fatal("expected error for non-RFC3339 --from")
	}
}

func TestActivityAggregate_InvalidToRFC3339(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "activity", "aggregate", "--group-by=action", "--to=tomorrow")
	if err == nil {
		t.Fatal("expected error for non-RFC3339 --to")
	}
}

func TestActivityAggregate_QueryStringHasFilters(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })

	var rawQuery, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		rawQuery = r.URL.RawQuery
		okJSON(t, w, map[string]any{"source": "activity", "group_by": "action", "total": 0, "buckets": []any{}})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "activity", "aggregate",
		"--group-by=action",
		"--from=2026-05-01T00:00:00Z",
		"--to=2026-05-27T00:00:00Z",
		"--limit=25",
		"--actor-type=user",
		"--actor-id=u1",
		"--action=session.branch",
		"--entity-type=session",
		"--entity-id=sess-1",
	); err != nil {
		t.Fatalf("activity aggregate: %v", err)
	}
	if path != "/v1/activity/aggregate" {
		t.Fatalf("path = %q", path)
	}
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	want := map[string]string{
		"group_by":    "action",
		"from":        "2026-05-01T00:00:00Z",
		"to":          "2026-05-27T00:00:00Z",
		"limit":       "25",
		"actor_type":  "user",
		"actor_id":    "u1",
		"action":      "session.branch",
		"entity_type": "session",
		"entity_id":   "sess-1",
	}
	for k, v := range want {
		if got := q.Get(k); got != v {
			t.Errorf("query[%s] = %q, want %q (full raw: %s)", k, got, v, rawQuery)
		}
	}
}

func TestActivityAggregate_JSONPreservesFields(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"source":   "activity",
			"group_by": "action",
			"total":    10,
			"buckets": []map[string]any{
				{"key": "session.branch", "count": 7, "last_seen": "2026-05-27T11:00:00Z"},
			},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "activity", "aggregate", "--group-by=action")
	})
	if err != nil {
		t.Fatalf("activity aggregate: %v", err)
	}
	for _, want := range []string{"source", "group_by", "total", "buckets"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q in: %s", want, out)
		}
	}
}

func TestActivityAggregate_TableHeaders(t *testing.T) {
	t.Cleanup(func() { resetActivityAggregateFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"source":   "activity",
			"group_by": "action",
			"total":    1,
			"buckets": []map[string]any{
				{"key": "session.branch", "count": 1, "last_seen": "2026-05-27T11:00:00Z"},
			},
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "activity", "aggregate", "--group-by=action")
	})
	if err != nil {
		t.Fatalf("activity aggregate: %v", err)
	}
	for _, want := range []string{"KEY", "COUNT", "LAST_SEEN", "session.branch"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q in:\n%s", want, out)
		}
	}
}
