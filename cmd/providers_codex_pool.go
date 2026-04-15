package cmd

import (
	"github.com/spf13/cobra"
)

// providers_codex_pool.go adds codex-pool-activity to providersCmd.
// Returns activity data for a provider's Codex pool (code execution sessions).

var providersCodexPoolActivityCmd = &cobra.Command{
	Use:   "codex-pool-activity <id>",
	Short: "Show Codex pool activity for a provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/providers/" + args[0] + "/codex-pool-activity")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	providersCmd.AddCommand(providersCodexPoolActivityCmd)
}
