package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// channels_writers.go holds the writers subcommand, extracted from channels.go
// to keep that file under 200 LoC.

var channelsWritersCmd = &cobra.Command{Use: "writers", Short: "Manage group writers"}

var channelsWritersListCmd = &cobra.Command{
	Use: "list <instanceID>", Short: "List writers", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/instances/" + args[0] + "/writers")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsWritersGroupsCmd = &cobra.Command{
	Use:   "groups <instanceID>",
	Short: "List writer groups",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/instances/" + args[0] + "/writers/groups")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsWritersAddCmd = &cobra.Command{
	Use: "add <instanceID>", Short: "Add writer", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		displayName, _ := cmd.Flags().GetString("display-name")
		_, err = c.Post("/v1/channels/instances/"+args[0]+"/writers",
			buildBody("user_id", user, "display_name", displayName))
		if err != nil {
			return err
		}
		printer.Success("Writer added")
		return nil
	},
}

var channelsWritersRemoveCmd = &cobra.Command{
	Use: "remove <instanceID>", Short: "Remove writer", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		user, _ := cmd.Flags().GetString("user")
		_, err = c.Delete("/v1/channels/instances/" + args[0] + "/writers/" + user)
		if err != nil {
			return err
		}
		printer.Success("Writer removed")
		return nil
	},
}

// channelsWritersTestCmd probes whether a (group, user) pair is permitted to
// write into a channel instance. Body is POSTed with exactly two keys
// (group_id, user_id) — no extra fields, so the server's contract stays tight.
//
// Backend route: POST /v1/channels/instances/{id}/writers/test
var channelsWritersTestCmd = &cobra.Command{
	Use:   "test <instanceID>",
	Short: "Test whether a (group, user) pair is allowed to write",
	Long: `Probe a channel instance's writer policy for a specific group/user pair
without mutating state.

Backend route: POST /v1/channels/instances/{id}/writers/test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")
		userID, _ := cmd.Flags().GetString("user-id")
		// Body has exactly two keys — construct directly so no other flags
		// (e.g. accidental future additions) leak into the request.
		body := map[string]any{
			"group_id": groupID,
			"user_id":  userID,
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/channels/instances/" + url.PathEscape(args[0]) + "/writers/test"
		data, err := c.Post(path, body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		tbl := output.NewTable("ALLOWED", "REASON", "WRITER_COUNT", "GROUP_ID", "USER_ID")
		tbl.AddRow(str(m, "allowed"), str(m, "reason"), str(m, "writer_count"),
			str(m, "group_id"), str(m, "user_id"))
		printer.Print(tbl)
		return nil
	},
}

func init() {
	channelsWritersAddCmd.Flags().String("user", "", "User ID")
	channelsWritersAddCmd.Flags().String("display-name", "", "Display name")
	_ = channelsWritersAddCmd.MarkFlagRequired("user")
	channelsWritersRemoveCmd.Flags().String("user", "", "User ID")
	_ = channelsWritersRemoveCmd.MarkFlagRequired("user")

	channelsWritersTestCmd.Flags().String("group-id", "", "Group identifier (required)")
	channelsWritersTestCmd.Flags().String("user-id", "", "User identifier (required)")
	_ = channelsWritersTestCmd.MarkFlagRequired("group-id")
	_ = channelsWritersTestCmd.MarkFlagRequired("user-id")

	channelsWritersCmd.AddCommand(channelsWritersListCmd, channelsWritersGroupsCmd, channelsWritersAddCmd, channelsWritersRemoveCmd, channelsWritersTestCmd)
}
