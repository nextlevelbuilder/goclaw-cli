package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DownloadSigned fetches a URL without any Authorization header (signed token flow).
// The server uses a time-limited token embedded in the URL so no auth header is needed.
// Progress callback receives bytes written so far; pass nil to skip.
func DownloadSigned(url string, dst io.Writer, insecure bool, progress func(written int64)) error {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	c := &http.Client{
		Timeout:   10 * time.Minute, // large files may take a while
		Transport: transport,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	// Intentionally NO Authorization header — token is part of URL.

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed [%d]: %s", resp.StatusCode, string(body))
	}

	written, err := copyWithProgress(dst, resp.Body, progress)
	if err != nil {
		return fmt.Errorf("write download (wrote %d bytes): %w", written, err)
	}
	return nil
}

// copyWithProgress copies from src to dst, calling progress after each chunk.
func copyWithProgress(dst io.Writer, src io.Reader, progress func(int64)) (int64, error) {
	if progress == nil {
		return io.Copy(dst, src)
	}
	buf := make([]byte, 32*1024)
	var total int64
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			total += int64(written)
			progress(total)
			if writeErr != nil {
				return total, writeErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return total, readErr
		}
	}
	return total, nil
}
