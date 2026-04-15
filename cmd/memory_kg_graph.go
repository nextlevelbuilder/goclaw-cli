package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// memory_kg_graph.go — KG traverse, stats, graph commands.
// Extracted from memory_kg.go to keep files <200 LoC.

var memoryKGTraverseCmd = &cobra.Command{
	Use:   "traverse <agentID>",
	Short: "Traverse the KG from an entity",
	Long: `Traverse the knowledge graph starting from a given entity ID up to --depth hops.

POST /v1/agents/{id}/kg/traverse

Example:
  goclaw memory kg traverse agent-1 --from=entity-42 --depth=2
  goclaw memory kg traverse agent-1 --from=entity-42 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		depth, _ := cmd.Flags().GetInt("depth")
		if from == "" {
			return fmt.Errorf("--from is required")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := map[string]any{"entity_id": from}
		if depth > 0 {
			body["depth"] = depth
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/kg/traverse", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGStatsCmd = &cobra.Command{
	Use:   "stats <agentID>",
	Short: "Get KG statistics for an agent",
	Long: `Retrieve entity count, relation count, and other KG statistics.

GET /v1/agents/{id}/kg/stats

Example:
  goclaw memory kg stats agent-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/kg/stats")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGGraphCmd = &cobra.Command{
	Use:   "graph <agentID>",
	Short: "Get full KG graph (or compact form)",
	Long: `Retrieve the full knowledge graph for an agent. Use --compact for a minimal representation.

GET /v1/agents/{id}/kg/graph
GET /v1/agents/{id}/kg/graph/compact  (with --compact)

Example:
  goclaw memory kg graph agent-1
  goclaw memory kg graph agent-1 --compact --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		compact, _ := cmd.Flags().GetBool("compact")
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/agents/" + args[0] + "/kg/graph"
		if compact {
			path += "/compact"
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	memoryKGTraverseCmd.Flags().String("from", "", "Starting entity ID")
	memoryKGTraverseCmd.Flags().Int("depth", 2, "Traversal depth (hops)")
	memoryKGGraphCmd.Flags().Bool("compact", false, "Return compact graph representation")

	memoryKGCmd.AddCommand(memoryKGTraverseCmd, memoryKGStatsCmd, memoryKGGraphCmd)
}
