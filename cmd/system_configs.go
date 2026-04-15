package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var systemConfigsCmd = &cobra.Command{
	Use:   "system-configs",
	Short: "Manage system configuration key-value pairs",
	Long:  "Read and write system-level configuration entries. Set/delete operations require admin role.",
}

var systemConfigsListCmd = &cobra.Command{
	Use: "list", Short: "List all system config entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system-configs")
		if err != nil {
			return err
		}
		// Response is a JSON object (map of key→value) or list depending on server impl.
		// Attempt map first, fall back to raw print.
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			if m != nil {
				printer.Print(m)
			} else {
				printer.Print(unmarshalList(data))
			}
			return nil
		}
		tbl := output.NewTable("KEY", "VALUE")
		for k, v := range m {
			tbl.AddRow(k, fmt.Sprintf("%v", v))
		}
		printer.Print(tbl)
		return nil
	},
}

var systemConfigsGetCmd = &cobra.Command{
	Use: "get <key>", Short: "Get a system config value by key", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system-configs/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var systemConfigsSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a system config value",
	Long: `Set a system config value. Use --json to parse value as JSON.

Example:
  goclaw system-configs set agent.default_model gpt-4o
  goclaw system-configs set feature.flags '{"beta":true}' --json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		key, rawVal := args[0], args[1]
		asJSON, _ := cmd.Flags().GetBool("json")

		var body map[string]any
		if asJSON {
			// Validate that the value is valid JSON; send as parsed object.
			var parsed any
			if err := json.Unmarshal([]byte(rawVal), &parsed); err != nil {
				return fmt.Errorf("--json flag set but value is not valid JSON: %w", err)
			}
			body = map[string]any{"value": parsed}
		} else {
			body = map[string]any{"value": rawVal}
		}

		data, err := c.Put("/v1/system-configs/"+key, body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var systemConfigsDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a system config entry",
	Long: `Delete a system config key. Requires --yes to confirm.

Example:
  goclaw system-configs delete feature.beta --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete system config key %q?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/system-configs/" + args[0])
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Config key %q deleted", args[0]))
		return nil
	},
}

func init() {
	systemConfigsSetCmd.Flags().Bool("json", false, "Parse <value> as JSON")
	// --yes is a root persistent flag; no need to redefine here.

	systemConfigsCmd.AddCommand(
		systemConfigsListCmd, systemConfigsGetCmd,
		systemConfigsSetCmd, systemConfigsDeleteCmd,
	)
	rootCmd.AddCommand(systemConfigsCmd)
}
