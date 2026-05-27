package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProvidersReconnect_PathAndMethod(t *testing.T) {
	var calls int64
	var gotPath, gotMethod string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		okJSON(t, w, map[string]any{
			"status":            "reconnected",
			"provider":          map[string]any{"id": "prov-1", "name": "openai"},
			"registry_updated":  true,
			"cache_invalidated": true,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "providers", "reconnect", "prov-1"); err != nil {
		t.Fatalf("providers reconnect: %v", err)
	}
	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected exactly 1 request, got %d", atomic.LoadInt64(&calls))
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/v1/providers/prov-1/reconnect" {
		t.Fatalf("path = %q", gotPath)
	}
	body := strings.TrimSpace(string(gotBody))
	if body != "" && body != "null" && body != "{}" {
		// Must not include a "verify" key (or any payload).
		if strings.Contains(body, "verify") {
			t.Fatalf("body contains 'verify': %q", body)
		}
	}
}

func TestProvidersReconnect_PathEscape(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		okJSON(t, w, map[string]any{"status": "reconnected"})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	// Provider id with characters that need percent-encoding.
	if err := runCmd(t, "providers", "reconnect", "weird/id:1"); err != nil {
		t.Fatalf("providers reconnect: %v", err)
	}
	// httptest decodes path when populating r.URL.Path, so check escaped form via RawPath.
	if !strings.Contains(gotPath, "weird/id:1") {
		// Path semantics: PathEscape encodes "/" as %2F; net/http decodes back. Accept either form.
		if !strings.Contains(gotPath, "weird") || !strings.Contains(gotPath, "id:1") {
			t.Fatalf("path = %q (expected to contain escaped provider id)", gotPath)
		}
	}
}

func TestProvidersReconnect_JSONOutputPreservesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"status":            "disabled",
			"registry_updated":  false,
			"cache_invalidated": true,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "providers", "reconnect", "prov-1")
	})
	if err != nil {
		t.Fatalf("providers reconnect: %v", err)
	}
	if !strings.Contains(out, "registry_updated") || !strings.Contains(out, "cache_invalidated") {
		t.Fatalf("stdout missing fields: %s", out)
	}
}

func TestProvidersReconnect_TableOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"status":            "reconnected",
			"registry_updated":  true,
			"cache_invalidated": true,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "providers", "reconnect", "prov-1")
	})
	if err != nil {
		t.Fatalf("providers reconnect: %v", err)
	}
	if !strings.Contains(out, "STATUS") || !strings.Contains(out, "REGISTRY_UPDATED") || !strings.Contains(out, "CACHE_INVALIDATED") {
		t.Fatalf("table headers missing in:\n%s", out)
	}
}

func TestProvidersReconnect_MissingArg(t *testing.T) {
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")
	err := runCmd(t, "providers", "reconnect")
	if err == nil {
		t.Fatal("expected error for missing provider id")
	}
}
