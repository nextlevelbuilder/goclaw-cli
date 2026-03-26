package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Manage heartbeat configuration and monitoring",
}

var heartbeatGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get heartbeat configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("heartbeat.get", map[string]any{})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set heartbeat configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		interval, _ := cmd.Flags().GetInt("interval")
		url, _ := cmd.Flags().GetString("url")
		params := buildBody("interval", interval, "url", url)
		data, err := ws.Call("heartbeat.set", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatToggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Enable or disable heartbeat",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		enabled, _ := cmd.Flags().GetBool("enabled")
		data, err := ws.Call("heartbeat.toggle", map[string]any{"enabled": enabled})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Trigger a test heartbeat",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("heartbeat.test", map[string]any{})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get heartbeat logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		limit, _ := cmd.Flags().GetInt("limit")
		data, err := ws.Call("heartbeat.logs", map[string]any{"limit": limit})
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("TIMESTAMP", "STATUS", "LATENCY", "ERROR")
		for _, l := range unmarshalList(data) {
			tbl.AddRow(str(l, "timestamp"), str(l, "status"), str(l, "latency"), str(l, "error"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	heartbeatSetCmd.Flags().Int("interval", 0, "Heartbeat interval in seconds")
	heartbeatSetCmd.Flags().String("url", "", "Heartbeat endpoint URL")
	heartbeatToggleCmd.Flags().Bool("enabled", true, "Enable or disable heartbeat")
	heartbeatLogsCmd.Flags().Int("limit", 20, "Number of log entries to return")

	heartbeatCmd.AddCommand(
		heartbeatGetCmd,
		heartbeatSetCmd,
		heartbeatToggleCmd,
		heartbeatTestCmd,
		heartbeatLogsCmd,
	)
	rootCmd.AddCommand(heartbeatCmd)
}
