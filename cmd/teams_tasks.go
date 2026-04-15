package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// teams_tasks.go — core task CRUD: list/get/get-light/create/assign/approve/reject/comment/comments.
// Advanced ops (delete/delete-bulk/events/active) → teams_tasks_advanced.go

var teamsTasksCmd = &cobra.Command{Use: "tasks", Short: "Manage team tasks"}

var teamsTasksListCmd = &cobra.Command{
	Use:   "list <teamID>",
	Short: "List team tasks",
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
		params := map[string]any{"team_id": args[0]}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			params["status"] = v
		}
		data, err := ws.Call("teams.tasks.list", params)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "TITLE", "STATUS", "ASSIGNEE")
		for _, t := range unmarshalList(data) {
			tbl.AddRow(str(t, "id"), str(t, "title"), str(t, "status"), str(t, "assignee_id"))
		}
		printer.Print(tbl)
		return nil
	},
}

var teamsTasksGetCmd = &cobra.Command{
	Use:   "get <teamID> <taskID>",
	Short: "Get task details",
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
		data, err := ws.Call("teams.tasks.get", map[string]any{
			"team_id": args[0], "task_id": args[1],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var teamsTasksGetLightCmd = &cobra.Command{
	Use:   "get-light <teamID> <taskID>",
	Short: "Get lightweight task summary",
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
		data, err := ws.Call("teams.tasks.get-light", map[string]any{
			"team_id": args[0], "task_id": args[1],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var teamsTasksCreateCmd = &cobra.Command{
	Use:   "create <teamID>",
	Short: "Create a task",
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
		title, _ := cmd.Flags().GetString("title")
		desc, _ := cmd.Flags().GetString("description")
		assignee, _ := cmd.Flags().GetString("assignee")
		params := buildBody("team_id", args[0], "title", title, "description", desc, "assignee_id", assignee)
		data, err := ws.Call("teams.tasks.create", params)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Task created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var teamsTasksAssignCmd = &cobra.Command{
	Use:   "assign <teamID> <taskID>",
	Short: "Assign a task to an agent",
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
		agent, _ := cmd.Flags().GetString("agent")
		_, err = ws.Call("teams.tasks.assign", map[string]any{
			"team_id": args[0], "task_id": args[1], "agent_id": agent,
		})
		if err != nil {
			return err
		}
		printer.Success("Task assigned")
		return nil
	},
}

// approve/reject/comment/comments registered in teams_tasks_review.go init()

func init() {
	teamsTasksListCmd.Flags().String("status", "", "Filter: open, assigned, approved, rejected")

	teamsTasksCreateCmd.Flags().String("title", "", "Task title")
	teamsTasksCreateCmd.Flags().String("description", "", "Task description")
	teamsTasksCreateCmd.Flags().String("assignee", "", "Assignee agent ID")
	_ = teamsTasksCreateCmd.MarkFlagRequired("title")

	teamsTasksAssignCmd.Flags().String("agent", "", "Agent ID")
	_ = teamsTasksAssignCmd.MarkFlagRequired("agent")

	// delete/delete-bulk/events/active → teams_tasks_advanced.go
	// approve/reject/comment/comments → teams_tasks_review.go
	teamsTasksCmd.AddCommand(
		teamsTasksListCmd, teamsTasksGetCmd, teamsTasksGetLightCmd,
		teamsTasksCreateCmd, teamsTasksAssignCmd,
	)
}
