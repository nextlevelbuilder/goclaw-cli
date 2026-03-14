package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check server health",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Server == "" {
			return client.ErrServerRequired
		}
		c := client.NewHTTPClient(cfg.Server, cfg.Token, cfg.Insecure)
		if err := c.HealthCheck(); err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Server %s is healthy", cfg.Server))
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status and metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Server == "" {
			return client.ErrServerRequired
		}
		if cfg.Token == "" {
			return client.ErrNotAuthenticated
		}

		// Use WebSocket to call status method
		ws := client.NewWSClient(cfg.Server, cfg.Token, "cli", cfg.Insecure)
		resp, err := ws.Connect()
		if err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		defer ws.Close()

		statusResp, err := ws.Call("status", nil)
		if err != nil {
			// Fallback: show connect response
			if resp != nil {
				printer.Print(jsonToMap(*resp))
				return nil
			}
			return fmt.Errorf("status: %w", err)
		}

		if cfg.OutputFormat == "json" || cfg.OutputFormat == "yaml" {
			printer.Print(jsonToMap(statusResp))
			return nil
		}

		// Parse and display as table
		var info map[string]any
		if err := json.Unmarshal(statusResp, &info); err != nil {
			printer.Print(jsonToMap(statusResp))
			return nil
		}

		tbl := output.NewTable("FIELD", "VALUE")
		for k, v := range info {
			tbl.AddRow(k, fmt.Sprintf("%v", v))
		}
		printer.Print(tbl)
		return nil
	},
}

// jsonToMap converts raw JSON to a map for printing.
func jsonToMap(data json.RawMessage) map[string]any {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]any{"raw": string(data)}
	}
	return m
}

func init() {
	rootCmd.AddCommand(healthCmd, statusCmd)
}
