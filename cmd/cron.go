package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var cronCmd = &cobra.Command{Use: "cron", Short: "Manage scheduled jobs"}

var cronListCmd = &cobra.Command{
	Use: "list", Short: "List cron jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		result, err := ws.CronList(client.CronListParams{})
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(result.Jobs)
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "AGENT", "SCHEDULE", "ENABLED", "DELIVER_CHANNEL")
		for _, j := range result.Jobs {
			schedule := j.Schedule.Expr
			if schedule == "" && j.Schedule.EveryMS != nil {
				schedule = fmt.Sprintf("%dms", *j.Schedule.EveryMS)
			}
			tbl.AddRow(j.ID, j.Name, j.AgentID, schedule, fmt.Sprintf("%v", j.Enabled), j.DeliverChannel)
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
		result, err := ws.CronList(client.CronListParams{IncludeDisabled: true})
		if err != nil {
			return err
		}
		for _, j := range result.Jobs {
			if j.ID == args[0] {
				printer.Print(j)
				return nil
			}
		}
		return fmt.Errorf("cron job %q not found", args[0])
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

		params := client.CronCreateParams{
			Name:    name,
			AgentID: agent,
			Message: message,
			Schedule: client.CronSchedule{
				Kind: "cron",
				Expr: schedule,
				TZ:   timezone,
			},
		}

		result, err := ws.CronCreate(params)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Cron job created: %s", result.Job.ID))
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
		patch := client.CronJobPatch{}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			patch.Name = v
		}
		if cmd.Flags().Changed("schedule") {
			v, _ := cmd.Flags().GetString("schedule")
			patch.Schedule = &client.CronSchedule{Kind: "cron", Expr: v}
		}
		_, err = ws.CronUpdate(client.CronUpdateParams{JobID: args[0], Patch: patch})
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
		_, err = ws.CronDelete(client.CronJobIDParams{JobID: args[0]})
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
		enabled, _ := cmd.Flags().GetBool("enabled")
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.CronToggle(client.CronToggleParams{JobID: args[0], Enabled: enabled})
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
		_, err = ws.CronRun(client.CronRunParams{JobID: args[0], Mode: "force"})
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
		result, err := ws.CronList(client.CronListParams{IncludeDisabled: true})
		if err != nil {
			return err
		}
		for _, j := range result.Jobs {
			if j.ID == args[0] {
				printer.Print(j.State)
				return nil
			}
		}
		return fmt.Errorf("cron job %q not found", args[0])
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
		result, err := ws.CronRuns(client.CronRunsParams{JobID: args[0], Limit: limit})
		if err != nil {
			return err
		}
		printer.Print(result)
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
	cronToggleCmd.Flags().Bool("enabled", true, "Enable or disable the job")
	cronRunsCmd.Flags().Int("limit", 20, "Max results")

	cronCmd.AddCommand(cronListCmd, cronGetCmd, cronCreateCmd, cronUpdateCmd,
		cronDeleteCmd, cronToggleCmd, cronRunCmd, cronStatusCmd, cronRunsCmd)
	rootCmd.AddCommand(cronCmd)
}
