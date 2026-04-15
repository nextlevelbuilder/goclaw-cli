package cmd

import (
	"github.com/spf13/cobra"
)

// skills_deps.go adds the single-dependency install command to skillsCmd.
// The plural install-deps (trigger full rescan+install) already exists in skills.go.
// install-dep installs one specific dependency without a full rescan.

var skillsInstallDepCmd = &cobra.Command{
	Use:   "install-dep <dep>",
	Short: "Install a single skill dependency (e.g. numpy, requests)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		runtime, _ := cmd.Flags().GetString("runtime")
		body := buildBody("dep", args[0], "runtime", runtime)
		data, err := c.Post("/v1/skills/install-dep", body)
		if err != nil {
			return err
		}
		printer.Success("Dependency install initiated")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	skillsInstallDepCmd.Flags().String("runtime", "", "Target runtime: python, node (default: auto-detect)")
	skillsCmd.AddCommand(skillsInstallDepCmd)
}
