package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"

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
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/traces/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
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
		resp, err := c.GetRaw("/v1/traces/" + url.PathEscape(args[0]) + "/export")
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

func init() {
	tracesListCmd.Flags().String("agent", "", "Filter by agent ID")
	tracesListCmd.Flags().String("status", "", "Filter: running, success, error")
	tracesListCmd.Flags().Int("limit", 20, "Max results")
	tracesExportCmd.Flags().StringP("output", "f", "", "Output file (default: <traceID>.json.gz)")

	tracesCmd.AddCommand(tracesListCmd, tracesGetCmd, tracesExportCmd)
	rootCmd.AddCommand(tracesCmd)
}
