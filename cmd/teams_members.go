package cmd

import (
	"github.com/spf13/cobra"
)

var teamsMembersCmd = &cobra.Command{Use: "members", Short: "Manage team members"}

var teamsMembersListCmd = &cobra.Command{
	Use: "list <teamID>", Short: "List team members", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.known_users", map[string]any{"team_id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var teamsMembersAddCmd = &cobra.Command{
	Use: "add <teamID>", Short: "Add team member", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		role, _ := cmd.Flags().GetString("role")
		_, err = ws.Call("teams.members.add", map[string]any{
			"team_id": args[0], "agent_id": agent, "role": role,
		})
		if err != nil {
			return err
		}
		printer.Success("Member added")
		return nil
	},
}

var teamsMembersRemoveCmd = &cobra.Command{
	Use: "remove <teamID>", Short: "Remove team member", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		_, err = ws.Call("teams.members.remove", map[string]any{
			"team_id": args[0], "agent_id": agent,
		})
		if err != nil {
			return err
		}
		printer.Success("Member removed")
		return nil
	},
}

func init() {
	teamsMembersAddCmd.Flags().String("agent", "", "Agent ID")
	teamsMembersAddCmd.Flags().String("role", "member", "Role: lead, member")
	_ = teamsMembersAddCmd.MarkFlagRequired("agent")
	teamsMembersRemoveCmd.Flags().String("agent", "", "Agent ID")
	_ = teamsMembersRemoveCmd.MarkFlagRequired("agent")

	teamsMembersCmd.AddCommand(teamsMembersListCmd, teamsMembersAddCmd, teamsMembersRemoveCmd)
	teamsCmd.AddCommand(teamsMembersCmd)
}
