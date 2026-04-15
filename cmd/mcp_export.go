package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var mcpExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all MCP servers as a portable archive",
	Long: `Export all MCP server configurations and grants as a .tar.gz archive.

Writes to -o file if specified, otherwise streams to stdout.

Examples:
  goclaw mcp export -o mcp.tar.gz
  goclaw mcp export > mcp.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("file")

		resp, err := c.GetRaw("/v1/mcp/export")
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
			printer.Success(fmt.Sprintf("MCP servers exported to %s", outFile))
			return nil
		}
		_, err = copyProgress(os.Stdout, resp)
		return err
	},
}

var mcpImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import MCP servers from an archive (preview by default)",
	Long: `Import MCP server configurations from a .tar.gz archive.

By default runs in PREVIEW mode — shows what would be imported without making changes.
Use --apply to perform the actual import.

Examples:
  goclaw mcp import mcp.tar.gz           # preview only
  goclaw mcp import mcp.tar.gz --apply   # perform import`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		c, err := newHTTP()
		if err != nil {
			return err
		}

		endpoint := "/v1/mcp/import/preview"
		if apply {
			endpoint = "/v1/mcp/import"
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
			printer.Success("MCP servers imported successfully")
		} else {
			printer.Success("Preview complete (use --apply to perform import)")
		}
		return nil
	},
}

func init() {
	mcpExportCmd.Flags().String("file", "", "Output file path (default: stdout)")
	mcpImportCmd.Flags().Bool("apply", false, "Perform the actual import (default: preview only)")

	mcpCmd.AddCommand(mcpExportCmd, mcpImportCmd)
}
