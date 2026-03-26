package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var packagesCmd = &cobra.Command{Use: "packages", Short: "Manage runtime packages"}

var packagesListCmd = &cobra.Command{
	Use: "list", Short: "List installed packages",
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
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("NAME", "VERSION", "RUNTIME", "STATUS")
		for _, row := range unmarshalList(data) {
			tbl.AddRow(str(row, "name"), str(row, "version"), str(row, "runtime"), str(row, "status"))
		}
		printer.Print(tbl)
		return nil
	},
}

var packagesRuntimesCmd = &cobra.Command{
	Use: "runtimes", Short: "List available runtimes",
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
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("NAME", "VERSION", "AVAILABLE")
		for _, row := range unmarshalList(data) {
			tbl.AddRow(str(row, "name"), str(row, "version"), str(row, "available"))
		}
		printer.Print(tbl)
		return nil
	},
}

var packagesInstallCmd = &cobra.Command{
	Use: "install", Short: "Install a package",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		runtime, _ := cmd.Flags().GetString("runtime")
		_, err = c.Post("/v1/packages/install", buildBody("name", name, "runtime", runtime))
		if err != nil {
			return err
		}
		printer.Success("Package installation started")
		return nil
	},
}

var packagesUninstallCmd = &cobra.Command{
	Use: "uninstall", Short: "Uninstall a package",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		_, err = c.Post("/v1/packages/uninstall", buildBody("name", name))
		if err != nil {
			return err
		}
		printer.Success("Package uninstalled")
		return nil
	},
}

func init() {
	packagesInstallCmd.Flags().String("name", "", "Package name")
	packagesInstallCmd.Flags().String("runtime", "", "Target runtime")
	_ = packagesInstallCmd.MarkFlagRequired("name")

	packagesUninstallCmd.Flags().String("name", "", "Package name")
	_ = packagesUninstallCmd.MarkFlagRequired("name")

	packagesCmd.AddCommand(packagesListCmd, packagesRuntimesCmd, packagesInstallCmd, packagesUninstallCmd)
	rootCmd.AddCommand(packagesCmd)
}
