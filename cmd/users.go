package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// usersCmd provides user search across tenant contacts and peers.
var usersCmd = &cobra.Command{Use: "users", Short: "Search and manage users"}

var usersSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search users by query string",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		query, _ := cmd.Flags().GetString("q")
		if query == "" {
			return fmt.Errorf("--q is required")
		}
		q.Set("q", query)
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", limit))
		}
		if peerKind, _ := cmd.Flags().GetString("peer-kind"); peerKind != "" {
			q.Set("peer_kind", peerKind)
		}
		data, err := c.Get("/v1/users/search?" + q.Encode())
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "USERNAME", "PEER_KIND")
		for _, u := range unmarshalList(data) {
			tbl.AddRow(str(u, "id"), str(u, "name"), str(u, "username"), str(u, "peer_kind"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	usersSearchCmd.Flags().StringP("q", "q", "", "Search query (required)")
	usersSearchCmd.Flags().Int("limit", 30, "Maximum results")
	usersSearchCmd.Flags().String("peer-kind", "", "Filter by peer kind (e.g. telegram, discord)")
	_ = usersSearchCmd.MarkFlagRequired("q")

	usersCmd.AddCommand(usersSearchCmd)
	rootCmd.AddCommand(usersCmd)
}
