package cmd

import (
	"fmt"
	"net/url"

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
		data, err := c.Get("/v1/mcp/servers/" + url.PathEscape(args[0]))
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
		_, err = c.Put("/v1/mcp/servers/"+url.PathEscape(args[0]), body)
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
		_, err = c.Delete("/v1/mcp/servers/" + url.PathEscape(args[0]))
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
		data, err := c.Post("/v1/mcp/servers/"+url.PathEscape(args[0])+"/test", nil)
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
		data, err := c.Get("/v1/mcp/servers/" + url.PathEscape(args[0]) + "/tools")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

// --- MCP Grants ---

var mcpGrantsCmd = &cobra.Command{Use: "grants", Short: "Manage MCP access grants"}

var mcpGrantsListCmd = &cobra.Command{
	Use: "list", Short: "List grants for an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		agent, _ := cmd.Flags().GetString("agent")
		data, err := c.Get("/v1/mcp/grants/agent/" + agent)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var mcpGrantsGrantCmd = &cobra.Command{
	Use: "grant", Short: "Grant MCP server access",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		server, _ := cmd.Flags().GetString("server")
		agent, _ := cmd.Flags().GetString("agent")
		user, _ := cmd.Flags().GetString("user")
		if agent != "" {
			_, err = c.Post(fmt.Sprintf("/v1/mcp/servers/%s/grants/agent/%s", url.PathEscape(server), url.PathEscape(agent)), nil)
		} else if user != "" {
			_, err = c.Post(fmt.Sprintf("/v1/mcp/servers/%s/grants/user/%s", url.PathEscape(server), url.PathEscape(user)), nil)
		} else {
			return fmt.Errorf("specify --agent or --user")
		}
		if err != nil {
			return err
		}
		printer.Success("Access granted")
		return nil
	},
}

var mcpGrantsRevokeCmd = &cobra.Command{
	Use: "revoke", Short: "Revoke MCP server access",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		server, _ := cmd.Flags().GetString("server")
		agent, _ := cmd.Flags().GetString("agent")
		user, _ := cmd.Flags().GetString("user")
		if agent != "" {
			_, err = c.Delete(fmt.Sprintf("/v1/mcp/servers/%s/grants/agent/%s", url.PathEscape(server), url.PathEscape(agent)))
		} else if user != "" {
			_, err = c.Delete(fmt.Sprintf("/v1/mcp/servers/%s/grants/user/%s", url.PathEscape(server), url.PathEscape(user)))
		} else {
			return fmt.Errorf("specify --agent or --user")
		}
		if err != nil {
			return err
		}
		printer.Success("Access revoked")
		return nil
	},
}

// --- MCP Access Requests ---

var mcpRequestsCmd = &cobra.Command{Use: "requests", Short: "Manage MCP access requests"}

var mcpRequestsListCmd = &cobra.Command{
	Use: "list", Short: "List access requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			q.Set("status", v)
		}
		path := "/v1/mcp/requests"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var mcpRequestsCreateCmd = &cobra.Command{
	Use: "create", Short: "Request access to an MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		server, _ := cmd.Flags().GetString("server")
		reason, _ := cmd.Flags().GetString("reason")
		_, err = c.Post("/v1/mcp/requests", buildBody("server_id", server, "reason", reason))
		if err != nil {
			return err
		}
		printer.Success("Access request submitted")
		return nil
	},
}

var mcpRequestsReviewCmd = &cobra.Command{
	Use: "review <id>", Short: "Approve or reject access request", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		action, _ := cmd.Flags().GetString("action")
		_, err = c.Post("/v1/mcp/requests/"+url.PathEscape(args[0])+"/review", map[string]any{"action": action})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Request %sd", action))
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

	// Grant flags
	for _, c := range []*cobra.Command{mcpGrantsListCmd, mcpGrantsGrantCmd, mcpGrantsRevokeCmd} {
		c.Flags().String("agent", "", "Agent ID")
		c.Flags().String("user", "", "User ID")
	}
	for _, c := range []*cobra.Command{mcpGrantsGrantCmd, mcpGrantsRevokeCmd} {
		c.Flags().String("server", "", "Server ID")
		_ = c.MarkFlagRequired("server")
	}

	// Request flags
	mcpRequestsListCmd.Flags().String("status", "", "Filter: pending, approved, rejected")
	mcpRequestsCreateCmd.Flags().String("server", "", "Server ID")
	mcpRequestsCreateCmd.Flags().String("reason", "", "Request reason")
	_ = mcpRequestsCreateCmd.MarkFlagRequired("server")
	mcpRequestsReviewCmd.Flags().String("action", "", "Action: approve or reject")
	_ = mcpRequestsReviewCmd.MarkFlagRequired("action")

	mcpServersCmd.AddCommand(mcpServersListCmd, mcpServersGetCmd, mcpServersCreateCmd,
		mcpServersUpdateCmd, mcpServersDeleteCmd, mcpServersTestCmd, mcpServersToolsCmd)
	mcpGrantsCmd.AddCommand(mcpGrantsListCmd, mcpGrantsGrantCmd, mcpGrantsRevokeCmd)
	mcpRequestsCmd.AddCommand(mcpRequestsListCmd, mcpRequestsCreateCmd, mcpRequestsReviewCmd)
	mcpCmd.AddCommand(mcpServersCmd, mcpGrantsCmd, mcpRequestsCmd)
	rootCmd.AddCommand(mcpCmd)
}
