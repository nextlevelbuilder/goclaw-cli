package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// copyProgress copies the HTTP response body to dst, printing progress to stderr.
// Uses streaming io.Copy — never buffers full body to RAM.
func copyProgress(dst io.Writer, resp *http.Response) (int64, error) {
	n, err := io.Copy(dst, resp.Body)
	if err != nil {
		return n, fmt.Errorf("write: %w", err)
	}
	return n, nil
}

// openFileForUpload opens a local file for streaming upload.
// Caller is responsible for closing the returned file.
func openFileForUpload(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return f, nil
}

// writeToFile streams src to a new file at path. Creates parent dirs as needed.
func writeToFile(path string, src io.Reader) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// printProgress prints a download/upload progress message to stderr.
// Kept on stderr so stdout remains clean for JSON/YAML output piping.
func printProgress(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
