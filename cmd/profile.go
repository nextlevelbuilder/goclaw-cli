package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage CLI connection profiles",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, active, err := config.ListProfiles()
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(map[string]any{"active": active, "profiles": profiles})
			return nil
		}
		tbl := output.NewTable("ACTIVE", "NAME", "SERVER", "OUTPUT")
		for _, p := range profiles {
			marker := ""
			if p.Name == active {
				marker = "*"
			}
			tbl.AddRow(marker, p.Name, p.Server, p.OutputFormat)
		}
		printer.Print(tbl)
		return nil
	},
}

var profileCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Print the active profile name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.OutputFormat != "table" {
			printer.Print(map[string]any{"profile": cfg.Profile})
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), cfg.Profile)
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.ValidateProfileName(name); err != nil {
			return err
		}
		if err := config.SetActiveProfile(name); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Switched to profile %q", name))
		return nil
	},
}

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.ValidateProfileName(name); err != nil {
			return err
		}
		server := cfg.Server
		outputFormat, _ := cmd.Flags().GetString("profile-output")
		copyFrom, _ := cmd.Flags().GetString("copy-from")

		if copyFrom != "" {
			if err := config.ValidateProfileName(copyFrom); err != nil {
				return err
			}
			profiles, _, err := config.ListProfiles()
			if err != nil {
				return err
			}
			found := false
			for _, p := range profiles {
				if p.Name == copyFrom {
					found = true
					if server == "" {
						server = p.Server
					}
					if outputFormat == "" {
						outputFormat = p.OutputFormat
					}
					break
				}
			}
			if !found {
				return &config.ProfileNotFoundError{Name: copyFrom}
			}
		}

		profile := config.Profile{Name: name, Server: server, OutputFormat: outputFormat}
		if err := config.Save(profile, false); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Created profile %q", name))
		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.ValidateProfileName(name); err != nil {
			return err
		}
		if name == cfg.Profile && !tui.Confirm(fmt.Sprintf("Delete active profile %s?", name), cfg.Yes) {
			return nil
		}
		if err := config.RemoveProfile(name); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Deleted profile %q", name))
		return nil
	},
}

func init() {
	profileCreateCmd.Flags().String("copy-from", "", "Copy server and output settings from an existing profile")
	profileCreateCmd.Flags().String("profile-output", "", "Default output format for this profile")

	profileCmd.AddCommand(profileListCmd, profileCurrentCmd, profileUseCmd, profileCreateCmd, profileDeleteCmd)
	rootCmd.AddCommand(profileCmd)
}
