package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// agents.go — root + list/get/create/update/delete (core CRUD only, <200 LoC).
// share/unshare/regenerate/resummon → agents_sharing.go
// Instances → agents_instances.go | Links → agents_links.go
// Lifecycle → agents_lifecycle.go | Evolution → agents_evolution.go
// Episodic → agents_episodic.go   | v3-flags → agents_v3_flags.go
// Orchestration/codex → agents_misc.go

// agents.go — root command + list/get/create/update/delete/share/unshare/regenerate/resummon.
// Instances extracted → agents_instances.go
// Links extracted → agents_links.go (unchanged)
// Lifecycle (wake/wait/identity/sync-workspace/prompt-preview) → agents_lifecycle.go
// Evolution → agents_evolution.go
// Episodic → agents_episodic.go
// v3-flags → agents_v3_flags.go
// Orchestration/codex-pool-activity → agents_misc.go

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage agents",
}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "KEY", "NAME", "PROVIDER", "MODEL", "STATUS", "TYPE")
		for _, a := range unmarshalList(data) {
			tbl.AddRow(str(a, "id"), str(a, "agent_key"), str(a, "display_name"),
				str(a, "provider"), str(a, "model"), str(a, "status"), str(a, "agent_type"))
		}
		printer.Print(tbl)
		return nil
	},
}

var agentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get agent details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var agentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		agentType, _ := cmd.Flags().GetString("type")
		contextWindow, _ := cmd.Flags().GetInt("context-window")
		workspace, _ := cmd.Flags().GetString("workspace")
		budget, _ := cmd.Flags().GetInt("budget")

		body := buildBody(
			"display_name", name,
			"provider", provider,
			"model", model,
			"agent_type", agentType,
			"context_window", contextWindow,
			"workspace", workspace,
			"monthly_cents", budget,
		)
		data, err := c.Post("/v1/agents", body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		printer.Success(fmt.Sprintf("Agent created: %s (ID: %s)", str(m, "display_name"), str(m, "id")))
		return nil
	},
}

var agentsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update agent configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		for _, flag := range []string{"name", "provider", "model", "workspace", "type"} {
			if cmd.Flags().Changed(flag) {
				val, _ := cmd.Flags().GetString(flag)
				key := flag
				if flag == "name" {
					key = "display_name"
				}
				if flag == "type" {
					key = "agent_type"
				}
				body[key] = val
			}
		}
		if cmd.Flags().Changed("context-window") {
			v, _ := cmd.Flags().GetInt("context-window")
			body["context_window"] = v
		}
		if cmd.Flags().Changed("budget") {
			v, _ := cmd.Flags().GetInt("budget")
			body["monthly_cents"] = v
		}
		if len(body) == 0 {
			return fmt.Errorf("no fields to update — use flags like --name, --model, etc.")
		}
		_, err = c.Put("/v1/agents/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Agent updated")
		return nil
	},
}

var agentsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete agent %s?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/agents/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Agent deleted")
		return nil
	},
}

func init() {
	// Agent CRUD flags
	for _, cmd := range []*cobra.Command{agentsCreateCmd, agentsUpdateCmd} {
		cmd.Flags().String("name", "", "Agent display name")
		cmd.Flags().String("provider", "", "LLM provider name")
		cmd.Flags().String("model", "", "Model identifier")
		cmd.Flags().String("type", "open", "Agent type: open or predefined")
		cmd.Flags().Int("context-window", 0, "Context window size")
		cmd.Flags().String("workspace", "", "Workspace directory")
		cmd.Flags().Int("budget", 0, "Monthly budget in cents")
	}

	// Wire up core CRUD commands.
	// share/unshare/regenerate/resummon → agents_sharing.go init()
	// agentsInstancesCmd → agents_instances.go init()
	// agentsLinksCmd     → agents_links.go init()
	// lifecycle/evolution/episodic/v3-flags/misc → respective files' init()
	agentsCmd.AddCommand(
		agentsListCmd, agentsGetCmd, agentsCreateCmd, agentsUpdateCmd, agentsDeleteCmd,
	)
	rootCmd.AddCommand(agentsCmd)
}
