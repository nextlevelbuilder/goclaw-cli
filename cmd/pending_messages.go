package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var pendingMessagesCmd = &cobra.Command{Use: "pending-messages", Short: "Manage pending messages"}

var pendingMessagesListCmd = &cobra.Command{
	Use: "list", Short: "List pending messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/pending-messages")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "CHANNEL", "FROM", "PREVIEW", "CREATED")
		for _, row := range unmarshalList(data) {
			tbl.AddRow(str(row, "id"), str(row, "channel"), str(row, "from"),
				str(row, "preview"), str(row, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var pendingMessagesCompactCmd = &cobra.Command{
	Use: "compact", Short: "Compact pending messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/pending-messages/compact", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var pendingMessagesDeleteCmd = &cobra.Command{
	Use: "delete", Short: "Delete all pending messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete all pending messages?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/pending-messages")
		if err != nil {
			return err
		}
		printer.Success("Pending messages deleted")
		return nil
	},
}

func init() {
	pendingMessagesCmd.AddCommand(pendingMessagesListCmd, pendingMessagesCompactCmd, pendingMessagesDeleteCmd)
	rootCmd.AddCommand(pendingMessagesCmd)
}
