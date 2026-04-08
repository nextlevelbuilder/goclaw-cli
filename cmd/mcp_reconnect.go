package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var mcpServersReconnectCmd = &cobra.Command{
	Use: "reconnect <id>", Short: "Reconnect MCP server", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/mcp/servers/"+url.PathEscape(args[0])+"/reconnect", nil)
		if err != nil {
			return err
		}
		printer.Success("MCP server reconnect triggered")
		return nil
	},
}

func init() {
	mcpServersCmd.AddCommand(mcpServersReconnectCmd)
}
