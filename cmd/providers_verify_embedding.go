package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var providersVerifyEmbeddingCmd = &cobra.Command{
	Use: "verify-embedding <id>", Short: "Verify embedding API credentials", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/providers/"+url.PathEscape(args[0])+"/verify-embedding", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	providersCmd.AddCommand(providersVerifyEmbeddingCmd)
}
