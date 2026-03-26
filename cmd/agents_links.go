package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"net/url"
	"github.com/spf13/cobra"
)

var agentsLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage agent delegation links",
}

var agentsLinksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List delegation links",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/links")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "SOURCE", "TARGET", "DIRECTION", "MAX_CONCURRENT")
		for _, l := range unmarshalList(data) {
			tbl.AddRow(str(l, "id"), str(l, "source_agent"), str(l, "target_agent"),
				str(l, "direction"), str(l, "max_concurrent"))
		}
		printer.Print(tbl)
		return nil
	},
}

var agentsLinksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a delegation link",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")
		direction, _ := cmd.Flags().GetString("direction")
		maxConc, _ := cmd.Flags().GetInt("max-concurrent")
		body := buildBody("source_agent", source, "target_agent", target,
			"direction", direction, "max_concurrent", maxConc)
		data, err := c.Post("/v1/agents/links", body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Link created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var agentsLinksUpdateCmd = &cobra.Command{
	Use:   "update <linkID>",
	Short: "Update a delegation link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("direction") {
			v, _ := cmd.Flags().GetString("direction")
			body["direction"] = v
		}
		if cmd.Flags().Changed("max-concurrent") {
			v, _ := cmd.Flags().GetInt("max-concurrent")
			body["max_concurrent"] = v
		}
		_, err = c.Put("/v1/agents/links/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Link updated")
		return nil
	},
}

var agentsLinksDeleteCmd = &cobra.Command{
	Use:   "delete <linkID>",
	Short: "Delete a delegation link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this link?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/agents/links/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("Link deleted")
		return nil
	},
}

func init() {
	agentsLinksCreateCmd.Flags().String("source", "", "Source agent ID")
	agentsLinksCreateCmd.Flags().String("target", "", "Target agent ID")
	agentsLinksCreateCmd.Flags().String("direction", "outbound", "Direction: outbound, inbound, bidirectional")
	agentsLinksCreateCmd.Flags().Int("max-concurrent", 3, "Max concurrent delegations")
	agentsLinksUpdateCmd.Flags().String("direction", "", "Direction")
	agentsLinksUpdateCmd.Flags().Int("max-concurrent", 0, "Max concurrent")

	agentsLinksCmd.AddCommand(agentsLinksListCmd, agentsLinksCreateCmd, agentsLinksUpdateCmd, agentsLinksDeleteCmd)
	agentsCmd.AddCommand(agentsLinksCmd)
}
