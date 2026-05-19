package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func parseToolInvokeParams(cmd *cobra.Command) (map[string]any, error) {
	paramPairs, _ := cmd.Flags().GetStringSlice("param")
	paramsJSON, _ := cmd.Flags().GetString("params")
	argsJSON, _ := cmd.Flags().GetString("args")
	paramsChanged := cmd.Flags().Changed("params")
	argsChanged := cmd.Flags().Changed("args")
	if paramsChanged && argsChanged && paramsJSON != argsJSON {
		return nil, fmt.Errorf("use only one of --params or --args")
	}

	rawJSON := paramsJSON
	flagName := "--params"
	if rawJSON == "" && argsJSON != "" {
		rawJSON = argsJSON
		flagName = "--args"
	}

	params := make(map[string]any)
	if rawJSON != "" {
		content, err := readContent(rawJSON)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(content), &params); err != nil {
			return nil, fmt.Errorf("invalid %s JSON: %w", flagName, err)
		}
	}
	for _, pair := range paramPairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}
	return params, nil
}
