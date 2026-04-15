package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// agentsInstancesCmd and subcommands — extracted from agents.go (Phase 4 split).
// Owns: list, get-file, set-file, update-metadata for per-user agent instances.

var agentsInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage per-user agent instances",
}

var agentsInstancesListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List user instances for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/instances")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var agentsInstancesGetFileCmd = &cobra.Command{
	Use:   "get-file <agentID>",
	Short: "Get an instance context file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		file, _ := cmd.Flags().GetString("file")
		data, err := c.Get(fmt.Sprintf("/v1/agents/%s/instances/%s/files/%s", args[0], user, file))
		if err != nil {
			return err
		}
		var content struct {
			Content string `json:"content"`
		}
		_ = json.Unmarshal(data, &content)
		fmt.Println(content.Content)
		return nil
	},
}

var agentsInstancesSetFileCmd = &cobra.Command{
	Use:   "set-file <agentID>",
	Short: "Set an instance context file",
	Long: `Set the content of a named context file for a specific user instance.

The --content flag accepts a literal string or @filepath to read from disk.

Examples:
  goclaw agents instances set-file agent-1 --user=user-42 --file=context.md --content="Hello"
  goclaw agents instances set-file agent-1 --user=user-42 --file=context.md --content=@./context.md`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		file, _ := cmd.Flags().GetString("file")
		contentVal, _ := cmd.Flags().GetString("content")
		content, err := readContent(contentVal)
		if err != nil {
			return err
		}
		_, err = c.Put(fmt.Sprintf("/v1/agents/%s/instances/%s/files/%s", args[0], user, file),
			map[string]any{"content": content})
		if err != nil {
			return err
		}
		printer.Success("File updated")
		return nil
	},
}

var agentsInstancesUpdateMetadataCmd = &cobra.Command{
	Use:   "update-metadata <agentID>",
	Short: "Patch instance metadata for a user",
	Long: `Patch arbitrary metadata for a specific user instance of an agent.

--metadata must be a valid JSON object string.

Example:
  goclaw agents instances update-metadata agent-1 --user=user-42 --metadata='{"tier":"premium"}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		metaStr, _ := cmd.Flags().GetString("metadata")
		var body map[string]any
		if err := json.Unmarshal([]byte(metaStr), &body); err != nil {
			return fmt.Errorf("invalid JSON metadata: %w", err)
		}
		_, err = c.Patch(fmt.Sprintf("/v1/agents/%s/instances/%s/metadata", args[0], user), body)
		if err != nil {
			return err
		}
		printer.Success("Metadata updated")
		return nil
	},
}

// agentsInstancesMetadataCmd is kept as legacy alias for get/patch (was in agents.go).
var agentsInstancesMetadataCmd = &cobra.Command{
	Use:   "metadata <agentID>",
	Short: "Get or patch instance metadata (use update-metadata for patching)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		patch, _ := cmd.Flags().GetString("patch")
		if patch != "" {
			var body map[string]any
			if err := json.Unmarshal([]byte(patch), &body); err != nil {
				return fmt.Errorf("invalid JSON patch: %w", err)
			}
			_, err = c.Patch(fmt.Sprintf("/v1/agents/%s/instances/%s/metadata", args[0], user), body)
			if err != nil {
				return err
			}
			printer.Success("Metadata updated")
			return nil
		}
		printer.Success("Use --patch to update metadata, or use 'update-metadata' subcommand")
		return nil
	},
}

func init() {
	// get-file flags
	agentsInstancesGetFileCmd.Flags().String("user", "", "User ID")
	_ = agentsInstancesGetFileCmd.MarkFlagRequired("user")
	agentsInstancesGetFileCmd.Flags().String("file", "", "File name")
	_ = agentsInstancesGetFileCmd.MarkFlagRequired("file")

	// set-file flags
	agentsInstancesSetFileCmd.Flags().String("user", "", "User ID")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("user")
	agentsInstancesSetFileCmd.Flags().String("file", "", "File name")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("file")
	agentsInstancesSetFileCmd.Flags().String("content", "", "Content string or @filepath")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("content")

	// update-metadata flags
	agentsInstancesUpdateMetadataCmd.Flags().String("user", "", "User ID")
	_ = agentsInstancesUpdateMetadataCmd.MarkFlagRequired("user")
	agentsInstancesUpdateMetadataCmd.Flags().String("metadata", "", "JSON object to merge into metadata")
	_ = agentsInstancesUpdateMetadataCmd.MarkFlagRequired("metadata")

	// legacy metadata flags
	agentsInstancesMetadataCmd.Flags().String("user", "", "User ID")
	_ = agentsInstancesMetadataCmd.MarkFlagRequired("user")
	agentsInstancesMetadataCmd.Flags().String("patch", "", "JSON patch object")

	agentsInstancesCmd.AddCommand(
		agentsInstancesListCmd,
		agentsInstancesGetFileCmd,
		agentsInstancesSetFileCmd,
		agentsInstancesUpdateMetadataCmd,
		agentsInstancesMetadataCmd,
	)
	agentsCmd.AddCommand(agentsInstancesCmd)
}
