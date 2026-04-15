package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// agents_v3_flags.go — v3 feature flag get and toggle
// HTTP endpoints: GET /v1/agents/{id}/v3-flags, PATCH /v1/agents/{id}/v3-flags
// WARNING: toggling v3 flags may enable experimental features — use with caution.

var agentsV3FlagsCmd = &cobra.Command{
	Use:   "v3-flags",
	Short: "Manage agent v3 feature flags (experimental)",
}

var agentsV3FlagsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get v3 feature flags for an agent",
	Long: `Get the current v3 feature flag state for an agent.

GET /v1/agents/{id}/v3-flags

Example:
  goclaw agents v3-flags get agent-1
  goclaw agents v3-flags get agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/v3-flags")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var agentsV3FlagsToggleCmd = &cobra.Command{
	Use:   "toggle <id>",
	Short: "Toggle a v3 feature flag for an agent",
	Long: `Toggle a named v3 feature flag on or off for an agent.

PATCH /v1/agents/{id}/v3-flags

WARNING: v3 flags may enable experimental features. Verify before toggling in production.

Example:
  goclaw agents v3-flags toggle agent-1 --flag=kg_auto_extract
  goclaw agents v3-flags toggle agent-1 --flag=multi_session`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flag, _ := cmd.Flags().GetString("flag")
		if flag == "" {
			return fmt.Errorf("--flag is required")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Patch("/v1/agents/"+args[0]+"/v3-flags",
			map[string]any{"flag": flag})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Flag %q toggled for agent %s", flag, args[0]))
		return nil
	},
}

func init() {
	agentsV3FlagsToggleCmd.Flags().String("flag", "", "Feature flag name to toggle")
	_ = agentsV3FlagsToggleCmd.MarkFlagRequired("flag")

	agentsV3FlagsCmd.AddCommand(agentsV3FlagsGetCmd, agentsV3FlagsToggleCmd)
	agentsCmd.AddCommand(agentsV3FlagsCmd)
}
