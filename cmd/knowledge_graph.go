package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var kgCmd = &cobra.Command{Use: "knowledge-graph", Aliases: []string{"kg"}, Short: "Knowledge graph operations"}

var kgEntitiesCmd = &cobra.Command{Use: "entities", Short: "Manage knowledge graph entities"}

var kgEntitiesListCmd = &cobra.Command{
	Use: "list <agent-id>", Short: "List KG entities", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/kg/entities")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "TYPE", "PROPERTIES_COUNT")
		for _, e := range unmarshalList(data) {
			tbl.AddRow(str(e, "id"), str(e, "name"), str(e, "type"), str(e, "properties_count"))
		}
		printer.Print(tbl)
		return nil
	},
}

var kgEntitiesGetCmd = &cobra.Command{
	Use: "get <agent-id> <id>", Short: "Get KG entity", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/kg/entities/" + url.PathEscape(args[1]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgEntitiesCreateCmd = &cobra.Command{
	Use: "create <agent-id>", Short: "Create KG entity", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		dataFlag, _ := cmd.Flags().GetString("data")
		content, err := readContent(dataFlag)
		if err != nil {
			return err
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(content), &body); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		data, err := c.Post("/v1/agents/"+url.PathEscape(args[0])+"/kg/entities", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgEntitiesDeleteCmd = &cobra.Command{
	Use: "delete <agent-id> <id>", Short: "Delete KG entity", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this entity?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/agents/" + url.PathEscape(args[0]) + "/kg/entities/" + url.PathEscape(args[1]))
		if err != nil {
			return err
		}
		printer.Success("Entity deleted")
		return nil
	},
}

var kgTraverseCmd = &cobra.Command{
	Use: "traverse <agent-id>", Short: "Traverse the knowledge graph", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		from, _ := cmd.Flags().GetString("from")
		data, err := c.Post("/v1/agents/"+url.PathEscape(args[0])+"/kg/traverse", buildBody("from", from))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgGraphCmd = &cobra.Command{
	Use: "graph <agent-id>", Short: "Get full knowledge graph", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/kg/graph")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgStatsCmd = &cobra.Command{
	Use: "stats <agent-id>", Short: "Get knowledge graph stats", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/kg/stats")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	kgEntitiesCreateCmd.Flags().String("data", "", "Entity JSON body (or @filepath)")
	_ = kgEntitiesCreateCmd.MarkFlagRequired("data")
	kgTraverseCmd.Flags().String("from", "", "Starting entity ID")
	_ = kgTraverseCmd.MarkFlagRequired("from")

	kgEntitiesCmd.AddCommand(kgEntitiesListCmd, kgEntitiesGetCmd, kgEntitiesCreateCmd, kgEntitiesDeleteCmd)
	kgCmd.AddCommand(kgEntitiesCmd, kgTraverseCmd, kgGraphCmd, kgStatsCmd)
	rootCmd.AddCommand(kgCmd)
}
