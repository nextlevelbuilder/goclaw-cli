package cmd

import (
	"github.com/spf13/cobra"
)

// heartbeatChecklistCmd groups checklist subcommands (split from heartbeat.go for LoC limit).
var heartbeatChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Manage agent HEARTBEAT.md checklist",
}

var heartbeatChecklistGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the HEARTBEAT.md checklist content for an agent",
	Long: `Get the HEARTBEAT.md context file used during heartbeat runs.

Example:
  goclaw heartbeat checklist get --agent=my-agent`,
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
		data, err := ws.Call("heartbeat.checklist.get", map[string]any{"agentId": agent})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatChecklistSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the HEARTBEAT.md checklist content for an agent",
	Long: `Set the HEARTBEAT.md context file used during heartbeat runs.
Use @filepath to read content from a file.

Example:
  goclaw heartbeat checklist set --agent=my-agent --content=@./checklist.md
  goclaw heartbeat checklist set --agent=my-agent --content="## Health Check\n- [ ] API responding"`,
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
		rawContent, _ := cmd.Flags().GetString("content")
		content, err := readContent(rawContent)
		if err != nil {
			return err
		}
		data, err := ws.Call("heartbeat.checklist.set", map[string]any{
			"agentId": agent,
			"content": content,
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	heartbeatChecklistGetCmd.Flags().String("agent", "", "Agent key or ID")
	_ = heartbeatChecklistGetCmd.MarkFlagRequired("agent")

	heartbeatChecklistSetCmd.Flags().String("agent", "", "Agent key or ID")
	heartbeatChecklistSetCmd.Flags().String("content", "", "Markdown content or @filepath")
	_ = heartbeatChecklistSetCmd.MarkFlagRequired("agent")
	_ = heartbeatChecklistSetCmd.MarkFlagRequired("content")

	heartbeatChecklistCmd.AddCommand(heartbeatChecklistGetCmd, heartbeatChecklistSetCmd)

	// Register under heartbeatCmd (defined in heartbeat.go).
	heartbeatCmd.AddCommand(heartbeatChecklistCmd)
}
