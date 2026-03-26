package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var channelsContactsCmd = &cobra.Command{Use: "contacts", Short: "Manage contacts"}

var channelsContactsListCmd = &cobra.Command{
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
		printer.Print(unmarshalList(data))
		return nil
	},
}

var channelsContactsResolveCmd = &cobra.Command{
	Use: "resolve <ids>", Short: "Resolve contacts by IDs", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/contacts/resolve?ids=" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	channelsContactsCmd.AddCommand(channelsContactsListCmd, channelsContactsResolveCmd)
	channelsCmd.AddCommand(channelsContactsCmd)
}
