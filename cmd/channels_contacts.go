package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// channels_contacts.go extends channelsContactsCmd with merge and merged-lookup
// operations. These are admin-scope endpoints for deduplicating contacts across
// channel instances.

var channelsContactsMergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two contacts (source into target) — requires --yes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Merge contacts? This action cannot be undone.", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")
		data, err := c.Post("/v1/contacts/merge", map[string]any{
			"source_id": source,
			"target_id": target,
		})
		if err != nil {
			return err
		}
		printer.Success("Contacts merged")
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var channelsContactsMergedCmd = &cobra.Command{
	Use:   "merged <tenantUserID>",
	Short: "List contacts merged into a tenant user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/contacts/merged/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	channelsContactsMergeCmd.Flags().String("source", "", "Source contact ID to merge from (required)")
	channelsContactsMergeCmd.Flags().String("target", "", "Target contact ID to merge into (required)")
	_ = channelsContactsMergeCmd.MarkFlagRequired("source")
	_ = channelsContactsMergeCmd.MarkFlagRequired("target")

	// Register as subcommands of channelsContactsCmd (defined in channels.go).
	channelsContactsCmd.AddCommand(channelsContactsMergeCmd, channelsContactsMergedCmd)
}
