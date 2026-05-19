package cmd

import "github.com/spf13/cobra"

var chatSessionsCmd = &cobra.Command{Use: "sessions", Short: "Chat session convenience commands"}

var chatSessionsResumeCmd = &cobra.Command{
	Use:   "resume <agent>",
	Short: "Resume a chat session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		session, _ := cmd.Flags().GetString("session")
		message, _ := cmd.Flags().GetString("message")
		noStream, _ := cmd.Flags().GetBool("no-stream")
		if message != "" {
			return chatSingleShot(args[0], message, session, noStream)
		}
		return chatInteractive(args[0], session)
	},
}

func init() {
	chatSessionsResumeCmd.Flags().String("session", "", "Session key to resume")
	_ = chatSessionsResumeCmd.MarkFlagRequired("session")
	chatSessionsResumeCmd.Flags().StringP("message", "m", "", "Message to send")
	chatSessionsResumeCmd.Flags().Bool("no-stream", false, "Disable streaming, wait for full response")
	chatSessionsCmd.AddCommand(chatSessionsResumeCmd)
	chatCmd.AddCommand(chatSessionsCmd)
}
