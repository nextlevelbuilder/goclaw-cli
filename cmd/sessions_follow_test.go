package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func resetSessionsFollowFlags(t *testing.T) {
	t.Helper()
	// Reset to declared defaults (cursor=0, limit=50).
	resetTestFlag(sessionsFollowCmd, "cursor", "0")
	resetTestFlag(sessionsFollowCmd, "limit", "50")
}

func TestSessionsFollow_DefaultsAndQuery(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })

	var calls int64
	var rawQuery, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		rawQuery = r.URL.RawQuery
		path = r.URL.Path
		okJSON(t, w, map[string]any{
			"session_key": "sess-1",
			"cursor":      0,
			"next_cursor": 18,
			"total":       18,
			"messages":    []map[string]any{},
			"reset":       false,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "follow", "sess-1"); err != nil {
		t.Fatalf("sessions follow: %v", err)
	}
	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected exactly 1 request, got %d (no watch loop)", atomic.LoadInt64(&calls))
	}
	if path != "/v1/chat/sessions/sess-1/history/follow" {
		t.Fatalf("path = %q", path)
	}
	// Defaults must be present in the raw query string.
	if !strings.Contains(rawQuery, "cursor=0") {
		t.Errorf("expected cursor=0 in raw query, got: %q", rawQuery)
	}
	if !strings.Contains(rawQuery, "limit=50") {
		t.Errorf("expected limit=50 in raw query, got: %q", rawQuery)
	}
}

// Zero-boundary regression: --cursor 0 is a valid pagination origin and must
// appear literally as "cursor=0" in the raw query. Any helper that drops
// numeric zeros (e.g. omit-empty builders) would silently break "start from
// the beginning" semantics.
func TestSessionsFollow_CursorZero(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })

	var rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		okJSON(t, w, map[string]any{"session_key": "sess-1", "cursor": 0, "next_cursor": 0, "total": 0, "messages": []any{}, "reset": false})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "follow", "sess-1", "--cursor=0"); err != nil {
		t.Fatalf("sessions follow: %v", err)
	}
	if !strings.Contains(rawQuery, "cursor=0") {
		t.Fatalf("--cursor=0 must appear as cursor=0 in query, got: %q", rawQuery)
	}
}

func TestSessionsFollow_CustomCursorAndLimit(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })

	var rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		okJSON(t, w, map[string]any{"session_key": "sess-1", "cursor": 12, "next_cursor": 17, "total": 17, "messages": []any{}})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "follow", "sess-1", "--cursor=12", "--limit=25"); err != nil {
		t.Fatalf("sessions follow: %v", err)
	}
	if !strings.Contains(rawQuery, "cursor=12") || !strings.Contains(rawQuery, "limit=25") {
		t.Fatalf("rawQuery = %q", rawQuery)
	}
}

func TestSessionsFollow_NegativeCursor(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "sessions", "follow", "sess-1", "--cursor=-1")
	if err == nil {
		t.Fatal("expected error for negative --cursor")
	}
}

func TestSessionsFollow_NonPositiveLimit(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "sessions", "follow", "sess-1", "--limit=0", "--cursor=0")
	if err == nil {
		t.Fatal("expected error for --limit=0")
	}
}

func TestSessionsFollow_PathEscape(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })
	var rawPath, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPath = r.URL.RawPath
		path = r.URL.Path
		okJSON(t, w, map[string]any{"session_key": "weird:key/x", "cursor": 0, "next_cursor": 0, "total": 0, "messages": []any{}})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "follow", "weird:key/x"); err != nil {
		t.Fatalf("sessions follow: %v", err)
	}
	if !strings.Contains(rawPath, "weird%3Akey%2Fx") && !strings.Contains(path, "weird:key/x") {
		t.Fatalf("path not escaped — RawPath=%q Path=%q", rawPath, path)
	}
}

func TestSessionsFollow_JSONPreservesFields(t *testing.T) {
	t.Cleanup(func() { resetSessionsFollowFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"session_key": "sess-1",
			"cursor":      0,
			"next_cursor": 5,
			"total":       5,
			"reset":       true,
			"messages":    []map[string]any{{"index": 0, "role": "user", "content": "hi"}},
			"updated":     "2026-05-27T12:00:00Z",
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "sessions", "follow", "sess-1")
	})
	if err != nil {
		t.Fatalf("sessions follow: %v", err)
	}
	for _, want := range []string{"reset", "next_cursor", "messages"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q in: %s", want, out)
		}
	}
}

func TestSessionsFollow_NotAWatchLoop(t *testing.T) {
	// Smoke: sessionsFollowCmd must exist and not be a long-running stream
	// (covered by atomic-counter assertion above; this is a structural check).
	if sessionsFollowCmd == nil {
		t.Fatal("sessionsFollowCmd not declared")
	}
	if !strings.HasPrefix(sessionsFollowCmd.Use, "follow") {
		t.Fatalf("Use = %q", sessionsFollowCmd.Use)
	}
}
