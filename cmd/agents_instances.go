package cmd

import (
	"encoding/json"
	"fmt"

	"net/url"

	"github.com/spf13/cobra"
)

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
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/instances")
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
		// file may contain path separators — don't escape it
		data, err := c.Get(fmt.Sprintf("/v1/agents/%s/instances/%s/files/%s",
			url.PathEscape(args[0]), url.PathEscape(user), file))
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
	Args:  cobra.ExactArgs(1),
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
		// file may contain path separators — don't escape it
		_, err = c.Put(fmt.Sprintf("/v1/agents/%s/instances/%s/files/%s",
			url.PathEscape(args[0]), url.PathEscape(user), file),
			map[string]any{"content": content})
		if err != nil {
			return err
		}
		printer.Success("File updated")
		return nil
	},
}

var agentsInstancesMetadataCmd = &cobra.Command{
	Use:   "metadata <agentID>",
	Short: "Get or patch instance metadata",
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
			_, err = c.Patch(fmt.Sprintf("/v1/agents/%s/instances/%s/metadata",
				url.PathEscape(args[0]), url.PathEscape(user)), body)
			if err != nil {
				return err
			}
			printer.Success("Metadata updated")
			return nil
		}
		printer.Success("Use --patch to update metadata")
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{agentsInstancesGetFileCmd, agentsInstancesSetFileCmd, agentsInstancesMetadataCmd} {
		cmd.Flags().String("user", "", "User ID")
		_ = cmd.MarkFlagRequired("user")
	}
	agentsInstancesGetFileCmd.Flags().String("file", "", "File name")
	_ = agentsInstancesGetFileCmd.MarkFlagRequired("file")
	agentsInstancesSetFileCmd.Flags().String("file", "", "File name")
	agentsInstancesSetFileCmd.Flags().String("content", "", "Content (or @filepath)")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("file")
	_ = agentsInstancesSetFileCmd.MarkFlagRequired("content")
	agentsInstancesMetadataCmd.Flags().String("patch", "", "JSON patch object")

	agentsInstancesCmd.AddCommand(agentsInstancesListCmd, agentsInstancesGetFileCmd,
		agentsInstancesSetFileCmd, agentsInstancesMetadataCmd)
	agentsCmd.AddCommand(agentsInstancesCmd)
}
