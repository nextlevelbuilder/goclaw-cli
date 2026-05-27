package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// sessionsBranchCmd posts to /v1/chat/sessions/{key}/branch.
//
// Backend route: POST /v1/chat/sessions/{key}/branch (chat domain).
// (Sibling commands under `sessions` parent target /v1/sessions/...; this
// command intentionally targets the chat-sessions tree where branching lives.)
//
// Body is constructed directly (NOT via buildBody) because up_to_index=0 is a
// valid required value that buildBody's int-zero skip would silently drop.
var sessionsBranchCmd = &cobra.Command{
	Use:   "branch <sessionKey>",
	Short: "Branch a chat session at a message index",
	Long: `Branch a chat session into a new session by copying messages up to a 1-based
index. The source session is unchanged.

Backend route: POST /v1/chat/sessions/{key}/branch (chat domain).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upTo, _ := cmd.Flags().GetInt("up-to-index")
		if upTo < 0 {
			return fmt.Errorf("--up-to-index must be >= 0 (got %d)", upTo)
		}

		// Validate metadata up front, BEFORE HTTP call.
		metaPairs, _ := cmd.Flags().GetStringArray("metadata")
		metadata := make(map[string]any)
		for _, kv := range metaPairs {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 || parts[0] == "" {
				return fmt.Errorf("--metadata must be key=value (got %q)", kv)
			}
			metadata[parts[0]] = parts[1]
		}

		// Build body directly so up_to_index=0 is preserved on the wire.
		body := map[string]any{"up_to_index": upTo}
		if v, _ := cmd.Flags().GetString("new-session-key"); v != "" {
			body["new_session_key"] = v
		}
		if v, _ := cmd.Flags().GetString("label"); v != "" {
			body["label"] = v
		}
		if len(metadata) > 0 {
			body["metadata"] = metadata
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/chat/sessions/"+url.PathEscape(args[0])+"/branch", body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		tbl := output.NewTable("SOURCE", "NEW_KEY", "COPIED", "TOTAL", "LABEL")
		tbl.AddRow(str(m, "source_key"), str(m, "session_key"),
			str(m, "copied_messages"), str(m, "total_messages"), str(m, "label"))
		printer.Print(tbl)
		return nil
	},
}

func init() {
	sessionsBranchCmd.Flags().Int("up-to-index", -1, "Copy messages 1..N into the new session (required, >=0)")
	sessionsBranchCmd.Flags().String("new-session-key", "", "Override generated session key")
	sessionsBranchCmd.Flags().String("label", "", "Label for the new session")
	sessionsBranchCmd.Flags().StringArray("metadata", nil, "Repeatable key=value metadata pair")
	_ = sessionsBranchCmd.MarkFlagRequired("up-to-index")
	sessionsCmd.AddCommand(sessionsBranchCmd)
}
