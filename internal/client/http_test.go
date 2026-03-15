package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHTTPClient(t *testing.T) {
	c := NewHTTPClient("https://example.com/", "tok123", false)
	if c.BaseURL != "https://example.com" {
		t.Errorf("expected trailing slash stripped, got %s", c.BaseURL)
	}
	if c.Token != "tok123" {
		t.Errorf("expected token tok123, got %s", c.Token)
	}
}

func TestHealthCheck_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "", false)
	if err := c.HealthCheck(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHealthCheck_Fail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "", false)
	if err := c.HealthCheck(); err == nil {
		t.Fatal("expected error for 503 status")
	}
}

func TestGet_EnvelopeResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer auth, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"id": "123", "name": "test"},
		})
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "test-token", false)
	data, err := c.Get("/v1/agents")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id=123, got %v", result["id"])
	}
}

func TestGet_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"ok": false,
			"error": map[string]any{
				"code":    "not_found",
				"message": "agent not found",
			},
		})
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "tok", false)
	_, err := c.Get("/v1/agents/999")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("expected code not_found, got %s", apiErr.Code)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestPost_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected JSON content type")
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test-agent" {
			t.Errorf("expected name=test-agent, got %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"id": "new-id"},
		})
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "tok", false)
	data, err := c.Post("/v1/agents", map[string]any{"name": "test-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["id"] != "new-id" {
		t.Errorf("expected id=new-id, got %v", result["id"])
	}
}

func TestNonEnvelopeResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`"OK"`))
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "tok", false)
	data, err := c.Get("/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `"OK"` {
		t.Errorf("expected raw response, got %s", string(data))
	}
}

func TestRetryOn429(t *testing.T) {
	attempt := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":      true,
			"payload": map[string]any{"attempt": attempt},
		})
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "tok", false)
	data, err := c.Get("/v1/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	json.Unmarshal(data, &result)
	if result["attempt"] != float64(2) {
		t.Errorf("expected attempt 2 after retry, got %v", result["attempt"])
	}
}
