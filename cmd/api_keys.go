package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var apiKeysCmd = &cobra.Command{Use: "api-keys", Short: "Manage API keys for scoped access"}

var apiKeysListCmd = &cobra.Command{
	Use: "list", Short: "List all API keys (masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/api-keys")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "PREFIX", "SCOPES", "EXPIRES", "LAST_USED", "REVOKED")
		for _, k := range unmarshalList(data) {
			scopes := ""
			if s, ok := k["scopes"].([]any); ok {
				parts := make([]string, len(s))
				for i, v := range s {
					parts[i] = fmt.Sprintf("%v", v)
				}
				scopes = strings.Join(parts, ",")
			}
			tbl.AddRow(
				str(k, "id"), str(k, "name"), str(k, "prefix"),
				scopes, str(k, "expires_at"), str(k, "last_used_at"),
				str(k, "revoked"),
			)
		}
		printer.Print(tbl)
		return nil
	},
}

var apiKeysCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a new API key (raw key shown once)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		scopesRaw, _ := cmd.Flags().GetString("scopes")
		expiresIn, _ := cmd.Flags().GetInt("expires-in")

		// Parse comma-separated scopes into slice
		var scopes []string
		for _, s := range strings.Split(scopesRaw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				scopes = append(scopes, s)
			}
		}
		if len(scopes) == 0 {
			return fmt.Errorf("at least one scope is required")
		}

		body := buildBody("name", name, "scopes", scopes)
		if expiresIn > 0 {
			body["expires_in"] = expiresIn
		}

		data, err := c.Post("/v1/api-keys", body)
		if err != nil {
			return err
		}

		result := unmarshalMap(data)

		// In table mode, highlight the show-once key
		if cfg.OutputFormat == "table" {
			fmt.Printf("API key created: %s\n", str(result, "id"))
			fmt.Printf("Name:    %s\n", str(result, "name"))
			fmt.Printf("Prefix:  %s\n", str(result, "prefix"))
			fmt.Printf("Scopes:  %v\n", result["scopes"])
			if v := str(result, "expires_at"); v != "" {
				fmt.Printf("Expires: %s\n", v)
			}
			fmt.Println()
			fmt.Println("--- IMPORTANT: Copy your API key now. It will not be shown again. ---")
			fmt.Printf("Key: %s\n", str(result, "key"))
			return nil
		}

		printer.Print(result)
		return nil
	},
}

var apiKeysRevokeCmd = &cobra.Command{
	Use: "revoke <id>", Short: "Revoke an API key", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Revoke this API key?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/api-keys/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("API key revoked")
		return nil
	},
}

func init() {
	apiKeysCreateCmd.Flags().String("name", "", "Human-readable key name")
	_ = apiKeysCreateCmd.MarkFlagRequired("name")
	apiKeysCreateCmd.Flags().String("scopes", "", "Comma-separated scopes (operator.admin,operator.read,operator.write,operator.approvals,operator.pairing)")
	_ = apiKeysCreateCmd.MarkFlagRequired("scopes")
	apiKeysCreateCmd.Flags().Int("expires-in", 0, "TTL in seconds (0 = no expiry)")

	apiKeysCmd.AddCommand(apiKeysListCmd, apiKeysCreateCmd, apiKeysRevokeCmd)
	rootCmd.AddCommand(apiKeysCmd)
}
