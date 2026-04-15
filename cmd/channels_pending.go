package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// channels_pending.go extends channelsPendingCmd with system-wide pending message
// management (groups, messages, delete, compact). These operate on /v1/pending-messages
// which is distinct from /v1/channels/{id}/pending (channel-specific queue).

var channelsPendingGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "List all pending message groups (system-wide)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/pending-messages")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsPendingMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "List messages in a pending group",
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID, _ := cmd.Flags().GetString("group")
		if groupID == "" {
			return fmt.Errorf("--group is required")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		q.Set("group_id", groupID)
		data, err := c.Get("/v1/pending-messages/messages?" + q.Encode())
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsPendingDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete pending messages for a group (requires --yes; --all to wipe everything)",
	Long: `Delete pending messages.

By default, you must specify --channel and/or --key to scope the deletion.
Use --all to delete EVERY pending message system-wide (extremely destructive).

Example:
  goclaw channels pending delete --channel=zalo --key=group123 --yes
  goclaw channels pending delete --all --yes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		ch, _ := cmd.Flags().GetString("channel")
		key, _ := cmd.Flags().GetString("key")

		if !all && ch == "" && key == "" {
			return fmt.Errorf("delete requires at least --channel/--key filter, or --all to wipe everything")
		}
		msg := "Delete pending messages?"
		if all {
			msg = "Delete ALL pending messages system-wide? This cannot be undone."
		}
		if !tui.Confirm(msg, cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if ch != "" {
			q.Set("channel", ch)
		}
		if key != "" {
			q.Set("key", key)
		}
		path := "/v1/pending-messages"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		_, err = c.Delete(path)
		if err != nil {
			return err
		}
		printer.Success("Pending messages deleted")
		return nil
	},
}

var channelsPendingCompactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Compact pending messages via LLM summarization (requires --yes)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Compact pending messages? This uses LLM tokens.", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/pending-messages/compact", nil)
		if err != nil {
			return err
		}
		printer.Success("Compact initiated")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	channelsPendingMessagesCmd.Flags().String("group", "", "Group ID to list messages for (required)")
	_ = channelsPendingMessagesCmd.MarkFlagRequired("group")

	channelsPendingDeleteCmd.Flags().String("channel", "", "Filter by channel name")
	channelsPendingDeleteCmd.Flags().String("key", "", "Filter by group key")
	channelsPendingDeleteCmd.Flags().Bool("all", false, "Delete ALL pending messages system-wide (requires --yes)")

	// Register as subcommands of the existing channelsPendingCmd (defined in channels.go).
	channelsPendingCmd.AddCommand(
		channelsPendingGroupsCmd,
		channelsPendingMessagesCmd,
		channelsPendingDeleteCmd,
		channelsPendingCompactCmd,
	)
}
