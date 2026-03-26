package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat <agent> [message]",
	Short: "Chat with an agent",
	Long: `Chat with an agent interactively or send a single message.

Interactive mode (default when no -m flag):
  goclaw chat myagent

Single-shot (automation):
  goclaw chat myagent -m "What is the status?"

Pipe stdin:
  echo "Analyze this" | goclaw chat myagent`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentKey := args[0]
		message, _ := cmd.Flags().GetString("message")
		session, _ := cmd.Flags().GetString("session")
		noStream, _ := cmd.Flags().GetBool("no-stream")

		// Check for piped stdin
		if message == "" && !tui.IsInteractive() {
			data, err := io.ReadAll(os.Stdin)
			if err == nil && len(data) > 0 {
				message = strings.TrimSpace(string(data))
			}
		}

		// Single-shot mode
		if message != "" {
			return chatSingleShot(agentKey, message, session, noStream)
		}

		// Interactive mode
		return chatInteractive(agentKey, session)
	},
}

func chatSingleShot(agentKey, message, session string, noStream bool) error {
	ws, err := newWS("cli")
	if err != nil {
		return err
	}
	if _, err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]any{
		"agent_key": agentKey,
		"message":   message,
	}
	if session != "" {
		params["session_key"] = session
	}

	if noStream || cfg.OutputFormat == "json" {
		// Non-streaming: collect full response
		resp, err := ws.Call("chat.send", params)
		if err != nil {
			return err
		}
		if cfg.OutputFormat == "json" {
			printer.Print(unmarshalMap(resp))
		} else {
			var result struct {
				Content string `json:"content"`
			}
			_ = json.Unmarshal(resp, &result)
			fmt.Println(result.Content)
		}
		return nil
	}

	// Streaming mode
	_, err = ws.Stream("chat.send", params, func(e *client.WSEvent) {
		if cfg.OutputFormat == "json" {
			// NDJSON output
			line := map[string]any{"event": e.Event, "data": unmarshalMap(e.Payload)}
			data, _ := json.Marshal(line)
			fmt.Println(string(data))
			return
		}
		switch e.Event {
		case "chunk":
			var chunk struct {
				Content string `json:"content"`
			}
			_ = json.Unmarshal(e.Payload, &chunk)
			fmt.Print(chunk.Content)
		case "tool.call":
			var tc struct {
				Name string `json:"name"`
			}
			_ = json.Unmarshal(e.Payload, &tc)
			fmt.Printf("\n[tool: %s]\n", tc.Name)
		case "tool.result":
			fmt.Print(".")
		case "run.completed":
			fmt.Println()
		}
	})
	return err
}

func chatInteractive(agentKey, session string) error {
	ws, err := newWS("cli")
	if err != nil {
		return err
	}
	if _, err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	fmt.Printf("Connected to agent: %s\n", agentKey)
	fmt.Println("Type your message and press Enter. Use /exit to quit, /abort to cancel.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle slash commands
		switch input {
		case "/exit", "/quit":
			fmt.Println("Goodbye!")
			return nil
		case "/abort":
			_, _ = ws.Call("chat.abort", map[string]any{"agent_key": agentKey})
			fmt.Println("[aborted]")
			continue
		case "/sessions":
			resp, err := ws.Call("sessions.list", map[string]any{"agent_key": agentKey})
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				continue
			}
			fmt.Println(string(resp))
			continue
		case "/clear":
			if session != "" {
				_, _ = ws.Call("sessions.reset", map[string]any{"session_key": session})
				fmt.Println("[session cleared]")
			}
			continue
		}

		params := map[string]any{
			"agent_key": agentKey,
			"message":   input,
		}
		if session != "" {
			params["session_key"] = session
		}

		_, err := ws.Stream("chat.send", params, func(e *client.WSEvent) {
			switch e.Event {
			case "chunk":
				var chunk struct {
					Content string `json:"content"`
				}
				_ = json.Unmarshal(e.Payload, &chunk)
				fmt.Print(chunk.Content)
			case "tool.call":
				var tc struct {
					Name string `json:"name"`
				}
				_ = json.Unmarshal(e.Payload, &tc)
				fmt.Printf("\n[calling: %s] ", tc.Name)
			case "tool.result":
				fmt.Print(".")
			case "run.completed":
				fmt.Print("\n\n")
			}
		})
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	}
	return nil
}

var chatInjectCmd = &cobra.Command{
	Use:   "inject <agent>",
	Short: "Inject message into running session",
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
		text, _ := cmd.Flags().GetString("text")
		session, _ := cmd.Flags().GetString("session")
		params := buildBody("agent_key", args[0], "text", text, "session_key", session)
		data, err := ws.Call("chat.inject", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var chatStatusCmd = &cobra.Command{
	Use:   "status <agent>",
	Short: "Get session/run status",
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
		session, _ := cmd.Flags().GetString("session")
		params := buildBody("agent_key", args[0], "session_key", session)
		data, err := ws.Call("chat.session.status", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var chatAbortCmd = &cobra.Command{
	Use:   "abort <agent>",
	Short: "Abort running agent",
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
		session, _ := cmd.Flags().GetString("session")
		params := buildBody("agent_key", args[0], "session_key", session)
		data, err := ws.Call("chat.abort", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	chatCmd.Flags().StringP("message", "m", "", "Message to send (single-shot mode)")
	chatCmd.Flags().String("session", "", "Session key to continue")
	chatCmd.Flags().Bool("no-stream", false, "Disable streaming, wait for full response")

	chatInjectCmd.Flags().String("text", "", "Text to inject")
	chatInjectCmd.Flags().String("session", "", "Session key")
	_ = chatInjectCmd.MarkFlagRequired("text")

	chatStatusCmd.Flags().String("session", "", "Session key")
	chatAbortCmd.Flags().String("session", "", "Session key")

	chatCmd.AddCommand(chatInjectCmd, chatStatusCmd, chatAbortCmd)
	rootCmd.AddCommand(chatCmd)
}
