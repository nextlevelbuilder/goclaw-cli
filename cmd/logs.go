package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{Use: "logs", Short: "View server logs"}

var logsTailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Stream server logs in real-time",
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		level, _ := cmd.Flags().GetString("level")
		follow, _ := cmd.Flags().GetBool("follow")
		quiet, _ := cmd.Flags().GetBool("quiet")

		params := make(map[string]any)
		if agent != "" {
			params["agent_id"] = agent
		}
		if level != "" {
			params["level"] = level
		}

		// Show banner only in TTY + non-quiet mode
		if !quiet && output.IsTTY(int(os.Stdout.Fd())) {
			fmt.Println("Streaming logs... (Ctrl+C to stop)")
		}

		handler := makeLogHandler(cfg.OutputFormat)

		if follow {
			// Use FollowStream for persistent streaming with reconnect
			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			return client.FollowStream(ctx, cfg.Server, cfg.Token, "cli", cfg.Insecure,
				"logs.tail", params, handler, nil)
		}

		// Non-follow: single connection, block until interrupt
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		ws.Subscribe("*", func(e *client.WSEvent) {
			_ = handler(e)
		})

		if _, err = ws.Call("logs.tail", params); err != nil {
			return err
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		if !quiet && output.IsTTY(int(os.Stdout.Fd())) {
			fmt.Println("\nStopping log stream...")
		}
		return nil
	},
}

// makeLogHandler returns a FollowHandler that formats log events per output format.
func makeLogHandler(format string) client.FollowHandler {
	return func(e *client.WSEvent) error {
		if format == "json" {
			line := map[string]any{"event": e.Event, "data": unmarshalMap(e.Payload)}
			data, _ := json.Marshal(line)
			fmt.Println(string(data))
			return nil
		}
		var log struct {
			Level   string `json:"level"`
			Message string `json:"message"`
			Time    string `json:"time"`
		}
		_ = json.Unmarshal(e.Payload, &log)
		if log.Message != "" {
			fmt.Printf("[%s] %s: %s\n", log.Time, log.Level, log.Message)
		} else {
			fmt.Printf("[%s] %s\n", e.Event, string(e.Payload))
		}
		return nil
	}
}

func init() {
	logsTailCmd.Flags().String("agent", "", "Filter by agent ID")
	logsTailCmd.Flags().String("level", "", "Filter: info, warn, error")
	logsTailCmd.Flags().Bool("follow", true, "Follow log output with auto-reconnect")
	logsTailCmd.Flags().Bool("quiet", false, "Suppress status messages")

	logsCmd.AddCommand(logsTailCmd)
	rootCmd.AddCommand(logsCmd)
}
