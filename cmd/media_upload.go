package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/client"
)

func uploadMediaFile(c *client.HTTPClient, filePath string) (*http.Response, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}
	pr, pw := io.Pipe()
	mw := newMultipartWriter(pw)
	ct := mw.contentType()

	go func() {
		defer f.Close()
		if err := mw.writeFile("file", filePath, f); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(mw.close())
	}()

	return c.PostRaw("/v1/media/upload", ct, pr)
}
