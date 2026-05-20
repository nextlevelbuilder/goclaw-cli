package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

// agents_evolution.go — evolution metrics, suggestions, update
// HTTP endpoints: GET/PATCH /v1/agents/{id}/evolution/...

var agentsEvolutionCmd = &cobra.Command{
	Use:   "evolution",
	Short: "Agent evolution metrics and suggestions",
}

var agentsEvolutionMetricsCmd = &cobra.Command{
	Use:   "metrics <id>",
	Short: "Get evolution metrics for an agent",
	Long: `Retrieve performance and evolution metrics for an agent.

GET /v1/agents/{id}/evolution/metrics

Example:
  goclaw agents evolution metrics agent-1
  goclaw agents evolution metrics agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/evolution/metrics")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var agentsEvolutionSuggestionsCmd = &cobra.Command{
	Use:   "suggestions <id>",
	Short: "List evolution suggestions for an agent",
	Long: `List pending evolution suggestions for an agent (e.g. prompt improvements, config changes).

GET /v1/agents/{id}/evolution/suggestions

Example:
  goclaw agents evolution suggestions agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/evolution/suggestions")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var agentsEvolutionUpdateCmd = &cobra.Command{
	Use:   "update <id> <suggestionID>",
	Short: "Accept or reject an evolution suggestion",
	Long: `Accept or reject a specific evolution suggestion for an agent.

PATCH /v1/agents/{id}/evolution/suggestions/{suggestionID}

--action must be "accept" or "reject".

Example:
  goclaw agents evolution update agent-1 sugg-42 --action=accept
  goclaw agents evolution update agent-1 sugg-42 --action=reject`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		action, _ := cmd.Flags().GetString("action")
		if action != "accept" && action != "reject" {
			return fmt.Errorf("--action must be 'accept' or 'reject', got %q", action)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		status := map[string]string{
			"accept": "approved",
			"reject": "rejected",
		}[action]
		_, err = c.Patch(
			fmt.Sprintf(
				"/v1/agents/%s/evolution/suggestions/%s",
				url.PathEscape(args[0]),
				url.PathEscape(args[1]),
			),
			map[string]any{"status": status},
		)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Suggestion %s %s", args[1], status))
		return nil
	},
}

var agentsEvolutionSkillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Apply skill evolution suggestions",
}

var agentsEvolutionSkillApplyCmd = &cobra.Command{
	Use:   "apply <id> <suggestionID>",
	Short: "Approve a skill_add evolution suggestion",
	Long: `Approve a skill_add evolution suggestion for an agent.

PATCH /v1/agents/{id}/evolution/suggestions/{suggestionID}

Example:
  goclaw agents evolution skill apply agent-1 sugg-42
  goclaw agents evolution skill apply agent-1 sugg-42 --skill-draft @./SKILL.md`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{"status": "approved"}
		if cmd.Flags().Changed("skill-draft") {
			draft, _ := cmd.Flags().GetString("skill-draft")
			content, err := readContent(draft)
			if err != nil {
				return err
			}
			body["skill_draft"] = content
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		if err := requireSkillAddSuggestion(c, args[0], args[1]); err != nil {
			return err
		}
		data, err := c.Patch(
			fmt.Sprintf(
				"/v1/agents/%s/evolution/suggestions/%s",
				url.PathEscape(args[0]),
				url.PathEscape(args[1]),
			),
			body,
		)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func requireSkillAddSuggestion(c interface {
	Get(path string) (json.RawMessage, error)
}, agentID, suggestionID string) error {
	data, err := c.Get("/v1/agents/" + url.PathEscape(agentID) + "/evolution/suggestions?status=pending&limit=500")
	if err != nil {
		return err
	}
	for _, suggestion := range unmarshalList(data) {
		if str(suggestion, "id") == suggestionID {
			if str(suggestion, "suggestion_type") != "skill_add" {
				return fmt.Errorf("suggestion %s is %q, not skill_add", suggestionID, str(suggestion, "suggestion_type"))
			}
			return nil
		}
	}
	return fmt.Errorf("suggestion %s not found in agent evolution suggestions", suggestionID)
}

func init() {
	agentsEvolutionUpdateCmd.Flags().String("action", "", "Action: accept or reject")
	_ = agentsEvolutionUpdateCmd.MarkFlagRequired("action")
	agentsEvolutionSkillApplyCmd.Flags().String("skill-draft", "", "Skill draft content or @file")
	agentsEvolutionSkillCmd.AddCommand(agentsEvolutionSkillApplyCmd)

	agentsEvolutionCmd.AddCommand(
		agentsEvolutionMetricsCmd,
		agentsEvolutionSuggestionsCmd,
		agentsEvolutionUpdateCmd,
		agentsEvolutionSkillCmd,
	)
	agentsCmd.AddCommand(agentsEvolutionCmd)
}
