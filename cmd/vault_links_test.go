package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVaultLinksCreate_CallsEndpoint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/vault/links" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(vaultEnvelope(map[string]any{
			"id":          "link-1",
			"from_doc_id": "doc-a",
			"to_doc_id":   "doc-b",
			"link_type":   "reference",
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "links", "create",
		"--from=doc-a", "--to=doc-b", "--type=reference")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["from_doc_id"] != "doc-a" {
		t.Errorf("expected from_doc_id=doc-a, got %v", gotBody["from_doc_id"])
	}
	if gotBody["to_doc_id"] != "doc-b" {
		t.Errorf("expected to_doc_id=doc-b, got %v", gotBody["to_doc_id"])
	}
}

func TestVaultLinksCreate_RequiresFromAndTo(t *testing.T) {
	err := runVaultArgs(t, "http://localhost:0", "links", "create")
	if err == nil {
		t.Fatal("expected error when --from and --to are missing")
	}
}

func TestVaultLinksDelete_RequiresYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			called = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "links", "delete", "link-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("DELETE should not be called without --yes in non-interactive mode")
	}
}

func TestVaultLinksDelete_WithYes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/vault/links/link-123" {
			called = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "links", "delete", "link-123", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected DELETE to be called with --yes")
	}
}

func TestVaultLinksBatchGet_CallsEndpoint(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/vault/links/batch" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write(vaultEnvelope([]map[string]any{
			{"id": "l1", "from_doc_id": "doc-a", "to_doc_id": "doc-b", "link_type": "ref"},
		}))
	}))
	defer srv.Close()

	err := runVaultArgs(t, srv.URL, "links", "batch-get", "doc-a", "doc-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ids, ok := gotBody["doc_ids"].([]any)
	if !ok {
		t.Fatalf("expected doc_ids array, got %T: %v", gotBody["doc_ids"], gotBody["doc_ids"])
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 doc IDs, got %d", len(ids))
	}
}

func TestVaultLinksBatchGet_RequiresAtLeastOneArg(t *testing.T) {
	err := runVaultArgs(t, "http://localhost:0", "links", "batch-get")
	if err == nil {
		t.Fatal("expected error when no doc IDs provided")
	}
}
