package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var vaultLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "Manage vault document links",
}

// --- create link ---

var vaultLinksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a link between two vault documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		linkType, _ := cmd.Flags().GetString("type")

		body := buildBody(
			"from_doc_id", from,
			"to_doc_id", to,
			"link_type", linkType,
		)
		data, err := c.Post("/v1/vault/links", body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		printer.Success(fmt.Sprintf("Link created: %s → %s (type: %s)", str(m, "from_doc_id"), str(m, "to_doc_id"), str(m, "link_type")))
		return nil
	},
}

// --- delete link ---

var vaultLinksDeleteCmd = &cobra.Command{
	Use:   "delete <linkID>",
	Short: "Delete a vault link",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete link %s?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/vault/links/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Link deleted")
		return nil
	},
}

// --- batch-get links ---

var vaultLinksBatchGetCmd = &cobra.Command{
	Use:   "batch-get <docID> [docID...]",
	Short: "Get outlinks for multiple document IDs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := map[string]any{"doc_ids": args}
		data, err := c.Post("/v1/vault/links/batch", body)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "FROM", "TO", "TYPE")
		for _, l := range unmarshalList(data) {
			tbl.AddRow(str(l, "id"), str(l, "from_doc_id"), str(l, "to_doc_id"), str(l, "link_type"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	vaultLinksCreateCmd.Flags().String("from", "", "Source document ID (required)")
	vaultLinksCreateCmd.Flags().String("to", "", "Target document ID (required)")
	vaultLinksCreateCmd.Flags().String("type", "reference", "Link type (e.g. reference, depends-on)")
	_ = vaultLinksCreateCmd.MarkFlagRequired("from")
	_ = vaultLinksCreateCmd.MarkFlagRequired("to")

	vaultLinksCmd.AddCommand(vaultLinksCreateCmd, vaultLinksDeleteCmd, vaultLinksBatchGetCmd)
	vaultCmd.AddCommand(vaultLinksCmd)
}
