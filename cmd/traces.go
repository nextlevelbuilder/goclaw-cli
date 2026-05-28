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
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
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
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			q.Set("status", v)
		}
		if v, _ := cmd.Flags().GetString("since"); v != "" {
			q.Set("since", v)
		}
		if v, _ := cmd.Flags().GetBool("root-only"); v {
			q.Set("root_only", "true")
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		path := "/v1/traces"
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
		tbl := output.NewTable("TRACE_ID", "AGENT", "STATUS", "DURATION_MS", "INPUT_TOKENS", "OUTPUT_TOKENS", "COST")
		for _, t := range unmarshalList(data) {
			tbl.AddRow(str(t, "trace_id"), str(t, "agent_id"), str(t, "status"),
				str(t, "duration_ms"), str(t, "input_tokens"), str(t, "output_tokens"), str(t, "cost"))
		}
		printer.Print(tbl)
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
		var trace map[string]any
		if err := json.Unmarshal(data, &trace); err != nil {
			return fmt.Errorf("decode trace payload: %w", err)
		}
		if cfg.OutputFormat != "table" {
			printer.Print(trace)
			return nil
		}
		renderTraceTable(trace, os.Stdout)
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

// renderTraceTable prints a human-readable summary: header card, span tree, events.
func renderTraceTable(t map[string]any, w io.Writer) {
	for _, row := range [][2]string{
		{"TRACE_ID", str(t, "trace_id")}, {"AGENT_ID", str(t, "agent_id")},
		{"SESSION_KEY", str(t, "session_key")}, {"STATUS", str(t, "status")},
		{"DURATION_MS", str(t, "duration_ms")},
	} {
		if row[1] != "" {
			fmt.Fprintf(w, "%-12s %s\n", row[0]+":", row[1])
		}
	}
	if in, out, cost := str(t, "input_tokens"), str(t, "output_tokens"), str(t, "cost"); in+out+cost != "" {
		fmt.Fprintf(w, "%-12s in=%s out=%s cost=%s\n", "TOKENS:", in, out, cost)
	}
	spans, _ := t["spans"].([]any)
	if len(spans) == 0 {
		fmt.Fprintln(w, "\nSPANS: (none)")
	} else {
		fmt.Fprintln(w, "\nSPANS:")
		output.PrintTreeRoot(buildSpanTree(spans), w)
	}
	events, _ := t["events"].([]any)
	fmt.Fprintf(w, "\nEVENTS (n=%d):\n", len(events))
	for _, e := range events {
		if m, ok := e.(map[string]any); ok {
			fmt.Fprintf(w, "  - %s\n", str(m, "type"))
		}
	}
}

// buildSpanTree links spans via parent_span_id; spans whose parent isn't in this
// trace attach to a virtual root. Children are kept in insertion order.
func buildSpanTree(spans []any) output.TreeNode {
	order := make([]string, 0, len(spans))
	labels := make(map[string]string, len(spans))
	children := make(map[string][]string, len(spans))
	parentOf := make(map[string]string, len(spans))
	for _, s := range spans {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		id := str(m, "span_id")
		if id == "" {
			continue
		}
		label := id
		if name := str(m, "name"); name != "" {
			label = name + " [" + id + "]"
		}
		if kind := str(m, "kind"); kind != "" {
			label += " kind=" + kind
		}
		if dur := str(m, "duration_ms"); dur != "" {
			label += " " + dur + "ms"
		}
		labels[id] = label
		order = append(order, id)
		parentOf[id], _ = m["parent_span_id"].(string)
	}
	for _, id := range order {
		if p := parentOf[id]; p != "" {
			if _, ok := labels[p]; ok {
				children[p] = append(children[p], id)
				continue
			}
		}
		children[""] = append(children[""], id)
	}
	var build func(id string) output.TreeNode
	build = func(id string) output.TreeNode {
		n := output.TreeNode{Name: labels[id]}
		for _, c := range children[id] {
			n.Children = append(n.Children, build(c))
		}
		return n
	}
	root := output.TreeNode{Name: "trace"}
	for _, id := range children[""] {
		root.Children = append(root.Children, build(id))
	}
	return root
}

var tracesExportCmd = &cobra.Command{
	Use: "export <traceID>", Short: "Export trace to file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		if outFile == "" {
			outFile = args[0] + ".json.gz"
		}
		resp, err := c.GetRaw("/v1/traces/" + args[0] + "/export")
		if err != nil {
			return err
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
	tracesListCmd.Flags().String("status", "", "Filter: running, success, error")
	tracesListCmd.Flags().String("since", "", "Filter by relative or ISO timestamp, e.g. 1h or 2026-05-19T00:00:00Z")
	tracesListCmd.Flags().Bool("root-only", false, "Only show root traces")
	tracesListCmd.Flags().Int("limit", 20, "Max results")
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
