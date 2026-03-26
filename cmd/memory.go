package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var memoryCmd = &cobra.Command{Use: "memory", Short: "Manage agent memory"}

var memoryListCmd = &cobra.Command{
	Use: "list <agentID>", Short: "List memory documents", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/memory/" + url.PathEscape(args[0])
		if v, _ := cmd.Flags().GetString("user"); v != "" {
			q := url.Values{}
			q.Set("user_id", v)
			path += "?" + q.Encode()
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
	Use: "get <agentID> <path>", Short: "Get memory document", Args: cobra.ExactArgs(2),
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
	Use: "store <agentID> <path>", Short: "Store memory document", Args: cobra.ExactArgs(2),
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
		_, err = c.Put("/v1/memory/"+url.PathEscape(args[0])+"/"+url.PathEscape(args[1]), map[string]any{"content": content})
		if err != nil {
			return err
		}
		printer.Success("Document stored")
		return nil
	},
}

var memoryDeleteCmd = &cobra.Command{
	Use: "delete <agentID> <path>", Short: "Delete memory document", Args: cobra.ExactArgs(2),
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
	Use: "search <agentID>", Short: "Semantic search memory", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		query, _ := cmd.Flags().GetString("query")
		user, _ := cmd.Flags().GetString("user")
		body := buildBody("query", query, "user_id", user)
		data, err := c.Post("/v1/memory/"+url.PathEscape(args[0])+"/search", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

// --- Knowledge Graph (legacy query/extract/link — kgCmd declared in knowledge_graph.go) ---

var kgQueryCmd = &cobra.Command{
	Use: "query <agentID>", Short: "Query knowledge graph", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/knowledge-graph/" + url.PathEscape(args[0])
		if v, _ := cmd.Flags().GetString("entity"); v != "" {
			q := url.Values{}
			q.Set("entity", v)
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

var kgExtractCmd = &cobra.Command{
	Use: "extract <agentID>", Short: "Extract entities from text", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		text, _ := cmd.Flags().GetString("text")
		content, err := readContent(text)
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/knowledge-graph/"+url.PathEscape(args[0])+"/extract", map[string]any{"text": content})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgLinkCmd = &cobra.Command{
	Use: "link <agentID>", Short: "Create entity link", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		relation, _ := cmd.Flags().GetString("relation")
		_, err = c.Post("/v1/knowledge-graph/"+url.PathEscape(args[0])+"/link",
			map[string]any{"from": from, "to": to, "relation": relation})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Link created: %s -[%s]-> %s", from, relation, to))
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

	kgQueryCmd.Flags().String("entity", "", "Entity name filter")
	kgExtractCmd.Flags().String("text", "", "Text to extract from (or @filepath)")
	_ = kgExtractCmd.MarkFlagRequired("text")
	kgLinkCmd.Flags().String("from", "", "Source entity")
	kgLinkCmd.Flags().String("to", "", "Target entity")
	kgLinkCmd.Flags().String("relation", "", "Relation type")
	_ = kgLinkCmd.MarkFlagRequired("from")
	_ = kgLinkCmd.MarkFlagRequired("to")
	_ = kgLinkCmd.MarkFlagRequired("relation")

	memoryCmd.AddCommand(memoryListCmd, memoryGetCmd, memoryStoreCmd, memoryDeleteCmd, memorySearchCmd)
	// Legacy kg subcommands added to kgCmd (declared in knowledge_graph.go)
	kgCmd.AddCommand(kgQueryCmd, kgExtractCmd, kgLinkCmd)
	rootCmd.AddCommand(memoryCmd)
}
