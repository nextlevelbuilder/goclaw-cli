package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// providersReconnectCmd POSTs an empty body to /v1/providers/{id}/reconnect.
// Backend verifies reconnect server-side — do NOT add a --verify flag.
var providersReconnectCmd = &cobra.Command{
	Use:   "reconnect <provider-id>",
	Short: "Force-reconnect a registered provider (admin-only)",
	Long: `Force-reconnect a registered provider. Admin-only on the server.

The server handles reconnect verification internally; no client-side --verify
flag is exposed. (Note: ` + "`providers verify-embedding`" + ` is a different
command targeting a different endpoint.)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/providers/"+url.PathEscape(args[0])+"/reconnect", nil)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		tbl := output.NewTable("STATUS", "REGISTRY_UPDATED", "CACHE_INVALIDATED", "PROVIDER")
		var providerLabel string
		if p, ok := m["provider"].(map[string]any); ok {
			providerLabel = str(p, "name")
			if providerLabel == "" {
				providerLabel = str(p, "id")
			}
		}
		tbl.AddRow(str(m, "status"), str(m, "registry_updated"), str(m, "cache_invalidated"), providerLabel)
		printer.Print(tbl)
		return nil
	},
}

func init() {
	providersCmd.AddCommand(providersReconnectCmd)
}
