package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var agentsFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage agent context files",
}

var agentsFilesListCmd = &cobra.Command{
	Use: "list <agentID>", Short: "List agent files", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("agents.files.list", map[string]any{"agent_id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var agentsFilesGetCmd = &cobra.Command{
	Use: "get <agentID> <fileName>", Short: "Get agent file content", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("agents.files.get", map[string]any{"agent_id": args[0], "file_name": args[1]})
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		if content := str(m, "content"); content != "" {
			fmt.Println(content)
		} else {
			printer.Print(m)
		}
		return nil
	},
}

var agentsFilesSetCmd = &cobra.Command{
	Use: "set <agentID> <fileName>", Short: "Set agent file content", Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		contentVal, _ := cmd.Flags().GetString("content")
		content, err := readContent(contentVal)
		if err != nil {
			return err
		}
		_, err = ws.Call("agents.files.set", map[string]any{
			"agent_id": args[0], "file_name": args[1], "content": content,
		})
		if err != nil {
			return err
		}
		printer.Success("File updated")
		return nil
	},
}

func init() {
	agentsFilesSetCmd.Flags().String("content", "", "Content (or @filepath)")
	_ = agentsFilesSetCmd.MarkFlagRequired("content")

	agentsFilesCmd.AddCommand(agentsFilesListCmd, agentsFilesGetCmd, agentsFilesSetCmd)
	agentsCmd.AddCommand(agentsFilesCmd)
}
