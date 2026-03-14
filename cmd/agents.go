package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

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

var agentsShareCmd = &cobra.Command{
	Use:   "share <agentID>",
	Short: "Share agent with a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		userID, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")
		body := buildBody("user_id", userID, "role", role)
		_, err = c.Post("/v1/agents/"+args[0]+"/shares", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Agent shared with %s (role: %s)", userID, role))
		return nil
	},
}

var agentsUnshareCmd = &cobra.Command{
	Use:   "unshare <agentID>",
	Short: "Revoke agent share from a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		userID, _ := cmd.Flags().GetString("user")
		_, err = c.Delete("/v1/agents/" + args[0] + "/shares/" + userID)
		if err != nil {
			return err
		}
		printer.Success("Share revoked")
		return nil
	},
}

var agentsRegenerateCmd = &cobra.Command{
	Use:   "regenerate <id>",
	Short: "Regenerate agent configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+args[0]+"/regenerate", nil)
		if err != nil {
			return err
		}
		printer.Success("Agent regenerated")
		return nil
	},
}

var agentsResummonCmd = &cobra.Command{
	Use:   "resummon <id>",
	Short: "Re-summon agent setup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+args[0]+"/resummon", nil)
		if err != nil {
			return err
		}
		printer.Success("Agent re-summoned")
		return nil
	},
}

// --- Agent Links ---

var agentsLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage agent delegation links",
}

var agentsLinksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List delegation links",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/links")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "SOURCE", "TARGET", "DIRECTION", "MAX_CONCURRENT")
		for _, l := range unmarshalList(data) {
			tbl.AddRow(str(l, "id"), str(l, "source_agent"), str(l, "target_agent"),
				str(l, "direction"), str(l, "max_concurrent"))
		}
		printer.Print(tbl)
		return nil
	},
}

var agentsLinksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a delegation link",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")
		direction, _ := cmd.Flags().GetString("direction")
		maxConc, _ := cmd.Flags().GetInt("max-concurrent")
		body := buildBody("source_agent", source, "target_agent", target,
			"direction", direction, "max_concurrent", maxConc)
		data, err := c.Post("/v1/agents/links", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Link created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var agentsLinksUpdateCmd = &cobra.Command{
	Use:   "update <linkID>",
	Short: "Update a delegation link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("direction") {
			v, _ := cmd.Flags().GetString("direction")
			body["direction"] = v
		}
		if cmd.Flags().Changed("max-concurrent") {
			v, _ := cmd.Flags().GetInt("max-concurrent")
			body["max_concurrent"] = v
		}
		_, err = c.Put("/v1/agents/links/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Link updated")
		return nil
	},
}

var agentsLinksDeleteCmd = &cobra.Command{
	Use:   "delete <linkID>",
	Short: "Delete a delegation link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this link?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/agents/links/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Link deleted")
		return nil
	},
}

// --- Agent Instances ---

var agentsInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage per-user agent instances",
}

var agentsInstancesListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List user instances for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/instances")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var agentsInstancesGetFileCmd = &cobra.Command{
	Use:   "get-file <agentID>",
	Short: "Get an instance context file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		file, _ := cmd.Flags().GetString("file")
		data, err := c.Get(fmt.Sprintf("/v1/agents/%s/instances/%s/files/%s", args[0], user, file))
		if err != nil {
			return err
		}
		var content struct {
			Content string `json:"content"`
		}
		_ = json.Unmarshal(data, &content)
		fmt.Println(content.Content)
		return nil
	},
}

var agentsInstancesSetFileCmd = &cobra.Command{
	Use:   "set-file <agentID>",
	Short: "Set an instance context file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		file, _ := cmd.Flags().GetString("file")
		contentVal, _ := cmd.Flags().GetString("content")
		content, err := readContent(contentVal)
		if err != nil {
			return err
		}
		_, err = c.Put(fmt.Sprintf("/v1/agents/%s/instances/%s/files/%s", args[0], user, file),
			map[string]any{"content": content})
		if err != nil {
			return err
		}
		printer.Success("File updated")
		return nil
	},
}

var agentsInstancesMetadataCmd = &cobra.Command{
	Use:   "metadata <agentID>",
	Short: "Get or patch instance metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		patch, _ := cmd.Flags().GetString("patch")
		if patch != "" {
			var body map[string]any
			if err := json.Unmarshal([]byte(patch), &body); err != nil {
				return fmt.Errorf("invalid JSON patch: %w", err)
			}
			_, err = c.Patch(fmt.Sprintf("/v1/agents/%s/instances/%s/metadata", args[0], user), body)
			if err != nil {
				return err
			}
			printer.Success("Metadata updated")
			return nil
		}
		// GET not directly available — show info via instances list
		printer.Success("Use --patch to update metadata")
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

	// Share flags
	agentsShareCmd.Flags().String("user", "", "User ID to share with")
	agentsShareCmd.Flags().String("role", "operator", "Role: admin, operator, viewer")
	_ = agentsShareCmd.MarkFlagRequired("user")
	agentsUnshareCmd.Flags().String("user", "", "User ID to revoke")
	_ = agentsUnshareCmd.MarkFlagRequired("user")

	// Link flags
	agentsLinksCreateCmd.Flags().String("source", "", "Source agent ID")
	agentsLinksCreateCmd.Flags().String("target", "", "Target agent ID")
	agentsLinksCreateCmd.Flags().String("direction", "outbound", "Direction: outbound, inbound, bidirectional")
	agentsLinksCreateCmd.Flags().Int("max-concurrent", 3, "Max concurrent delegations")
	agentsLinksUpdateCmd.Flags().String("direction", "", "Direction")
	agentsLinksUpdateCmd.Flags().Int("max-concurrent", 0, "Max concurrent")

	// Instance flags
	for _, cmd := range []*cobra.Command{agentsInstancesGetFileCmd, agentsInstancesSetFileCmd, agentsInstancesMetadataCmd} {
		cmd.Flags().String("user", "", "User ID")
		_ = cmd.MarkFlagRequired("user")
	}
	agentsInstancesGetFileCmd.Flags().String("file", "", "File name")
	_ = agentsInstancesGetFileCmd.MarkFlagRequired("file")
	agentsInstancesSetFileCmd.Flags().String("file", "", "File name")
	agentsInstancesSetFileCmd.Flags().String("content", "", "Content (or @filepath)")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("file")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("content")
	agentsInstancesMetadataCmd.Flags().String("patch", "", "JSON patch object")

	// Wire up subcommands
	agentsLinksCmd.AddCommand(agentsLinksListCmd, agentsLinksCreateCmd, agentsLinksUpdateCmd, agentsLinksDeleteCmd)
	agentsInstancesCmd.AddCommand(agentsInstancesListCmd, agentsInstancesGetFileCmd, agentsInstancesSetFileCmd, agentsInstancesMetadataCmd)
	agentsCmd.AddCommand(agentsListCmd, agentsGetCmd, agentsCreateCmd, agentsUpdateCmd, agentsDeleteCmd,
		agentsShareCmd, agentsUnshareCmd, agentsRegenerateCmd, agentsResummonCmd,
		agentsLinksCmd, agentsInstancesCmd)
	rootCmd.AddCommand(agentsCmd)
}
