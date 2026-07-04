package client

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// UploadFile performs a streaming multipart POST to path using the client's auth.
// fieldName is the form field name (e.g. "archive", "file").
// Returns the raw *http.Response for caller to read/handle.
func (c *HTTPClient) UploadFile(path, fieldName, filePath string) (*http.Response, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	// Stream file into pipe in goroutine to avoid buffering full file.
	go func() {
		part, partErr := mw.CreateFormFile(fieldName, filepath.Base(filePath))
		if partErr != nil {
			pw.CloseWithError(partErr)
			return
		}
		if _, copyErr := io.Copy(part, f); copyErr != nil {
			pw.CloseWithError(copyErr)
			return
		}
		pw.CloseWithError(mw.Close())
	}()

	resp, err := c.PostRaw(path, mw.FormDataContentType(), pr)
	if err != nil {
		return nil, fmt.Errorf("multipart POST: %w", err)
	}
	return resp, nil
}

// DrainResponse reads the response body and returns an error for HTTP 4xx/5xx.
func DrainResponse(resp *http.Response) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error [%d]: %s", resp.StatusCode, string(body))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
