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

// teams_workspace.go — workspace list/read/delete, extracted from teams.go (Phase 4 split).

var teamsWorkspaceCmd = &cobra.Command{Use: "workspace", Short: "Team workspace files"}

var teamsWorkspaceListCmd = &cobra.Command{
	Use:   "list <teamID>",
	Short: "List workspace files",
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
		data, err := ws.Call("teams.workspace.list", map[string]any{"team_id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var teamsWorkspaceReadCmd = &cobra.Command{
	Use:   "read <teamID> <path>",
	Short: "Read a workspace file",
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
		data, err := ws.Call("teams.workspace.read", map[string]any{
			"team_id": args[0], "path": args[1],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var teamsWorkspaceDeleteCmd = &cobra.Command{
	Use:   "delete <teamID> <path>",
	Short: "Delete a workspace file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this file?", cfg.Yes) {
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
		_, err = ws.Call("teams.workspace.delete", map[string]any{
			"team_id": args[0], "path": args[1],
		})
		if err != nil {
			return err
		}
		printer.Success("File deleted")
		return nil
	},
}

var teamsWorkspaceUploadCmd = &cobra.Command{
	Use:   "upload <teamID> <local-file>",
	Short: "Upload a file to team workspace (POST /v1/teams/{id}/workspace/upload)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[1]
		if _, err := os.Stat(filePath); err != nil {
			return fmt.Errorf("file not found: %s", filePath)
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open %s: %w", filePath, err)
		}

		pr, pw := io.Pipe()
		mw := newMultipartWriter(pw)
		ct := mw.contentType()

		go func() {
			defer f.Close()
			if err := mw.writeFile("file", filePath, f); err != nil {
				pw.CloseWithError(err)
				return
			}
			pw.CloseWithError(mw.close())
		}()

		path := "/v1/teams/" + args[0] + "/workspace/upload"
		if chat, _ := cmd.Flags().GetString("chat"); chat != "" {
			path += "?chat_id=" + url.QueryEscape(chat)
		}
		resp, err := c.PostRaw(path, ct, pr)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("upload failed [%d]: %s", resp.StatusCode, string(body))
		}
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			printer.Success("File uploaded")
			return nil
		}
		printer.Print(result)
		return nil
	},
}

var teamsWorkspaceMoveCmd = &cobra.Command{
	Use:   "move <teamID>",
	Short: "Rename/move a workspace file (PUT /v1/teams/{id}/workspace/move)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		if from == "" || to == "" {
			return fmt.Errorf("--from and --to are required")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{"from": {from}, "to": {to}}
		if chat, _ := cmd.Flags().GetString("chat"); chat != "" {
			q.Set("chat_id", chat)
		}
		path := "/v1/teams/" + args[0] + "/workspace/move?" + q.Encode()
		data, err := c.Put(path, nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	teamsWorkspaceUploadCmd.Flags().String("chat", "", "Chat ID (required for isolated workspaces)")
	teamsWorkspaceMoveCmd.Flags().String("from", "", "Source filename")
	teamsWorkspaceMoveCmd.Flags().String("to", "", "Destination filename")
	teamsWorkspaceMoveCmd.Flags().String("chat", "", "Chat ID (required for isolated workspaces)")
	_ = teamsWorkspaceMoveCmd.MarkFlagRequired("from")
	_ = teamsWorkspaceMoveCmd.MarkFlagRequired("to")

	teamsWorkspaceCmd.AddCommand(teamsWorkspaceListCmd, teamsWorkspaceReadCmd,
		teamsWorkspaceDeleteCmd, teamsWorkspaceUploadCmd, teamsWorkspaceMoveCmd)
}
