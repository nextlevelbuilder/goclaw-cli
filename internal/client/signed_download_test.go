package client_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
)

func TestDownloadSigned_NoAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("binary-content"))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	err := client.DownloadSigned(srv.URL+"/download/token123", &buf, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
	if buf.String() != "binary-content" {
		t.Errorf("unexpected body: %q", buf.String())
	}
}

func TestDownloadSigned_HTTP4xx_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer srv.Close()

	err := client.DownloadSigned(srv.URL+"/download/bad-token", &bytes.Buffer{}, false, nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

func TestDownloadSigned_ProgressCallback(t *testing.T) {
	payload := strings.Repeat("x", 1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	var lastWritten int64
	var buf bytes.Buffer
	err := client.DownloadSigned(srv.URL+"/file", &buf, false, func(n int64) {
		lastWritten = n
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastWritten != int64(len(payload)) {
		t.Errorf("expected progress %d, got %d", len(payload), lastWritten)
	}
}
