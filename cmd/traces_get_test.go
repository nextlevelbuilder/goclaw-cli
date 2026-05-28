package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

// loadTraceDetailFixture reads the captured trace detail envelope from testdata.
// The fixture is a single trace map (not wrapped). Tests wrap it via okJSON.
func loadTraceDetailFixture(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile("testdata/trace_detail_get.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return m
}

// errJSON writes an error envelope mimicking the server shape.
func errJSON(t *testing.T, w http.ResponseWriter, status int, code, message string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]any{
		"ok":    false,
		"error": map[string]any{"code": code, "message": message},
	})
	_, _ = w.Write(body)
}

// TestTracesGet_PathAndMethod locks the wire contract: GET /v1/traces/{id}.
func TestTracesGet_PathAndMethod(t *testing.T) {
	var calls int64
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		gotPath = r.URL.Path
		gotMethod = r.Method
		okJSON(t, w, loadTraceDetailFixture(t))
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "json"); err != nil {
		t.Fatalf("traces get: %v", err)
	}
	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected 1 request, got %d", atomic.LoadInt64(&calls))
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/v1/traces/trace_FIXTURE_001" {
		t.Errorf("path = %q, want /v1/traces/trace_FIXTURE_001", gotPath)
	}
}

// TestTracesGet_HappyPath_JSON_LocksFixture round-trips the JSON envelope.
func TestTracesGet_HappyPath_JSON_LocksFixture(t *testing.T) {
	fixture := loadTraceDetailFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, fixture)
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "json")
	})
	if err != nil {
		t.Fatalf("traces get: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\nstdout: %q", err, out)
	}
	if got["trace_id"] != "trace_FIXTURE_001" {
		t.Errorf("trace_id = %v", got["trace_id"])
	}
	if got["agent_id"] != "agent_FIXTURE_001" {
		t.Errorf("agent_id = %v", got["agent_id"])
	}
	if got["status"] != "success" {
		t.Errorf("status = %v", got["status"])
	}
	spans, ok := got["spans"].([]any)
	if !ok || len(spans) != 3 {
		t.Errorf("spans = %v (want 3 entries)", got["spans"])
	}
}

// TestTracesGet_TableMode_HumanReadable_RED — the issue #17 repro (now green after fix).
func TestTracesGet_TableMode_HumanReadable_RED(t *testing.T) {
	fixture := loadTraceDetailFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, fixture)
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "table")
	})
	if err != nil {
		t.Fatalf("traces get: %v", err)
	}
	trimmed := strings.TrimSpace(out)
	if strings.HasPrefix(trimmed, "{") {
		t.Fatalf("table mode rendered raw JSON (starts with '{'): %q", out)
	}
	wantAny := []string{"TRACE", "SPAN", "EVENT", "trace_id", "agent_id"}
	hit := false
	for _, m := range wantAny {
		if strings.Contains(out, m) {
			hit = true
			break
		}
	}
	if !hit {
		t.Fatalf("table mode missing human-readable markers; got: %q", out)
	}
}

// TestTracesGet_TableMode_HasHeaderAndSpanMarkers — verifies header card + tree drawing.
func TestTracesGet_TableMode_HasHeaderAndSpanMarkers(t *testing.T) {
	fixture := loadTraceDetailFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, fixture)
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "table")
	})
	if err != nil {
		t.Fatalf("traces get: %v", err)
	}
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("table mode rendered raw JSON: %q", out)
	}
	if !strings.Contains(out, "TRACE_ID") {
		t.Errorf("missing TRACE_ID header in: %q", out)
	}
	// At least one tree connector must appear (├─ or └─).
	if !strings.Contains(out, "├") && !strings.Contains(out, "└") {
		t.Errorf("missing span tree connectors (├ / └) in: %q", out)
	}
	if !strings.Contains(out, "EVENTS") {
		t.Errorf("missing EVENTS section in: %q", out)
	}
}

// TestTracesGet_JSONMode_PreservesStructure — all top-level fixture keys present in JSON output.
func TestTracesGet_JSONMode_PreservesStructure(t *testing.T) {
	fixture := loadTraceDetailFixture(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, fixture)
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "json")
	})
	if err != nil {
		t.Fatalf("traces get: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	for k := range fixture {
		if _, ok := got[k]; !ok {
			t.Errorf("top-level key %q missing from JSON output", k)
		}
	}
	// Nested reachability: spans[0].name
	spans, ok := got["spans"].([]any)
	if !ok || len(spans) == 0 {
		t.Fatalf("spans not reachable")
	}
	first, ok := spans[0].(map[string]any)
	if !ok || first["name"] == nil {
		t.Errorf("spans[0].name not reachable: %v", spans[0])
	}
}

// TestTracesGet_NotFound_ExitCode3 — 404 → ExitNotFound.
func TestTracesGet_NotFound_ExitCode3(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errJSON(t, w, http.StatusNotFound, "NOT_FOUND", "trace not found")
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "traces", "get", "doesnotexist", "--output", "json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := output.FromError(err); code != output.ExitNotFound {
		t.Errorf("exit code = %d, want %d (ExitNotFound)", code, output.ExitNotFound)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") {
		t.Errorf("error message should mention 'not found': %q", err.Error())
	}
}

// TestTracesGet_PermissionDenied_ExitCode2 — 403 → ExitAuth.
func TestTracesGet_PermissionDenied_ExitCode2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errJSON(t, w, http.StatusForbidden, "TENANT_ACCESS_REVOKED", "tenant access revoked")
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := output.FromError(err); code != output.ExitAuth {
		t.Errorf("exit code = %d, want %d (ExitAuth)", code, output.ExitAuth)
	}
}

// TestTracesGet_MalformedID_NoHTTPCall — id validation runs before HTTP; exit 4.
func TestTracesGet_MalformedID_NoHTTPCall(t *testing.T) {
	var calls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		okJSON(t, w, map[string]any{})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	bad := []string{"", "  ", "..", ".", "../etc/passwd", "a/b", "a\\b", "a\x00b", "a b"}
	for _, id := range bad {
		t.Run("id="+strings.ReplaceAll(id, "\x00", "NUL"), func(t *testing.T) {
			before := atomic.LoadInt64(&calls)
			err := runCmd(t, "traces", "get", id, "--output", "json")
			if err == nil {
				t.Fatalf("expected validation error for %q, got nil", id)
			}
			if code := output.FromError(err); code != output.ExitValidation {
				t.Errorf("id=%q: exit = %d, want %d (ExitValidation)", id, code, output.ExitValidation)
			}
			if got := atomic.LoadInt64(&calls); got != before {
				t.Errorf("id=%q: HTTP call made (calls %d -> %d); should be blocked client-side", id, before, got)
			}
		})
	}
}

// TestTracesGet_ServerError_ExitCode5 — 5xx → ExitServer; client retries so calls >= 1.
func TestTracesGet_ServerError_ExitCode5(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short: HTTP client backs off ~3s between 5xx retries")
	}
	var calls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		errJSON(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if code := output.FromError(err); code != output.ExitServer {
		t.Errorf("exit code = %d, want %d (ExitServer)", code, output.ExitServer)
	}
	if n := atomic.LoadInt64(&calls); n < 1 {
		t.Errorf("calls = %d, want >= 1 (client retries 5xx)", n)
	}
}

// TestTracesGet_MalformedResponse_SurfacesError — bad JSON body → wrapped decode error.
func TestTracesGet_MalformedResponse_SurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("this is not json"))
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "traces", "get", "trace_FIXTURE_001", "--output", "json")
	})
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "decode") && !strings.Contains(msg, "unmarshal") && !strings.Contains(msg, "invalid") {
		t.Errorf("error should mention decode/unmarshal/invalid; got: %q", err.Error())
	}
	if strings.TrimSpace(out) == "{}" {
		t.Errorf("stdout is empty JSON object — silent failure regressed: %q", out)
	}
}
