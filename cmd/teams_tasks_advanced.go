package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// teams_tasks_advanced.go — delete, delete-bulk, events (follow), get-light, active-by-session.
// Extracted from teams_tasks.go to keep files <200 LoC.

var teamsTasksDeleteCmd = &cobra.Command{
	Use:   "delete <teamID> <taskID>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete task %s?", args[1]), cfg.Yes) {
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

var teamsTasksDeleteBulkCmd = &cobra.Command{
	Use:   "delete-bulk <teamID>",
	Short: "Delete multiple tasks by ID",
	Long: `Delete multiple tasks at once. Accepts a comma-separated list of task IDs.

WS method: teams.tasks.delete-bulk

Example:
  goclaw teams tasks delete-bulk team-1 --ids=task-1,task-2,task-3 --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idsStr, _ := cmd.Flags().GetString("ids")
		if idsStr == "" {
			return fmt.Errorf("--ids is required")
		}
		ids := strings.Split(idsStr, ",")
		for i, id := range ids {
			ids[i] = strings.TrimSpace(id)
		}

		if !tui.Confirm(fmt.Sprintf("Delete %d tasks from team %s?", len(ids), args[0]), cfg.Yes) {
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

		_, err = ws.Call("teams.tasks.delete-bulk", map[string]any{
			"team_id":  args[0],
			"task_ids": ids,
		})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted %d tasks", len(ids)))
		return nil
	},
}

var teamsTasksEventsCmd = &cobra.Command{
	Use:   "events <teamID> <taskID>",
	Short: "Stream events for a task",
	Long: `Stream real-time events for a specific task.

Without --follow: returns current event snapshot.
With --follow: streams continuously until Ctrl+C.

WS method: teams.tasks.events

Example:
  goclaw teams tasks events team-1 task-42
  goclaw teams tasks events team-1 task-42 --follow`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")
		params := map[string]any{"team_id": args[0], "task_id": args[1]}

		handler := func(event *client.WSEvent) error {
			var payload map[string]any
			if err := json.Unmarshal(event.Payload, &payload); err == nil {
				printer.Print(payload)
			}
			return nil
		}

		if follow {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			if output.IsTTY(int(os.Stdout.Fd())) {
				fmt.Println("Streaming task events... (Ctrl+C to stop)")
			}

			err := client.FollowStream(
				ctx, cfg.Server, cfg.Token, "cli", cfg.Insecure,
				"teams.tasks.events", params, handler, nil,
			)
			if err != nil && ctx.Err() != nil {
				return nil // graceful SIGINT
			}
			return err
		}

		// One-shot snapshot
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		data, err := ws.Call("teams.tasks.events", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var teamsTasksActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "Get active tasks for a session",
	Long: `Get tasks currently active for a given session key.

WS method: teams.tasks.active-by-session

Example:
  goclaw teams tasks active --session=my-session-key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session, _ := cmd.Flags().GetString("session")
		if session == "" {
			return fmt.Errorf("--session is required")
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.tasks.active-by-session", map[string]any{
			"session_key": session,
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	teamsTasksDeleteBulkCmd.Flags().String("ids", "", "Comma-separated task IDs to delete")
	teamsTasksEventsCmd.Flags().Bool("follow", false, "Stream events continuously (Ctrl+C to stop)")
	teamsTasksActiveCmd.Flags().String("session", "", "Session key")

	teamsTasksCmd.AddCommand(
		teamsTasksDeleteCmd,
		teamsTasksDeleteBulkCmd,
		teamsTasksEventsCmd,
		teamsTasksActiveCmd,
	)
}
