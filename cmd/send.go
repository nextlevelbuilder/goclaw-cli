package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sendCmd is the inter-agent messaging primitive for AI orchestration.
// It delivers a one-shot message to a channel recipient without creating
// a full chat session. Useful for AI agents signalling each other or
// posting notifications programmatically.
//
// WS method: "send"
// Params:    { channel, to, message }
// Response:  { ok, channel, to }
//
// JSON schema for --content:
//
//	string — any UTF-8 message text; server does NOT apply Markdown rendering.
//
// Security note: send bypasses normal chat session logging. Server-side
// audit logs capture the event but conversation history is not stored.
// Do not use for user-facing chat flows; use `chat` commands instead.
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a one-shot message to a channel recipient (inter-agent messaging)",
	Long: `Send delivers a message directly to a channel recipient via WebSocket RPC.

This is the inter-agent messaging primitive — designed for AI orchestration
pipelines where one agent needs to notify another or post to a channel
without creating a full chat session.

Parameters schema:
  --channel  string  Channel instance name (e.g. "telegram", "discord"). REQUIRED.
  --to       string  Recipient identifier within the channel (chat_id / user_id). REQUIRED.
  --content  string  Message text to deliver. REQUIRED.

The server publishes an outbound message event on the message bus.
No retry is performed on delivery failure — the caller is responsible
for idempotency if needed.

Examples:
  # Notify a Telegram group from an AI agent
  goclaw send --channel=telegram --to="-1001234567890" --content="Deploy complete ✓"

  # AI-to-AI coordination: signal a downstream agent via its chat ID
  goclaw send --channel=discord --to="987654321" --content='{"event":"task_done","id":"abc"}'

  # Pipe content from a file
  goclaw send --channel=telegram --to="12345" --content="@/tmp/report.txt"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		channel, _ := cmd.Flags().GetString("channel")
		to, _ := cmd.Flags().GetString("to")
		rawContent, _ := cmd.Flags().GetString("content")

		// Validate required fields before opening WS connection.
		if channel == "" {
			return fmt.Errorf("--channel is required")
		}
		if to == "" {
			return fmt.Errorf("--to is required")
		}
		if rawContent == "" {
			return fmt.Errorf("--content is required")
		}

		// Support @filepath syntax for content from file.
		content, err := readContent(rawContent)
		if err != nil {
			return err
		}
		if content == "" {
			return fmt.Errorf("--content must not be empty")
		}

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		data, err := ws.Call("send", map[string]any{
			"channel": channel,
			"to":      to,
			"message": content,
		})
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if cfg.OutputFormat == "table" {
			printer.Success(fmt.Sprintf("Sent → channel=%s to=%s", str(m, "channel"), str(m, "to")))
			return nil
		}
		printer.Print(m)
		return nil
	},
}

func init() {
	sendCmd.Flags().String("channel", "", "Channel instance name, e.g. telegram, discord (required)")
	sendCmd.Flags().String("to", "", "Recipient identifier within the channel, e.g. chat_id (required)")
	sendCmd.Flags().StringP("content", "m", "", `Message text. Prefix with @ to read from file, e.g. @report.txt (required)`)
	_ = sendCmd.MarkFlagRequired("channel")
	_ = sendCmd.MarkFlagRequired("to")
	_ = sendCmd.MarkFlagRequired("content")

	rootCmd.AddCommand(sendCmd)
}
