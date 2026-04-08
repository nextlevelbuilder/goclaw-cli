package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var memoryIndexCmd = &cobra.Command{
	Use: "index <agentID> <path>", Short: "Index a memory document", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+url.PathEscape(args[0])+"/memory/index",
			map[string]any{"path": args[1]})
		if err != nil {
			return err
		}
		printer.Success("Document indexed")
		return nil
	},
}

var memoryIndexAllCmd = &cobra.Command{
	Use: "index-all <agentID>", Short: "Index all memory documents", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+url.PathEscape(args[0])+"/memory/index-all", nil)
		if err != nil {
			return err
		}
		printer.Success("All documents indexed")
		return nil
	},
}

var memoryChunksCmd = &cobra.Command{
	Use: "chunks <agentID>", Short: "List memory chunks", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/memory/chunks")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "DOCUMENT", "CONTENT", "CREATED")
		for _, ch := range unmarshalList(data) {
			tbl.AddRow(str(ch, "id"), str(ch, "document_path"), str(ch, "content"), str(ch, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	memoryCmd.AddCommand(memoryIndexCmd, memoryIndexAllCmd, memoryChunksCmd)
}
