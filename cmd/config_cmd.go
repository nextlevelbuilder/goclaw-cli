package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
		params := map[string]any{}
		if v, _ := cmd.Flags().GetString("key"); v != "" {
			params["key"] = v
		}
		data, err := ws.Call("config.get", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
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
		var cfg map[string]any
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
		_, err = ws.Call("config.apply", cfg)
		if err != nil {
			return err
		}
		printer.Success("Configuration applied")
		return nil
	},
}

var configPatchCmd = &cobra.Command{
	Use: "patch", Short: "Patch a config key",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		// Try to parse value as JSON, fall back to string
		var parsedVal any
		if err := json.Unmarshal([]byte(value), &parsedVal); err != nil {
			parsedVal = value
		}

		_, err = ws.Call("config.patch", map[string]any{"key": key, "value": parsedVal})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Config %s updated", key))
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
		data, err := ws.Call("config.schema", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
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
	configGetCmd.Flags().String("key", "", "Config key path")
	configApplyCmd.Flags().String("file", "", "Config JSON file")
	_ = configApplyCmd.MarkFlagRequired("file")
	configPatchCmd.Flags().String("key", "", "Config key")
	configPatchCmd.Flags().String("value", "", "Config value")
	_ = configPatchCmd.MarkFlagRequired("key")
	_ = configPatchCmd.MarkFlagRequired("value")

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
