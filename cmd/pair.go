package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// pairCmd manages device pairing records (list, request, approve, deny, revoke).
// Distinct from `auth login --pair` which initiates the browser pairing flow.
var pairCmd = &cobra.Command{Use: "pair", Short: "Manage device pairings"}

var pairListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pairings (pending and approved)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("device.pair.list", nil)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		// Print pending section
		tbl := output.NewTable("CODE/SENDER", "CHANNEL", "STATUS", "CHAT_ID")
		if pending, ok := m["pending"].([]any); ok {
			for _, raw := range pending {
				if row, ok := raw.(map[string]any); ok {
					tbl.AddRow(str(row, "sender_id"), str(row, "channel"), "pending", str(row, "chat_id"))
				}
			}
		}
		if paired, ok := m["paired"].([]any); ok {
			for _, raw := range paired {
				if row, ok := raw.(map[string]any); ok {
					tbl.AddRow(str(row, "sender_id"), str(row, "channel"), "approved", str(row, "chat_id"))
				}
			}
		}
		printer.Print(tbl)
		return nil
	},
}

var pairRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request a new device pairing",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		senderID, _ := cmd.Flags().GetString("sender-id")
		channel, _ := cmd.Flags().GetString("channel")
		chatID, _ := cmd.Flags().GetString("chat-id")
		purpose, _ := cmd.Flags().GetString("purpose")
		params := buildBody("senderId", senderID, "channel", channel, "chatId", chatID, "purpose", purpose)
		data, err := ws.Call("device.pair.request", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var pairApproveCmd = &cobra.Command{
	Use:   "approve <code>",
	Short: "Approve a pairing request by code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		approvedBy, _ := cmd.Flags().GetString("approved-by")
		data, err := ws.Call("device.pair.approve", buildBody("code", args[0], "approvedBy", approvedBy))
		if err != nil {
			return err
		}
		printer.Success("Pairing approved")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var pairDenyCmd = &cobra.Command{
	Use:   "deny <code>",
	Short: "Deny a pairing request by code",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("device.pair.deny", map[string]any{"code": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Pairing denied")
		return nil
	},
}

var pairRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke an active pairing (--sender-id + --channel required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Revoke this pairing?", cfg.Yes) {
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
		senderID, _ := cmd.Flags().GetString("sender-id")
		channel, _ := cmd.Flags().GetString("channel")
		_, err = ws.Call("device.pair.revoke", buildBody("senderId", senderID, "channel", channel))
		if err != nil {
			return err
		}
		printer.Success("Pairing revoked")
		return nil
	},
}

func init() {
	pairRequestCmd.Flags().String("sender-id", "", "Sender identifier (required)")
	pairRequestCmd.Flags().String("channel", "", "Channel type, e.g. telegram (required)")
	pairRequestCmd.Flags().String("chat-id", "", "Chat ID")
	pairRequestCmd.Flags().String("purpose", "", "Purpose description")
	_ = pairRequestCmd.MarkFlagRequired("sender-id")
	_ = pairRequestCmd.MarkFlagRequired("channel")

	pairApproveCmd.Flags().String("approved-by", "operator", "Who is approving (default: operator)")

	pairRevokeCmd.Flags().String("sender-id", "", "Sender identifier (required)")
	pairRevokeCmd.Flags().String("channel", "", "Channel type (required)")
	_ = pairRevokeCmd.MarkFlagRequired("sender-id")
	_ = pairRevokeCmd.MarkFlagRequired("channel")

	pairCmd.AddCommand(pairListCmd, pairRequestCmd, pairApproveCmd, pairDenyCmd, pairRevokeCmd)
	rootCmd.AddCommand(pairCmd)
}
