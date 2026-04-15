package cmd

import (
	"github.com/spf13/cobra"
)

// agents_misc.go — orchestration mode and codex pool activity
// HTTP endpoints: GET /v1/agents/{id}/orchestration, GET /v1/agents/{id}/codex-pool-activity

var agentsOrchestrationCmd = &cobra.Command{
	Use:   "orchestration <id>",
	Short: "Get orchestration configuration for an agent",
	Long: `Retrieve the orchestration mode and delegation configuration for an agent.

GET /v1/agents/{id}/orchestration

Example:
  goclaw agents orchestration agent-1
  goclaw agents orchestration agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/orchestration")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var agentsCodexPoolActivityCmd = &cobra.Command{
	Use:   "codex-pool-activity <id>",
	Short: "Get codex pool activity for an agent",
	Long: `Retrieve recent codex (context pool) activity for an agent.

GET /v1/agents/{id}/codex-pool-activity

Example:
  goclaw agents codex-pool-activity agent-1
  goclaw agents codex-pool-activity agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/codex-pool-activity")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	agentsCmd.AddCommand(agentsOrchestrationCmd, agentsCodexPoolActivityCmd)
}
