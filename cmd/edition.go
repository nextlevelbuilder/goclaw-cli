package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var editionCmd = &cobra.Command{
	Use:   "edition",
	Short: "Show the server edition and feature set",
	Long: `Display the current GoClaw server edition info (community, pro, enterprise).
No authentication required.

Example:
  goclaw edition`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Server == "" {
			return client.ErrServerRequired
		}
		// Token is not required for this endpoint.
		c := client.NewHTTPClient(cfg.Server, cfg.Token, cfg.Insecure)
		data, err := c.Get("/v1/edition")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editionCmd)
}
