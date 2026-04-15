package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// memory_kg.go — KG entities CRUD, traverse, stats, graph.
// Legacy (query/extract/link) → memory_kg_legacy.go
// Dedup → memory_kg_dedup.go

// memory_kg.go — Knowledge Graph: entities CRUD, traverse, stats, graph, legacy query/extract/link.
// All HTTP. Registered under memoryCmd as "kg" subcommand group.

var memoryKGCmd = &cobra.Command{
	Use:     "kg",
	Aliases: []string{"knowledge-graph"},
	Short:   "Knowledge graph operations",
}

// --- Entities ---

var memoryKGEntitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "Manage KG entities",
}

var memoryKGEntitiesListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List all KG entities for an agent",
	Long: `List all knowledge graph entities for an agent.

GET /v1/agents/{id}/kg/entities

Example:
  goclaw memory kg entities list agent-1
  goclaw memory kg entities list agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/kg/entities")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var memoryKGEntitiesGetCmd = &cobra.Command{
	Use:   "get <agentID> <entityID>",
	Short: "Get a KG entity by ID",
	Long: `Retrieve a specific knowledge graph entity.

GET /v1/agents/{id}/kg/entities/{entityID}

Example:
  goclaw memory kg entities get agent-1 entity-42`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get(fmt.Sprintf("/v1/agents/%s/kg/entities/%s", args[0], args[1]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGEntitiesUpsertCmd = &cobra.Command{
	Use:   "upsert <agentID>",
	Short: "Upsert KG entities from a JSON file",
	Long: `Create or update KG entities for an agent. Reads entity data from a JSON file.

POST /v1/agents/{id}/kg/entities

The file must contain a valid JSON object or array of entity objects.

Example:
  goclaw memory kg entities upsert agent-1 --file=./entities.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		content, err := readContent("@" + filePath)
		if err != nil {
			return fmt.Errorf("read entity file: %w", err)
		}
		// Parse as either object or array — surface JSON errors instead of
		// silently sending an empty body.
		var body any
		var arr []any
		if err := json.Unmarshal([]byte(content), &arr); err == nil && len(arr) > 0 {
			body = arr
		} else {
			var obj map[string]any
			if err := json.Unmarshal([]byte(content), &obj); err != nil {
				return fmt.Errorf("invalid JSON in %s: %w", filePath, err)
			}
			if len(obj) == 0 {
				return fmt.Errorf("empty entity payload in %s", filePath)
			}
			body = obj
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/kg/entities", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var memoryKGEntitiesDeleteCmd = &cobra.Command{
	Use:   "delete <agentID> <entityID>",
	Short: "Delete a KG entity",
	Long: `Delete a knowledge graph entity. This may cascade to linked relations.

DELETE /v1/agents/{id}/kg/entities/{entityID}

Example:
  goclaw memory kg entities delete agent-1 entity-42 --yes`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete entity %s?", args[1]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete(fmt.Sprintf("/v1/agents/%s/kg/entities/%s", args[0], args[1]))
		if err != nil {
			return err
		}
		printer.Success("Entity deleted")
		return nil
	},
}

func init() {
	// entity flags
	memoryKGEntitiesUpsertCmd.Flags().String("file", "", "Path to JSON entity file")
	_ = memoryKGEntitiesUpsertCmd.MarkFlagRequired("file")

	memoryKGEntitiesCmd.AddCommand(
		memoryKGEntitiesListCmd, memoryKGEntitiesGetCmd,
		memoryKGEntitiesUpsertCmd, memoryKGEntitiesDeleteCmd,
	)

	memoryKGCmd.AddCommand(memoryKGEntitiesCmd)
	// traverse/stats/graph → memory_kg_graph.go init()
	// dedup             → memory_kg_dedup.go init()
	// legacy            → memory_kg_legacy.go init()
}
