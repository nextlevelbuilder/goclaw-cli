package cmd

import (
	"encoding/json"
	"fmt"
	neturl "net/url"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage Knowledge Vault (documents, links, search, graph)",
}

// --- tree ---

var vaultTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show vault document tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		treePath, _ := cmd.Flags().GetString("path")
		urlPath := "/v1/vault/tree"
		if treePath != "" {
			urlPath += "?path=" + neturl.QueryEscape(treePath)
		}
		data, err := c.Get(urlPath)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalMap(data))
			return nil
		}
		// Render ASCII tree
		m := unmarshalMap(data)
		var entries []map[string]any
		if raw, ok := m["entries"]; ok {
			if b, err2 := json.Marshal(raw); err2 == nil {
				_ = json.Unmarshal(b, &entries)
			}
		}
		root := output.VaultEntriesToTreeNode("/", entries)
		output.PrintTreeRoot(root, cmd.OutOrStdout())
		return nil
	},
}

// --- search ---

var vaultSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search vault documents (FTS + vector)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		body := map[string]any{
			"query":       args[0],
			"max_results": limit,
		}
		if offset > 0 {
			body["offset"] = offset
		}
		data, err := c.Post("/v1/vault/search", body)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "TITLE", "PATH", "SCORE")
		for _, r := range unmarshalList(data) {
			tbl.AddRow(str(r, "id"), str(r, "title"), str(r, "path"), str(r, "score"))
		}
		printer.Print(tbl)
		return nil
	},
}

// --- rescan ---

var vaultRescanCmd = &cobra.Command{
	Use:   "rescan",
	Short: "Trigger vault workspace rescan (admin)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Trigger vault rescan?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/vault/rescan", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// --- graph ---

var vaultGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Get vault knowledge graph",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		format, _ := cmd.Flags().GetString("format")
		data, err := c.Get("/v1/vault/graph")
		if err != nil {
			return err
		}
		if format == "dot" {
			dot, dotErr := graphJSONToDOT(data)
			if dotErr != nil {
				return fmt.Errorf("dot transform failed: %w", dotErr)
			}
			fmt.Println(dot)
			return nil
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// graphJSONToDOT converts a vault graph JSON payload to Graphviz DOT format.
// Expected shape: {"nodes":[{"id":"...","label":"..."},...], "edges":[{"from":"...","to":"...","label":"..."},...]}
func graphJSONToDOT(data json.RawMessage) (string, error) {
	var g struct {
		Nodes []map[string]any `json:"nodes"`
		Edges []map[string]any `json:"edges"`
	}
	if err := json.Unmarshal(data, &g); err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.WriteString("digraph vault {\n")
	sb.WriteString(`  node [shape=box];` + "\n")
	for _, e := range g.Edges {
		from := str(e, "from_doc_id")
		if from == "" {
			from = str(e, "from")
		}
		to := str(e, "to_doc_id")
		if to == "" {
			to = str(e, "to")
		}
		label := str(e, "link_type")
		if label == "" {
			label = str(e, "label")
		}
		if from == "" || to == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("  %q -> %q", from, to))
		if label != "" {
			sb.WriteString(fmt.Sprintf(` [label=%q]`, label))
		}
		sb.WriteString(";\n")
	}
	sb.WriteString("}\n")
	return sb.String(), nil
}

func init() {
	vaultTreeCmd.Flags().String("path", "", "Path prefix to list (optional)")

	vaultSearchCmd.Flags().Int("limit", 20, "Max results to return")
	vaultSearchCmd.Flags().Int("offset", 0, "Pagination offset")

	vaultGraphCmd.Flags().String("format", "json", "Output format: json or dot (Graphviz)")

	vaultCmd.AddCommand(vaultTreeCmd, vaultSearchCmd, vaultRescanCmd, vaultGraphCmd)
	rootCmd.AddCommand(vaultCmd)
}
