package cmd

import (
	"github.com/spf13/cobra"
)

var toolsCmd = &cobra.Command{Use: "tools", Short: "Manage custom and built-in tools"}

// toolsCustomCmd, toolsInvokeCmd assembled in tools_custom.go.

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
		data, err := c.Get("/v1/tools/builtin/" + args[0])
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
		_, err = c.Put("/v1/tools/builtin/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Built-in tool updated")
		return nil
	},
}

func init() {
	// toolsCustomCmd assembled in tools_custom.go init().
	// Built-in tool flags
	toolsBuiltinUpdateCmd.Flags().Bool("enabled", true, "Enable/disable")

	toolsBuiltinCmd.AddCommand(toolsBuiltinListCmd, toolsBuiltinGetCmd, toolsBuiltinUpdateCmd)
	// toolsInvokeCmd defined in tools_custom.go
	toolsCmd.AddCommand(toolsCustomCmd, toolsBuiltinCmd, toolsInvokeCmd)
	rootCmd.AddCommand(toolsCmd)
}
