package cmd

import (
	"github.com/spf13/cobra"
)

// agents_admin.go — admin-level agent operations: sync-workspace, prompt-preview.
// Extracted from agents_lifecycle.go to keep file sizes <200 LoC.

var agentsSyncWorkspaceCmd = &cobra.Command{
	Use:   "sync-workspace",
	Short: "Sync workspace for all agents (admin)",
	Long: `Trigger a workspace sync for all agents. This is an admin-only operation.

POST /v1/agents/sync-workspace

Example:
  goclaw agents sync-workspace`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/sync-workspace", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var agentsPromptPreviewCmd = &cobra.Command{
	Use:   "prompt-preview <id>",
	Short: "Preview the assembled system prompt for an agent",
	Long: `Retrieve the fully rendered system prompt for an agent as the server would send it to the LLM.

GET /v1/agents/{id}/system-prompt-preview

Example:
  goclaw agents prompt-preview agent-1
  goclaw agents prompt-preview agent-1 --output=json | jq '.prompt' -r`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/system-prompt-preview")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	agentsCmd.AddCommand(agentsSyncWorkspaceCmd, agentsPromptPreviewCmd)
}
