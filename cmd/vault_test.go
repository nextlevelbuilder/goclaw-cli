package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// runVaultArgs executes vault subcommands via rootCmd using env-based config.
// It resets the persistent --yes flag after execution to prevent state leaking
// between tests that share the same cobra command tree in one process.
func runVaultArgs(t *testing.T, serverURL string, args ...string) error {
	t.Helper()
	t.Setenv("GOCLAW_SERVER", serverURL)
	t.Setenv("GOCLAW_TOKEN", "test-token")
	t.Setenv("GOCLAW_OUTPUT", "json")
	rootCmd.SetArgs(append([]string{"vault"}, args...))
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	// Reset persistent --yes flag so it doesn't leak into subsequent tests.
	_ = rootCmd.PersistentFlags().Set("yes", "false")
	return err
}

// vaultEnvelope wraps payload in the standard GoClaw API envelope.
func vaultEnvelope(payload any) []byte {
	b, _ := json.Marshal(map[string]any{"ok": true, "payload": payload})
	return b
}

func TestVaultSearch_CallsEndpoint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/vault/search" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope([]map[string]any{
			{"id": "doc1", "title": "Auth Guide", "path": "notes/auth.md"},
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "search", "authentication")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["query"] != "authentication" {
		t.Errorf("expected query=authentication, got %v", gotBody["query"])
	}
}

func TestVaultSearch_DefaultLimit(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope([]any{}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "search", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default limit=20 should be sent as max_results.
	if gotBody["max_results"] != float64(20) {
		t.Errorf("expected max_results=20, got %v", gotBody["max_results"])
	}
}

func TestVaultTree_CallsEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vault/tree" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{
			"entries": []map[string]any{
				{"name": "agents/", "type": "folder"},
				{"name": "notes/", "type": "folder"},
			},
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "tree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultRescan_RequiresYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/vault/rescan" {
			called = true
		}
		w.Write(vaultEnvelope(map[string]any{"new": 0}))
	}))
	defer srv.Close()

	// Without --yes, non-interactive mode should refuse.
	err := runVaultArgs(t, srv.URL, "rescan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("rescan should not be called without --yes in non-interactive mode")
	}
}

func TestVaultRescan_WithYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/vault/rescan" {
			called = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{"new": 3, "updated": 1}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "rescan", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected rescan endpoint to be called with --yes")
	}
}

func TestVaultGraph_JSONFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vault/graph" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{
			"nodes": []map[string]any{{"id": "n1"}},
			"edges": []map[string]any{{"from_doc_id": "n1", "to_doc_id": "n2", "link_type": "ref"}},
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "graph", "--format=json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultGraph_DOTFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{
			"nodes": []map[string]any{{"id": "docA"}, {"id": "docB"}},
			"edges": []map[string]any{
				{"from_doc_id": "docA", "to_doc_id": "docB", "link_type": "depends-on"},
			},
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "graph", "--format=dot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- graphJSONToDOT unit tests (no server needed) ---

func TestGraphJSONToDOT_ValidEdge(t *testing.T) {
	input := map[string]any{
		"nodes": []map[string]any{{"id": "a"}, {"id": "b"}},
		"edges": []map[string]any{
			{"from_doc_id": "a", "to_doc_id": "b", "link_type": "ref"},
		},
	}
	raw, _ := json.Marshal(input)
	dot, err := graphJSONToDOT(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(dot, "digraph vault") {
		t.Error("DOT output must contain 'digraph vault'")
	}
	if !strings.Contains(dot, `"a" -> "b"`) {
		t.Errorf("expected edge a->b in DOT, got:\n%s", dot)
	}
	if !strings.Contains(dot, `[label="ref"]`) {
		t.Errorf("expected label=ref in DOT edge, got:\n%s", dot)
	}
}

func TestGraphJSONToDOT_EmptyGraph(t *testing.T) {
	input := map[string]any{"nodes": []any{}, "edges": []any{}}
	raw, _ := json.Marshal(input)
	dot, err := graphJSONToDOT(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(dot, "digraph vault {") {
		t.Error("empty graph must still produce valid DOT header")
	}
}

func TestGraphJSONToDOT_SkipsIncompleteEdges(t *testing.T) {
	input := map[string]any{
		"nodes": []any{},
		"edges": []map[string]any{
			{"from_doc_id": "", "to_doc_id": "x"},  // empty from — skip
			{"from_doc_id": "y", "to_doc_id": ""},  // empty to — skip
		},
	}
	raw, _ := json.Marshal(input)
	dot, err := graphJSONToDOT(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(dot, "->") {
		t.Error("expected no edges emitted for incomplete edge data")
	}
}

func TestGraphJSONToDOT_NoLabel(t *testing.T) {
	input := map[string]any{
		"edges": []map[string]any{
			{"from_doc_id": "x", "to_doc_id": "y"}, // no link_type
		},
	}
	raw, _ := json.Marshal(input)
	dot, err := graphJSONToDOT(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(dot, `"x" -> "y"`) {
		t.Error("expected x->y edge without label")
	}
	if strings.Contains(dot, "[label=") {
		t.Error("should not emit label attribute when link_type is empty")
	}
}
