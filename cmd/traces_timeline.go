package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var tracesTimelineCmd = &cobra.Command{
	Use:   "timeline <runID>",
	Short: "Read archived run timeline items",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runID := strings.TrimSpace(args[0])
		if err := validateRunID(runID); err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("session-key"); v != "" {
			q.Set("session_key", v)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		if v, _ := cmd.Flags().GetInt("offset"); v > 0 {
			q.Set("offset", fmt.Sprintf("%d", v))
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/runs/" + url.PathEscape(runID) + "/timeline"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		var envelope map[string]any
		if err := json.Unmarshal(data, &envelope); err != nil {
			return fmt.Errorf("decode run timeline payload: %w", err)
		}
		if cfg.OutputFormat != "table" {
			printer.Print(envelope)
			return nil
		}
		items, _ := envelope["items"].([]any)
		tbl := output.NewTable("SEQ", "TYPE", "STATUS", "TITLE", "TOOL", "TRACE_ID", "SPAN_ID", "CREATED_AT")
		for _, raw := range items {
			item, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			tbl.AddRow(str(item, "seq"), str(item, "item_type"), str(item, "status"),
				str(item, "title"), str(item, "tool_name"), str(item, "trace_id"),
				str(item, "span_id"), str(item, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

func validateRunID(id string) error {
	if id == "" || id == "." || id == ".." {
		return &client.APIError{Code: "INVALID_REQUEST", Message: "run id is empty or reserved"}
	}
	if !traceIDPattern.MatchString(id) {
		return &client.APIError{Code: "INVALID_REQUEST", Message: "run id contains invalid characters (allowed: A-Z a-z 0-9 . _ -)"}
	}
	return nil
}

func init() {
	tracesTimelineCmd.Flags().String("session-key", "", "Optional session key scope")
	tracesTimelineCmd.Flags().Int("limit", 0, "Max timeline items (server default 200, max 500)")
	tracesTimelineCmd.Flags().Int("offset", 0, "Pagination offset")
	tracesCmd.AddCommand(tracesTimelineCmd)
}
