package cmd

import (
	"net/url"
	"github.com/spf13/cobra"
)

var channelsPendingCmd = &cobra.Command{Use: "pending", Short: "Manage pending messages"}

var channelsPendingListCmd = &cobra.Command{
	Use: "list <channelID>", Short: "List pending messages", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/" + url.PathEscape(args[0]) + "/pending")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsPendingRetryCmd = &cobra.Command{
	Use: "retry <channelID> <messageID>", Short: "Retry pending message", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Patch("/v1/channels/"+args[0]+"/pending/"+args[1], map[string]any{"action": "retry"})
		if err != nil {
			return err
		}
		printer.Success("Message retried")
		return nil
	},
}

func init() {
	channelsPendingCmd.AddCommand(channelsPendingListCmd, channelsPendingRetryCmd)
	channelsCmd.AddCommand(channelsPendingCmd)
}
