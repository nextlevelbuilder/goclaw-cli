package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// tools_custom.go holds the custom tools subcommand and tool invocation,
// extracted from tools.go to keep that file under 200 LoC.

var toolsCustomCmd = &cobra.Command{Use: "custom", Short: "Manage custom tools"}

var toolsCustomListCmd = &cobra.Command{
	Use: "list", Short: "List custom tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		return customToolsUnsupported()
	},
}

var toolsCustomGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get custom tool details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return customToolsUnsupported()
	},
}

var toolsCustomCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a custom tool",
	RunE: func(cmd *cobra.Command, args []string) error {
		return customToolsUnsupported()
	},
}

var toolsCustomUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update custom tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return customToolsUnsupported()
	},
}

var toolsCustomDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete custom tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return customToolsUnsupported()
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
		params, err := parseToolInvokeParams(cmd)
		if err != nil {
			return err
		}
		agentID, _ := cmd.Flags().GetString("agent")
		action, _ := cmd.Flags().GetString("action")
		sessionKey, _ := cmd.Flags().GetString("session")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		body := buildBody(
			"tool", args[0],
			"args", params,
			"agentId", agentID,
			"action", action,
			"sessionKey", sessionKey,
			"dryRun", dryRun,
		)
		data, err := c.Post("/v1/tools/invoke", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func customToolsUnsupported() error {
	return &output.ErrorDetail{
		Code:    "INVALID_REQUEST",
		Message: "custom tool management is not supported by this GoClaw server; use `tools builtin` or `tools invoke`",
	}
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
	toolsInvokeCmd.Flags().String("args", "", "Alias for --params; accepts literal JSON or @filepath")
	toolsInvokeCmd.Flags().String("agent", "", "Agent key or ID for tool context")
	toolsInvokeCmd.Flags().String("action", "", "Optional action to pass to the tool")
	toolsInvokeCmd.Flags().String("session", "", "Optional session key for tool context")
	toolsInvokeCmd.Flags().Bool("dry-run", false, "Validate tool and return schema without executing it")

	toolsCustomCmd.AddCommand(toolsCustomListCmd, toolsCustomGetCmd, toolsCustomCreateCmd,
		toolsCustomUpdateCmd, toolsCustomDeleteCmd)
}
