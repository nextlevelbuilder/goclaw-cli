package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// teams_attachments.go — team task attachment downloads.
// HTTP endpoint: GET /v1/teams/{teamId}/attachments/{attachmentId}/download

var teamsAttachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage team task attachments",
}

var teamsAttachmentsDownloadCmd = &cobra.Command{
	Use:   "download <team-id> <attachment-id>",
	Short: "Download a team task attachment",
	Long: `Download a team task attachment to an explicit local output file.

GET /v1/teams/{teamId}/attachments/{attachmentId}/download

Example:
  goclaw teams attachments download team-1 att-42 --output ./artifact.bin
  goclaw teams attachments download team-1 att-42 -o ./artifact.bin --force`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		outFile, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")
		if outFile == "" {
			return fmt.Errorf("--output is required")
		}
		if err := os.MkdirAll(filepath.Dir(outFile), 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		if !force {
			if _, err := os.Stat(outFile); err == nil {
				return fmt.Errorf("output file already exists: %s (use --force to overwrite)", outFile)
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("check output file: %w", err)
			}
		}

		flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		if force {
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := fmt.Sprintf(
			"/v1/teams/%s/attachments/%s/download",
			url.PathEscape(args[0]),
			url.PathEscape(args[1]),
		)
		resp, err := c.GetRaw(path)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return rawResponseError(resp)
		}
		defer resp.Body.Close()

		f, err := os.OpenFile(outFile, flags, 0644)
		if err != nil {
			if os.IsExist(err) {
				return fmt.Errorf("output file already exists: %s (use --force to overwrite)", outFile)
			}
			return fmt.Errorf("open output file: %w", err)
		}
		defer f.Close()

		n, err := io.Copy(f, resp.Body)
		if err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		printer.Success(fmt.Sprintf("Downloaded %d bytes to %s", n, outFile))
		return nil
	},
}

func init() {
	teamsAttachmentsDownloadCmd.Flags().StringP("output", "o", "", "Output file path")
	teamsAttachmentsDownloadCmd.Flags().Bool("force", false, "Overwrite output file if it exists")
	teamsAttachmentsCmd.AddCommand(teamsAttachmentsDownloadCmd)
}
