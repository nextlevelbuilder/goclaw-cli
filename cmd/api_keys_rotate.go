package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var apiKeysRotateCmd = &cobra.Command{
	Use:   "rotate <id>",
	Short: "Create a replacement API key and revoke the old key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Rotate this API key?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		scopesRaw, _ := cmd.Flags().GetString("scopes")
		expiresIn, _ := cmd.Flags().GetInt("expires-in")
		scopes, err := parseAPIKeyScopes(scopesRaw)
		if err != nil {
			return err
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
		printAPIKeyRotateResult(args[0], result)

		_, err = c.Post("/v1/api-keys/"+url.PathEscape(args[0])+"/revoke", nil)
		if err != nil {
			return apiKeyRotatePartialError(args[0], result, err)
		}
		return nil
	},
}

func printAPIKeyRotateResult(oldKeyID string, result map[string]any) {
	if cfg.OutputFormat == "table" {
		fmt.Printf("API key rotated: %s\n", str(result, "id"))
		fmt.Println("--- IMPORTANT: Copy your replacement API key now. It will not be shown again. ---")
		fmt.Printf("Key: %s\n", str(result, "key"))
		return
	}
	result["old_key_id"] = oldKeyID
	result["old_revoke_status"] = "pending"
	printer.Print(result)
}

func apiKeyRotatePartialError(oldKeyID string, result map[string]any, revokeErr error) error {
	details := map[string]any{
		"new_key_id":        str(result, "id"),
		"old_key_id":        oldKeyID,
		"old_revoke_status": "failed",
		"revoke_error":      revokeErr.Error(),
	}
	return &output.ErrorDetail{
		Code:    "INTERNAL",
		Message: "replacement API key was created, but revoking the old key failed",
		Details: details,
	}
}

func init() {
	apiKeysRotateCmd.Flags().String("name", "", "Human-readable replacement key name")
	_ = apiKeysRotateCmd.MarkFlagRequired("name")
	apiKeysRotateCmd.Flags().String("scopes", "", "Comma-separated replacement key scopes")
	_ = apiKeysRotateCmd.MarkFlagRequired("scopes")
	apiKeysRotateCmd.Flags().Int("expires-in", 0, "TTL in seconds (0 = no expiry)")
	apiKeysCmd.AddCommand(apiKeysRotateCmd)
}
