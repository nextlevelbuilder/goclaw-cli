package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var teamsCmd = &cobra.Command{Use: "teams", Short: "Manage agent teams"}

var teamsListCmd = &cobra.Command{
	Use: "list", Short: "List teams",
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
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "MEMBERS", "TASKS")
		for _, t := range unmarshalList(data) {
			tbl.AddRow(str(t, "id"), str(t, "name"), str(t, "member_count"), str(t, "task_count"))
		}
		printer.Print(tbl)
		return nil
	},
}

var teamsGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get team details", Args: cobra.ExactArgs(1),
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
	Use: "create", Short: "Create a team",
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
	Use: "update <id>", Short: "Update team", Args: cobra.ExactArgs(1),
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
	Use: "delete <id>", Short: "Delete team", Args: cobra.ExactArgs(1),
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

// --- Team Members ---

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

// --- Team Tasks ---

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

// --- Team Workspace ---

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
	teamsCreateCmd.Flags().String("name", "", "Team name")
	teamsCreateCmd.Flags().StringSlice("agents", nil, "Agent IDs")
	_ = teamsCreateCmd.MarkFlagRequired("name")
	teamsUpdateCmd.Flags().String("name", "", "Team name")

	teamsMembersAddCmd.Flags().String("agent", "", "Agent ID")
	teamsMembersAddCmd.Flags().String("role", "member", "Role: lead, member")
	_ = teamsMembersAddCmd.MarkFlagRequired("agent")
	teamsMembersRemoveCmd.Flags().String("agent", "", "Agent ID")
	_ = teamsMembersRemoveCmd.MarkFlagRequired("agent")

	teamsTasksListCmd.Flags().String("status", "", "Filter: open, assigned, approved, rejected")
	teamsTasksCreateCmd.Flags().String("title", "", "Task title")
	teamsTasksCreateCmd.Flags().String("description", "", "Task description")
	teamsTasksCreateCmd.Flags().String("assignee", "", "Assignee agent ID")
	_ = teamsTasksCreateCmd.MarkFlagRequired("title")
	teamsTasksAssignCmd.Flags().String("agent", "", "Agent ID")
	_ = teamsTasksAssignCmd.MarkFlagRequired("agent")
	teamsTasksRejectCmd.Flags().String("reason", "", "Rejection reason")
	teamsTasksCommentCmd.Flags().String("body", "", "Comment text")
	_ = teamsTasksCommentCmd.MarkFlagRequired("body")

	teamsMembersCmd.AddCommand(teamsMembersListCmd, teamsMembersAddCmd, teamsMembersRemoveCmd)
	teamsTasksCmd.AddCommand(teamsTasksListCmd, teamsTasksGetCmd, teamsTasksCreateCmd,
		teamsTasksAssignCmd, teamsTasksApproveCmd, teamsTasksRejectCmd,
		teamsTasksCommentCmd, teamsTasksCommentsCmd)
	teamsWorkspaceCmd.AddCommand(teamsWorkspaceListCmd, teamsWorkspaceReadCmd, teamsWorkspaceDeleteCmd)
	teamsCmd.AddCommand(teamsListCmd, teamsGetCmd, teamsCreateCmd, teamsUpdateCmd, teamsDeleteCmd,
		teamsMembersCmd, teamsTasksCmd, teamsWorkspaceCmd)
	rootCmd.AddCommand(teamsCmd)
}
