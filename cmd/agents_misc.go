package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// agents_misc.go — orchestration mode, codex pool activity, cancel-summon, skills.
// HTTP endpoints: GET /v1/agents/{id}/orchestration, GET /v1/agents/{id}/codex-pool-activity,
// POST /v1/agents/{id}/cancel-summon, GET /v1/agents/{agentID}/skills

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

var agentsCancelSummonCmd = &cobra.Command{
	Use:   "cancel-summon <id>",
	Short: "Cancel an in-progress agent summon (requires --yes)",
	Long: `Cancel a pending or running summon for an agent.

POST /v1/agents/{id}/cancel-summon`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Cancel the in-progress summon?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+args[0]+"/cancel-summon", nil)
		if err != nil {
			return err
		}
		printer.Success("Summon cancelled")
		return nil
	},
}

var agentsSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage skills granted to agents",
}

var agentsSkillsListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List skills granted to an agent",
	Long:  `GET /v1/agents/{agentID}/skills`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/skills")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	agentsSkillsCmd.AddCommand(agentsSkillsListCmd)
	agentsCmd.AddCommand(agentsOrchestrationCmd, agentsCodexPoolActivityCmd,
		agentsCancelSummonCmd, agentsSkillsCmd)
}
