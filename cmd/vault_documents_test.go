package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestVaultDocsList_CallsEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vault/documents" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{
			"documents": []map[string]any{
				{"id": "d1", "title": "Guide", "path": "notes/guide.md", "doc_type": "note", "scope": "shared"},
			},
			"total": 1,
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "documents", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultDocsGet_CallsEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vault/documents/doc-123" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{"id": "doc-123", "title": "My Doc"}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "documents", "get", "doc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultDocsCreate_RequiresTitleAndPath(t *testing.T) {
	// Flag validation must fail before any HTTP call.
	err := runVaultArgs(t, "http://localhost:0", "documents", "create")
	if err == nil {
		t.Fatal("expected error when --title and --path are missing")
	}
}

func TestVaultDocsCreate_CallsEndpoint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/vault/documents" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(vaultEnvelope(map[string]any{"id": "new-doc", "title": "My Guide"}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "documents", "create",
		"--title=My Guide", "--path=notes/guide.md",
		"--content=# Guide\nhello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["title"] != "My Guide" {
		t.Errorf("expected title=My Guide, got %v", gotBody["title"])
	}
	if gotBody["path"] != "notes/guide.md" {
		t.Errorf("expected path=notes/guide.md, got %v", gotBody["path"])
	}
	if gotBody["content"] == nil || gotBody["content"] == "" {
		t.Errorf("expected non-empty content field in request body, got %v", gotBody["content"])
	}
}

func TestVaultDocsCreate_RequiresContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called when content missing")
	}))
	defer srv.Close()

	// Reset persistent flag state from prior tests in same process.
	_ = vaultDocsCreateCmd.Flags().Set("content", "")
	_ = vaultDocsCreateCmd.Flags().Set("file", "")
	_ = vaultDocsCreateCmd.Flags().Set("title", "")
	_ = vaultDocsCreateCmd.Flags().Set("path", "")

	err := runVaultArgs(t, srv.URL, "documents", "create",
		"--title=X", "--path=p.md")
	if err == nil {
		t.Fatal("expected error when --content/--file missing")
	}
}

func TestVaultDocsDelete_RequiresYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			called = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	// Without --yes in non-interactive mode, delete must be refused.
	err := runVaultArgs(t, srv.URL, "documents", "delete", "doc-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("DELETE should not be called without --yes in non-interactive mode")
	}
}

func TestVaultDocsDelete_WithYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/vault/documents/doc-abc" {
			called = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "documents", "delete", "doc-abc", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DELETE to be called with --yes")
	}
}

func TestVaultDocsLinks_CallsEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vault/documents/doc-abc/links" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope(map[string]any{
			"outlinks":  []map[string]any{{"id": "l1", "to_doc_id": "doc-xyz", "link_type": "ref"}},
			"backlinks": []map[string]any{},
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "documents", "links", "doc-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- unit tests for helpers ---

func TestExtractDocsList_WithDocuments(t *testing.T) {
	m := map[string]any{
		"documents": []any{
			map[string]any{"id": "d1", "title": "Doc 1"},
			map[string]any{"id": "d2", "title": "Doc 2"},
		},
		"total": 2,
	}
	docs := extractDocsList(m)
	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(docs))
	}
	if docs[0]["id"] != "d1" {
		t.Errorf("expected first doc id=d1, got %v", docs[0]["id"])
	}
}

func TestExtractDocsList_MissingKey(t *testing.T) {
	m := map[string]any{"total": 0}
	docs := extractDocsList(m)
	if docs != nil {
		t.Errorf("expected nil when documents key missing, got %v", docs)
	}
}

func TestReadFileOrStdin_AtPrefix(t *testing.T) {
	tmp := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmp, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	content, err := readFileOrStdin("@" + tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "hello world" {
		t.Errorf("expected 'hello world', got %q", content)
	}
}

func TestReadFileOrStdin_DirectPath(t *testing.T) {
	tmp := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmp, []byte("direct read"), 0o644); err != nil {
		t.Fatal(err)
	}
	content, err := readFileOrStdin(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "direct read" {
		t.Errorf("expected 'direct read', got %q", content)
	}
}

func TestReadFileOrStdin_NotFound(t *testing.T) {
	_, err := readFileOrStdin("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
