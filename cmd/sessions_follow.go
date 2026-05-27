package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// sessionsFollowCmd issues one HTTP GET to /v1/chat/sessions/{key}/history/follow.
// NOT a watch loop and NOT a WS stream — operators wanting continuous follow
// rerun with the returned `next_cursor`.
//
// Backend route: GET /v1/chat/sessions/{key}/history/follow (chat domain).
//
// Query string is built directly via url.Values so cursor=0 is preserved
// (buildBody would drop int v == 0).
var sessionsFollowCmd = &cobra.Command{
	Use:   "follow <sessionKey>",
	Short: "Poll cursor-based session history (one shot)",
	Long: `Poll the next batch of session-history messages from a cursor. One-shot
polling — no watch loop, no SSE, no WebSocket stream. Re-invoke with the
returned ` + "`next_cursor`" + ` to advance.

Backend route: GET /v1/chat/sessions/{key}/history/follow (chat domain).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cursor, _ := cmd.Flags().GetInt("cursor")
		limit, _ := cmd.Flags().GetInt("limit")
		if cursor < 0 {
			return fmt.Errorf("--cursor must be >= 0 (got %d)", cursor)
		}
		if limit <= 0 {
			return fmt.Errorf("--limit must be > 0 (got %d)", limit)
		}

		q := url.Values{}
		// Build query directly so cursor=0 appears literally.
		q.Set("cursor", fmt.Sprintf("%d", cursor))
		q.Set("limit", fmt.Sprintf("%d", limit))

		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/chat/sessions/" + url.PathEscape(args[0]) + "/history/follow?" + q.Encode()
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		// Summary row first.
		summary := output.NewTable("SESSION", "CURSOR", "NEXT_CURSOR", "TOTAL", "RESET", "UPDATED")
		summary.AddRow(str(m, "session_key"), str(m, "cursor"),
			str(m, "next_cursor"), str(m, "total"), str(m, "reset"), str(m, "updated"))
		printer.Print(summary)
		// Compact message rows.
		msgs, _ := m["messages"].([]any)
		if len(msgs) > 0 {
			tbl := output.NewTable("INDEX", "ROLE", "CONTENT")
			for _, raw := range msgs {
				row, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				tbl.AddRow(str(row, "index"), str(row, "role"), str(row, "content"))
			}
			printer.Print(tbl)
		}
		return nil
	},
}

func init() {
	sessionsFollowCmd.Flags().Int("cursor", 0, "Starting cursor (>=0, default 0)")
	sessionsFollowCmd.Flags().Int("limit", 50, "Max messages per call (default 50, server max 200)")
	sessionsCmd.AddCommand(sessionsFollowCmd)
}
