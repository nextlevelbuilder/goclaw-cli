package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{Use: "logs", Short: "View server logs"}

var logsTailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Stream server logs in real-time",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		agent, _ := cmd.Flags().GetString("agent")
		level, _ := cmd.Flags().GetString("level")

		params := make(map[string]any)
		if agent != "" {
			params["agent_id"] = agent
		}
		if level != "" {
			params["level"] = level
		}

		fmt.Println("Streaming logs... (Ctrl+C to stop)")

		// Subscribe to log events
		ws.Subscribe("*", func(e *client.WSEvent) {
			if cfg.OutputFormat == "json" {
				line := map[string]any{"event": e.Event, "data": unmarshalMap(e.Payload)}
				data, _ := json.Marshal(line)
				fmt.Println(string(data))
				return
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
		})

		// Call logs.tail — server will push events
		_, err = ws.Call("logs.tail", params)
		if err != nil {
			return err
		}

		// Block until interrupt
		select {}
	},
}

func init() {
	logsTailCmd.Flags().String("agent", "", "Filter by agent ID")
	logsTailCmd.Flags().String("level", "", "Filter: info, warn, error")
	logsTailCmd.Flags().Bool("follow", true, "Follow log output")

	logsCmd.AddCommand(logsTailCmd)
	rootCmd.AddCommand(logsCmd)
}
