package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// memory_kg_dedup.go — KG deduplication: scan, list, merge, dismiss.
// All HTTP endpoints under /v1/agents/{id}/kg/dedup* and /v1/agents/{id}/kg/merge.

var memoryKGDedupCmd = &cobra.Command{
	Use:   "dedup",
	Short: "Knowledge graph deduplication",
}

var memoryKGDedupScanCmd = &cobra.Command{
	Use:   "scan <agentID>",
	Short: "Scan for duplicate KG entities",
	Long: `Trigger a deduplication scan for an agent's knowledge graph.
The server identifies entity pairs with high similarity scores.

POST /v1/agents/{id}/kg/dedup/scan

Example:
  goclaw memory kg dedup scan agent-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/kg/dedup/scan", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGDedupListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List pending deduplication candidates",
	Long: `List entity pairs identified as potential duplicates from the last dedup scan.

GET /v1/agents/{id}/kg/dedup

Example:
  goclaw memory kg dedup list agent-1
  goclaw memory kg dedup list agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/kg/dedup")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var memoryKGDedupMergeCmd = &cobra.Command{
	Use:   "merge <agentID> <entityA> <entityB>",
	Short: "Merge two KG entities",
	Long: `Merge entityB into entityA, consolidating relations and attributes.
EntityB is soft-deleted after merge. This operation is irreversible without a restore.

POST /v1/agents/{id}/kg/merge

Example:
  goclaw memory kg dedup merge agent-1 entity-1 entity-2 --yes`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Merge entity %s into %s?", args[2], args[1]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/kg/merge", map[string]any{
			"entity_a": args[1],
			"entity_b": args[2],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGDedupDismissCmd = &cobra.Command{
	Use:   "dismiss <agentID> <candidateID>",
	Short: "Dismiss a deduplication candidate",
	Long: `Mark a deduplication candidate as dismissed (not a true duplicate).

POST /v1/agents/{id}/kg/dedup/dismiss

Example:
  goclaw memory kg dedup dismiss agent-1 candidate-99`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/agents/"+args[0]+"/kg/dedup/dismiss",
			map[string]any{"candidate_id": args[1]})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("Candidate %s dismissed", args[1]))
		return nil
	},
}

func init() {
	memoryKGDedupCmd.AddCommand(
		memoryKGDedupScanCmd,
		memoryKGDedupListCmd,
		memoryKGDedupMergeCmd,
		memoryKGDedupDismissCmd,
	)
	// Registered onto memoryKGCmd in memory_kg.go init() is done via AddCommand in memory.go
	memoryKGCmd.AddCommand(memoryKGDedupCmd)
}
