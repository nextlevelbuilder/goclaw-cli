package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func resetSessionsBranchFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"new-session-key", "label"} {
		resetTestFlag(sessionsBranchCmd, name, "")
	}
	resetTestFlag(sessionsBranchCmd, "up-to-index", "-1")
	// metadata is a StringArray; reset via SliceValue.Replace.
	if f := sessionsBranchCmd.Flags().Lookup("metadata"); f != nil {
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
		}
		f.Changed = false
	}
}

func TestSessionsBranch_BodyShapeWithMetadata(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })

	var gotPath, gotMethod string
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		okJSON(t, w, map[string]any{
			"ok":              true,
			"source_key":      "sess-1",
			"session_key":     "sess-1-branch",
			"copied_messages": 12,
			"total_messages":  24,
			"label":           "demo",
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "branch", "sess-1",
		"--up-to-index=12",
		"--new-session-key=sess-1-branch",
		"--label=demo",
		"--metadata=foo=bar",
		"--metadata=baz=qux"); err != nil {
		t.Fatalf("sessions branch: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/v1/chat/sessions/sess-1/branch" {
		t.Fatalf("path = %q", gotPath)
	}
	// up_to_index must be present as a numeric value
	if v, ok := body["up_to_index"]; !ok {
		t.Fatalf("body missing up_to_index: %#v", body)
	} else if n, ok := v.(float64); !ok || int(n) != 12 {
		t.Fatalf("up_to_index = %#v", v)
	}
	if body["new_session_key"] != "sess-1-branch" {
		t.Errorf("new_session_key = %#v", body["new_session_key"])
	}
	if body["label"] != "demo" {
		t.Errorf("label = %#v", body["label"])
	}
	meta, ok := body["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata is not object: %#v", body["metadata"])
	}
	if meta["foo"] != "bar" || meta["baz"] != "qux" {
		t.Fatalf("metadata = %#v", meta)
	}
}

// Zero-boundary regression: --up-to-index 0 means "branch with zero copied
// messages" (an empty branch from the session start) and must appear as
// "up_to_index":0 in the wire body. Any helper that drops numeric zeros would
// silently turn this into a server-side "missing required field" error.
func TestSessionsBranch_UpToIndexZero(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })

	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		okJSON(t, w, map[string]any{"ok": true, "source_key": "sess-1", "session_key": "sess-1-branch", "copied_messages": 0, "total_messages": 5})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "branch", "sess-1", "--up-to-index=0", "--new-session-key=sess-1-branch"); err != nil {
		t.Fatalf("sessions branch: %v", err)
	}
	v, ok := body["up_to_index"]
	if !ok {
		t.Fatalf("body missing up_to_index when --up-to-index=0: %#v", body)
	}
	n, isNum := v.(float64)
	if !isNum || n != 0 {
		t.Fatalf("up_to_index = %#v, expected 0", v)
	}
}

func TestSessionsBranch_MissingUpToIndex(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "sessions", "branch", "sess-1")
	if err == nil {
		t.Fatal("expected error for missing --up-to-index")
	}
}

func TestSessionsBranch_NegativeUpToIndex(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "sessions", "branch", "sess-1", "--up-to-index=-1")
	if err == nil {
		t.Fatal("expected error for negative --up-to-index")
	}
}

func TestSessionsBranch_MalformedMetadata(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "sessions", "branch", "sess-1", "--up-to-index=0", "--metadata=foobar")
	if err == nil {
		t.Fatal("expected error for malformed --metadata (no '=')")
	}
}

func TestSessionsBranch_PathEscapesSessionKey(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })

	var gotRawPath, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawPath = r.URL.RawPath
		gotPath = r.URL.Path
		okJSON(t, w, map[string]any{"ok": true})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "sessions", "branch", "weird:key/with-slash", "--up-to-index=0"); err != nil {
		t.Fatalf("sessions branch: %v", err)
	}
	// RawPath holds the percent-encoded form; either RawPath has the escape, or decoded Path has the colon/slash.
	// What matters: client did NOT inject literal "/" into path segment.
	if !strings.Contains(gotRawPath, "weird%3Akey%2Fwith-slash") && !strings.Contains(gotPath, "weird:key/with-slash") {
		t.Fatalf("path not escaped — RawPath=%q Path=%q", gotRawPath, gotPath)
	}
}

func TestSessionsBranch_JSONPreservesCopiedAndTotal(t *testing.T) {
	t.Cleanup(func() { resetSessionsBranchFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"ok":              true,
			"copied_messages": 12,
			"total_messages":  24,
			"session_key":     "sess-new",
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "sessions", "branch", "sess-1", "--up-to-index=12")
	})
	if err != nil {
		t.Fatalf("sessions branch: %v", err)
	}
	if !strings.Contains(out, "copied_messages") || !strings.Contains(out, "total_messages") {
		t.Fatalf("stdout missing fields: %s", out)
	}
}
