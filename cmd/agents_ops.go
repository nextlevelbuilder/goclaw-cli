package cmd

import (
	"fmt"

	"net/url"

	"github.com/spf13/cobra"
)

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
		_, err = c.Post("/v1/agents/"+url.PathEscape(args[0])+"/shares", body)
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
		_, err = c.Delete("/v1/agents/" + url.PathEscape(args[0]) + "/shares/" + userID)
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
		_, err = c.Post("/v1/agents/"+url.PathEscape(args[0])+"/regenerate", nil)
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
		_, err = c.Post("/v1/agents/"+url.PathEscape(args[0])+"/resummon", nil)
		if err != nil {
			return err
		}
		printer.Success("Agent re-summoned")
		return nil
	},
}

var agentsWaitCmd = &cobra.Command{
	Use:   "wait <agent-id>",
	Short: "Wait for agent to complete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		session, _ := cmd.Flags().GetString("session")
		timeout, _ := cmd.Flags().GetInt("timeout")
		params := buildBody("agent_id", args[0], "session_key", session, "timeout", timeout)
		data, err := ws.Call("agent.wait", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	agentsShareCmd.Flags().String("user", "", "User ID to share with")
	agentsShareCmd.Flags().String("role", "operator", "Role: admin, operator, viewer")
	_ = agentsShareCmd.MarkFlagRequired("user")
	agentsUnshareCmd.Flags().String("user", "", "User ID to revoke")
	_ = agentsUnshareCmd.MarkFlagRequired("user")
	agentsWaitCmd.Flags().String("session", "", "Session key to wait on")
	agentsWaitCmd.Flags().Int("timeout", 0, "Timeout in seconds (0 = no timeout)")

	agentsCmd.AddCommand(agentsShareCmd, agentsUnshareCmd, agentsRegenerateCmd,
		agentsResummonCmd, agentsWaitCmd)
}
