package cmd

import (
	"fmt"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"net/url"
	"github.com/spf13/cobra"
)

var channelsCmd = &cobra.Command{Use: "channels", Short: "Manage messaging channels"}

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
		data, err := c.Get("/v1/channels/instances/" + url.PathEscape(args[0]))
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
		_, err = c.Delete("/v1/channels/instances/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("Channel deleted")
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

	channelsInstancesCmd.AddCommand(channelsInstancesListCmd, channelsInstancesGetCmd,
		channelsInstancesCreateCmd, channelsInstancesUpdateCmd, channelsInstancesDeleteCmd)
	// contacts, pending, writers registered from their own files
	channelsCmd.AddCommand(channelsInstancesCmd)
	rootCmd.AddCommand(channelsCmd)
}
