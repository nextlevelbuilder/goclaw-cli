package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var agentsWakeCmd = &cobra.Command{
	Use:   "wake <id>",
	Short: "Wake up a sleeping agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+url.PathEscape(args[0])+"/wake", nil)
		if err != nil {
			return err
		}
		printer.Success("Agent woken up")
		return nil
	},
}

func init() {
	agentsCmd.AddCommand(agentsWakeCmd)
}
