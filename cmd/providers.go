package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
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
		data, err := c.Get("/v1/providers/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var providersCreateCmd = &cobra.Command{
	Use: "create", Short: "Add a new LLM provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		displayName, _ := cmd.Flags().GetString("display-name")
		provType, _ := cmd.Flags().GetString("type")
		apiBase, _ := cmd.Flags().GetString("api-base")
		apiKey, _ := cmd.Flags().GetString("api-key")

		// Prompt for API key securely in interactive mode
		if apiKey == "" && tui.IsInteractive() {
			var promptErr error
			apiKey, promptErr = tui.Password("API Key")
			if promptErr != nil {
				return promptErr
			}
		}

		body := buildBody("name", name, "display_name", displayName,
			"provider_type", provType, "api_base", apiBase, "api_key", apiKey)
		data, err := c.Post("/v1/providers", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Provider created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var providersUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update provider", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		for _, f := range []string{"name", "display-name", "type", "api-base", "api-key"} {
			if cmd.Flags().Changed(f) {
				v, _ := cmd.Flags().GetString(f)
				key := f
				switch f {
				case "display-name":
					key = "display_name"
				case "type":
					key = "provider_type"
				case "api-base":
					key = "api_base"
				case "api-key":
					key = "api_key"
				}
				body[key] = v
			}
		}
		_, err = c.Put("/v1/providers/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Provider updated")
		return nil
	},
}

var providersDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete provider", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this provider?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/providers/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Provider deleted")
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
		data, err := c.Get("/v1/providers/" + args[0] + "/models")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var providersVerifyCmd = &cobra.Command{
	Use: "verify <id>", Short: "Verify provider API credentials", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/providers/"+args[0]+"/verify", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{providersCreateCmd, providersUpdateCmd} {
		c.Flags().String("name", "", "Provider name")
		c.Flags().String("display-name", "", "Display name")
		c.Flags().String("type", "openai_compat", "Provider type")
		c.Flags().String("api-base", "", "API base URL")
		c.Flags().String("api-key", "", "API key (prefer env GOCLAW_PROVIDER_API_KEY)")
	}

	providersCmd.AddCommand(providersListCmd, providersGetCmd, providersCreateCmd,
		providersUpdateCmd, providersDeleteCmd, providersModelsCmd, providersVerifyCmd)
	rootCmd.AddCommand(providersCmd)
}
