package cmd

import (
	"fmt"

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
		_, err = c.Patch(
			fmt.Sprintf("/v1/agents/%s/evolution/suggestions/%s", args[0], args[1]),
			map[string]any{"action": action},
		)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Suggestion %s: %sd", args[1], action))
		return nil
	},
}

func init() {
	agentsEvolutionUpdateCmd.Flags().String("action", "", "Action: accept or reject")
	_ = agentsEvolutionUpdateCmd.MarkFlagRequired("action")

	agentsEvolutionCmd.AddCommand(
		agentsEvolutionMetricsCmd,
		agentsEvolutionSuggestionsCmd,
		agentsEvolutionUpdateCmd,
	)
	agentsCmd.AddCommand(agentsEvolutionCmd)
}
