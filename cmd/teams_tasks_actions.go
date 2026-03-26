package cmd

import (
	"github.com/spf13/cobra"
)

var teamsTasksApproveCmd = &cobra.Command{
	Use: "approve <teamID> <taskID>", Short: "Approve task", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("teams.tasks.approve", map[string]any{
			"team_id": args[0], "task_id": args[1],
		})
		if err != nil {
			return err
		}
		printer.Success("Task approved")
		return nil
	},
}

var teamsTasksRejectCmd = &cobra.Command{
	Use: "reject <teamID> <taskID>", Short: "Reject task", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		reason, _ := cmd.Flags().GetString("reason")
		_, err = ws.Call("teams.tasks.reject", map[string]any{
			"team_id": args[0], "task_id": args[1], "reason": reason,
		})
		if err != nil {
			return err
		}
		printer.Success("Task rejected")
		return nil
	},
}

var teamsTasksCommentCmd = &cobra.Command{
	Use: "comment <teamID> <taskID>", Short: "Add task comment", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		text, _ := cmd.Flags().GetString("text")
		_, err = ws.Call("teams.tasks.comment", map[string]any{
			"team_id": args[0], "task_id": args[1], "text": text,
		})
		if err != nil {
			return err
		}
		printer.Success("Comment added")
		return nil
	},
}

var teamsTasksCommentsCmd = &cobra.Command{
	Use: "comments <teamID> <taskID>", Short: "List task comments", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.tasks.comments", map[string]any{
			"team_id": args[0], "task_id": args[1],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var teamsTasksEventsCmd = &cobra.Command{
	Use: "events <teamID> <taskID>", Short: "List task events", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.tasks.events", map[string]any{
			"team_id": args[0], "task_id": args[1],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	teamsTasksRejectCmd.Flags().String("reason", "", "Rejection reason")
	teamsTasksCommentCmd.Flags().String("text", "", "Comment text")
	_ = teamsTasksCommentCmd.MarkFlagRequired("text")

	teamsTasksCmd.AddCommand(teamsTasksApproveCmd, teamsTasksRejectCmd,
		teamsTasksCommentCmd, teamsTasksCommentsCmd, teamsTasksEventsCmd)
}
