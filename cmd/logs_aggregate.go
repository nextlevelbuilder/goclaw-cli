package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// validLogsGroupBy enumerates allowed --group-by values for the runtime logs
// aggregate endpoint.
var validLogsGroupBy = map[string]bool{
	"level":  true,
	"source": true,
}

// logsAggregateCmd queries the runtime log ring-buffer aggregate endpoint.
// Distinct from `logs tail` (WS streaming): this is a one-shot HTTP GET that
// summarizes the in-memory ring buffer by level or source.
//
// Backend route: GET /v1/logs/runtime/aggregate (admin-only on server side).
var logsAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Summarize runtime logs (ring buffer) by level or source",
	Long: `Aggregate the in-memory runtime log ring buffer by --group-by (level or
source). This is a one-shot HTTP query — not a stream. Use 'logs tail' for
real-time streaming.

Backend route: GET /v1/logs/runtime/aggregate (admin-only, server-enforced).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groupBy, _ := cmd.Flags().GetString("group-by")
		if groupBy == "" {
			groupBy = "level"
		}
		if !validLogsGroupBy[groupBy] {
			return fmt.Errorf("--group-by must be one of level, source (got %q)", groupBy)
		}
		from, _ := cmd.Flags().GetString("from")
		if from != "" {
			if _, err := time.Parse(time.RFC3339, from); err != nil {
				return fmt.Errorf("--from must be RFC3339: %w", err)
			}
		}

		q := url.Values{}
		q.Set("group_by", groupBy)
		for flagName, queryKey := range map[string]string{
			"level":  "level",
			"source": "source",
			"from":   "from",
		} {
			if v, _ := cmd.Flags().GetString(flagName); v != "" {
				q.Set(queryKey, v)
			}
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/logs/runtime/aggregate?" + q.Encode())
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		// Summary row (source, retention, capacity, sample_size).
		summary := output.NewTable("SOURCE", "RETENTION", "CAPACITY", "SAMPLE_SIZE")
		summary.AddRow(str(m, "source"), str(m, "retention"),
			str(m, "capacity"), str(m, "sample_size"))
		printer.Print(summary)
		// Bucket rows.
		buckets, _ := m["buckets"].([]any)
		tbl := output.NewTable("KEY", "COUNT", "LAST_SEEN")
		for _, raw := range buckets {
			row, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			tbl.AddRow(str(row, "key"), str(row, "count"), formatLastSeen(row["last_seen"]))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	logsAggregateCmd.Flags().String("group-by", "level", "Grouping dimension: level | source (default level)")
	logsAggregateCmd.Flags().String("level", "", "Filter by level: debug | info | warn | error")
	logsAggregateCmd.Flags().String("source", "", "Filter by source")
	logsAggregateCmd.Flags().String("from", "", "RFC3339 start of time window")
	logsCmd.AddCommand(logsAggregateCmd)
}
