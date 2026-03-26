package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var contactsCmd = &cobra.Command{Use: "contacts", Short: "Manage contacts"}

var contactsListCmd = &cobra.Command{
	Use: "list", Short: "List contacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/contacts")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "IDENTIFIER", "TYPE", "CREATED")
		for _, row := range unmarshalList(data) {
			tbl.AddRow(str(row, "id"), str(row, "name"), str(row, "identifier"),
				str(row, "type"), str(row, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var contactsResolveCmd = &cobra.Command{
	Use: "resolve", Short: "Resolve contact by phone or email",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("identifier"); v != "" {
			q.Set("identifier", v)
		}
		path := "/v1/contacts/resolve"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var contactsMergeCmd = &cobra.Command{
	Use: "merge <id1> <id2>", Short: "Merge two contacts", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/contacts/merge",
			map[string]any{"id1": args[0], "id2": args[1]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var contactsUnmergeCmd = &cobra.Command{
	Use: "unmerge <tenant-user-id>", Short: "Unmerge a contact", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/contacts/unmerge",
			map[string]any{"tenant_user_id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Contact unmerged")
		return nil
	},
}

var contactsMergedCmd = &cobra.Command{
	Use: "merged <tenant-user-id>", Short: "List merged contacts for a tenant user", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/contacts/merged/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	contactsResolveCmd.Flags().String("identifier", "", "Phone number or email address")
	_ = contactsResolveCmd.MarkFlagRequired("identifier")

	contactsCmd.AddCommand(contactsListCmd, contactsResolveCmd, contactsMergeCmd,
		contactsUnmergeCmd, contactsMergedCmd)
	rootCmd.AddCommand(contactsCmd)
}
