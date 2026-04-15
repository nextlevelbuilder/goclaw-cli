package cmd

import (
	"io"
	"mime/multipart"
	"path/filepath"
)

// vaultMultipartWriter wraps multipart.Writer with helpers for streaming uploads.
type vaultMultipartWriter struct {
	mw *multipart.Writer
}

// newMultipartWriter creates a multipart writer writing to w (e.g. an io.PipeWriter).
func newMultipartWriter(w io.Writer) *vaultMultipartWriter {
	return &vaultMultipartWriter{mw: multipart.NewWriter(w)}
}

// contentType returns the Content-Type header value including the boundary.
// Must be called before any goroutine writes to capture the correct boundary.
func (v *vaultMultipartWriter) contentType() string {
	return v.mw.FormDataContentType()
}

// writeField writes a plain text form field.
func (v *vaultMultipartWriter) writeField(key, value string) error {
	w, err := v.mw.CreateFormField(key)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, value)
	return err
}

// writeFile streams r into a multipart file part named by basename of filePath.
func (v *vaultMultipartWriter) writeFile(fieldName, filePath string, r io.Reader) error {
	part, err := v.mw.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, r)
	return err
}

// close finalises the multipart boundary.
func (v *vaultMultipartWriter) close() error {
	return v.mw.Close()
}
