package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

var toolsCmd = &cobra.Command{Use: "tools", Short: "Manage built-in tools"}

// --- Built-in Tools ---

var toolsBuiltinCmd = &cobra.Command{Use: "builtin", Short: "Manage built-in tools"}

var toolsBuiltinListCmd = &cobra.Command{
	Use: "list", Short: "List built-in tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tools/builtin")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var toolsBuiltinGetCmd = &cobra.Command{
	Use: "get <name>", Short: "Get built-in tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tools/builtin/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var toolsBuiltinUpdateCmd = &cobra.Command{
	Use: "update <name>", Short: "Update built-in tool settings", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = v
		}
		_, err = c.Put("/v1/tools/builtin/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		printer.Success("Built-in tool updated")
		return nil
	},
}

// --- Builtin Tenant Config ---

var toolsBuiltinTenantConfigCmd = &cobra.Command{Use: "tenant-config", Short: "Manage tenant config for built-in tool"}

var toolsBuiltinTenantConfigSetCmd = &cobra.Command{
	Use: "set <name>", Short: "Set tenant config for built-in tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		enabled, _ := cmd.Flags().GetBool("enabled")
		_, err = c.Put("/v1/tools/builtin/"+url.PathEscape(args[0])+"/tenant-config",
			map[string]any{"enabled": enabled})
		if err != nil {
			return err
		}
		printer.Success("Tenant config updated")
		return nil
	},
}

var toolsBuiltinTenantConfigDeleteCmd = &cobra.Command{
	Use: "delete <name>", Short: "Delete tenant config for built-in tool", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/tools/builtin/" + url.PathEscape(args[0]) + "/tenant-config")
		if err != nil {
			return err
		}
		printer.Success("Tenant config deleted")
		return nil
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
		paramPairs, _ := cmd.Flags().GetStringSlice("param")
		paramsJSON, _ := cmd.Flags().GetString("params")

		params := make(map[string]any)
		if paramsJSON != "" {
			if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
				return fmt.Errorf("invalid --params JSON: %w", err)
			}
		}
		for _, pair := range paramPairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				params[parts[0]] = parts[1]
			}
		}

		body := map[string]any{"name": args[0], "parameters": params}
		data, err := c.Post("/v1/tools/invoke", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	// Built-in tool flags
	toolsBuiltinUpdateCmd.Flags().Bool("enabled", true, "Enable/disable")

	// Tenant config flags
	toolsBuiltinTenantConfigSetCmd.Flags().Bool("enabled", true, "Enable/disable for tenant")

	// Invoke flags
	toolsInvokeCmd.Flags().StringSlice("param", nil, "Parameter key=value pairs")
	toolsInvokeCmd.Flags().String("params", "", "Parameters as JSON object")

	toolsBuiltinTenantConfigCmd.AddCommand(toolsBuiltinTenantConfigSetCmd, toolsBuiltinTenantConfigDeleteCmd)
	toolsBuiltinCmd.AddCommand(toolsBuiltinListCmd, toolsBuiltinGetCmd, toolsBuiltinUpdateCmd,
		toolsBuiltinTenantConfigCmd)
	toolsCmd.AddCommand(toolsBuiltinCmd, toolsInvokeCmd)
	rootCmd.AddCommand(toolsCmd)
}
