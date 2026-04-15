package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// mcp_servers.go extends mcpServersCmd with reconnect and test-connection subcommands.
// reconnect triggers a live reconnection of an existing server.
// test-connection validates a config before creating a server (dry-run).

var mcpServersReconnectCmd = &cobra.Command{
	Use:   "reconnect <id>",
	Short: "Reconnect an existing MCP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/mcp/servers/"+args[0]+"/reconnect", nil)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Reconnect triggered for server %s", args[0]))
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var mcpServersTestConnectionCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test an MCP server config before creating it (dry-run)",
	Long: `Test an MCP server configuration without persisting it.

Provide the same JSON body you would pass to 'mcp servers create' via --config.
The server attempts to connect and returns tool listing or an error.

Example:
  goclaw mcp servers test-connection \
    --config '{"transport":"stdio","command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","/tmp"]}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configJSON, _ := cmd.Flags().GetString("config")
		if configJSON == "" {
			return fmt.Errorf("--config is required (JSON object)")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(configJSON), &body); err != nil {
			return fmt.Errorf("invalid --config JSON: %w", err)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/mcp/servers/test", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	mcpServersTestConnectionCmd.Flags().String("config", "", "Server config as JSON object (required)")
	_ = mcpServersTestConnectionCmd.MarkFlagRequired("config")

	// Register into existing mcpServersCmd (defined in mcp.go).
	mcpServersCmd.AddCommand(mcpServersReconnectCmd, mcpServersTestConnectionCmd)
}
