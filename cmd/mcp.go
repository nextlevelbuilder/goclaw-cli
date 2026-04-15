package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{Use: "mcp", Short: "Manage MCP servers and grants"}

// --- MCP Servers ---

var mcpServersCmd = &cobra.Command{Use: "servers", Short: "Manage MCP servers"}

var mcpServersListCmd = &cobra.Command{
	Use: "list", Short: "List MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/mcp/servers")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "TRANSPORT", "ENABLED")
		for _, s := range unmarshalList(data) {
			tbl.AddRow(str(s, "id"), str(s, "name"), str(s, "transport"), str(s, "enabled"))
		}
		printer.Print(tbl)
		return nil
	},
}

var mcpServersGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get MCP server details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/mcp/servers/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var mcpServersCreateCmd = &cobra.Command{
	Use: "create", Short: "Register a new MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		transport, _ := cmd.Flags().GetString("transport")
		command, _ := cmd.Flags().GetString("command")
		mcpArgs, _ := cmd.Flags().GetStringSlice("args")
		mcpURL, _ := cmd.Flags().GetString("url")
		prefix, _ := cmd.Flags().GetString("prefix")
		timeout, _ := cmd.Flags().GetInt("timeout")

		body := buildBody("name", name, "transport", transport,
			"command", command, "url", mcpURL, "tool_prefix", prefix, "timeout_sec", timeout)
		if len(mcpArgs) > 0 {
			body["args"] = mcpArgs
		}
		data, err := c.Post("/v1/mcp/servers", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("MCP server created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var mcpServersUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update MCP server", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		for _, f := range []string{"name", "command", "url", "prefix"} {
			if cmd.Flags().Changed(f) {
				v, _ := cmd.Flags().GetString(f)
				key := f
				if f == "prefix" {
					key = "tool_prefix"
				}
				body[key] = v
			}
		}
		if cmd.Flags().Changed("timeout") {
			v, _ := cmd.Flags().GetInt("timeout")
			body["timeout_sec"] = v
		}
		_, err = c.Put("/v1/mcp/servers/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("MCP server updated")
		return nil
	},
}

var mcpServersDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete MCP server", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this MCP server?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/mcp/servers/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("MCP server deleted")
		return nil
	},
}

var mcpServersTestCmd = &cobra.Command{
	Use: "test <id>", Short: "Test MCP server connection", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/mcp/servers/"+args[0]+"/test", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var mcpServersToolsCmd = &cobra.Command{
	Use: "tools <id>", Short: "List tools from MCP server", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/mcp/servers/" + args[0] + "/tools")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	// Server flags
	for _, c := range []*cobra.Command{mcpServersCreateCmd, mcpServersUpdateCmd} {
		c.Flags().String("name", "", "Server name")
		c.Flags().String("transport", "stdio", "Transport: stdio, sse, streamable-http")
		c.Flags().String("command", "", "Command (stdio)")
		c.Flags().StringSlice("args", nil, "Command args (stdio)")
		c.Flags().String("url", "", "URL (sse/http)")
		c.Flags().String("prefix", "", "Tool prefix")
		c.Flags().Int("timeout", 60, "Timeout seconds")
	}

	// mcpGrantsCmd and mcpRequestsCmd assembled in mcp_grants_requests.go init().
	mcpServersCmd.AddCommand(mcpServersListCmd, mcpServersGetCmd, mcpServersCreateCmd,
		mcpServersUpdateCmd, mcpServersDeleteCmd, mcpServersTestCmd, mcpServersToolsCmd)
	mcpCmd.AddCommand(mcpServersCmd, mcpGrantsCmd, mcpRequestsCmd)
	rootCmd.AddCommand(mcpCmd)
}
