package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// tracesFollowCmd issues a single polling GET to /v1/traces/follow. NOT a watch loop.
// Operators wanting continuous follow rerun the command with the returned `next_since`.
var tracesFollowCmd = &cobra.Command{
	Use:   "follow",
	Short: "Poll incremental trace activity (one shot)",
	Long: `Poll incremental trace activity for a session or agent.

Exactly one of --session-key or --agent must be provided. This is a one-shot
polling request — no watch loop. Use the returned ` + "`next_since`" + ` to re-poll.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionKey, _ := cmd.Flags().GetString("session-key")
		agent, _ := cmd.Flags().GetString("agent")
		if sessionKey == "" && agent == "" {
			return fmt.Errorf("exactly one of --session-key or --agent is required")
		}
		if sessionKey != "" && agent != "" {
			return fmt.Errorf("--session-key and --agent are mutually exclusive")
		}

		since, _ := cmd.Flags().GetString("since")
		if since != "" {
			if _, err := time.Parse(time.RFC3339, since); err != nil {
				return fmt.Errorf("--since must be RFC3339: %w", err)
			}
		}

		q := url.Values{}
		if sessionKey != "" {
			q.Set("session_key", sessionKey)
		}
		if agent != "" {
			q.Set("agent_id", agent)
		}
		if since != "" {
			q.Set("since", since)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			q.Set("status", v)
		}
		if v, _ := cmd.Flags().GetString("channel"); v != "" {
			q.Set("channel", v)
		}
		if v, _ := cmd.Flags().GetBool("include-spans"); v {
			q.Set("include_spans", "true")
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/traces/follow"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		envelope := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(envelope)
			return nil
		}
		traces, _ := envelope["traces"].([]any)
		tbl := output.NewTable("TRACE_ID", "AGENT", "STATUS", "DURATION_MS", "INPUT_TOKENS", "OUTPUT_TOKENS", "COST")
		for _, raw := range traces {
			t, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			tbl.AddRow(str(t, "trace_id"), str(t, "agent_id"), str(t, "status"),
				str(t, "duration_ms"), str(t, "input_tokens"), str(t, "output_tokens"), str(t, "cost"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	tracesFollowCmd.Flags().String("session-key", "", "Session key to follow")
	tracesFollowCmd.Flags().String("agent", "", "Agent id or key to follow")
	tracesFollowCmd.Flags().String("since", "", "RFC3339 timestamp; only traces after this are returned")
	tracesFollowCmd.Flags().Int("limit", 0, "Max traces (server default 50, max 200)")
	tracesFollowCmd.Flags().String("status", "", "Filter by status")
	tracesFollowCmd.Flags().String("channel", "", "Filter by channel")
	tracesFollowCmd.Flags().Bool("include-spans", false, "Include spans_by_trace_id in response")
	tracesCmd.AddCommand(tracesFollowCmd)
}
