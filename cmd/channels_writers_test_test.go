package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resetChannelsWritersTestFlags(t *testing.T) {
	t.Helper()
	for _, name := range []string{"group-id", "user-id"} {
		resetTestFlag(channelsWritersTestCmd, name, "")
	}
}

func TestChannelsWritersTest_BodyShape(t *testing.T) {
	t.Cleanup(func() { resetChannelsWritersTestFlags(t) })

	var gotPath, gotMethod string
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		okJSON(t, w, map[string]any{
			"allowed":      true,
			"reason":       "writer",
			"instance_id":  "inst-1",
			"agent_id":     "agent-1",
			"group_id":     "group:telegram:-100123",
			"user_id":      "386246614",
			"writer_count": 3,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "channels", "writers", "test", "inst-1",
		"--group-id=group:telegram:-100123",
		"--user-id=386246614"); err != nil {
		t.Fatalf("channels writers test: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/v1/channels/instances/inst-1/writers/test" {
		t.Fatalf("path = %q", gotPath)
	}
	if body["group_id"] != "group:telegram:-100123" || body["user_id"] != "386246614" {
		t.Fatalf("body fields wrong: %#v", body)
	}
	// Body must contain ONLY these two keys.
	if len(body) != 2 {
		t.Fatalf("body has extra keys: %#v", body)
	}
}

func TestChannelsWritersTest_MissingGroupID(t *testing.T) {
	t.Cleanup(func() { resetChannelsWritersTestFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "channels", "writers", "test", "inst-1", "--user-id=u1")
	if err == nil {
		t.Fatal("expected error for missing --group-id")
	}
}

func TestChannelsWritersTest_MissingUserID(t *testing.T) {
	t.Cleanup(func() { resetChannelsWritersTestFlags(t) })
	t.Setenv("GOCLAW_SERVER", "http://localhost:9")
	t.Setenv("GOCLAW_TOKEN", "test-token")

	err := runCmd(t, "channels", "writers", "test", "inst-1", "--group-id=g1")
	if err == nil {
		t.Fatal("expected error for missing --user-id")
	}
}

func TestChannelsWritersTest_PathEscape(t *testing.T) {
	t.Cleanup(func() { resetChannelsWritersTestFlags(t) })

	var rawPath, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPath = r.URL.RawPath
		path = r.URL.Path
		okJSON(t, w, map[string]any{"allowed": true, "reason": "writer", "writer_count": 1})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")

	if err := runCmd(t, "channels", "writers", "test", "weird/id:1", "--group-id=g1", "--user-id=u1"); err != nil {
		t.Fatalf("channels writers test: %v", err)
	}
	if !strings.Contains(rawPath, "weird%2Fid%3A1") && !strings.Contains(path, "weird/id:1") {
		t.Fatalf("path not escaped — RawPath=%q Path=%q", rawPath, path)
	}
}

func TestChannelsWritersTest_JSONOutput(t *testing.T) {
	t.Cleanup(func() { resetChannelsWritersTestFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"allowed":      false,
			"reason":       "not_writer",
			"writer_count": 2,
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "channels", "writers", "test", "inst-1", "--group-id=g1", "--user-id=u1")
	})
	if err != nil {
		t.Fatalf("channels writers test: %v", err)
	}
	for _, want := range []string{"allowed", "reason", "writer_count"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q in: %s", want, out)
		}
	}
}

func TestChannelsWritersTest_TableHeaders(t *testing.T) {
	t.Cleanup(func() { resetChannelsWritersTestFlags(t) })
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okJSON(t, w, map[string]any{
			"allowed":      true,
			"reason":       "writer",
			"writer_count": 3,
			"group_id":     "g1",
			"user_id":      "u1",
		})
	}))
	defer srv.Close()
	t.Setenv("GOCLAW_SERVER", srv.URL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "table")

	out, err := captureStdout(t, func() error {
		return runCmd(t, "channels", "writers", "test", "inst-1", "--group-id=g1", "--user-id=u1")
	})
	if err != nil {
		t.Fatalf("channels writers test: %v", err)
	}
	for _, want := range []string{"ALLOWED", "REASON", "WRITER_COUNT", "GROUP_ID", "USER_ID"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing header %q in:\n%s", want, out)
		}
	}
}
