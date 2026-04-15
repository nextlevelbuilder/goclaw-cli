package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// memory_kg_legacy.go — legacy KG commands (query/extract/link) kept for backward compat.
// New code should use "memory kg entities" and related subcommands instead.
// Extracted from memory_kg.go to keep files <200 LoC.

var memoryKGQueryCmd = &cobra.Command{
	Use:   "query <agentID>",
	Short: "Query knowledge graph (legacy — prefer 'memory kg entities list')",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/knowledge-graph/" + args[0]
		if v, _ := cmd.Flags().GetString("entity"); v != "" {
			path += "?entity=" + v
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGExtractCmd = &cobra.Command{
	Use:   "extract <agentID>",
	Short: "Extract entities from text (legacy)",
	Args:  cobra.ExactArgs(1),
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
		data, err := c.Post("/v1/knowledge-graph/"+args[0]+"/extract",
			map[string]any{"text": content})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGLinkCmd = &cobra.Command{
	Use:   "link <agentID>",
	Short: "Create entity link (legacy)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		relation, _ := cmd.Flags().GetString("relation")
		_, err = c.Post("/v1/knowledge-graph/"+args[0]+"/link",
			map[string]any{"from": from, "to": to, "relation": relation})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Link created: %s -[%s]-> %s", from, relation, to))
		return nil
	},
}

func init() {
	memoryKGQueryCmd.Flags().String("entity", "", "Entity name filter")

	memoryKGExtractCmd.Flags().String("text", "", "Text to extract from (or @filepath)")
	_ = memoryKGExtractCmd.MarkFlagRequired("text")

	memoryKGLinkCmd.Flags().String("from", "", "Source entity")
	memoryKGLinkCmd.Flags().String("to", "", "Target entity")
	memoryKGLinkCmd.Flags().String("relation", "", "Relation type")
	_ = memoryKGLinkCmd.MarkFlagRequired("from")
	_ = memoryKGLinkCmd.MarkFlagRequired("to")
	_ = memoryKGLinkCmd.MarkFlagRequired("relation")

	memoryKGCmd.AddCommand(memoryKGQueryCmd, memoryKGExtractCmd, memoryKGLinkCmd)
}
