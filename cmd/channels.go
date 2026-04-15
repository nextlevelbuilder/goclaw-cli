package cmd

import (
	"github.com/spf13/cobra"
)

var channelsCmd = &cobra.Command{Use: "channels", Short: "Manage messaging channels"}

// channelsInstancesCmd assembled in channels_instances.go init().

// --- Contacts ---

var channelsContactsCmd = &cobra.Command{Use: "contacts", Short: "Manage contacts"}

var channelsContactsListCmd = &cobra.Command{
	Use: "list", Short: "List contacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/contacts")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsContactsResolveCmd = &cobra.Command{
	Use: "resolve <ids>", Short: "Resolve contacts by IDs", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/contacts/resolve?ids=" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

// --- Pending Messages ---

var channelsPendingCmd = &cobra.Command{Use: "pending", Short: "Manage pending messages"}

var channelsPendingListCmd = &cobra.Command{
	Use: "list <channelID>", Short: "List pending messages", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/" + args[0] + "/pending")
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
	// channelsInstancesCmd assembled in channels_instances.go init().
	// channelsWritersCmd assembled in channels_writers.go init().
	channelsContactsCmd.AddCommand(channelsContactsListCmd, channelsContactsResolveCmd)
	channelsPendingCmd.AddCommand(channelsPendingListCmd, channelsPendingRetryCmd)
	channelsCmd.AddCommand(channelsInstancesCmd, channelsContactsCmd, channelsPendingCmd, channelsWritersCmd)
	rootCmd.AddCommand(channelsCmd)
}
