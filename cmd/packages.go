package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// packagesCmd manages skill runtime packages (Python/Node via uv/npm).
var packagesCmd = &cobra.Command{Use: "packages", Short: "Manage skill runtime packages"}

var packagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/packages")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(rawPayload(data))
			return nil
		}
		tbl := output.NewTable("SOURCE", "NAME", "VERSION", "DETAIL")
		addInstalledPackageRows(tbl, data)
		printer.Print(tbl)
		return nil
	},
}

var packagesInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a package (admin)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		runtime, _ := cmd.Flags().GetString("runtime")
		spec, err := packageSpecFromRuntime(args[0], runtime)
		if err != nil {
			return err
		}
		body := buildBody("package", spec)
		data, err := c.Post("/v1/packages/install", body)
		if err != nil {
			return err
		}
		printer.Success("Package install initiated")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var packagesUninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Uninstall a package — affects shared runtimes (requires --yes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Uninstall package "+args[0]+"? This affects shared runtimes.", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		runtime, _ := cmd.Flags().GetString("runtime")
		spec, err := packageSpecFromRuntime(args[0], runtime)
		if err != nil {
			return err
		}
		body := buildBody("package", spec)
		data, err := c.Post("/v1/packages/uninstall", body)
		if err != nil {
			return err
		}
		printer.Success("Package uninstall initiated")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var packagesRuntimesCmd = &cobra.Command{
	Use:   "runtimes",
	Short: "List available skill runtimes (python, node, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/packages/runtimes")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(rawPayload(data))
			return nil
		}
		tbl := output.NewTable("NAME", "AVAILABLE", "VERSION", "READY")
		addRuntimeRows(tbl, data)
		printer.Print(tbl)
		return nil
	},
}

var packagesDenyGroupsCmd = &cobra.Command{
	Use:   "deny-groups",
	Short: "List shell deny groups (blocked command groups)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/shell-deny-groups")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(rawPayload(data))
			return nil
		}
		tbl := output.NewTable("NAME", "DESCRIPTION", "DEFAULT")
		addDenyGroupRows(tbl, listFromResponse(data, "groups"))
		printer.Print(tbl)
		return nil
	},
}

var packagesGitHubReleasesCmd = &cobra.Command{
	Use:   "github-releases",
	Short: "List GitHub releases for tracked packages (GET /v1/packages/github-releases)",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, _ := cmd.Flags().GetString("repo")
		limit, _ := cmd.Flags().GetInt("limit")
		path, err := githubReleasesPath(repo, limit)
		if err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(rawPayload(data))
			return nil
		}
		tbl := output.NewTable("TAG", "NAME", "PRERELEASE", "ASSETS")
		addGitHubReleaseRows(tbl, listFromResponse(data, "releases"))
		printer.Print(tbl)
		return nil
	},
}

func init() {
	packagesInstallCmd.Flags().String("runtime", "", "Target runtime: python, node")
	packagesUninstallCmd.Flags().String("runtime", "", "Target runtime: python, node")
	packagesGitHubReleasesCmd.Flags().String("repo", "", "GitHub repository in owner/name format")
	packagesGitHubReleasesCmd.Flags().Int("limit", 10, "Maximum releases to fetch")

	packagesCmd.AddCommand(packagesListCmd, packagesInstallCmd, packagesUninstallCmd,
		packagesRuntimesCmd, packagesDenyGroupsCmd, packagesGitHubReleasesCmd)
	rootCmd.AddCommand(packagesCmd)
}
