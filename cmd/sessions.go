package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage chat sessions",
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		if v, _ := cmd.Flags().GetString("user"); v != "" {
			q.Set("user_id", v)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		path := "/v1/sessions"
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
		tbl := output.NewTable("KEY", "AGENT", "USER", "LABEL", "INPUT_TOKENS", "OUTPUT_TOKENS")
		for _, s := range unmarshalList(data) {
			tbl.AddRow(str(s, "session_key"), str(s, "agent_id"), str(s, "user_id"),
				str(s, "label"), str(s, "input_tokens"), str(s, "output_tokens"))
		}
		printer.Print(tbl)
		return nil
	},
}

var sessionsPreviewCmd = &cobra.Command{
	Use:   "preview <sessionKey>",
	Short: "Preview session messages",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/sessions/"+args[0]+"/preview", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete <sessionKey>",
	Short: "Delete a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete session %s?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/sessions/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Session deleted")
		return nil
	},
}

var sessionsResetCmd = &cobra.Command{
	Use:   "reset <sessionKey>",
	Short: "Reset a session (clear messages)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Reset session %s?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/sessions/"+args[0]+"/reset", nil)
		if err != nil {
			return err
		}
		printer.Success("Session reset")
		return nil
	},
}

var sessionsLabelCmd = &cobra.Command{
	Use:   "label <sessionKey>",
	Short: "Set session label",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		label, _ := cmd.Flags().GetString("label")
		_, err = c.Patch("/v1/sessions/"+args[0], map[string]any{"label": label})
		if err != nil {
			return err
		}
		printer.Success("Session labeled")
		return nil
	},
}

func init() {
	sessionsListCmd.Flags().String("agent", "", "Filter by agent ID")
	sessionsListCmd.Flags().String("user", "", "Filter by user ID")
	sessionsListCmd.Flags().Int("limit", 0, "Max results")
	sessionsLabelCmd.Flags().String("label", "", "Session label")
	_ = sessionsLabelCmd.MarkFlagRequired("label")

	sessionsCmd.AddCommand(sessionsListCmd, sessionsPreviewCmd, sessionsDeleteCmd, sessionsResetCmd, sessionsLabelCmd)
	rootCmd.AddCommand(sessionsCmd)
}
