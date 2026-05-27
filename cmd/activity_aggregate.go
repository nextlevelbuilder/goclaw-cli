package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// validActivityGroupBy enumerates allowed --group-by values for the activity
// aggregate endpoint. Server enforces admin-only for actor_id; the CLI does
// not pre-check that — it only validates the enum.
var validActivityGroupBy = map[string]bool{
	"action":      true,
	"actor_type":  true,
	"entity_type": true,
	"actor_id":    true,
}

// formatLastSeen renders an aggregate bucket's last_seen field as RFC3339.
//
// The activity aggregate endpoint returns last_seen as an RFC3339 string,
// but the logs runtime aggregate endpoint returns last_seen as epoch millis
// (a number). `unmarshalMap` decodes JSON numbers as float64, and the shared
// `str()` helper renders large float64 as scientific notation (e.g.
// "1.76e+12"). This helper type-switches so both endpoints render
// consistently as RFC3339 strings in the table view.
func formatLastSeen(v any) string {
	switch t := v.(type) {
	case nil:
		return "-"
	case string:
		if t == "" {
			return "-"
		}
		return t
	case float64:
		if t == 0 {
			return "-"
		}
		return time.UnixMilli(int64(t)).UTC().Format(time.RFC3339)
	case int64:
		if t == 0 {
			return "-"
		}
		return time.UnixMilli(t).UTC().Format(time.RFC3339)
	case int:
		if t == 0 {
			return "-"
		}
		return time.UnixMilli(int64(t)).UTC().Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// activityAggregateCmd groups audit-log activity by a dimension and returns
// bucket counts. Attached as a subcommand of the existing activityCmd
// (declared in cmd/admin.go) so the top-level command surface is unchanged.
//
// Backend route: GET /v1/activity/aggregate
var activityAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate audit-log activity by a grouping dimension",
	Long: `Group activity log entries by a dimension (action, actor_type, entity_type,
or actor_id) and return bucket counts with last_seen timestamps.

Optional filters narrow the result set: --from/--to (RFC3339 window),
--actor-type, --actor-id, --action, --entity-type, --entity-id, --limit.

Backend route: GET /v1/activity/aggregate
Note: --group-by=actor_id requires admin privileges (enforced server-side).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groupBy, _ := cmd.Flags().GetString("group-by")
		if groupBy == "" {
			return fmt.Errorf("--group-by is required (one of action, actor_type, entity_type, actor_id)")
		}
		if !validActivityGroupBy[groupBy] {
			return fmt.Errorf("--group-by must be one of action, actor_type, entity_type, actor_id (got %q)", groupBy)
		}
		from, _ := cmd.Flags().GetString("from")
		if from != "" {
			if _, err := time.Parse(time.RFC3339, from); err != nil {
				return fmt.Errorf("--from must be RFC3339: %w", err)
			}
		}
		to, _ := cmd.Flags().GetString("to")
		if to != "" {
			if _, err := time.Parse(time.RFC3339, to); err != nil {
				return fmt.Errorf("--to must be RFC3339: %w", err)
			}
		}

		q := url.Values{}
		q.Set("group_by", groupBy)
		if from != "" {
			q.Set("from", from)
		}
		if to != "" {
			q.Set("to", to)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		for flagName, queryKey := range map[string]string{
			"actor-type":  "actor_type",
			"actor-id":    "actor_id",
			"action":      "action",
			"entity-type": "entity_type",
			"entity-id":   "entity_id",
		} {
			if v, _ := cmd.Flags().GetString(flagName); v != "" {
				q.Set(queryKey, v)
			}
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/activity/aggregate?" + q.Encode())
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
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
	activityAggregateCmd.Flags().String("group-by", "", "Grouping dimension: action | actor_type | entity_type | actor_id (required)")
	activityAggregateCmd.Flags().String("from", "", "RFC3339 start of time window")
	activityAggregateCmd.Flags().String("to", "", "RFC3339 end of time window")
	activityAggregateCmd.Flags().Int("limit", 0, "Maximum buckets to return (server default applied if 0)")
	activityAggregateCmd.Flags().String("actor-type", "", "Filter by actor type")
	activityAggregateCmd.Flags().String("actor-id", "", "Filter by actor id")
	activityAggregateCmd.Flags().String("action", "", "Filter by action")
	activityAggregateCmd.Flags().String("entity-type", "", "Filter by entity type")
	activityAggregateCmd.Flags().String("entity-id", "", "Filter by entity id")
	_ = activityAggregateCmd.MarkFlagRequired("group-by")
	activityCmd.AddCommand(activityAggregateCmd)
}
