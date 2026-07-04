package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/client"
	"github.com/spf13/cobra"
)

// chat_ai_commands.go — AI-critical chat extensions: history, inject, session-status.
// MAX POLISH: JSON schemas in --help, full error validation, ≥80% test coverage.
// Extracted from chat.go to keep files <200 LoC.

var chatHistoryCmd = &cobra.Command{
	Use:   "history <agent>",
	Short: "Retrieve chat history for an agent session",
	Long: `Retrieve the conversation history for an agent via WebSocket.

WS method: chat.history

Response schema:
  {
    "messages": [
      {"role": "user|assistant|system", "content": "string", ...},
      ...
    ]
  }

Flags:
  --limit  N       Maximum messages to return (default: 50)
  --before <ts>    Return messages before this RFC3339 timestamp
  --session <key>  Filter to a specific session key

Examples:
  goclaw chat history my-agent --limit=20
  goclaw chat history my-agent --output=json | jq '.[].content'
  goclaw chat history my-agent --before=2024-01-15T00:00:00Z --limit=100`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		before, _ := cmd.Flags().GetString("before")
		session, _ := cmd.Flags().GetString("session")

		return runChatHistory(args[0], session, before, limit)
	},
}

var chatInjectCmd = &cobra.Command{
	Use:   "inject <agent>",
	Short: "Inject a message into agent context without triggering a response",
	Long: `Inject a message directly into the agent's conversation context.
The agent does NOT process or respond to injected messages — they are inserted
into context as-is. Useful for AI orchestration tools that need to seed context.

SECURITY: Injecting system-role messages can alter agent behavior.
This endpoint requires admin-level permissions on the server.

WS method: chat.inject

Request fields:
  sessionKey string  Target session key (required)
  message    string  Message content (required)
  label      string  Optional label (role is used as label)

Response schema:
  {
    "ok":        true,
    "messageId": "string"
  }

Examples:
  goclaw chat inject my-agent --role=system --content="You are a helpful assistant."
  goclaw chat inject my-agent --role=user --content="Previous context here" --session=sess-1
  goclaw chat inject my-agent --role=assistant --content=@./prior-response.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		contentVal, _ := cmd.Flags().GetString("content")
		session, _ := cmd.Flags().GetString("session")

		if role != "user" && role != "assistant" && role != "system" {
			return fmt.Errorf("--role must be 'user', 'assistant', or 'system', got %q", role)
		}

		content, err := readContent(contentVal)
		if err != nil {
			return err
		}
		if content == "" {
			return fmt.Errorf("--content is required and must not be empty")
		}
		if session == "" {
			return fmt.Errorf("--session is required")
		}

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		result, err := ws.ChatInject(client.ChatInjectParams{
			SessionKey: session,
			Message:    content,
			Label:      role,
		})
		if err != nil {
			return err
		}
		printer.Print(result)
		return nil
	},
}

var chatSessionStatusCmd = &cobra.Command{
	Use:   "session-status <agent>",
	Short: "Get current session state for an agent",
	Long: `Retrieve the current session status for an agent's active session.

WS method: chat.session.status

Response schema:
  {
    "isRunning": true,
    "runId":     "string",
    "activity":  {"phase": "string", "tool": "string", "iteration": 1}
  }

Examples:
  goclaw chat session-status my-agent --session=sess-key-1
  goclaw chat session-status my-agent --output=json | jq '.isRunning'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		session, _ := cmd.Flags().GetString("session")
		if session == "" {
			return fmt.Errorf("--session is required")
		}

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		result, err := ws.ChatSessionStatus(client.ChatSessionStatusParams{SessionKey: session})
		if err != nil {
			return err
		}
		printer.Print(result)
		return nil
	},
}

func init() {
	// chat history flags
	chatHistoryCmd.Flags().Int("limit", 50, "Maximum number of messages to return")
	chatHistoryCmd.Flags().String("before", "", "Return messages before this RFC3339 timestamp")
	chatHistoryCmd.Flags().String("session", "", "Filter to a specific session key")

	// chat inject flags
	chatInjectCmd.Flags().String("role", "", "Message role: user, assistant, or system")
	_ = chatInjectCmd.MarkFlagRequired("role")
	chatInjectCmd.Flags().String("content", "", "Message content (or @filepath)")
	_ = chatInjectCmd.MarkFlagRequired("content")
	chatInjectCmd.Flags().String("session", "", "Target session key")
	_ = chatInjectCmd.MarkFlagRequired("session")

	// chat session-status flags
	chatSessionStatusCmd.Flags().String("session", "", "Session key to query")
	_ = chatSessionStatusCmd.MarkFlagRequired("session")

	chatCmd.AddCommand(chatHistoryCmd, chatInjectCmd, chatSessionStatusCmd)
}
