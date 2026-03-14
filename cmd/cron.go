package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var cronCmd = &cobra.Command{Use: "cron", Short: "Manage scheduled jobs"}

var cronListCmd = &cobra.Command{
	Use: "list", Short: "List cron jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		ws, wsErr := newWS("cli")
		if wsErr != nil {
			return wsErr
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("cron.list", nil)
		if err != nil {
			// Fallback to HTTP if WS method unavailable
			_ = c
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "AGENT", "SCHEDULE", "ENABLED", "LAST_STATUS", "NEXT_RUN")
		for _, j := range unmarshalList(data) {
			schedule := str(j, "cron_expression")
			if schedule == "" {
				schedule = str(j, "interval_ms") + "ms"
			}
			tbl.AddRow(str(j, "id"), str(j, "name"), str(j, "agent_id"),
				schedule, str(j, "enabled"), str(j, "last_status"), str(j, "next_run_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var cronGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get cron job details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("cron.status", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var cronCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a cron job",
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
		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		message, _ := cmd.Flags().GetString("message")
		timezone, _ := cmd.Flags().GetString("timezone")

		params := buildBody("agent_id", agent, "name", name,
			"cron_expression", schedule, "timezone", timezone)
		if message != "" {
			params["payload"] = map[string]any{"message": message}
		}

		data, err := ws.Call("cron.create", params)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Cron job created: %s", str(unmarshalMap(data), "id")))
		return nil
	},
}

var cronUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update cron job", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		params := map[string]any{"id": args[0]}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			params["name"] = v
		}
		if cmd.Flags().Changed("schedule") {
			v, _ := cmd.Flags().GetString("schedule")
			params["cron_expression"] = v
		}
		_, err = ws.Call("cron.update", params)
		if err != nil {
			return err
		}
		printer.Success("Cron job updated")
		return nil
	},
}

var cronDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete cron job", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this cron job?", cfg.Yes) {
			return nil
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("cron.delete", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Cron job deleted")
		return nil
	},
}

var cronToggleCmd = &cobra.Command{
	Use: "toggle <id>", Short: "Enable/disable cron job", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("cron.toggle", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Cron job toggled")
		return nil
	},
}

var cronRunCmd = &cobra.Command{
	Use: "run <id>", Short: "Manually trigger cron job", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("cron.run", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Cron job triggered")
		return nil
	},
}

var cronStatusCmd = &cobra.Command{
	Use: "status <id>", Short: "Check cron job status", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("cron.status", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var cronRunsCmd = &cobra.Command{
	Use: "runs <id>", Short: "List cron run history", Args: cobra.ExactArgs(1),
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
		params := map[string]any{"id": args[0]}
		if limit > 0 {
			params["limit"] = limit
		}
		data, err := ws.Call("cron.runs", params)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("RUN_ID", "STATUS", "STARTED", "DURATION_MS", "TOKENS")
		for _, r := range unmarshalList(data) {
			tbl.AddRow(str(r, "id"), str(r, "status"), str(r, "started_at"),
				str(r, "duration_ms"), str(r, "total_tokens"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	cronCreateCmd.Flags().String("agent", "", "Agent ID")
	cronCreateCmd.Flags().String("name", "", "Job name")
	cronCreateCmd.Flags().String("schedule", "", "Cron expression")
	cronCreateCmd.Flags().String("message", "", "Message payload")
	cronCreateCmd.Flags().String("timezone", "", "Timezone")
	_ = cronCreateCmd.MarkFlagRequired("agent")
	_ = cronCreateCmd.MarkFlagRequired("name")
	_ = cronCreateCmd.MarkFlagRequired("schedule")

	cronUpdateCmd.Flags().String("name", "", "Job name")
	cronUpdateCmd.Flags().String("schedule", "", "Cron expression")
	cronRunsCmd.Flags().Int("limit", 20, "Max results")

	cronCmd.AddCommand(cronListCmd, cronGetCmd, cronCreateCmd, cronUpdateCmd,
		cronDeleteCmd, cronToggleCmd, cronRunCmd, cronStatusCmd, cronRunsCmd)
	rootCmd.AddCommand(cronCmd)
}
