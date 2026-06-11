package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// admin_credentials_agents.go adds per-agent credential material management.
// Routes: GET/PUT/DELETE /v1/cli-credentials/{id}/agent-credentials[/{agentId}]

var adminCredAgentCredentialsCmd = &cobra.Command{
	Use:   "agent-credentials",
	Short: "Manage per-agent credentials for a CLI credential",
}

var adminCredAgentCredentialsListCmd = &cobra.Command{
	Use:   "list <credID>",
	Short: "List agent credentials for a CLI credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + args[0] + "/agent-credentials")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(rawPayload(data))
			return nil
		}
		printer.Print(agentCredentialsTable(data))
		return nil
	},
}

var adminCredAgentCredentialsGetCmd = &cobra.Command{
	Use:   "get <credID> <agentID>",
	Short: "Get an agent credential entry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + args[0] + "/agent-credentials/" + args[1])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var adminCredAgentCredentialsSetCmd = &cobra.Command{
	Use:   "set <credID> <agentID>",
	Short: "Create or update an agent credential entry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := jsonObjectFlag(cmd, "body", true)
		if err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Put("/v1/cli-credentials/"+args[0]+"/agent-credentials/"+args[1], body)
		if err != nil {
			return err
		}
		printer.Success("Agent credential set")
		return nil
	},
}

var adminCredAgentCredentialsDeleteCmd = &cobra.Command{
	Use:   "delete <credID> <agentID>",
	Short: "Delete an agent credential entry (requires --yes)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete agent credential?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/cli-credentials/" + args[0] + "/agent-credentials/" + args[1])
		if err != nil {
			return err
		}
		printer.Success("Agent credential deleted")
		return nil
	},
}

func init() {
	adminCredAgentCredentialsSetCmd.Flags().String("body", "", "Credential payload as JSON object (required)")
	_ = adminCredAgentCredentialsSetCmd.MarkFlagRequired("body")

	adminCredAgentCredentialsCmd.AddCommand(
		adminCredAgentCredentialsListCmd,
		adminCredAgentCredentialsGetCmd,
		adminCredAgentCredentialsSetCmd,
		adminCredAgentCredentialsDeleteCmd,
	)
	adminCredentialsCmd.AddCommand(adminCredAgentCredentialsCmd)
}
