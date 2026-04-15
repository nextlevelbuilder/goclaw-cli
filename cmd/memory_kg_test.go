package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
)

func setupMemoryKGTest(serverURL string) {
	cfg = &config.Config{
		Server:       serverURL,
		Token:        "test-token",
		OutputFormat: "json",
	}
	printer = output.NewPrinter("json")
}

func mockMemoryHTTPServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

// --- kg entities ---

func TestMemoryKGEntitiesList(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/kg/entities") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"ent-1","name":"GoClaw","type":"software"}]`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryKGEntitiesListCmd.RunE(memoryKGEntitiesListCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryKGEntitiesGet(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/kg/entities/ent-1") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"ent-1","name":"GoClaw","type":"software","description":"AI gateway"}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryKGEntitiesGetCmd.RunE(memoryKGEntitiesGetCmd, []string{"agent-1", "ent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryKGEntitiesDelete_RequiresConfirm(t *testing.T) {
	// cfg.Yes=false and non-interactive → tui.Confirm returns false → no-op (nil error)
	cfg = &config.Config{Server: "http://localhost", Token: "tok", OutputFormat: "json", Yes: false}
	printer = output.NewPrinter("json")

	err := memoryKGEntitiesDeleteCmd.RunE(memoryKGEntitiesDeleteCmd, []string{"agent-1", "ent-1"})
	if err != nil {
		t.Fatalf("expected nil (confirm declined), got: %v", err)
	}
}

func TestMemoryKGEntitiesDelete_WithYes(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()
	cfg = &config.Config{
		Server:       srv.URL,
		Token:        "test-token",
		OutputFormat: "json",
		Yes:          true,
	}
	printer = output.NewPrinter("json")

	if err := memoryKGEntitiesDeleteCmd.RunE(memoryKGEntitiesDeleteCmd, []string{"agent-1", "ent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- kg traverse ---

func TestMemoryKGTraverse_RequiresFrom(t *testing.T) {
	cfg = &config.Config{Server: "http://localhost", Token: "tok", OutputFormat: "json"}
	printer = output.NewPrinter("json")
	memoryKGTraverseCmd.Flags().Set("from", "")
	memoryKGTraverseCmd.Flags().Set("depth", "2")

	err := memoryKGTraverseCmd.RunE(memoryKGTraverseCmd, []string{"agent-1"})
	if err == nil {
		t.Fatal("expected error when --from is empty")
	}
	if !strings.Contains(err.Error(), "--from is required") {
		t.Fatalf("expected from error, got: %v", err)
	}
}

func TestMemoryKGTraverse_Success(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/kg/traverse") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"nodes":[{"id":"ent-1"},{"id":"ent-2"}],"edges":[{"from":"ent-1","to":"ent-2"}]}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	memoryKGTraverseCmd.Flags().Set("from", "ent-1")
	memoryKGTraverseCmd.Flags().Set("depth", "2")

	if err := memoryKGTraverseCmd.RunE(memoryKGTraverseCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- kg stats ---

func TestMemoryKGStats(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/kg/stats") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"entity_count":42,"relation_count":87}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryKGStatsCmd.RunE(memoryKGStatsCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- kg graph ---

func TestMemoryKGGraph_Full(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agents/agent-1/kg/graph" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"nodes":[],"edges":[]}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	memoryKGGraphCmd.Flags().Set("compact", "false")
	if err := memoryKGGraphCmd.RunE(memoryKGGraphCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryKGGraph_Compact(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/kg/graph/compact") {
			t.Errorf("expected /kg/graph/compact path, got: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"summary":"compact graph"}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	memoryKGGraphCmd.Flags().Set("compact", "true")
	if err := memoryKGGraphCmd.RunE(memoryKGGraphCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- kg dedup ---

func TestMemoryKGDedupScan(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/kg/dedup/scan") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"candidates_found":3}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryKGDedupScanCmd.RunE(memoryKGDedupScanCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryKGDedupList(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.Contains(r.URL.Path, "/kg/dedup") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"cand-1","entity_a":"ent-1","entity_b":"ent-2","score":0.95}]`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryKGDedupListCmd.RunE(memoryKGDedupListCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryKGDedupMerge_WithYes(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/kg/merge") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"merged":true,"surviving_id":"ent-1"}`))
	})
	defer srv.Close()
	cfg = &config.Config{
		Server:       srv.URL,
		Token:        "test-token",
		OutputFormat: "json",
		Yes:          true,
	}
	printer = output.NewPrinter("json")

	if err := memoryKGDedupMergeCmd.RunE(memoryKGDedupMergeCmd, []string{"agent-1", "ent-1", "ent-2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryKGDedupDismiss(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/kg/dedup/dismiss") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"dismissed":true}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryKGDedupDismissCmd.RunE(memoryKGDedupDismissCmd, []string{"agent-1", "cand-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- memory chunks / index / index-all / documents-global ---

func TestMemoryChunks(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/memory/chunks") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":"chunk-1","content":"...","size":512}]`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryChunksCmd.RunE(memoryChunksCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryIndex(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/memory/index") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"indexed":true,"chunks_created":5}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryIndexCmd.RunE(memoryIndexCmd, []string{"agent-1", "docs/readme.md"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryIndexAll(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/memory/index-all") {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"queued":true}`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryIndexAllCmd.RunE(memoryIndexAllCmd, []string{"agent-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemoryDocumentsGlobal(t *testing.T) {
	srv := mockMemoryHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/memory/documents" {
			t.Errorf("expected /v1/memory/documents, got: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"agent_id":"agent-1","path":"docs/a.md"},{"agent_id":"agent-2","path":"docs/b.md"}]`))
	})
	defer srv.Close()
	setupMemoryKGTest(srv.URL)

	if err := memoryDocumentsGlobalCmd.RunE(memoryDocumentsGlobalCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
