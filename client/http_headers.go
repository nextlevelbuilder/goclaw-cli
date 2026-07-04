package client

import (
	"io"
	"net/http"
)

// GetRawWithHeaders sends a GET with additional headers and returns the raw response.
func (c *HTTPClient) GetRawWithHeaders(path string, headers map[string]string) (*http.Response, error) {
	return c.doRawWithHeaders("GET", path, "", nil, headers)
}

// PostRawWithHeaders sends a POST with additional headers and returns the raw response.
func (c *HTTPClient) PostRawWithHeaders(path, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return c.doRawWithHeaders("POST", path, contentType, body, headers)
}

func (c *HTTPClient) doRawWithHeaders(method, path, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	return c.HTTPClient.Do(req)
}
