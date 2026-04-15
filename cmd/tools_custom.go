package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// tools_custom.go holds the custom tools subcommand and tool invocation,
// extracted from tools.go to keep that file under 200 LoC.

var toolsCustomCmd = &cobra.Command{Use: "custom", Short: "Manage custom tools"}

var toolsCustomListCmd = &cobra.Command{
	Use: "list", Short: "List custom tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/tools/custom"
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			path += "?agent_id=" + url.QueryEscape(v)
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "DESCRIPTION", "ENABLED", "TIMEOUT")
		for _, t := range unmarshalList(data) {
			tbl.AddRow(str(t, "id"), str(t, "name"), str(t, "description"),
				str(t, "enabled"), str(t, "timeout_seconds"))
		}
		printer.Print(tbl)
		return nil
	},
}

var toolsCustomGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get custom tool details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tools/custom/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var toolsCustomCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a custom tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		desc, _ := cmd.Flags().GetString("description")
		command, _ := cmd.Flags().GetString("command")
		timeout, _ := cmd.Flags().GetInt("timeout")
		agent, _ := cmd.Flags().GetString("agent")
		paramsJSON, _ := cmd.Flags().GetString("parameters")
		body := buildBody("name", name, "description", desc,
			"command", command, "timeout_seconds", timeout, "agent_id", agent, "enabled", true)
		if paramsJSON != "" {
			var params any
			if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
				return fmt.Errorf("invalid parameters JSON: %w", err)
			}
			body["parameters"] = params
		}
		data, err := c.Post("/v1/tools/custom", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Tool created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var toolsCustomUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update custom tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		for _, f := range []string{"name", "description", "command"} {
			if cmd.Flags().Changed(f) {
				v, _ := cmd.Flags().GetString(f)
				body[f] = v
			}
		}
		if cmd.Flags().Changed("timeout") {
			v, _ := cmd.Flags().GetInt("timeout")
			body["timeout_seconds"] = v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = v
		}
		_, err = c.Put("/v1/tools/custom/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Tool updated")
		return nil
	},
}

var toolsCustomDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete custom tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this tool?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/tools/custom/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Tool deleted")
		return nil
	},
}

// --- Tool Invocation ---

var toolsInvokeCmd = &cobra.Command{
	Use:   "invoke <name>",
	Short: "Invoke a tool directly",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		paramPairs, _ := cmd.Flags().GetStringSlice("param")
		paramsJSON, _ := cmd.Flags().GetString("params")
		params := make(map[string]any)
		if paramsJSON != "" {
			if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
				return fmt.Errorf("invalid --params JSON: %w", err)
			}
		}
		for _, pair := range paramPairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				params[parts[0]] = parts[1]
			}
		}
		body := map[string]any{"name": args[0], "parameters": params}
		data, err := c.Post("/v1/tools/invoke", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	toolsCustomListCmd.Flags().String("agent", "", "Filter by agent ID")
	for _, c := range []*cobra.Command{toolsCustomCreateCmd, toolsCustomUpdateCmd} {
		c.Flags().String("name", "", "Tool name")
		c.Flags().String("description", "", "Tool description")
		c.Flags().String("command", "", "Shell command template")
		c.Flags().Int("timeout", 60, "Timeout seconds")
		c.Flags().String("agent", "", "Agent ID (empty=global)")
		c.Flags().String("parameters", "", "JSON Schema for parameters")
		c.Flags().Bool("enabled", true, "Enable tool")
	}
	toolsInvokeCmd.Flags().StringSlice("param", nil, "Parameter key=value pairs")
	toolsInvokeCmd.Flags().String("params", "", "Parameters as JSON object")

	toolsCustomCmd.AddCommand(toolsCustomListCmd, toolsCustomGetCmd, toolsCustomCreateCmd,
		toolsCustomUpdateCmd, toolsCustomDeleteCmd)
}
