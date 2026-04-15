package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var teamsExportCmd = &cobra.Command{
	Use:   "export <id>",
	Short: "Export a team as a portable archive",
	Long: `Export a team's configuration, tasks, and member assignments as a .tar.gz archive.

Writes to -o file if specified, otherwise streams to stdout.

Examples:
  goclaw teams export team-123 -o team-123.tar.gz
  goclaw teams export team-123 > team-123.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("file")

		resp, err := c.GetRaw("/v1/teams/" + args[0] + "/export")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("export failed [%d]", resp.StatusCode)
		}

		if outFile != "" {
			if err := writeToFile(outFile, resp.Body); err != nil {
				return err
			}
			printer.Success(fmt.Sprintf("Team exported to %s", outFile))
			return nil
		}
		_, err = copyProgress(os.Stdout, resp)
		return err
	},
}

var teamsImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import a team from an archive (preview by default)",
	Long: `Import a team from a .tar.gz archive.

By default runs in PREVIEW mode — shows what would be imported without making changes.
Use --apply to perform the actual import.

Examples:
  goclaw teams import team.tar.gz           # preview only
  goclaw teams import team.tar.gz --apply   # perform import`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		c, err := newHTTP()
		if err != nil {
			return err
		}

		endpoint := "/v1/teams/import/preview"
		if apply {
			endpoint = "/v1/teams/import"
		}

		resp, err := c.UploadFile(endpoint, "file", args[0])
		if err != nil {
			return fmt.Errorf("upload: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("import failed [%d]", resp.StatusCode)
		}

		if apply {
			printer.Success("Team imported successfully")
		} else {
			printer.Success("Preview complete (use --apply to perform import)")
		}
		return nil
	},
}

func init() {
	teamsExportCmd.Flags().String("file", "", "Output file path (default: stdout)")
	teamsImportCmd.Flags().Bool("apply", false, "Perform the actual import (default: preview only)")

	teamsCmd.AddCommand(teamsExportCmd, teamsImportCmd)
}
