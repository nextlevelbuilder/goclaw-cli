package cmd

import (
	"fmt"

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

Response schema (array of message objects):
  [
    {
      "role":        "user|assistant|system",
      "content":     "string",
      "created_at":  "RFC3339 timestamp",
      "session_key": "string"
    },
    ...
  ]

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

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		params := map[string]any{
			"agent_key": args[0],
			"limit":     limit,
		}
		if before != "" {
			params["before"] = before
		}
		if session != "" {
			params["session_key"] = session
		}

		data, err := ws.Call("chat.history", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
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
  agent_key   string  Agent key (required)
  role        string  "user", "assistant", or "system" (required)
  content     string  Message content (required)
  session_key string  Target session key (optional)

Response schema:
  {
    "injected":    true,
    "message_id":  "string",
    "session_key": "string"
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

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		params := map[string]any{
			"agent_key": args[0],
			"role":      role,
			"content":   content,
		}
		if session != "" {
			params["session_key"] = session
		}

		data, err := ws.Call("chat.inject", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
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
    "agent_key":   "string",
    "session_key": "string",
    "state":       "idle|running|waiting|error",
    "turn_count":  42,
    "last_active": "RFC3339 timestamp",
    "model":       "string"
  }

Examples:
  goclaw chat session-status my-agent
  goclaw chat session-status my-agent --output=json | jq '.state'
  goclaw chat session-status my-agent --session=sess-key-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		session, _ := cmd.Flags().GetString("session")

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		params := map[string]any{"agent_key": args[0]}
		if session != "" {
			params["session_key"] = session
		}

		data, err := ws.Call("chat.session.status", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
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

	// chat session-status flags
	chatSessionStatusCmd.Flags().String("session", "", "Session key to query")

	chatCmd.AddCommand(chatHistoryCmd, chatInjectCmd, chatSessionStatusCmd)
}
