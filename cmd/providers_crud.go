package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"net/url"
	"github.com/spf13/cobra"
)

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
		_, err = c.Delete("/v1/providers/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("Provider deleted")
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

var providersEmbeddingStatusCmd = &cobra.Command{
	Use: "embedding-status", Short: "Get embedding provider status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/embedding/status")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var providersClaudeAuthStatusCmd = &cobra.Command{
	Use: "claude-auth-status", Short: "Get Claude CLI auth status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/providers/claude-cli/auth-status")
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

	providersCmd.AddCommand(providersCreateCmd, providersUpdateCmd, providersDeleteCmd,
		providersVerifyCmd, providersEmbeddingStatusCmd, providersClaudeAuthStatusCmd)
}
