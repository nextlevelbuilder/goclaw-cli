package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// tools_builtin_tenant.go adds tenant-config subcommands to toolsBuiltinCmd.
// Allows per-tenant overrides of built-in tool settings without modifying
// the global tool configuration (admin-scope).

var toolsBuiltinTenantConfigCmd = &cobra.Command{
	Use:   "tenant-config",
	Short: "Manage per-tenant config for a built-in tool",
}

var toolsBuiltinTenantConfigGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get tenant config for a built-in tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tools/builtin/" + args[0] + "/tenant-config")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var toolsBuiltinTenantConfigSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set tenant config for a built-in tool",
	Args:  cobra.ExactArgs(1),
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
		_, err = c.Put("/v1/tools/builtin/"+args[0]+"/tenant-config", body)
		if err != nil {
			return err
		}
		printer.Success("Tenant config updated")
		return nil
	},
}

var toolsBuiltinTenantConfigDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete tenant config for a built-in tool (reverts to defaults) — requires --yes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete tenant config for tool "+args[0]+"?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/tools/builtin/" + args[0] + "/tenant-config")
		if err != nil {
			return err
		}
		printer.Success("Tenant config deleted")
		return nil
	},
}

func init() {
	toolsBuiltinTenantConfigSetCmd.Flags().String("config", "", "Tenant config as JSON object (required)")
	_ = toolsBuiltinTenantConfigSetCmd.MarkFlagRequired("config")

	toolsBuiltinTenantConfigCmd.AddCommand(
		toolsBuiltinTenantConfigGetCmd,
		toolsBuiltinTenantConfigSetCmd,
		toolsBuiltinTenantConfigDeleteCmd,
	)
	// Register into existing toolsBuiltinCmd (defined in tools.go).
	toolsBuiltinCmd.AddCommand(toolsBuiltinTenantConfigCmd)
}
