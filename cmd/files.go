package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// files.go — file signing helper. POST /v1/files/sign returns a signed URL.

var filesCmd = &cobra.Command{Use: "files", Short: "File operations"}

var filesSignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Generate a signed URL for a server-side file",
	Long: `POST /v1/files/sign

Flags:
  --path=<path>   Required. Server-side file path to sign.
  --ttl=<sec>     Optional. Validity in seconds.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			return fmt.Errorf("--path is required")
		}
		ttl, _ := cmd.Flags().GetInt("ttl")
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := map[string]any{"path": path}
		if ttl > 0 {
			body["ttl"] = ttl
		}
		data, err := c.Post("/v1/files/sign", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	filesSignCmd.Flags().String("path", "", "Server-side file path")
	filesSignCmd.Flags().Int("ttl", 0, "Validity in seconds (optional)")
	_ = filesSignCmd.MarkFlagRequired("path")
	filesCmd.AddCommand(filesSignCmd)
	rootCmd.AddCommand(filesCmd)
}
