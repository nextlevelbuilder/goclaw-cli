package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// hooks.go — manage event hooks via WS RPC `hooks.*`.
// Server: internal/gateway/methods/hooks.go (list/create/update/delete/toggle/test/history).
// `test` runner is handled in hooks_test_runner.go to keep this file under 200 LoC.

var hooksCmd = &cobra.Command{Use: "hooks", Short: "Manage event hooks"}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List hooks",
	Long: `List hooks with optional filters.

Filters:
  --event=<event>     Filter by hook event (e.g. pre_tool_use, post_chat)
  --scope=<scope>     Filter by scope: global, tenant, agent
  --agent=<agentId>   Filter by agent UUID
  --enabled           Only enabled hooks (omit flag for all)`,
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
		if v, _ := cmd.Flags().GetString("event"); v != "" {
			params["event"] = v
		}
		if v, _ := cmd.Flags().GetString("scope"); v != "" {
			params["scope"] = v
		}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			params["agentId"] = v
		}
		if cmd.Flags().Changed("enabled") {
			v, _ := cmd.Flags().GetBool("enabled")
			params["enabled"] = v
		}

		data, err := ws.Call("hooks.list", params)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		hooks, _ := m["hooks"].([]any)
		if cfg.OutputFormat != "table" {
			printer.Print(hooks)
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "EVENT", "SCOPE", "HANDLER", "PRIORITY", "ENABLED")
		for _, h := range hooks {
			row, _ := h.(map[string]any)
			if row == nil {
				continue
			}
			tbl.AddRow(str(row, "id"), str(row, "name"), str(row, "event"),
				str(row, "scope"), str(row, "handler_type"),
				str(row, "priority"), str(row, "enabled"))
		}
		printer.Print(tbl)
		return nil
	},
}

var hooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a hook",
	Long: `Create a hook from a JSON config file.

Required fields: event, scope, handler_type, name.
Optional: matcher, if_expr, timeout_ms, on_timeout, priority, agent_ids, config, metadata.

Example:
  goclaw hooks create --config=@hook.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("config")
		if path == "" {
			return fmt.Errorf("--config is required")
		}
		raw, err := readContent(path)
		if err != nil {
			return err
		}
		var cfg map[string]any
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			return fmt.Errorf("invalid JSON in --config: %w", err)
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("hooks.create", cfg)
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Hook created: %s", str(unmarshalMap(data), "hookId")))
		return nil
	},
}

var hooksUpdateCmd = &cobra.Command{
	Use:   "update <hookId>",
	Short: "Update hook fields (partial patch)",
	Long: `Apply a partial JSON patch to a hook.

Allowed fields: name, agent_ids, event, scope, handler_type, matcher, if_expr,
timeout_ms, on_timeout, priority, enabled, config, metadata.

Example:
  goclaw hooks update <id> --patch=@updates.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("patch")
		if path == "" {
			return fmt.Errorf("--patch is required")
		}
		raw, err := readContent(path)
		if err != nil {
			return err
		}
		var updates map[string]any
		if err := json.Unmarshal([]byte(raw), &updates); err != nil {
			return fmt.Errorf("invalid JSON in --patch: %w", err)
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("hooks.update", map[string]any{"hookId": args[0], "updates": updates})
		if err != nil {
			return err
		}
		printer.Success("Hook updated")
		return nil
	},
}

var hooksDeleteCmd = &cobra.Command{
	Use:   "delete <hookId>",
	Short: "Delete a hook (requires --yes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this hook?", cfg.Yes) {
			return nil
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("hooks.delete", map[string]any{"hookId": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Hook deleted")
		return nil
	},
}

var hooksToggleCmd = &cobra.Command{
	Use:   "toggle <hookId>",
	Short: "Enable/disable a hook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enabled, _ := cmd.Flags().GetBool("enabled")
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("hooks.toggle", map[string]any{"hookId": args[0], "enabled": enabled})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Hook toggled (enabled=%v)", enabled))
		return nil
	},
}

var hooksHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show hook execution history",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("hooks.history", map[string]any{})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	hooksListCmd.Flags().String("event", "", "Filter by event")
	hooksListCmd.Flags().String("scope", "", "Filter by scope (global|tenant|agent)")
	hooksListCmd.Flags().String("agent", "", "Filter by agent UUID")
	hooksListCmd.Flags().Bool("enabled", false, "Only enabled hooks")

	hooksCreateCmd.Flags().String("config", "", "Hook config JSON (@filepath or literal)")
	_ = hooksCreateCmd.MarkFlagRequired("config")

	hooksUpdateCmd.Flags().String("patch", "", "Partial JSON patch (@filepath or literal)")
	_ = hooksUpdateCmd.MarkFlagRequired("patch")

	hooksToggleCmd.Flags().Bool("enabled", true, "Set enabled state")

	hooksCmd.AddCommand(hooksListCmd, hooksCreateCmd, hooksUpdateCmd,
		hooksDeleteCmd, hooksToggleCmd, hooksHistoryCmd, hooksTestCmd)
	rootCmd.AddCommand(hooksCmd)
}
