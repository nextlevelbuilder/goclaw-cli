package cmd

import (
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var packagesUpdatesCmd = &cobra.Command{Use: "updates", Short: "Manage runtime package updates"}

var packagesUpdatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached package updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/packages/updates")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var packagesUpdatesRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh package update cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/packages/updates/refresh", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var packagesUpdateApplyCmd = &cobra.Command{
	Use:   "apply <package>",
	Short: "Apply one package update",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		toVersion, _ := cmd.Flags().GetString("to-version")
		data, err := c.Post("/v1/packages/update", buildBody("package", args[0], "toVersion", toVersion))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var packagesUpdatesApplyAllCmd = &cobra.Command{
	Use:   "apply-all [packages...]",
	Short: "Apply all cached package updates",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		packagesRaw, _ := cmd.Flags().GetString("packages")
		body := map[string]any{}
		packages := splitCSV(packagesRaw)
		packages = append(packages, args...)
		if len(packages) > 0 {
			body["packages"] = packages
		}
		data, err := c.Post("/v1/packages/updates/apply-all", body)
		if err != nil {
			return err
		}
		result := unmarshalMap(data)
		allowPartial, _ := cmd.Flags().GetBool("allow-partial")
		if hasNonEmptyList(result, "failed") && !allowPartial {
			err := &client.APIError{
				Code:    "FAILED_PRECONDITION",
				Message: "one or more package updates failed; rerun with --allow-partial to accept partial success",
				Details: result,
			}
			if cfg.OutputFormat != "table" {
				return err
			}
			printer.Print(result)
			return err
		}
		printer.Print(result)
		return nil
	},
}

func splitCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func hasNonEmptyList(m map[string]any, key string) bool {
	if v, ok := m[key].([]any); ok {
		return len(v) > 0
	}
	return false
}

func init() {
	packagesUpdateApplyCmd.Flags().String("to-version", "", "Target version (default: latest cached version)")
	packagesUpdatesApplyAllCmd.Flags().String("packages", "", "Comma-separated package specs to apply")
	packagesUpdatesApplyAllCmd.Flags().Bool("allow-partial", false, "Exit 0 even when failed[] is non-empty")
	packagesUpdatesCmd.AddCommand(packagesUpdatesListCmd, packagesUpdatesRefreshCmd, packagesUpdateApplyCmd, packagesUpdatesApplyAllCmd)
	packagesCmd.AddCommand(packagesUpdatesCmd)
}
