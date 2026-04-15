package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// teams.go — root command + list/get/create/update/delete (core CRUD only, <200 LoC).
// Members extracted → teams_members.go
// Tasks extracted/extended → teams_tasks.go
// Workspace extracted → teams_workspace.go
// Events → teams_events.go
// Scopes → teams_scopes.go

var teamsCmd = &cobra.Command{Use: "teams", Short: "Manage agent teams"}

var teamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.list", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var teamsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get team details",
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
		data, err := ws.Call("teams.get", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var teamsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a team",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		name, _ := cmd.Flags().GetString("name")
		agents, _ := cmd.Flags().GetStringSlice("agents")
		params := map[string]any{"name": name, "agent_ids": agents}
		data, err := ws.Call("teams.create", params)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Team created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var teamsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update team",
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
		params := map[string]any{"id": args[0]}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			params["name"] = v
		}
		_, err = ws.Call("teams.update", params)
		if err != nil {
			return err
		}
		printer.Success("Team updated")
		return nil
	},
}

var teamsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this team?", cfg.Yes) {
			return nil
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("teams.delete", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Team deleted")
		return nil
	},
}

func init() {
	teamsCreateCmd.Flags().String("name", "", "Team name")
	teamsCreateCmd.Flags().StringSlice("agents", nil, "Agent IDs")
	_ = teamsCreateCmd.MarkFlagRequired("name")
	teamsUpdateCmd.Flags().String("name", "", "Team name")

	teamsCmd.AddCommand(
		teamsListCmd, teamsGetCmd, teamsCreateCmd, teamsUpdateCmd, teamsDeleteCmd,
		teamsMembersCmd, teamsTasksCmd, teamsWorkspaceCmd,
		teamsEventsCmd, teamsScopesCmd,
	)
	rootCmd.AddCommand(teamsCmd)
}
