package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// admin_credentials.go owns the credentials subcommand tree under adminCmd.
// Extracted from admin.go to keep file sizes under 200 LoC.
// Covers: list, create, delete (existing) + update, test, presets, check-binary (new).
// User-credentials subtree → admin_credentials_users.go
// Agent-grants subtree    → admin_credentials_grants.go

var adminCredentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage CLI credentials store",
}

var adminCredentialsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored credentials",
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

var adminCredentialsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a CLI credential",
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

var adminCredentialsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a CLI credential",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bodyJSON, _ := cmd.Flags().GetString("body")
		if bodyJSON == "" {
			return fmt.Errorf("--body is required (JSON object)")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
			return fmt.Errorf("invalid --body JSON: %w", err)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Put("/v1/cli-credentials/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Credential updated")
		return nil
	},
}

var adminCredentialsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a CLI credential (requires --yes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this credential?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/cli-credentials/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Credential deleted")
		return nil
	},
}

var adminCredentialsTestCmd = &cobra.Command{
	Use:   "test <id>",
	Short: "Dry-run test a CLI credential (executes binary, no side-effects)",
	Args:  cobra.ExactArgs(1),
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

var adminCredentialsPresetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available credential presets",
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

var adminCredentialsCheckBinaryCmd = &cobra.Command{
	Use:   "check-binary",
	Short: "Verify a CLI binary is accessible on the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		bodyJSON, _ := cmd.Flags().GetString("body")
		var body map[string]any
		if bodyJSON != "" {
			if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
				return fmt.Errorf("invalid --body JSON: %w", err)
			}
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/cli-credentials/check-binary", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	adminCredentialsCreateCmd.Flags().String("name", "", "Credential name (required)")
	_ = adminCredentialsCreateCmd.MarkFlagRequired("name")

	adminCredentialsUpdateCmd.Flags().String("body", "", "Update payload as JSON object (required)")
	adminCredentialsCheckBinaryCmd.Flags().String("body", "", "Check payload as JSON object")

	adminCredentialsCmd.AddCommand(
		adminCredentialsListCmd,
		adminCredentialsCreateCmd,
		adminCredentialsUpdateCmd,
		adminCredentialsDeleteCmd,
		adminCredentialsTestCmd,
		adminCredentialsPresetsCmd,
		adminCredentialsCheckBinaryCmd,
	)
}
