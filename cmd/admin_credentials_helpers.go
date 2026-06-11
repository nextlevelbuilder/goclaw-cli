package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

func jsonObjectFlag(cmd *cobra.Command, flag string, required bool) (map[string]any, error) {
	bodyJSON, _ := cmd.Flags().GetString(flag)
	if bodyJSON == "" {
		if required {
			return nil, fmt.Errorf("--%s is required (JSON object)", flag)
		}
		return nil, nil
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
		return nil, fmt.Errorf("invalid --%s JSON: %w", flag, err)
	}
	return body, nil
}

func credentialCreateBody(cmd *cobra.Command) (map[string]any, error) {
	if body, err := jsonObjectFlag(cmd, "body", false); err != nil || body != nil {
		return body, err
	}
	if preset, _ := cmd.Flags().GetString("preset"); preset != "" {
		return map[string]any{"preset": preset}, nil
	}
	if name, _ := cmd.Flags().GetString("name"); name != "" {
		return map[string]any{"binary_name": name, "name": name}, nil
	}
	return nil, fmt.Errorf("--body, --preset, or --name is required")
}

func cliCredentialsTable(data []byte) *output.TableData {
	tbl := output.NewTable("ID", "BINARY", "DESCRIPTION", "ENABLED", "CREATED")
	for _, cr := range listFromResponse(data, "items") {
		tbl.AddRow(str(cr, "id"), credentialBinaryName(cr), str(cr, "description"), str(cr, "enabled"), str(cr, "created_at"))
	}
	return tbl
}

func credentialBinaryName(cr map[string]any) string {
	if binary := str(cr, "binary_name"); binary != "" {
		return binary
	}
	return str(cr, "name")
}

func credentialPresetsTable(data []byte) *output.TableData {
	tbl := output.NewTable("NAME", "BINARY", "DESCRIPTION")
	if list := unmarshalList(data); len(list) > 0 {
		for _, preset := range list {
			tbl.AddRow(str(preset, "name"), str(preset, "binary_name"), str(preset, "description"))
		}
		return tbl
	}
	presets, _ := unmarshalMap(data)["presets"].(map[string]any)
	for _, name := range sortedKeys(presets) {
		preset, _ := presets[name].(map[string]any)
		tbl.AddRow(name, str(preset, "binary_name"), str(preset, "description"))
	}
	return tbl
}

func credentialGrantsTable(data []byte) *output.TableData {
	tbl := output.NewTable("ID", "AGENT", "SCOPE", "ENABLED")
	for _, grant := range listFromResponse(data, "grants") {
		tbl.AddRow(str(grant, "id"), str(grant, "agent_id"), str(grant, "scope"), str(grant, "enabled"))
	}
	return tbl
}

func userCredentialsTable(data []byte) *output.TableData {
	tbl := output.NewTable("USER", "ENV_KEYS", "UPDATED")
	for _, item := range listFromResponse(data, "user_credentials") {
		tbl.AddRow(str(item, "user_id"), joinValueList(item["env_keys"]), str(item, "updated_at"))
	}
	return tbl
}

func agentCredentialsTable(data []byte) *output.TableData {
	tbl := output.NewTable("AGENT", "TYPE", "ENV_KEYS", "UPDATED")
	for _, item := range listFromResponse(data, "agent_credentials") {
		tbl.AddRow(str(item, "agent_id"), str(item, "credential_type"), joinValueList(item["env_keys"]), str(item, "updated_at"))
	}
	return tbl
}

func joinValueList(value any) string {
	return strings.Join(stringListFromValue(value), ",")
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
