package cmd

import (
	"github.com/spf13/cobra"
)

// providers_verify.go extends providersCmd with embedding verification
// and embedding status endpoints.

var providersVerifyEmbeddingCmd = &cobra.Command{
	Use:   "verify-embedding <id>",
	Short: "Verify embedding capability of a provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/providers/"+args[0]+"/verify-embedding", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var providersEmbeddingStatusCmd = &cobra.Command{
	Use:   "embedding-status",
	Short: "Show global embedding provider status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/embedding/status")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	providersCmd.AddCommand(providersVerifyEmbeddingCmd, providersEmbeddingStatusCmd)
}
