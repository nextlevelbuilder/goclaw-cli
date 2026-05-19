package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var codexPoolCmd = &cobra.Command{Use: "codex-pool", Short: "Inspect Codex pool activity"}

var codexPoolActivityCmd = &cobra.Command{
	Use:   "activity",
	Short: "Show Codex pool activity for an agent or provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetString("agent")
		providerID, _ := cmd.Flags().GetString("provider")
		if (agentID == "") == (providerID == "") {
			return fmt.Errorf("provide exactly one of --agent or --provider")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := ""
		if agentID != "" {
			path = "/v1/agents/" + url.PathEscape(agentID) + "/codex-pool-activity"
		} else {
			path = "/v1/providers/" + url.PathEscape(providerID) + "/codex-pool-activity"
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	codexPoolActivityCmd.Flags().String("agent", "", "Agent ID")
	codexPoolActivityCmd.Flags().String("provider", "", "Provider ID")
	codexPoolCmd.AddCommand(codexPoolActivityCmd)
	rootCmd.AddCommand(codexPoolCmd)
}
