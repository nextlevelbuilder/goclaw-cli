package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var tracesCmd = &cobra.Command{Use: "traces", Short: "View LLM traces"}

var tracesListCmd = &cobra.Command{
	Use: "list", Short: "List traces",
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
		if v, _ := cmd.Flags().GetString("session-key"); v != "" {
			q.Set("session_key", v)
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			q.Set("status", v)
		}
		if v, _ := cmd.Flags().GetString("channel"); v != "" {
			q.Set("channel", v)
		}
		setStringQueryFlag(cmd, q, "q", "q")
		setStringQueryFlag(cmd, q, "agent-query", "agent")
		setStringQueryFlag(cmd, q, "channel-query", "channel_query")
		setStringQueryFlag(cmd, q, "from", "from")
		setStringQueryFlag(cmd, q, "to", "to")
		setChangedIntQueryFlag(cmd, q, "min-input-tokens", "min_input_tokens")
		setChangedIntQueryFlag(cmd, q, "max-input-tokens", "max_input_tokens")
		setChangedIntQueryFlag(cmd, q, "min-output-tokens", "min_output_tokens")
		setChangedIntQueryFlag(cmd, q, "max-output-tokens", "max_output_tokens")
		setChangedIntQueryFlag(cmd, q, "min-tool-calls", "min_tool_calls")
		setChangedIntQueryFlag(cmd, q, "max-tool-calls", "max_tool_calls")
		setStringQueryFlag(cmd, q, "tool-name", "tool_name")
		setStringQueryFlag(cmd, q, "has-tool-calls", "has_tool_calls")
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		if v, _ := cmd.Flags().GetInt("offset"); v > 0 {
			q.Set("offset", fmt.Sprintf("%d", v))
		}
		if cmd.Flags().Changed("since") {
			return &client.APIError{Code: "INVALID_REQUEST", Message: "traces list no longer supports --since; use traces follow --since for incremental polling"}
		}
		if cmd.Flags().Changed("root-only") {
			return &client.APIError{Code: "INVALID_REQUEST", Message: "traces list no longer supports --root-only; the server trace list has no root-only filter"}
		}
		path := "/v1/traces"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		envelope, rows, err := decodeTraceListPayload(data)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(envelope)
			return nil
		}
		printTraceRowsTable(rows)
		return nil
	},
}

var tracesGetCmd = &cobra.Command{
	Use: "get <traceID>", Short: "Get trace with span tree", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := strings.TrimSpace(args[0])
		if err := validateTraceID(id); err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/traces/" + url.PathEscape(id))
		if err != nil {
			return err
		}
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("decode trace payload: %w", err)
		}
		if cfg.OutputFormat != "table" {
			printer.Print(payload)
			return nil
		}
		renderTraceTable(payload, os.Stdout)
		return nil
	},
}

// traceIDPattern restricts trace ids to a safe, URL-safe allowlist.
// Blocks path-traversal (`..`, `/`, `\`), control characters, and whitespace
// before any HTTP call is issued. PathEscape is still applied on top.
var traceIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func validateTraceID(id string) error {
	if id == "" || id == "." || id == ".." {
		return &client.APIError{Code: "INVALID_REQUEST", Message: "trace id is empty or reserved"}
	}
	if !traceIDPattern.MatchString(id) {
		return &client.APIError{Code: "INVALID_REQUEST", Message: "trace id contains invalid characters (allowed: A-Z a-z 0-9 . _ -)"}
	}
	return nil
}

func setStringQueryFlag(cmd *cobra.Command, q url.Values, flagName, queryName string) {
	v, _ := cmd.Flags().GetString(flagName)
	if v != "" {
		q.Set(queryName, v)
	}
}

func setChangedIntQueryFlag(cmd *cobra.Command, q url.Values, flagName, queryName string) {
	if !cmd.Flags().Changed(flagName) {
		return
	}
	v, _ := cmd.Flags().GetInt(flagName)
	q.Set(queryName, fmt.Sprintf("%d", v))
}

var tracesExportCmd = &cobra.Command{
	Use: "export <traceID>", Short: "Export trace to file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := strings.TrimSpace(args[0])
		if err := validateTraceID(id); err != nil {
			return err
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		if outFile == "" {
			outFile = id + ".json.gz"
		}
		resp, err := c.GetRaw("/v1/traces/" + url.PathEscape(id) + "/export")
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return rawResponseError(resp)
		}
		defer resp.Body.Close()
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer f.Close()
		n, err := io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Exported %d bytes to %s", n, outFile))
		return nil
	},
}

// --- Usage/Costs ---

var usageCmd = &cobra.Command{Use: "usage", Short: "View usage and cost analytics"}

var usageSummaryCmd = &cobra.Command{
	Use: "summary", Short: "Usage summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("from"); v != "" {
			q.Set("from", v)
		}
		if v, _ := cmd.Flags().GetString("to"); v != "" {
			q.Set("to", v)
		}
		path := "/v1/usage/summary"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var usageDetailCmd = &cobra.Command{
	Use: "detail", Short: "Detailed usage breakdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		if v, _ := cmd.Flags().GetString("provider"); v != "" {
			q.Set("provider", v)
		}
		if v, _ := cmd.Flags().GetString("from"); v != "" {
			q.Set("from", v)
		}
		if v, _ := cmd.Flags().GetString("to"); v != "" {
			q.Set("to", v)
		}
		path := "/v1/usage"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var usageCostsCmd = &cobra.Command{
	Use: "costs", Short: "Cost summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/costs/summary")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var usageTimeseriesCmd = &cobra.Command{
	Use: "timeseries", Short: "Token usage over time (GET /v1/usage/timeseries)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("start"); v != "" {
			q.Set("from", normalizeUsageTimestamp(v))
		}
		if v, _ := cmd.Flags().GetString("end"); v != "" {
			q.Set("to", normalizeUsageTimestamp(v))
		}
		if v, _ := cmd.Flags().GetString("granularity"); v != "" {
			q.Set("group_by", v)
		}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		path := "/v1/usage/timeseries"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var usageBreakdownCmd = &cobra.Command{
	Use: "breakdown", Short: "Usage broken down by dimension (GET /v1/usage/breakdown)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("by"); v != "" {
			q.Set("group_by", v)
		}
		if v, _ := cmd.Flags().GetString("start"); v != "" {
			q.Set("from", normalizeUsageTimestamp(v))
		}
		if v, _ := cmd.Flags().GetString("end"); v != "" {
			q.Set("to", normalizeUsageTimestamp(v))
		}
		path := "/v1/usage/breakdown"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func normalizeUsageTimestamp(v string) string {
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t.Format(time.RFC3339)
	}
	return v
}

func init() {
	tracesListCmd.Flags().String("agent", "", "Filter by agent ID")
	tracesListCmd.Flags().String("user", "", "Filter by user ID")
	tracesListCmd.Flags().String("session-key", "", "Filter by session key")
	tracesListCmd.Flags().String("status", "", "Filter: running, success, error")
	tracesListCmd.Flags().String("channel", "", "Filter by channel")
	tracesListCmd.Flags().String("q", "", "Search trace id, previews, session/channel labels, agent/channel labels, and span previews")
	tracesListCmd.Flags().String("agent-query", "", "Search agent display name or key")
	tracesListCmd.Flags().String("channel-query", "", "Search channel instance name, display name, or type")
	tracesListCmd.Flags().String("from", "", "Filter traces from RFC3339 start time")
	tracesListCmd.Flags().String("to", "", "Filter traces up to RFC3339 start time")
	tracesListCmd.Flags().Int("min-input-tokens", 0, "Minimum total input tokens")
	tracesListCmd.Flags().Int("max-input-tokens", 0, "Maximum total input tokens")
	tracesListCmd.Flags().Int("min-output-tokens", 0, "Minimum total output tokens")
	tracesListCmd.Flags().Int("max-output-tokens", 0, "Maximum total output tokens")
	tracesListCmd.Flags().Int("min-tool-calls", 0, "Minimum tool-call count")
	tracesListCmd.Flags().Int("max-tool-calls", 0, "Maximum tool-call count")
	tracesListCmd.Flags().String("tool-name", "", "Search span tool names")
	tracesListCmd.Flags().String("has-tool-calls", "", "Filter traces by tool-call presence: true or false")
	tracesListCmd.Flags().Int("limit", 20, "Max results")
	tracesListCmd.Flags().Int("offset", 0, "Pagination offset")
	tracesListCmd.Flags().String("since", "", "Deprecated: use traces follow --since")
	tracesListCmd.Flags().Bool("root-only", false, "Deprecated: unsupported by server trace list")
	_ = tracesListCmd.Flags().MarkHidden("since")
	_ = tracesListCmd.Flags().MarkHidden("root-only")
	tracesExportCmd.Flags().StringP("output", "f", "", "Output file (default: <traceID>.json.gz)")

	usageSummaryCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	usageSummaryCmd.Flags().String("to", "", "End date")
	usageDetailCmd.Flags().String("agent", "", "Agent ID")
	usageDetailCmd.Flags().String("provider", "", "Provider name")
	usageDetailCmd.Flags().String("from", "", "Start date")
	usageDetailCmd.Flags().String("to", "", "End date")

	usageTimeseriesCmd.Flags().String("start", "", "Start date or RFC3339 timestamp")
	usageTimeseriesCmd.Flags().String("end", "", "End date or RFC3339 timestamp")
	usageTimeseriesCmd.Flags().String("granularity", "day", "Group by: provider|model|channel|agent|day")
	usageTimeseriesCmd.Flags().String("agent", "", "Filter by agent")
	usageTimeseriesCmd.Flags().String("user", "", "Filter by user")
	usageTimeseriesCmd.Flags().String("tenant", "", "Filter by tenant")
	usageBreakdownCmd.Flags().String("by", "agent", "Dimension: provider|model|channel|agent|day")
	usageBreakdownCmd.Flags().String("start", "", "Start date or RFC3339 timestamp")
	usageBreakdownCmd.Flags().String("end", "", "End date or RFC3339 timestamp")

	tracesCmd.AddCommand(tracesListCmd, tracesGetCmd, tracesExportCmd)
	usageCmd.AddCommand(usageSummaryCmd, usageDetailCmd, usageCostsCmd,
		usageTimeseriesCmd, usageBreakdownCmd)
	rootCmd.AddCommand(tracesCmd, usageCmd)
}
