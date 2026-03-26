package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var sysConfigCmd = &cobra.Command{Use: "system-config", Short: "Per-tenant key-value configuration"}

var sysConfigListCmd = &cobra.Command{
	Use: "list", Short: "List all system config keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system-configs")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("KEY", "VALUE", "UPDATED")
		for _, item := range unmarshalList(data) {
			tbl.AddRow(str(item, "key"), str(item, "value"), str(item, "updated_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var sysConfigGetCmd = &cobra.Command{
	Use: "get <key>", Short: "Get a config value", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system-configs/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var sysConfigSetCmd = &cobra.Command{
	Use: "set <key>", Short: "Set a config value", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		value, _ := cmd.Flags().GetString("value")
		body := buildBody("value", value)
		data, err := c.Put("/v1/system-configs/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var sysConfigDeleteCmd = &cobra.Command{
	Use: "delete <key>", Short: "Delete a config key", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete config key?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/system-configs/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("Config key deleted")
		return nil
	},
}

func init() {
	sysConfigSetCmd.Flags().String("value", "", "Config value")
	_ = sysConfigSetCmd.MarkFlagRequired("value")

	sysConfigCmd.AddCommand(sysConfigListCmd, sysConfigGetCmd, sysConfigSetCmd, sysConfigDeleteCmd)
	rootCmd.AddCommand(sysConfigCmd)
}
