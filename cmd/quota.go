package cmd

import (
	"github.com/spf13/cobra"
)

// quotaCmd provides quota inspection via WebSocket.
var quotaCmd = &cobra.Command{Use: "quota", Short: "Inspect quota usage"}

var quotaUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show quota consumption (all users or filtered by agent)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		var params map[string]any
		if agent, _ := cmd.Flags().GetString("agent"); agent != "" {
			params = map[string]any{"agent": agent}
		}
		data, err := ws.Call("quota.usage", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	quotaUsageCmd.Flags().String("agent", "", "Filter by agent key")
	quotaCmd.AddCommand(quotaUsageCmd)
	rootCmd.AddCommand(quotaCmd)
}
