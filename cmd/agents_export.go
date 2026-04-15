package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// agentsExportCmd exports a single agent archive to file or stdout.
var agentsExportCmd = &cobra.Command{
	Use:   "export <id>",
	Short: "Export an agent as a portable archive",
	Long: `Export an agent configuration, context files, memory, and skills as a .tar.gz archive.

Writes to -o file if specified, otherwise streams to stdout.

Examples:
  goclaw agents export agent-123 -o agent-123.tar.gz
  goclaw agents export agent-123 > agent-123.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("file")

		resp, err := c.GetRaw("/v1/agents/"+args[0]+"/export")
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
			printer.Success(fmt.Sprintf("Agent exported to %s", outFile))
			return nil
		}
		// Stream to stdout — useful for piping.
		_, err = copyProgress(os.Stdout, resp)
		return err
	},
}

// agentsImportCmd uploads an agent archive (preview by default, --apply to commit).
var agentsImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import an agent from an archive (preview by default)",
	Long: `Import an agent from a .tar.gz archive.

By default runs in PREVIEW mode — shows what would be imported without making changes.
Use --apply to perform the actual import.

Examples:
  goclaw agents import agent.tar.gz            # preview only
  goclaw agents import agent.tar.gz --apply    # perform import`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		c, err := newHTTP()
		if err != nil {
			return err
		}

		endpoint := "/v1/agents/import/preview"
		if apply {
			endpoint = "/v1/agents/import"
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
			printer.Success("Agent imported successfully")
		} else {
			printer.Success("Preview complete (use --apply to perform import)")
		}
		return nil
	},
}

// agentsImportMergeCmd merges an agent archive into an existing agent.
var agentsImportMergeCmd = &cobra.Command{
	Use:   "import-merge <id> <file>",
	Short: "Merge an agent archive into an existing agent",
	Long: `Merge exported agent data into an existing agent without replacing it.

By default runs in PREVIEW mode. Use --apply to commit the merge.

Examples:
  goclaw agents import-merge agent-123 updates.tar.gz
  goclaw agents import-merge agent-123 updates.tar.gz --apply`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, filePath := args[0], args[1]
		apply, _ := cmd.Flags().GetBool("apply")

		c, err := newHTTP()
		if err != nil {
			return err
		}

		path := "/v1/agents/" + agentID + "/import"
		if !apply {
			path += "?preview=true"
		}

		resp, err := c.UploadFile(path, "file", filePath)
		if err != nil {
			return fmt.Errorf("upload: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("import-merge failed [%d]", resp.StatusCode)
		}

		if apply {
			printer.Success(fmt.Sprintf("Agent %s merge import complete", agentID))
		} else {
			printer.Success("Preview complete (use --apply to perform merge)")
		}
		return nil
	},
}

func init() {
	agentsExportCmd.Flags().String("file", "", "Output file path (default: stdout)")

	agentsImportCmd.Flags().Bool("apply", false, "Perform the actual import (default: preview only)")
	agentsImportMergeCmd.Flags().Bool("apply", false, "Perform the actual merge (default: preview only)")

	agentsCmd.AddCommand(agentsExportCmd, agentsImportCmd, agentsImportMergeCmd)
}
