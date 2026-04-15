package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// admin_credentials_users.go adds the user-credentials subtree to adminCredentialsCmd.
// Routes: GET/PUT/DELETE /v1/cli-credentials/{id}/user-credentials/{userId}

var adminCredUserCmd = &cobra.Command{
	Use:   "user-credentials",
	Short: "Manage per-user credentials for a CLI credential",
}

var adminCredUserListCmd = &cobra.Command{
	Use:   "list <credID>",
	Short: "List user credentials for a CLI credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + args[0] + "/user-credentials")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var adminCredUserGetCmd = &cobra.Command{
	Use:   "get <credID> <userID>",
	Short: "Get user credential entry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + args[0] + "/user-credentials/" + args[1])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var adminCredUserSetCmd = &cobra.Command{
	Use:   "set <credID> <userID>",
	Short: "Create or update a user credential entry",
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
		_, err = c.Put("/v1/cli-credentials/"+args[0]+"/user-credentials/"+args[1], body)
		if err != nil {
			return err
		}
		printer.Success("User credential set")
		return nil
	},
}

var adminCredUserDeleteCmd = &cobra.Command{
	Use:   "delete <credID> <userID>",
	Short: "Delete a user credential entry (requires --yes)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete user credential?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/cli-credentials/" + args[0] + "/user-credentials/" + args[1])
		if err != nil {
			return err
		}
		printer.Success("User credential deleted")
		return nil
	},
}

func init() {
	adminCredUserSetCmd.Flags().String("body", "", "Credential payload as JSON object (required)")
	_ = adminCredUserSetCmd.MarkFlagRequired("body")

	adminCredUserCmd.AddCommand(
		adminCredUserListCmd,
		adminCredUserGetCmd,
		adminCredUserSetCmd,
		adminCredUserDeleteCmd,
	)
	adminCredentialsCmd.AddCommand(adminCredUserCmd)
}
