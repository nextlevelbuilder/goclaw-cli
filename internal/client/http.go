package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient wraps net/http.Client with auth and error handling.
type HTTPClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Verbose    bool
}

// NewHTTPClient creates a client for the given server URL.
func NewHTTPClient(baseURL, token string, insecure bool) *HTTPClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	return &HTTPClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// apiResponse is the standard envelope returned by GoClaw.
type apiResponse struct {
	OK      bool            `json:"ok"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// Get performs an HTTP GET request.
func (c *HTTPClient) Get(path string) (json.RawMessage, error) {
	return c.do("GET", path, nil)
}

// Post performs an HTTP POST with a JSON body.
func (c *HTTPClient) Post(path string, body any) (json.RawMessage, error) {
	return c.do("POST", path, body)
}

// Put performs an HTTP PUT with a JSON body.
func (c *HTTPClient) Put(path string, body any) (json.RawMessage, error) {
	return c.do("PUT", path, body)
}

// Patch performs an HTTP PATCH with a JSON body.
func (c *HTTPClient) Patch(path string, body any) (json.RawMessage, error) {
	return c.do("PATCH", path, body)
}

// Delete performs an HTTP DELETE.
func (c *HTTPClient) Delete(path string) (json.RawMessage, error) {
	return c.do("DELETE", path, nil)
}

// PostRaw sends a POST and returns the raw *http.Response (for multipart uploads).
func (c *HTTPClient) PostRaw(path string, contentType string, body io.Reader) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return c.HTTPClient.Do(req)
}

// GetRaw sends a GET and returns the raw *http.Response (for file downloads).
func (c *HTTPClient) GetRaw(path string) (*http.Response, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return c.HTTPClient.Do(req)
}

// HealthCheck performs a simple GET /health check.
func (c *HTTPClient) HealthCheck() error {
	url := c.BaseURL + "/health"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPClient) do(method, path string, body any) (json.RawMessage, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	// Retry on 429 / 5xx
	var resp *http.Response
	for attempt := range 3 {
		resp, err = c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		if resp.StatusCode != 429 && resp.StatusCode < 500 {
			break
		}
		resp.Body.Close()
		if attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			// Re-create request body for retry
			if body != nil {
				data, _ := json.Marshal(body)
				req.Body = io.NopCloser(bytes.NewReader(data))
			}
		}
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Try to parse as API envelope
	var apiResp apiResponse
	if err := json.Unmarshal(respData, &apiResp); err == nil {
		if !apiResp.OK && apiResp.Error != nil {
			apiResp.Error.StatusCode = resp.StatusCode
			return nil, apiResp.Error
		}
		if apiResp.Payload != nil {
			return apiResp.Payload, nil
		}
	}

	// Non-envelope response (health, raw endpoints)
	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Code:       http.StatusText(resp.StatusCode),
			Message:    string(respData),
		}
	}

	return respData, nil
}
