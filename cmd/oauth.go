package cmd

import (
	"fmt"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// oauthCmd manages OAuth provider pool authentication (ChatGPT / OpenAI).
// Routes: /v1/auth/chatgpt/{provider}/... and /v1/auth/openai/...
var oauthCmd = &cobra.Command{Use: "oauth", Short: "Manage OAuth provider pool (chatgpt, openai)"}

// validOAuthProviders lists accepted --provider values.
// Extend when server adds new providers.
var validOAuthProviders = map[string]bool{
	"openai":   true,
	"chatgpt":  true,
	"claude":   true, // chatgpt sub-provider — keep if server supports
	"gemini":   true,
}

// validateProvider returns an error if provider is not supported.
func validateProvider(provider string) error {
	if provider == "" {
		return fmt.Errorf("--provider is required (one of: openai, chatgpt, claude, gemini)")
	}
	if !validOAuthProviders[provider] {
		return fmt.Errorf("unsupported --provider %q (valid: openai, chatgpt, claude, gemini)", provider)
	}
	return nil
}

// oauthPath returns the base URL path for the given provider.
// chatgpt uses /v1/auth/chatgpt/{provider}, openai uses /v1/auth/openai.
func oauthPath(provider, action string) string {
	switch provider {
	case "openai":
		return fmt.Sprintf("/v1/auth/openai/%s", action)
	default:
		// chatgpt and any other chatgpt sub-providers
		return fmt.Sprintf("/v1/auth/chatgpt/%s/%s", provider, action)
	}
}

var oauthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show OAuth provider authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		if err := validateProvider(provider); err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get(oauthPath(provider, "status"))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var oauthQuotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Show OAuth provider quota usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		if err := validateProvider(provider); err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get(oauthPath(provider, "quota"))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var oauthStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start OAuth flow — prints browser URL to stderr, auth_id to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		if err := validateProvider(provider); err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post(oauthPath(provider, "start"), nil)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		// Print auth URL to stderr so the user can open it in a browser.
		// auth_id goes to stdout for piping / automation.
		if authURL, ok := m["url"].(string); ok && authURL != "" {
			fmt.Fprintln(os.Stderr, "Open this URL in your browser:")
			fmt.Fprintln(os.Stderr, authURL)
		}
		if authID, ok := m["auth_id"].(string); ok && authID != "" {
			fmt.Println(authID)
		} else {
			printer.Print(m)
		}
		return nil
	},
}

var oauthCallbackCmd = &cobra.Command{
	Use:   "callback",
	Short: "Complete OAuth flow by submitting the auth code from the browser redirect",
	Long: `Complete OAuth authentication after the browser redirect.

After running 'oauth start', open the URL in a browser.
When the browser redirects to the callback page, copy the code= parameter
and pass it here via --code.

Example:
  goclaw oauth start --provider=chatgpt
  # Open printed URL → browser redirects → copy code from URL
  goclaw oauth callback --provider=chatgpt --code=<code>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, _ := cmd.Flags().GetString("provider")
		if err := validateProvider(provider); err != nil {
			return err
		}
		code, _ := cmd.Flags().GetString("code")
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post(oauthPath(provider, "callback"), map[string]any{"code": code})
		if err != nil {
			return err
		}
		printer.Success("OAuth authentication completed")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var oauthLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke OAuth provider token (requires --yes)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Revoke OAuth token for this provider?", cfg.Yes) {
			return nil
		}
		provider, _ := cmd.Flags().GetString("provider")
		if err := validateProvider(provider); err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post(oauthPath(provider, "logout"), nil)
		if err != nil {
			return err
		}
		printer.Success("OAuth token revoked")
		return nil
	},
}

func init() {
	providers := "chatgpt|openai"
	for _, c := range []*cobra.Command{
		oauthStatusCmd, oauthQuotaCmd, oauthStartCmd, oauthCallbackCmd, oauthLogoutCmd,
	} {
		c.Flags().String("provider", "chatgpt", "OAuth provider: "+providers)
		_ = c.MarkFlagRequired("provider")
	}

	oauthCallbackCmd.Flags().String("code", "", "Auth code from browser redirect URL (required)")
	_ = oauthCallbackCmd.MarkFlagRequired("code")

	oauthCmd.AddCommand(oauthStatusCmd, oauthQuotaCmd, oauthStartCmd, oauthCallbackCmd, oauthLogoutCmd)
	rootCmd.AddCommand(oauthCmd)
}
