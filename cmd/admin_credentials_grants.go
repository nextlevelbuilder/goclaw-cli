package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// admin_credentials_grants.go adds the agent-grants subtree to adminCredentialsCmd.
// Routes: CRUD /v1/cli-credentials/{id}/agent-grants[/{grantId}]

var adminCredGrantsCmd = &cobra.Command{
	Use:   "agent-grants",
	Short: "Manage agent grants for a CLI credential",
}

var adminCredGrantsListCmd = &cobra.Command{
	Use:   "list <credID>",
	Short: "List agent grants for a CLI credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + args[0] + "/agent-grants")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var adminCredGrantsCreateCmd = &cobra.Command{
	Use:   "create <credID>",
	Short: "Create an agent grant for a CLI credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bodyJSON, _ := cmd.Flags().GetString("body")
		if bodyJSON == "" {
			return fmt.Errorf("--body is required (JSON object)")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
			return fmt.Errorf("invalid --body JSON: %w", err)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/cli-credentials/"+args[0]+"/agent-grants", body)
		if err != nil {
			return err
		}
		printer.Success("Agent grant created")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var adminCredGrantsGetCmd = &cobra.Command{
	Use:   "get <credID> <grantID>",
	Short: "Get an agent grant",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + args[0] + "/agent-grants/" + args[1])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var adminCredGrantsUpdateCmd = &cobra.Command{
	Use:   "update <credID> <grantID>",
	Short: "Update an agent grant",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		bodyJSON, _ := cmd.Flags().GetString("body")
		if bodyJSON == "" {
			return fmt.Errorf("--body is required (JSON object)")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
			return fmt.Errorf("invalid --body JSON: %w", err)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Put("/v1/cli-credentials/"+args[0]+"/agent-grants/"+args[1], body)
		if err != nil {
			return err
		}
		printer.Success("Agent grant updated")
		return nil
	},
}

var adminCredGrantsDeleteCmd = &cobra.Command{
	Use:   "delete <credID> <grantID>",
	Short: "Delete an agent grant (requires --yes)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this agent grant?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/cli-credentials/" + args[0] + "/agent-grants/" + args[1])
		if err != nil {
			return err
		}
		printer.Success("Agent grant deleted")
		return nil
	},
}

func init() {
	adminCredGrantsCreateCmd.Flags().String("body", "", "Grant payload as JSON object (required)")
	_ = adminCredGrantsCreateCmd.MarkFlagRequired("body")
	adminCredGrantsUpdateCmd.Flags().String("body", "", "Update payload as JSON object (required)")
	_ = adminCredGrantsUpdateCmd.MarkFlagRequired("body")

	adminCredGrantsCmd.AddCommand(
		adminCredGrantsListCmd,
		adminCredGrantsCreateCmd,
		adminCredGrantsGetCmd,
		adminCredGrantsUpdateCmd,
		adminCredGrantsDeleteCmd,
	)
	adminCredentialsCmd.AddCommand(adminCredGrantsCmd)
}
