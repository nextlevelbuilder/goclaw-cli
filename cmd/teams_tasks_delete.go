package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var teamsTasksDeleteCmd = &cobra.Command{
	Use: "delete <teamID> <taskID>", Short: "Delete a task", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this task?", cfg.Yes) {
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
		_, err = ws.Call("teams.tasks.delete", map[string]any{
			"team_id": args[0], "task_id": args[1],
		})
		if err != nil {
			return err
		}
		printer.Success("Task deleted")
		return nil
	},
}

func init() {
	teamsTasksCmd.AddCommand(teamsTasksDeleteCmd)
}
