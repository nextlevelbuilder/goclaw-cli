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

// teams_events.go — team-level event stream (list with optional --follow).

var teamsEventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Team event stream",
}

var teamsEventsListCmd = &cobra.Command{
	Use:   "list <teamID>",
	Short: "List or stream team events",
	Long: `List recent team events or stream them live with --follow.

WS method: teams.events.list

Example:
  goclaw teams events list team-1
  goclaw teams events list team-1 --follow`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")
		params := map[string]any{"team_id": args[0]}

		handler := func(event *client.WSEvent) error {
			var payload map[string]any
			if err := json.Unmarshal(event.Payload, &payload); err == nil {
				printer.Print(payload)
			}
			return nil
		}

		if follow {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			if output.IsTTY(int(os.Stdout.Fd())) {
				fmt.Println("Streaming team events... (Ctrl+C to stop)")
			}

			err := client.FollowStream(
				ctx, cfg.Server, cfg.Token, "cli", cfg.Insecure,
				"teams.events.list", params, handler, nil,
			)
			if err != nil && ctx.Err() != nil {
				return nil // graceful SIGINT
			}
			return err
		}

		// One-shot
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.events.list", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	teamsEventsListCmd.Flags().Bool("follow", false, "Stream events continuously (Ctrl+C to stop)")
	teamsEventsCmd.AddCommand(teamsEventsListCmd)
}
