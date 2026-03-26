package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var skillsTenantConfigCmd = &cobra.Command{Use: "tenant-config", Short: "Manage tenant config for a skill"}

var skillsTenantConfigSetCmd = &cobra.Command{
	Use:   "set <skill-id>",
	Short: "Set tenant config for a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		enabled, _ := cmd.Flags().GetBool("enabled")
		_, err = c.Put("/v1/skills/"+url.PathEscape(args[0])+"/tenant-config",
			map[string]any{"enabled": enabled})
		if err != nil {
			return err
		}
		printer.Success("Tenant config updated")
		return nil
	},
}

var skillsTenantConfigDeleteCmd = &cobra.Command{
	Use:   "delete <skill-id>",
	Short: "Delete tenant config for a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/skills/" + url.PathEscape(args[0]) + "/tenant-config")
		if err != nil {
			return err
		}
		printer.Success("Tenant config deleted")
		return nil
	},
}

func init() {
	skillsTenantConfigSetCmd.Flags().Bool("enabled", true, "Enable/disable skill for tenant")
	skillsTenantConfigCmd.AddCommand(skillsTenantConfigSetCmd, skillsTenantConfigDeleteCmd)
	skillsCmd.AddCommand(skillsTenantConfigCmd)
}
