package cmd

import (
	"fmt"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{Use: "config", Short: "Manage server configuration"}

var configGetCmd = &cobra.Command{
	Use: "get", Short: "Get server configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		result, err := ws.ConfigGet()
		if err != nil {
			return err
		}
		printer.Print(result)
		return nil
	},
}

var configApplyCmd = &cobra.Command{
	Use: "apply", Short: "Apply configuration from file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		filePath, _ := cmd.Flags().GetString("file")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read config file: %w", err)
		}
		baseHash, _ := cmd.Flags().GetString("base-hash")
		_, err = ws.ConfigApply(client.ConfigApplyParams{
			Raw:      string(data),
			BaseHash: baseHash,
		})
		if err != nil {
			return err
		}
		printer.Success("Configuration applied")
		return nil
	},
}

var configPatchCmd = &cobra.Command{
	Use: "patch", Short: "Patch config with a partial JSON5 file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		filePath, _ := cmd.Flags().GetString("file")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read patch file: %w", err)
		}
		baseHash, _ := cmd.Flags().GetString("base-hash")
		_, err = ws.ConfigPatch(client.ConfigApplyParams{
			Raw:      string(data),
			BaseHash: baseHash,
		})
		if err != nil {
			return err
		}
		printer.Success("Configuration patched")
		return nil
	},
}

var configSchemaCmd = &cobra.Command{
	Use: "schema", Short: "Get config schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		result, err := ws.ConfigSchema()
		if err != nil {
			return err
		}
		printer.Print(result)
		return nil
	},
}

// --- Config Permissions ---

var configPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Manage config access permissions for agents",
}

var configPermissionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List config permissions for an agent",
	Long: `List config permissions granted to an agent.

Example:
  goclaw config permissions list --agent=my-agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		configType, _ := cmd.Flags().GetString("config-type")
		params := map[string]any{"agentId": agent}
		if configType != "" {
			params["configType"] = configType
		}
		data, err := ws.Call("config.permissions.list", params)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		printer.Print(toList(m["permissions"]))
		return nil
	},
}

var configPermissionsGrantCmd = &cobra.Command{
	Use:   "grant",
	Short: "Grant a config permission to a user for an agent",
	Long: `Grant config access permission.

Example:
  goclaw config permissions grant --agent=my-agent --user=user-123 \
    --scope=agent --config-type=system --permission=read`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		user, _ := cmd.Flags().GetString("user")
		scope, _ := cmd.Flags().GetString("scope")
		configType, _ := cmd.Flags().GetString("config-type")
		permission, _ := cmd.Flags().GetString("permission")
		params := map[string]any{
			"agentId":    agent,
			"userId":     user,
			"scope":      scope,
			"configType": configType,
			"permission": permission,
		}
		_, err = ws.Call("config.permissions.grant", params)
		if err != nil {
			return err
		}
		printer.Success("Permission granted")
		return nil
	},
}

var configPermissionsRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a config permission from a user for an agent",
	Long: `Revoke config access permission. Requires --yes to confirm.

Example:
  goclaw config permissions revoke --agent=my-agent --user=user-123 \
    --scope=agent --config-type=system --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		user, _ := cmd.Flags().GetString("user")
		scope, _ := cmd.Flags().GetString("scope")
		configType, _ := cmd.Flags().GetString("config-type")

		msg := fmt.Sprintf("Revoke %s permission on %s for user %s (agent %s)?", configType, scope, user, agent)
		if !tui.Confirm(msg, cfg.Yes) {
			return fmt.Errorf("cancelled")
		}

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		params := map[string]any{
			"agentId":    agent,
			"userId":     user,
			"scope":      scope,
			"configType": configType,
		}
		_, err = ws.Call("config.permissions.revoke", params)
		if err != nil {
			return err
		}
		printer.Success("Permission revoked")
		return nil
	},
}

func init() {
	configApplyCmd.Flags().String("file", "", "Config JSON5 file (full config)")
	_ = configApplyCmd.MarkFlagRequired("file")
	configApplyCmd.Flags().String("base-hash", "", "Base config hash for optimistic concurrency (from config get)")
	configPatchCmd.Flags().String("file", "", "Partial config JSON5 file to merge")
	_ = configPatchCmd.MarkFlagRequired("file")
	configPatchCmd.Flags().String("base-hash", "", "Base config hash for optimistic concurrency (from config get)")

	// Permissions list
	configPermissionsListCmd.Flags().String("agent", "", "Agent key or ID")
	configPermissionsListCmd.Flags().String("config-type", "", "Filter by config type")
	_ = configPermissionsListCmd.MarkFlagRequired("agent")

	// Permissions grant
	configPermissionsGrantCmd.Flags().String("agent", "", "Agent key or ID")
	configPermissionsGrantCmd.Flags().String("user", "", "User ID to grant permission to")
	configPermissionsGrantCmd.Flags().String("scope", "agent", "Permission scope")
	configPermissionsGrantCmd.Flags().String("config-type", "system", "Config type")
	configPermissionsGrantCmd.Flags().String("permission", "read", "Permission level: read, write")
	_ = configPermissionsGrantCmd.MarkFlagRequired("agent")
	_ = configPermissionsGrantCmd.MarkFlagRequired("user")

	// Permissions revoke
	configPermissionsRevokeCmd.Flags().String("agent", "", "Agent key or ID")
	configPermissionsRevokeCmd.Flags().String("user", "", "User ID to revoke permission from")
	configPermissionsRevokeCmd.Flags().String("scope", "agent", "Permission scope")
	configPermissionsRevokeCmd.Flags().String("config-type", "system", "Config type")
	_ = configPermissionsRevokeCmd.MarkFlagRequired("agent")
	_ = configPermissionsRevokeCmd.MarkFlagRequired("user")

	configPermissionsCmd.AddCommand(
		configPermissionsListCmd,
		configPermissionsGrantCmd,
		configPermissionsRevokeCmd,
	)
	configCmd.AddCommand(configGetCmd, configApplyCmd, configPatchCmd, configSchemaCmd, configPermissionsCmd)
	rootCmd.AddCommand(configCmd)
}
