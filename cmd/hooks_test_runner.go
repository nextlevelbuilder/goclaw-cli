package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// hooks_test_runner.go — `hooks test` dry-run subcommand.
// RPC: hooks.test takes {config, sampleEvent} where both are raw JSON.

var hooksTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Dry-run a hook config against a sample event",
	Long: `Test a hook config against a sample event without persisting it.

Required:
  --config=<@file|literal>      Hook config JSON
  --event=<@file|literal>       Sample event JSON

Example:
  goclaw hooks test --config=@hook.json --event=@event.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		evPath, _ := cmd.Flags().GetString("event")
		if cfgPath == "" || evPath == "" {
			return fmt.Errorf("--config and --event are required")
		}
		cfgRaw, err := readContent(cfgPath)
		if err != nil {
			return err
		}
		evRaw, err := readContent(evPath)
		if err != nil {
			return err
		}
		var cfgObj, evObj map[string]any
		if err := json.Unmarshal([]byte(cfgRaw), &cfgObj); err != nil {
			return fmt.Errorf("invalid --config JSON: %w", err)
		}
		if err := json.Unmarshal([]byte(evRaw), &evObj); err != nil {
			return fmt.Errorf("invalid --event JSON: %w", err)
		}

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		data, err := ws.Call("hooks.test", map[string]any{
			"config":      cfgObj,
			"sampleEvent": evObj,
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	hooksTestCmd.Flags().String("config", "", "Hook config JSON (@filepath or literal)")
	hooksTestCmd.Flags().String("event", "", "Sample event JSON (@filepath or literal)")
	_ = hooksTestCmd.MarkFlagRequired("config")
	_ = hooksTestCmd.MarkFlagRequired("event")
}
