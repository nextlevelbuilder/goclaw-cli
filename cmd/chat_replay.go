package cmd

import "github.com/spf13/cobra"

var chatReplayCmd = &cobra.Command{
	Use:   "replay <agent>",
	Short: "Replay history for a specific chat session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		session, _ := cmd.Flags().GetString("session")
		before, _ := cmd.Flags().GetString("before")
		limit, _ := cmd.Flags().GetInt("limit")
		return runChatHistory(args[0], session, before, limit)
	},
}

func runChatHistory(agent, session, before string, limit int) error {
	ws, err := newWS("cli")
	if err != nil {
		return err
	}
	if _, err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	params := map[string]any{
		"agentId": agent,
		"limit":   limit,
	}
	if before != "" {
		params["before"] = before
	}
	if session != "" {
		params["sessionKey"] = session
	}

	data, err := ws.Call("chat.history", params)
	if err != nil {
		return err
	}
	printer.Print(unmarshalList(data))
	return nil
}

func init() {
	chatReplayCmd.Flags().String("session", "", "Session key to replay")
	_ = chatReplayCmd.MarkFlagRequired("session")
	chatReplayCmd.Flags().String("before", "", "Return messages before this RFC3339 timestamp")
	chatReplayCmd.Flags().Int("limit", 50, "Maximum number of messages to return")
	chatCmd.AddCommand(chatReplayCmd)
}
