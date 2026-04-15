package cmd

import (
	"github.com/spf13/cobra"
)

// providers_claude_cli.go adds the claude-cli subgroup to providersCmd.
// Currently exposes auth-status for the Claude CLI provider integration.

var providersClaudeCLICmd = &cobra.Command{
	Use:   "claude-cli",
	Short: "Manage Claude CLI provider integration",
}

var providersClaudeCLIAuthStatusCmd = &cobra.Command{
	Use:   "auth-status",
	Short: "Show Claude CLI authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/providers/claude-cli/auth-status")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	providersClaudeCLICmd.AddCommand(providersClaudeCLIAuthStatusCmd)
	providersCmd.AddCommand(providersClaudeCLICmd)
}
