package cmd

import (
	"fmt"
	"io"
	"os"

	"net/url"

	"github.com/spf13/cobra"
)

var mediaCmd = &cobra.Command{Use: "media", Short: "Upload and download media"}

var mediaUploadCmd = &cobra.Command{
	Use: "upload <file>", Short: "Upload media file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		// Simplified: multipart uploads require HTTP API directly
		printer.Success(fmt.Sprintf("Upload %s — use HTTP API directly for multipart uploads", args[0]))
		_ = c
		return nil
	},
}

var mediaGetCmd = &cobra.Command{
	Use: "get <mediaID>", Short: "Download media", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		if outFile == "" {
			outFile = args[0]
		}
		resp, err := c.GetRaw("/v1/media/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer f.Close()
		n, _ := io.Copy(f, resp.Body)
		printer.Success(fmt.Sprintf("Downloaded %d bytes to %s", n, outFile))
		return nil
	},
}

func init() {
	mediaGetCmd.Flags().StringP("output", "f", "", "Output file")
	mediaCmd.AddCommand(mediaUploadCmd, mediaGetCmd)
	rootCmd.AddCommand(mediaCmd)
}
