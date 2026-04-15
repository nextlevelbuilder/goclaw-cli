package cmd

import (
	"fmt"
	"io"
	neturl "net/url"
	"os"
	"strings"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var vaultDocsCmd = &cobra.Command{
	Use:   "documents",
	Short: "Manage vault documents",
}

// --- list ---

var vaultDocsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List vault documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q, _ := cmd.Flags().GetString("q")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		urlPath := fmt.Sprintf("/v1/vault/documents?limit=%d&offset=%d", limit, offset)
		if q != "" {
			urlPath += "&q=" + neturl.QueryEscape(q)
		}
		data, err := c.Get(urlPath)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalMap(data))
			return nil
		}
		m := unmarshalMap(data)
		docs := extractDocsList(m)
		tbl := output.NewTable("ID", "TITLE", "PATH", "TYPE", "SCOPE")
		for _, d := range docs {
			tbl.AddRow(str(d, "id"), str(d, "title"), str(d, "path"), str(d, "doc_type"), str(d, "scope"))
		}
		printer.Print(tbl)
		return nil
	},
}

// --- get ---

var vaultDocsGetCmd = &cobra.Command{
	Use:   "get <docID>",
	Short: "Get a vault document by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/vault/documents/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// --- create ---

var vaultDocsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a vault document",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		title, _ := cmd.Flags().GetString("title")
		contentVal, _ := cmd.Flags().GetString("content")
		fileVal, _ := cmd.Flags().GetString("file")
		docPath, _ := cmd.Flags().GetString("path")
		docType, _ := cmd.Flags().GetString("doc-type")
		scope, _ := cmd.Flags().GetString("scope")

		if contentVal != "" && fileVal != "" {
			return fmt.Errorf("--content and --file are mutually exclusive")
		}

		// Read content — from --content literal, or --file path (or "-" for stdin)
		if fileVal != "" {
			data, err := readFileOrStdin(fileVal)
			if err != nil {
				return err
			}
			contentVal = string(data)
			if docPath == "" && fileVal != "-" {
				docPath = fileVal
			}
		}

		if docPath == "" {
			return fmt.Errorf("--path is required")
		}
		if title == "" {
			return fmt.Errorf("--title is required")
		}
		if contentVal == "" {
			return fmt.Errorf("--content or --file is required")
		}

		body := buildBody(
			"path", docPath,
			"title", title,
			"content", contentVal,
			"doc_type", docType,
			"scope", scope,
		)
		data, err := c.Post("/v1/vault/documents", body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		printer.Success(fmt.Sprintf("Document created: %s (ID: %s)", str(m, "title"), str(m, "id")))
		return nil
	},
}

// --- update ---

var vaultDocsUpdateCmd = &cobra.Command{
	Use:   "update <docID>",
	Short: "Update a vault document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			body["title"] = v
		}
		if cmd.Flags().Changed("doc-type") {
			v, _ := cmd.Flags().GetString("doc-type")
			body["doc_type"] = v
		}
		if cmd.Flags().Changed("scope") {
			v, _ := cmd.Flags().GetString("scope")
			body["scope"] = v
		}
		if len(body) == 0 {
			return fmt.Errorf("no fields to update — use --title, --doc-type, --scope")
		}
		_, err = c.Put("/v1/vault/documents/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Document updated")
		return nil
	},
}

// --- delete ---

var vaultDocsDeleteCmd = &cobra.Command{
	Use:   "delete <docID>",
	Short: "Delete a vault document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete document %s?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/vault/documents/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Document deleted")
		return nil
	},
}

// --- links (doc links listing) ---

var vaultDocsLinksCmd = &cobra.Command{
	Use:   "links <docID>",
	Short: "Show links for a vault document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/vault/documents/" + args[0] + "/links")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalMap(data))
			return nil
		}
		m := unmarshalMap(data)
		tbl := output.NewTable("DIRECTION", "DOC_ID", "LINK_TYPE")
		if outlinks, ok := m["outlinks"]; ok {
			if links, ok2 := toMapSlice(outlinks); ok2 {
				for _, l := range links {
					tbl.AddRow("out", str(l, "to_doc_id"), str(l, "link_type"))
				}
			}
		}
		if backlinks, ok := m["backlinks"]; ok {
			if links, ok2 := toMapSlice(backlinks); ok2 {
				for _, l := range links {
					tbl.AddRow("in", str(l, "from_doc_id"), str(l, "link_type"))
				}
			}
		}
		printer.Print(tbl)
		return nil
	},
}

// extractDocsList pulls []map[string]any from a {documents:[...], total:N} envelope.
func extractDocsList(m map[string]any) []map[string]any {
	raw, ok := m["documents"]
	if !ok {
		return nil
	}
	docs, _ := toMapSlice(raw)
	return docs
}

// toMapSlice converts an any value to []map[string]any, returning (slice, ok).
func toMapSlice(v any) ([]map[string]any, bool) {
	slice, ok := v.([]any)
	if !ok {
		return nil, false
	}
	result := make([]map[string]any, 0, len(slice))
	for _, item := range slice {
		if m, ok2 := item.(map[string]any); ok2 {
			result = append(result, m)
		}
	}
	return result, true
}

// readFileOrStdin reads content from a file path or stdin ("-").
func readFileOrStdin(path string) (string, error) {
	if path == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(b), nil
	}
	if strings.HasPrefix(path, "@") {
		path = path[1:]
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return string(b), nil
}

func init() {
	vaultDocsListCmd.Flags().String("q", "", "Filter query")
	vaultDocsListCmd.Flags().Int("limit", 20, "Max results")
	vaultDocsListCmd.Flags().Int("offset", 0, "Pagination offset")

	vaultDocsCreateCmd.Flags().String("title", "", "Document title (required)")
	vaultDocsCreateCmd.Flags().String("path", "", "Document path in vault (required)")
	vaultDocsCreateCmd.Flags().String("content", "", "Inline content (mutually exclusive with --file)")
	vaultDocsCreateCmd.Flags().String("file", "", "File to read content from; use - for stdin")
	vaultDocsCreateCmd.Flags().String("doc-type", "note", "Document type: note, context, memory, skill, episodic, media")
	vaultDocsCreateCmd.Flags().String("scope", "shared", "Scope: personal, team, shared")
	_ = vaultDocsCreateCmd.MarkFlagRequired("title")
	_ = vaultDocsCreateCmd.MarkFlagRequired("path")

	vaultDocsUpdateCmd.Flags().String("title", "", "New title")
	vaultDocsUpdateCmd.Flags().String("doc-type", "", "New doc type")
	vaultDocsUpdateCmd.Flags().String("scope", "", "New scope")

	vaultDocsCmd.AddCommand(
		vaultDocsListCmd,
		vaultDocsGetCmd,
		vaultDocsCreateCmd,
		vaultDocsUpdateCmd,
		vaultDocsDeleteCmd,
		vaultDocsLinksCmd,
	)
	vaultCmd.AddCommand(vaultDocsCmd)
}
