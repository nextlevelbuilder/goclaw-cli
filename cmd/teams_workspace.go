package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var teamsWorkspaceCmd = &cobra.Command{Use: "workspace", Short: "Team workspace files"}

var teamsWorkspaceListCmd = &cobra.Command{
	Use: "list <teamID>", Short: "List workspace files", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.workspace.list", map[string]any{"team_id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var teamsWorkspaceReadCmd = &cobra.Command{
	Use: "read <teamID> <path>", Short: "Read workspace file", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.workspace.read", map[string]any{
			"team_id": args[0], "path": args[1],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var teamsWorkspaceDeleteCmd = &cobra.Command{
	Use: "delete <teamID> <path>", Short: "Delete workspace file", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this file?", cfg.Yes) {
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
		_, err = ws.Call("teams.workspace.delete", map[string]any{
			"team_id": args[0], "path": args[1],
		})
		if err != nil {
			return err
		}
		printer.Success("File deleted")
		return nil
	},
}

func init() {
	teamsWorkspaceCmd.AddCommand(teamsWorkspaceListCmd, teamsWorkspaceReadCmd, teamsWorkspaceDeleteCmd)
	teamsCmd.AddCommand(teamsWorkspaceCmd)
}
