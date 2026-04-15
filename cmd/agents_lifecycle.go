package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// agents_lifecycle.go — wake, wait, identity (AI-critical MAX POLISH).
// sync-workspace + prompt-preview → agents_admin.go (split for LoC limit).

var agentsWakeCmd = &cobra.Command{
	Use:   "wake <id>",
	Short: "Wake a sleeping agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/wake", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var agentsWaitCmd = &cobra.Command{
	Use:   "wait <key>",
	Short: "Block until an agent reaches a target state",
	Long: `Block until the agent identified by <key> reaches the target state.
Exits 0 when the state matches. Exits 6 (RESOURCE_EXHAUSTED) on timeout.
Handles SIGINT/SIGTERM gracefully.

WS method: agent.wait

Response schema:
  {
    "agent_key": "string",
    "state":     "online|running|idle|offline",
    "reached_at": "RFC3339 timestamp"
  }

Examples:
  # Wait up to 60 s for agent to go idle
  goclaw agents wait my-agent --state=idle --timeout=60s

  # Wait indefinitely (default timeout = 5 min)
  goclaw agents wait my-agent --state=online

  # Automation: exit code 6 on timeout
  goclaw agents wait my-agent --state=running --timeout=30s || echo "timed out"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		state, _ := cmd.Flags().GetString("state")
		timeoutStr, _ := cmd.Flags().GetString("timeout")

		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid --timeout %q: %w", timeoutStr, err)
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		// Apply timeout on top of signal context
		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, timeout)
		defer timeoutCancel()

		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		params := map[string]any{"agent_key": args[0]}
		if state != "" {
			params["state"] = state
		}

		// Call is blocking on the server side until state matches or server times out.
		// We wrap with our own context deadline.
		type result struct {
			data []byte
			err  error
		}
		done := make(chan result, 1)
		go func() {
			data, err := ws.Call("agent.wait", params)
			done <- result{data, err}
		}()

		select {
		case <-timeoutCtx.Done():
			// Distinguish signal vs timeout
			if ctx.Err() != nil {
				// SIGINT/SIGTERM — graceful stop, not an error
				return nil
			}
			// Timeout elapsed. Close ws cleanly before exiting so server-side
			// pending Call goroutine unblocks (defer ws.Close would also run,
			// but os.Exit bypasses defers — close explicitly here).
			ws.Close()
			fmt.Fprintf(os.Stderr, "timeout: agent %q did not reach state %q within %s\n",
				args[0], state, timeoutStr)
			output.Exit(output.ExitResource)
			return nil // unreachable after Exit; satisfies compiler
		case r := <-done:
			if r.err != nil {
				return r.err
			}
			printer.Print(unmarshalMap(r.data))
			return nil
		}
	},
}

var agentsIdentityCmd = &cobra.Command{
	Use:   "identity <key>",
	Short: "Get agent identity and persona",
	Long: `Retrieve the identity configuration for an agent via WebSocket.

WS method: agent.identity.get

Response schema:
  {
    "agent_key":    "string",
    "display_name": "string",
    "persona":      "string",
    "traits":       ["string", ...],
    "goals":        ["string", ...],
    "constraints":  ["string", ...]
  }

Examples:
  goclaw agents identity my-agent
  goclaw agents identity my-agent --output=json | jq '.persona'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()

		data, err := ws.Call("agent.identity.get", map[string]any{"agent_key": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	agentsWaitCmd.Flags().String("state", "online", "Target state: online, running, idle, offline")
	agentsWaitCmd.Flags().String("timeout", "5m", "Maximum wait duration (e.g. 30s, 5m, 1h)")

	// sync-workspace + prompt-preview → agents_admin.go
	agentsCmd.AddCommand(agentsWakeCmd, agentsWaitCmd, agentsIdentityCmd)
}
