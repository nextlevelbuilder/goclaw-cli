package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var channelsWritersCmd = &cobra.Command{Use: "writers", Short: "Manage group writers"}

var channelsWritersListCmd = &cobra.Command{
	Use: "list <instanceID>", Short: "List writers", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/instances/" + url.PathEscape(args[0]) + "/writers")
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
		_, err = c.Post("/v1/channels/instances/"+url.PathEscape(args[0])+"/writers",
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
		_, err = c.Delete("/v1/channels/instances/" + url.PathEscape(args[0]) + "/writers/" + user)
		if err != nil {
			return err
		}
		printer.Success("Writer removed")
		return nil
	},
}

var channelsWritersGroupsCmd = &cobra.Command{
	Use: "groups <instanceID>", Short: "List writer groups", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/instances/" + url.PathEscape(args[0]) + "/writers/groups")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	channelsWritersAddCmd.Flags().String("user", "", "User ID")
	channelsWritersAddCmd.Flags().String("display-name", "", "Display name")
	_ = channelsWritersAddCmd.MarkFlagRequired("user")
	channelsWritersRemoveCmd.Flags().String("user", "", "User ID")
	_ = channelsWritersRemoveCmd.MarkFlagRequired("user")

	channelsWritersCmd.AddCommand(channelsWritersListCmd, channelsWritersAddCmd,
		channelsWritersRemoveCmd, channelsWritersGroupsCmd)
	channelsCmd.AddCommand(channelsWritersCmd)
}
