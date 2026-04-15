package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var vaultEnrichmentCmd = &cobra.Command{
	Use:   "enrichment",
	Short: "Manage vault enrichment pipeline",
}

var vaultEnrichmentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get current enrichment pipeline status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/vault/enrichment/status")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var vaultEnrichmentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running enrichment pipeline (admin)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Stop enrichment pipeline?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/vault/enrichment/stop", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	vaultEnrichmentCmd.AddCommand(vaultEnrichmentStatusCmd, vaultEnrichmentStopCmd)
	vaultCmd.AddCommand(vaultEnrichmentCmd)
}
