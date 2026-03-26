package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"net/url"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{Use: "providers", Short: "Manage LLM providers"}

var providersListCmd = &cobra.Command{
	Use: "list", Short: "List providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/providers")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "DISPLAY_NAME", "TYPE", "ENABLED")
		for _, p := range unmarshalList(data) {
			tbl.AddRow(str(p, "id"), str(p, "name"), str(p, "display_name"),
				str(p, "provider_type"), str(p, "enabled"))
		}
		printer.Print(tbl)
		return nil
	},
}

var providersGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get provider details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/providers/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var providersModelsCmd = &cobra.Command{
	Use: "models <id>", Short: "List models from provider", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/providers/" + url.PathEscape(args[0]) + "/models")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	// create/update/delete/verify/status registered from providers_crud.go
	providersCmd.AddCommand(providersListCmd, providersGetCmd, providersModelsCmd)
	rootCmd.AddCommand(providersCmd)
}
