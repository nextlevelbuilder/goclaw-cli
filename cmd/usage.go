package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var usageCmd = &cobra.Command{Use: "usage", Short: "View usage and cost analytics"}

var usageSummaryCmd = &cobra.Command{
	Use: "summary", Short: "Usage summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("from"); v != "" {
			q.Set("from", v)
		}
		if v, _ := cmd.Flags().GetString("to"); v != "" {
			q.Set("to", v)
		}
		path := "/v1/usage/summary"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var usageDetailCmd = &cobra.Command{
	Use: "detail", Short: "Detailed usage breakdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		if v, _ := cmd.Flags().GetString("provider"); v != "" {
			q.Set("provider", v)
		}
		if v, _ := cmd.Flags().GetString("from"); v != "" {
			q.Set("from", v)
		}
		if v, _ := cmd.Flags().GetString("to"); v != "" {
			q.Set("to", v)
		}
		path := "/v1/usage"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var usageCostsCmd = &cobra.Command{
	Use: "costs", Short: "Cost summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/costs/summary")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var usageBreakdownCmd = &cobra.Command{
	Use: "breakdown", Short: "Usage breakdown by model/agent/day",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		if v, _ := cmd.Flags().GetString("group-by"); v != "" {
			q.Set("group_by", v)
		}
		path := "/v1/usage/breakdown"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("GROUP", "TOKENS_IN", "TOKENS_OUT", "COST")
		for _, row := range unmarshalList(data) {
			tbl.AddRow(str(row, "group"), str(row, "tokens_in"), str(row, "tokens_out"), str(row, "cost"))
		}
		printer.Print(tbl)
		return nil
	},
}

var usageTimeseriesCmd = &cobra.Command{
	Use: "timeseries", Short: "Usage timeseries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		if v, _ := cmd.Flags().GetString("interval"); v != "" {
			q.Set("interval", v)
		}
		path := "/v1/usage/timeseries"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("PERIOD", "TOKENS_IN", "TOKENS_OUT", "REQUESTS")
		for _, row := range unmarshalList(data) {
			tbl.AddRow(str(row, "period"), str(row, "tokens_in"), str(row, "tokens_out"), str(row, "requests"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	usageSummaryCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	usageSummaryCmd.Flags().String("to", "", "End date")
	usageDetailCmd.Flags().String("agent", "", "Agent ID")
	usageDetailCmd.Flags().String("provider", "", "Provider name")
	usageDetailCmd.Flags().String("from", "", "Start date")
	usageDetailCmd.Flags().String("to", "", "End date")
	usageBreakdownCmd.Flags().String("agent", "", "Agent ID")
	usageBreakdownCmd.Flags().String("group-by", "", "Group by: model, agent, day")
	usageTimeseriesCmd.Flags().String("agent", "", "Agent ID")
	usageTimeseriesCmd.Flags().String("interval", "", "Interval: hour, day, week")

	usageCmd.AddCommand(usageSummaryCmd, usageDetailCmd, usageCostsCmd,
		usageBreakdownCmd, usageTimeseriesCmd)
	rootCmd.AddCommand(usageCmd)
}
