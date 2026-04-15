package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// skills_tenant_config.go adds tenant-config subcommands to skillsCmd.
// Tenant config allows per-tenant customization of skill settings without
// modifying the skill itself (admin-scope, PUT /v1/skills/{id}/tenant-config).

var skillsTenantConfigCmd = &cobra.Command{
	Use:   "tenant-config",
	Short: "Manage per-tenant skill configuration overrides",
}

var skillsTenantConfigSetCmd = &cobra.Command{
	Use:   "set <skillID>",
	Short: "Set tenant config for a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configJSON, _ := cmd.Flags().GetString("config")
		if configJSON == "" {
			return fmt.Errorf("--config is required (JSON object)")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(configJSON), &body); err != nil {
			return fmt.Errorf("invalid --config JSON: %w", err)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Put("/v1/skills/"+args[0]+"/tenant-config", body)
		if err != nil {
			return err
		}
		printer.Success("Tenant config updated")
		return nil
	},
}

var skillsTenantConfigDeleteCmd = &cobra.Command{
	Use:   "delete <skillID>",
	Short: "Delete tenant config for a skill (reverts to defaults) — requires --yes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete tenant config for skill "+args[0]+"?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/skills/" + args[0] + "/tenant-config")
		if err != nil {
			return err
		}
		printer.Success("Tenant config deleted")
		return nil
	},
}

func init() {
	skillsTenantConfigSetCmd.Flags().String("config", "", "Tenant config as JSON object (required)")
	_ = skillsTenantConfigSetCmd.MarkFlagRequired("config")

	skillsTenantConfigCmd.AddCommand(skillsTenantConfigSetCmd, skillsTenantConfigDeleteCmd)
	skillsCmd.AddCommand(skillsTenantConfigCmd)
}
