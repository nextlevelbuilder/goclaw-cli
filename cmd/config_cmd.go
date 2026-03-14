package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{Use: "config", Short: "Manage server configuration"}

var configGetCmd = &cobra.Command{
	Use: "get", Short: "Get server configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		params := map[string]any{}
		if v, _ := cmd.Flags().GetString("key"); v != "" {
			params["key"] = v
		}
		data, err := ws.Call("config.get", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var configApplyCmd = &cobra.Command{
	Use: "apply", Short: "Apply configuration from file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		filePath, _ := cmd.Flags().GetString("file")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read config file: %w", err)
		}
		var cfg map[string]any
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
		_, err = ws.Call("config.apply", cfg)
		if err != nil {
			return err
		}
		printer.Success("Configuration applied")
		return nil
	},
}

var configPatchCmd = &cobra.Command{
	Use: "patch", Short: "Patch a config key",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		// Try to parse value as JSON, fall back to string
		var parsedVal any
		if err := json.Unmarshal([]byte(value), &parsedVal); err != nil {
			parsedVal = value
		}

		_, err = ws.Call("config.patch", map[string]any{"key": key, "value": parsedVal})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Config %s updated", key))
		return nil
	},
}

var configSchemaCmd = &cobra.Command{
	Use: "schema", Short: "Get config schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("config.schema", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	configGetCmd.Flags().String("key", "", "Config key path")
	configApplyCmd.Flags().String("file", "", "Config JSON file")
	_ = configApplyCmd.MarkFlagRequired("file")
	configPatchCmd.Flags().String("key", "", "Config key")
	configPatchCmd.Flags().String("value", "", "Config value")
	_ = configPatchCmd.MarkFlagRequired("key")
	_ = configPatchCmd.MarkFlagRequired("value")

	configCmd.AddCommand(configGetCmd, configApplyCmd, configPatchCmd, configSchemaCmd)
	rootCmd.AddCommand(configCmd)
}
