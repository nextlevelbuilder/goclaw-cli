package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{Use: "export", Short: "Export resources (agents, teams, skills, mcp)"}
var importCmd = &cobra.Command{Use: "import", Short: "Import resources (agents, teams, skills, mcp)"}

// --- Agent Export/Import ---

var exportAgentPreviewCmd = &cobra.Command{
	Use: "agent-preview <agentID>", Short: "Preview agent export", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/export/preview")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var exportAgentCmd = &cobra.Command{
	Use: "agent <agentID>", Short: "Export an agent", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		resp, err := c.GetRaw("/v1/agents/" + url.PathEscape(args[0]) + "/export")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return writeExportFile(resp.Body, outFile, "agent-"+args[0]+".json")
	},
}

var importAgentPreviewCmd = &cobra.Command{
	Use: "agent-preview <file>", Short: "Preview agent import", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/import/preview", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var importAgentCmd = &cobra.Command{
	Use: "agent <file>", Short: "Import agents from file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Import agents from "+args[0]+"?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/import", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// --- Team Export/Import ---

var exportTeamPreviewCmd = &cobra.Command{
	Use: "team-preview <teamID>", Short: "Preview team export", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/teams/" + url.PathEscape(args[0]) + "/export/preview")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var exportTeamCmd = &cobra.Command{
	Use: "team <teamID>", Short: "Export a team", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		resp, err := c.GetRaw("/v1/teams/" + url.PathEscape(args[0]) + "/export")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return writeExportFile(resp.Body, outFile, "team-"+args[0]+".json")
	},
}

var importTeamPreviewCmd = &cobra.Command{
	Use: "team-preview <file>", Short: "Preview team import", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/teams/import/preview", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var importTeamCmd = &cobra.Command{
	Use: "team <file>", Short: "Import team from file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Import team from "+args[0]+"?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/teams/import", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// --- Skills Export/Import ---

var exportSkillsPreviewCmd = &cobra.Command{
	Use: "skills-preview", Short: "Preview skills export",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/skills/export/preview")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var exportSkillsCmd = &cobra.Command{
	Use: "skills", Short: "Export skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		resp, err := c.GetRaw("/v1/skills/export")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return writeExportFile(resp.Body, outFile, "skills-export.json")
	},
}

var importSkillsPreviewCmd = &cobra.Command{
	Use: "skills-preview <file>", Short: "Preview skills import", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/skills/import/preview", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var importSkillsCmd = &cobra.Command{
	Use: "skills <file>", Short: "Import skills from file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Import skills from "+args[0]+"?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/skills/import", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// --- MCP Export/Import ---

var exportMCPPreviewCmd = &cobra.Command{
	Use: "mcp-preview", Short: "Preview MCP servers export",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/mcp/export/preview")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var exportMCPCmd = &cobra.Command{
	Use: "mcp", Short: "Export MCP servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		resp, err := c.GetRaw("/v1/mcp/export")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return writeExportFile(resp.Body, outFile, "mcp-export.json")
	},
}

var importMCPPreviewCmd = &cobra.Command{
	Use: "mcp-preview <file>", Short: "Preview MCP servers import", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/mcp/import/preview", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var importMCPCmd = &cobra.Command{
	Use: "mcp <file>", Short: "Import MCP servers from file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Import MCP servers from "+args[0]+"?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body, err := readJSONFile(args[0])
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/mcp/import", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// writeExportFile saves response body to a file.
func writeExportFile(body io.Reader, outFile, defaultName string) error {
	if outFile == "" {
		outFile = defaultName
	}
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, body)
	if err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close export file: %w", err)
	}
	printer.Success(fmt.Sprintf("Exported %d bytes to %s", n, outFile))
	return nil
}

// readJSONFile reads and parses a JSON file.
func readJSONFile(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var body any
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return body, nil
}

func init() {
	for _, c := range []*cobra.Command{exportAgentCmd, exportTeamCmd, exportSkillsCmd, exportMCPCmd} {
		c.Flags().StringP("output", "o", "", "Output file path")
	}

	exportCmd.AddCommand(exportAgentPreviewCmd, exportAgentCmd, exportTeamPreviewCmd, exportTeamCmd,
		exportSkillsPreviewCmd, exportSkillsCmd, exportMCPPreviewCmd, exportMCPCmd)
	importCmd.AddCommand(importAgentPreviewCmd, importAgentCmd, importTeamPreviewCmd, importTeamCmd,
		importSkillsPreviewCmd, importSkillsCmd, importMCPPreviewCmd, importMCPCmd)
	rootCmd.AddCommand(exportCmd, importCmd)
}
