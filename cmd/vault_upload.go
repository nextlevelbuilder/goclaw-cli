package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var vaultUploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a file to the vault (multipart streaming)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("file not found: %s", filePath)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		title, _ := cmd.Flags().GetString("title")
		tagsRaw, _ := cmd.Flags().GetString("tags")

		resp, err := uploadVaultFile(c, filePath, title, tagsRaw)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("upload failed [%d]: %s", resp.StatusCode, string(body))
		}
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			printer.Success("File uploaded successfully")
			return nil
		}
		printer.Print(result)
		return nil
	},
}

// uploadVaultFile streams a multipart POST to /v1/vault/upload.
// Content-Type (with boundary) is captured before the goroutine starts
// to avoid a race between writer and PostRaw header setup.
func uploadVaultFile(c *client.HTTPClient, filePath, title, tags string) (*http.Response, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}
	// f is closed inside the goroutine after streaming completes.

	pr, pw := io.Pipe()
	mw := newMultipartWriter(pw)

	// Capture content-type (includes boundary) BEFORE goroutine writes anything.
	ct := mw.contentType()

	go func() {
		defer f.Close()
		if title != "" {
			if err := mw.writeField("title", title); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
		for _, tag := range strings.Split(tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				if err := mw.writeField("tags", tag); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
		}
		if err := mw.writeFile("files", filePath, f); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.CloseWithError(mw.close())
	}()

	return c.PostRaw("/v1/vault/upload", ct, pr)
}

func init() {
	vaultUploadCmd.Flags().String("title", "", "Document title (optional, inferred from filename if omitted)")
	vaultUploadCmd.Flags().String("tags", "", "Comma-separated tags (e.g. go,api,docs)")
	vaultCmd.AddCommand(vaultUploadCmd)
}
