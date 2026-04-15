package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// agents_sharing.go — share/unshare/regenerate/resummon commands, split from agents.go for LoC.

var agentsShareCmd = &cobra.Command{
	Use:   "share <agentID>",
	Short: "Share agent with a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		userID, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")
		body := buildBody("user_id", userID, "role", role)
		_, err = c.Post("/v1/agents/"+args[0]+"/shares", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Agent shared with %s (role: %s)", userID, role))
		return nil
	},
}

var agentsUnshareCmd = &cobra.Command{
	Use:   "unshare <agentID>",
	Short: "Revoke agent share from a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		userID, _ := cmd.Flags().GetString("user")
		_, err = c.Delete("/v1/agents/" + args[0] + "/shares/" + userID)
		if err != nil {
			return err
		}
		printer.Success("Share revoked")
		return nil
	},
}

var agentsRegenerateCmd = &cobra.Command{
	Use:   "regenerate <id>",
	Short: "Regenerate agent configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+args[0]+"/regenerate", nil)
		if err != nil {
			return err
		}
		printer.Success("Agent regenerated")
		return nil
	},
}

var agentsResummonCmd = &cobra.Command{
	Use:   "resummon <id>",
	Short: "Re-summon agent setup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+args[0]+"/resummon", nil)
		if err != nil {
			return err
		}
		printer.Success("Agent re-summoned")
		return nil
	},
}

func init() {
	agentsShareCmd.Flags().String("user", "", "User ID to share with")
	agentsShareCmd.Flags().String("role", "operator", "Role: admin, operator, viewer")
	_ = agentsShareCmd.MarkFlagRequired("user")

	agentsUnshareCmd.Flags().String("user", "", "User ID to revoke")
	_ = agentsUnshareCmd.MarkFlagRequired("user")

	agentsCmd.AddCommand(agentsShareCmd, agentsUnshareCmd, agentsRegenerateCmd, agentsResummonCmd)
}
