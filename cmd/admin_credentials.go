package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"net/url"
	"github.com/spf13/cobra"
)

var credentialsCmd = &cobra.Command{Use: "credentials", Short: "Manage CLI credentials store"}

var credentialsListCmd = &cobra.Command{
	Use: "list", Short: "List stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "CREATED")
		for _, cr := range unmarshalList(data) {
			tbl.AddRow(str(cr, "id"), str(cr, "name"), str(cr, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var credentialsGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get credential details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var credentialsCreateCmd = &cobra.Command{
	Use: "create", Short: "Create CLI credential",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		data, err := c.Post("/v1/cli-credentials", map[string]any{"name": name})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var credentialsUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update CLI credential", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		_, err = c.Put("/v1/cli-credentials/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Credential updated")
		return nil
	},
}

var credentialsDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete CLI credential", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this credential?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/cli-credentials/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("Credential deleted")
		return nil
	},
}

var credentialsTestCmd = &cobra.Command{
	Use: "test <id>", Short: "Test CLI credential", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/cli-credentials/"+args[0]+"/test", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var credentialsPresetsCmd = &cobra.Command{
	Use: "presets", Short: "List credential presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials/presets")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	credentialsCreateCmd.Flags().String("name", "", "Credential name")
	_ = credentialsCreateCmd.MarkFlagRequired("name")
	credentialsUpdateCmd.Flags().String("name", "", "New credential name")

	credentialsCmd.AddCommand(credentialsListCmd, credentialsGetCmd, credentialsCreateCmd,
		credentialsUpdateCmd, credentialsDeleteCmd, credentialsTestCmd, credentialsPresetsCmd)
	rootCmd.AddCommand(credentialsCmd)
}
