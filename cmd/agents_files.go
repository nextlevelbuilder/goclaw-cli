package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

// agents_files.go — manage global agent context files via WS RPC `agents.files.*`.
// Allowed names (server-validated): AGENTS.md, SOUL.md, IDENTITY.md, USER.md,
// USER_PREDEFINED.md, CAPABILITIES.md, BOOTSTRAP.md, MEMORY.json, HEARTBEAT.

var agentsFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage global agent context files (AGENTS.md, SOUL.md, ...)",
}

var agentsFilesListCmd = &cobra.Command{
	Use:   "list <agentID>",
	Short: "List context files for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("agents.files.list", map[string]any{"agentId": args[0]})
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		files, _ := m["files"].([]any)
		if cfg.OutputFormat != "table" {
			printer.Print(files)
			return nil
		}
		tbl := output.NewTable("NAME", "MISSING", "SIZE")
		for _, f := range files {
			row, _ := f.(map[string]any)
			if row == nil {
				continue
			}
			tbl.AddRow(str(row, "name"), str(row, "missing"), str(row, "size"))
		}
		printer.Print(tbl)
		return nil
	},
}

var agentsFilesGetCmd = &cobra.Command{
	Use:   "get <agentID> <name>",
	Short: "Read a context file (e.g. AGENTS.md, SOUL.md)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("agents.files.get", map[string]any{
			"agentId": args[0], "name": args[1],
		})
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		file, _ := m["file"].(map[string]any)
		if file == nil {
			return fmt.Errorf("unexpected response shape")
		}
		// In raw/text mode (--output=text or future raw flag) we'd dump just content.
		// For now: JSON/YAML get full envelope; table mode dumps content to stdout.
		if cfg.OutputFormat == "table" {
			if missing, _ := file["missing"].(bool); missing {
				fmt.Fprintf(cmd.ErrOrStderr(), "File %s is not set\n", args[1])
				return nil
			}
			fmt.Println(str(file, "content"))
			return nil
		}
		printer.Print(file)
		return nil
	},
}

var agentsFilesSetCmd = &cobra.Command{
	Use:   "set <agentID> <name>",
	Short: "Write a context file (requires --content and --yes)",
	Long: `Write content to a global agent context file.

The change affects the canonical agent definition. Use --propagate to also push
the new content to every existing user instance for this agent.

Examples:
  goclaw agents files set my-agent SOUL.md --content=@./soul.md --yes
  echo "..." | goclaw agents files set my-agent AGENTS.md --content=@- --yes`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("content")
		if path == "" {
			return fmt.Errorf("--content is required")
		}
		content, err := readContent(path)
		if err != nil {
			return err
		}
		propagate, _ := cmd.Flags().GetBool("propagate")
		msg := fmt.Sprintf("Overwrite %s for agent %s?", args[1], args[0])
		if !tui.Confirm(msg, cfg.Yes) {
			return nil
		}
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("agents.files.set", map[string]any{
			"agentId":   args[0],
			"name":      args[1],
			"content":   content,
			"propagate": propagate,
		})
		if err != nil {
			return err
		}
		printer.Success(fmt.Sprintf("File %s saved", args[1]))
		return nil
	},
}

func init() {
	agentsFilesSetCmd.Flags().String("content", "", "Content (@filepath or literal; @- for stdin)")
	agentsFilesSetCmd.Flags().Bool("propagate", false, "Push change to all existing user instances")
	_ = agentsFilesSetCmd.MarkFlagRequired("content")

	agentsFilesCmd.AddCommand(agentsFilesListCmd, agentsFilesGetCmd, agentsFilesSetCmd)
	agentsCmd.AddCommand(agentsFilesCmd)
}
