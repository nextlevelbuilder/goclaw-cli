package cmd

import (
	"fmt"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var channelsCmd = &cobra.Command{Use: "channels", Short: "Manage messaging channels"}

// --- Channel Instances ---

var channelsInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage channel instances"}

var channelsInstancesListCmd = &cobra.Command{
	Use: "list", Short: "List channel instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/channels/instances"
		if v, _ := cmd.Flags().GetString("type"); v != "" {
			path += "?channel_type=" + v
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "TYPE", "AGENT", "ENABLED")
		for _, ch := range unmarshalList(data) {
			tbl.AddRow(str(ch, "id"), str(ch, "name"), str(ch, "channel_type"),
				str(ch, "agent_id"), str(ch, "enabled"))
		}
		printer.Print(tbl)
		return nil
	},
}

var channelsInstancesGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get channel instance", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/channels/instances/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var channelsInstancesCreateCmd = &cobra.Command{
	Use: "create", Short: "Create channel instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		chType, _ := cmd.Flags().GetString("type")
		agent, _ := cmd.Flags().GetString("agent")
		body := buildBody("name", name, "channel_type", chType, "agent_id", agent, "enabled", true)
		data, err := c.Post("/v1/channels/instances", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Channel created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var channelsInstancesUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update channel instance", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			body["enabled"] = v
		}
		_, err = c.Put("/v1/channels/instances/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Channel updated")
		return nil
	},
}

var channelsInstancesDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete channel instance", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this channel?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/channels/instances/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Channel deleted")
		return nil
	},
}

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

// --- Writers ---

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

func init() {
	channelsInstancesListCmd.Flags().String("type", "", "Filter: "+strings.Join(
		[]string{"telegram", "discord", "slack", "zalo-oa", "zalo-personal", "feishu", "whatsapp"}, ", "))
	channelsInstancesCreateCmd.Flags().String("name", "", "Instance name")
	channelsInstancesCreateCmd.Flags().String("type", "", "Channel type")
	channelsInstancesCreateCmd.Flags().String("agent", "", "Agent ID")
	_ = channelsInstancesCreateCmd.MarkFlagRequired("name")
	_ = channelsInstancesCreateCmd.MarkFlagRequired("type")
	_ = channelsInstancesCreateCmd.MarkFlagRequired("agent")
	channelsInstancesUpdateCmd.Flags().String("name", "", "Name")
	channelsInstancesUpdateCmd.Flags().Bool("enabled", true, "Enable/disable")

	channelsWritersAddCmd.Flags().String("user", "", "User ID")
	channelsWritersAddCmd.Flags().String("display-name", "", "Display name")
	_ = channelsWritersAddCmd.MarkFlagRequired("user")
	channelsWritersRemoveCmd.Flags().String("user", "", "User ID")
	_ = channelsWritersRemoveCmd.MarkFlagRequired("user")

	channelsInstancesCmd.AddCommand(channelsInstancesListCmd, channelsInstancesGetCmd,
		channelsInstancesCreateCmd, channelsInstancesUpdateCmd, channelsInstancesDeleteCmd)
	channelsContactsCmd.AddCommand(channelsContactsListCmd, channelsContactsResolveCmd)
	channelsPendingCmd.AddCommand(channelsPendingListCmd, channelsPendingRetryCmd)
	channelsWritersCmd.AddCommand(channelsWritersListCmd, channelsWritersAddCmd, channelsWritersRemoveCmd)
	channelsCmd.AddCommand(channelsInstancesCmd, channelsContactsCmd, channelsPendingCmd, channelsWritersCmd)
	rootCmd.AddCommand(channelsCmd)
}
