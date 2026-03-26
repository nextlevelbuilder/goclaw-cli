package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var skillsVersionsCmd = &cobra.Command{
	Use: "versions <id>", Short: "List skill versions", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/skills/" + url.PathEscape(args[0]) + "/versions")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var skillsRuntimesCmd = &cobra.Command{
	Use: "runtimes", Short: "List available skill runtimes",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/skills/runtimes")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var skillsFilesCmd = &cobra.Command{
	Use: "files <id>", Short: "Browse skill files", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		p, _ := cmd.Flags().GetString("path")
		if p == "" {
			p = "."
		}
		data, err := c.Get(fmt.Sprintf("/v1/skills/%s/files/%s", args[0], url.PathEscape(p)))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var skillsRescanDepsCmd = &cobra.Command{
	Use: "rescan-deps", Short: "Rescan skill dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		skillID, _ := cmd.Flags().GetString("skill-id")
		_, err = c.Post("/v1/skills/rescan-deps", buildBody("skill_id", skillID))
		if err != nil {
			return err
		}
		printer.Success("Dependencies rescanned")
		return nil
	},
}

var skillsInstallDepCmd = &cobra.Command{
	Use: "install-dep", Short: "Install a skill dependency",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		skillID, _ := cmd.Flags().GetString("skill-id")
		dep, _ := cmd.Flags().GetString("dep")
		_, err = c.Post("/v1/skills/install-dep", buildBody("skill_id", skillID, "dep", dep))
		if err != nil {
			return err
		}
		printer.Success("Dependency installed")
		return nil
	},
}

func init() {
	skillsFilesCmd.Flags().String("path", "", "Sub-path to browse")
	skillsRescanDepsCmd.Flags().String("skill-id", "", "Skill ID")
	skillsInstallDepCmd.Flags().String("skill-id", "", "Skill ID")
	skillsInstallDepCmd.Flags().String("dep", "", "Dependency name/spec")
	_ = skillsInstallDepCmd.MarkFlagRequired("skill-id")
	_ = skillsInstallDepCmd.MarkFlagRequired("dep")

	skillsCmd.AddCommand(skillsVersionsCmd, skillsRuntimesCmd, skillsFilesCmd,
		skillsRescanDepsCmd, skillsInstallDepCmd)
}
