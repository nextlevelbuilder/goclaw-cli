package cmd

import (
	"fmt"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// memory_index.go — memory chunks, index, index-all, documents-global.
// HTTP endpoints under /v1/agents/{id}/memory/* and /v1/memory/documents (global).

var memoryChunksCmd = &cobra.Command{
	Use:   "chunks <agentID>",
	Short: "List memory chunks for an agent",
	Long: `List raw memory chunks stored for an agent (post-chunking, pre-embedding).

GET /v1/agents/{id}/memory/chunks

Example:
  goclaw memory chunks agent-1
  goclaw memory chunks agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/memory/chunks")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var memoryIndexCmd = &cobra.Command{
	Use:   "index <agentID> <path>",
	Short: "Trigger memory re-index for a document path",
	Long: `Trigger a memory re-index for a specific document path.
This is an expensive server-side operation — use sparingly.

POST /v1/agents/{id}/memory/index

Example:
  goclaw memory index agent-1 docs/architecture.md`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if output.IsTTY(int(os.Stdout.Fd())) {
			fmt.Fprintln(os.Stderr, "Warning: memory index triggers an expensive server-side operation.")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/memory/index",
			map[string]any{"path": args[1]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryIndexAllCmd = &cobra.Command{
	Use:   "index-all <agentID>",
	Short: "Trigger full memory re-index for an agent",
	Long: `Trigger a full memory re-index for all documents of an agent.
This is a heavy server-side operation. Server-side rate limiting applies.

POST /v1/agents/{id}/memory/index-all

Example:
  goclaw memory index-all agent-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if output.IsTTY(int(os.Stdout.Fd())) {
			fmt.Fprintln(os.Stderr, "Warning: index-all triggers a full re-index. This may take time.")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/memory/index-all", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryDocumentsGlobalCmd = &cobra.Command{
	Use:   "documents-global",
	Short: "List all memory documents (global, no agent scope)",
	Long: `List all memory documents across all agents (tenant-wide).
Requires admin permissions.

GET /v1/memory/documents

Example:
  goclaw memory documents-global
  goclaw memory documents-global --output=json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/memory/documents")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	// All registered onto memoryCmd in memory.go init()
}
