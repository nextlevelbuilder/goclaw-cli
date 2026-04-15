package cmd

import (
	"github.com/spf13/cobra"
)

// teams_tasks_review.go — approve/reject/comment/comments task review workflow.
// Extracted from teams_tasks.go to keep files <200 LoC.

var teamsTasksApproveCmd = &cobra.Command{
	Use:   "approve <teamID> <taskID>",
	Short: "Approve a task",
	Args:  cobra.ExactArgs(2),
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
	Use:   "reject <teamID> <taskID>",
	Short: "Reject a task",
	Args:  cobra.ExactArgs(2),
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
	Use:   "comment <teamID> <taskID>",
	Short: "Add a comment to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		body, _ := cmd.Flags().GetString("body")
		_, err = ws.Call("teams.tasks.comment", map[string]any{
			"team_id": args[0], "task_id": args[1], "body": body,
		})
		if err != nil {
			return err
		}
		printer.Success("Comment added")
		return nil
	},
}

var teamsTasksCommentsCmd = &cobra.Command{
	Use:   "comments <teamID> <taskID>",
	Short: "List task comments",
	Args:  cobra.ExactArgs(2),
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

func init() {
	teamsTasksRejectCmd.Flags().String("reason", "", "Rejection reason")
	teamsTasksCommentCmd.Flags().String("body", "", "Comment text")
	_ = teamsTasksCommentCmd.MarkFlagRequired("body")

	teamsTasksCmd.AddCommand(
		teamsTasksApproveCmd, teamsTasksRejectCmd,
		teamsTasksCommentCmd, teamsTasksCommentsCmd,
	)
}
