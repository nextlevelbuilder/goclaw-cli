package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var teamsTasksCmd = &cobra.Command{Use: "tasks", Short: "Manage team tasks"}

var teamsTasksListCmd = &cobra.Command{
	Use: "list <teamID>", Short: "List team tasks", Args: cobra.ExactArgs(1),
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
	Use: "get <teamID> <taskID>", Short: "Get task details", Args: cobra.ExactArgs(2),
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

var teamsTasksCreateCmd = &cobra.Command{
	Use: "create <teamID>", Short: "Create task", Args: cobra.ExactArgs(1),
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
	Use: "assign <teamID> <taskID>", Short: "Assign task", Args: cobra.ExactArgs(2),
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

func init() {
	teamsTasksListCmd.Flags().String("status", "", "Filter: open, assigned, approved, rejected")
	teamsTasksCreateCmd.Flags().String("title", "", "Task title")
	teamsTasksCreateCmd.Flags().String("description", "", "Task description")
	teamsTasksCreateCmd.Flags().String("assignee", "", "Assignee agent ID")
	_ = teamsTasksCreateCmd.MarkFlagRequired("title")
	teamsTasksAssignCmd.Flags().String("agent", "", "Agent ID")
	_ = teamsTasksAssignCmd.MarkFlagRequired("agent")

	// approve/reject/comment/comments/events registered from teams_tasks_actions.go
	teamsTasksCmd.AddCommand(teamsTasksListCmd, teamsTasksGetCmd, teamsTasksCreateCmd, teamsTasksAssignCmd)
	teamsCmd.AddCommand(teamsTasksCmd)
}
