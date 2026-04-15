package cmd

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestVaultUpload_RequiresFile(t *testing.T) {
	// ExactArgs(1) — no args should return usage error.
	err := runVaultArgs(t, "http://localhost:0", "upload")
	if err == nil {
		t.Fatal("expected error when no file argument provided")
	}
}

func TestVaultUpload_FileNotFound(t *testing.T) {
	err := runVaultArgs(t, "http://localhost:0", "upload", "/nonexistent/file.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected 'file not found' error, got: %v", err)
	}
}

func TestVaultUpload_StreamsMultipart(t *testing.T) {
	var receivedFilename string
	var receivedTitle string
	var receivedField string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/vault/upload" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Parse the multipart body.
		ct := r.Header.Get("Content-Type")
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			t.Errorf("invalid Content-Type: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			switch part.FormName() {
			case "title":
				b, _ := io.ReadAll(part)
				receivedTitle = string(b)
			case "tags":
				b, _ := io.ReadAll(part)
				receivedField = string(b)
			case "files":
				receivedFilename = part.FileName()
				io.Copy(io.Discard, part)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"count": 1, "documents": []any{}},
		})
	}))
	defer srv.Close()

	// Create a temp file to upload.
	tmp, err := os.CreateTemp(t.TempDir(), "upload-*.md")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("# Test content")
	tmp.Close()

	err = runVaultArgs(t, srv.URL, "upload", tmp.Name(),
		"--title=My Upload", "--tags=go,test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedTitle != "My Upload" {
		t.Errorf("expected title=My Upload, got %q", receivedTitle)
	}
	if receivedFilename == "" {
		t.Error("expected filename in multipart, got empty")
	}
	// At least one tag should have been sent.
	if receivedField == "" {
		t.Error("expected tags field in multipart form")
	}
}

func TestVaultUpload_NoTitleOrTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/vault/upload" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"count": 1},
		})
	}))
	defer srv.Close()

	tmp, err := os.CreateTemp(t.TempDir(), "upload-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("content")
	tmp.Close()

	// No --title or --tags flags — should still succeed.
	err = runVaultArgs(t, srv.URL, "upload", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error without optional flags: %v", err)
	}
}
