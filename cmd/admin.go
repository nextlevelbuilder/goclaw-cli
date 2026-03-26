package cmd

import (
	"fmt"
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// --- Approvals ---

var approvalsCmd = &cobra.Command{Use: "approvals", Short: "Manage execution approvals"}

var approvalsListCmd = &cobra.Command{
	Use: "list", Short: "List pending approvals",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("exec.approval.list", nil)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "AGENT", "TOOL", "STATUS")
		for _, a := range unmarshalList(data) {
			tbl.AddRow(str(a, "id"), str(a, "agent_id"), str(a, "tool_name"), str(a, "status"))
		}
		printer.Print(tbl)
		return nil
	},
}

var approvalsApproveCmd = &cobra.Command{
	Use: "approve <id>", Short: "Approve execution", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("exec.approval.approve", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Execution approved")
		return nil
	},
}

var approvalsDenyCmd = &cobra.Command{
	Use: "deny <id>", Short: "Deny execution", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		reason, _ := cmd.Flags().GetString("reason")
		_, err = ws.Call("exec.approval.deny", map[string]any{"id": args[0], "reason": reason})
		if err != nil {
			return err
		}
		printer.Success("Execution denied")
		return nil
	},
}

// --- Delegations ---

var delegationsCmd = &cobra.Command{Use: "delegations", Short: "View delegation history"}

var delegationsListCmd = &cobra.Command{
	Use: "list", Short: "List delegations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("agent"); v != "" {
			q.Set("agent_id", v)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		path := "/v1/delegations"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var delegationsGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get delegation details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/delegations/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	// Approvals
	approvalsDenyCmd.Flags().String("reason", "", "Denial reason")
	approvalsCmd.AddCommand(approvalsListCmd, approvalsApproveCmd, approvalsDenyCmd)

	// Delegations
	delegationsListCmd.Flags().String("agent", "", "Agent ID")
	delegationsListCmd.Flags().Int("limit", 20, "Max results")
	delegationsCmd.AddCommand(delegationsListCmd, delegationsGetCmd)

	rootCmd.AddCommand(approvalsCmd, delegationsCmd)
}
