package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// memory.go — root command + agent-scoped documents (list/get/store/delete/search).
// KG operations → memory_kg.go + memory_kg_dedup.go
// Index/chunks/global → memory_index.go
// Legacy kgCmd (knowledge-graph alias) removed — use "memory kg" instead.

var memoryCmd = &cobra.Command{Use: "memory", Short: "Manage agent memory"}

var memoryListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List memory documents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/memory/" + args[0]
		if v, _ := cmd.Flags().GetString("user"); v != "" {
			path += "?user_id=" + v
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("PATH", "SIZE", "UPDATED")
		for _, d := range unmarshalList(data) {
			tbl.AddRow(str(d, "path"), str(d, "size"), str(d, "updated_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var memoryGetCmd = &cobra.Command{
	Use:   "get <agentID> <path>",
	Short: "Get a memory document",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/memory/" + url.PathEscape(args[0]) + "/" + url.PathEscape(args[1]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryStoreCmd = &cobra.Command{
	Use:   "store <agentID> <path>",
	Short: "Store a memory document",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		contentVal, _ := cmd.Flags().GetString("content")
		content, err := readContent(contentVal)
		if err != nil {
			return err
		}
		_, err = c.Put("/v1/memory/"+url.PathEscape(args[0])+"/"+url.PathEscape(args[1]),
			map[string]any{"content": content})
		if err != nil {
			return err
		}
		printer.Success("Document stored")
		return nil
	},
}

var memoryDeleteCmd = &cobra.Command{
	Use:   "delete <agentID> <path>",
	Short: "Delete a memory document",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this document?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/memory/" + url.PathEscape(args[0]) + "/" + url.PathEscape(args[1]))
		if err != nil {
			return err
		}
		printer.Success("Document deleted")
		return nil
	},
}

var memorySearchCmd = &cobra.Command{
	Use:   "search <agentID>",
	Short: "Semantic search over agent memory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		query, _ := cmd.Flags().GetString("query")
		user, _ := cmd.Flags().GetString("user")
		body := buildBody("query", query, "user_id", user)
		data, err := c.Post("/v1/memory/"+args[0]+"/search", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	memoryListCmd.Flags().String("user", "", "Filter by user ID")
	memoryStoreCmd.Flags().String("content", "", "Content (or @filepath)")
	_ = memoryStoreCmd.MarkFlagRequired("content")
	memorySearchCmd.Flags().String("query", "", "Search query")
	memorySearchCmd.Flags().String("user", "", "User ID")
	_ = memorySearchCmd.MarkFlagRequired("query")

	memoryCmd.AddCommand(
		memoryListCmd, memoryGetCmd, memoryStoreCmd, memoryDeleteCmd, memorySearchCmd,
		memoryKGCmd,
		// index/chunks registered here
		memoryChunksCmd, memoryIndexCmd, memoryIndexAllCmd, memoryDocumentsGlobalCmd,
	)
	rootCmd.AddCommand(memoryCmd)
}
