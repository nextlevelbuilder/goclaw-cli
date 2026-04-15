package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

// mcp_grants_requests.go holds MCP grants and access-request subcommands,
// extracted from mcp.go to keep that file under 200 LoC.

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
			_, err = c.Post(fmt.Sprintf("/v1/mcp/servers/%s/grants/agent/%s", server, agent), nil)
		} else if user != "" {
			_, err = c.Post(fmt.Sprintf("/v1/mcp/servers/%s/grants/user/%s", server, user), nil)
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
			_, err = c.Delete(fmt.Sprintf("/v1/mcp/servers/%s/grants/agent/%s", server, agent))
		} else if user != "" {
			_, err = c.Delete(fmt.Sprintf("/v1/mcp/servers/%s/grants/user/%s", server, user))
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
		_, err = c.Post("/v1/mcp/requests/"+args[0]+"/review", map[string]any{"action": action})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Request %sd", action))
		return nil
	},
}

func init() {
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

	mcpGrantsCmd.AddCommand(mcpGrantsListCmd, mcpGrantsGrantCmd, mcpGrantsRevokeCmd)
	mcpRequestsCmd.AddCommand(mcpRequestsListCmd, mcpRequestsCreateCmd, mcpRequestsReviewCmd)
}
