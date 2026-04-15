package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Manage agent heartbeat monitoring",
	Long:  "Configure and monitor periodic heartbeat checks for GoClaw agents.",
}

var heartbeatGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get heartbeat configuration for an agent",
	Long: `Get heartbeat configuration for an agent.

Example:
  goclaw heartbeat get --agent=my-agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		params := map[string]any{}
		if agent != "" {
			params["agentId"] = agent
		}
		data, err := ws.Call("heartbeat.get", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set heartbeat configuration for an agent",
	Long: `Set heartbeat configuration for an agent.

Example:
  goclaw heartbeat set --agent=my-agent --interval=3600`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		interval, _ := cmd.Flags().GetInt("interval")
		params := buildBody("agentId", agent, "intervalSec", interval)
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
	Short: "Enable or disable heartbeat for an agent",
	Long: `Toggle heartbeat on/off for an agent.

Example:
  goclaw heartbeat toggle --agent=my-agent --enabled=true`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		enabled, _ := cmd.Flags().GetBool("enabled")
		params := map[string]any{"agentId": agent, "enabled": enabled}
		data, err := ws.Call("heartbeat.toggle", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Trigger an immediate heartbeat run",
	Long: `Trigger an immediate heartbeat run for an agent.

Example:
  goclaw heartbeat test --agent=my-agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		agent, _ := cmd.Flags().GetString("agent")
		data, err := ws.Call("heartbeat.test", map[string]any{"agentId": agent})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatTargetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "List heartbeat delivery targets for the current tenant",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		// agentId kept for server backward-compat but targets are tenant-scoped
		data, err := ws.Call("heartbeat.targets", map[string]any{"agentId": ""})
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		items := toList(m["targets"])
		if cfg.OutputFormat != "table" {
			printer.Print(items)
			return nil
		}
		tbl := output.NewTable("CHANNEL", "CHAT_ID", "ENABLED")
		for _, t := range items {
			tbl.AddRow(str(t, "channel"), str(t, "chat_id"), str(t, "enabled"))
		}
		printer.Print(tbl)
		return nil
	},
}

var heartbeatLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View heartbeat execution logs",
	Long: `View heartbeat execution logs for an agent.

Example:
  goclaw heartbeat logs --agent=my-agent --tail=50
  goclaw heartbeat logs --agent=my-agent --follow`,
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		follow, _ := cmd.Flags().GetBool("follow")
		tail, _ := cmd.Flags().GetInt("tail")

		if follow {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet && output.IsTTY(int(os.Stdout.Fd())) {
				fmt.Println("Streaming heartbeat logs... (Ctrl+C to stop)")
			}

			err := client.FollowStream(
				ctx,
				cfg.Server, cfg.Token, "cli", cfg.Insecure,
				"heartbeat.logs",
				map[string]any{"agentId": agent, "limit": tail},
				func(event *client.WSEvent) error {
					var payload map[string]any
					if err := json.Unmarshal(event.Payload, &payload); err == nil {
						printer.Print(payload)
					}
					return nil
				},
				nil,
			)
			// ctx.Err() on graceful Ctrl+C → not an error
			if err != nil && ctx.Err() != nil {
				return nil
			}
			return err
		}

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		params := map[string]any{"agentId": agent}
		if tail > 0 {
			params["limit"] = tail
		}
		data, err := ws.Call("heartbeat.logs", params)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		items := toList(m["logs"])
		if cfg.OutputFormat != "table" {
			printer.Print(items)
			return nil
		}
		tbl := output.NewTable("ID", "STATUS", "STARTED_AT", "DURATION_MS")
		for _, l := range items {
			tbl.AddRow(str(l, "id"), str(l, "status"), str(l, "started_at"), str(l, "duration_ms"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	heartbeatGetCmd.Flags().String("agent", "", "Agent key or ID")

	heartbeatSetCmd.Flags().String("agent", "", "Agent key or ID")
	heartbeatSetCmd.Flags().Int("interval", 0, "Heartbeat interval in seconds (min 300)")
	_ = heartbeatSetCmd.MarkFlagRequired("agent")

	heartbeatToggleCmd.Flags().String("agent", "", "Agent key or ID")
	heartbeatToggleCmd.Flags().Bool("enabled", true, "Enable or disable heartbeat")
	_ = heartbeatToggleCmd.MarkFlagRequired("agent")

	heartbeatTestCmd.Flags().String("agent", "", "Agent key or ID")
	_ = heartbeatTestCmd.MarkFlagRequired("agent")

	heartbeatLogsCmd.Flags().String("agent", "", "Agent key or ID")
	heartbeatLogsCmd.Flags().Bool("follow", false, "Stream logs continuously")
	heartbeatLogsCmd.Flags().Int("tail", 20, "Number of recent log entries to show")

	heartbeatCmd.AddCommand(
		heartbeatGetCmd, heartbeatSetCmd, heartbeatToggleCmd,
		heartbeatTestCmd, heartbeatTargetsCmd, heartbeatLogsCmd,
	)
	rootCmd.AddCommand(heartbeatCmd)
}

// heartbeatChecklistCmd is defined in heartbeat_checklist.go (split due to LoC limit).
// Its init() registers itself under heartbeatCmd directly.
