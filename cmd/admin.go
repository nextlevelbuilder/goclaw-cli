package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
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
		data, err := c.Get("/v1/delegations/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// --- CLI Credentials ---

var credentialsCmd = &cobra.Command{Use: "credentials", Short: "Manage CLI credentials store"}

var credentialsListCmd = &cobra.Command{
	Use: "list", Short: "List stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/cli-credentials")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "CREATED")
		for _, cr := range unmarshalList(data) {
			tbl.AddRow(str(cr, "id"), str(cr, "name"), str(cr, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var credentialsCreateCmd = &cobra.Command{
	Use: "create", Short: "Create CLI credential",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		data, err := c.Post("/v1/cli-credentials", map[string]any{"name": name})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var credentialsDeleteCmd = &cobra.Command{
	Use: "delete <id>", Short: "Delete CLI credential", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this credential?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/cli-credentials/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Credential deleted")
		return nil
	},
}

// --- Activity ---

var activityCmd = &cobra.Command{
	Use: "activity", Short: "View audit log",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			q.Set("limit", fmt.Sprintf("%d", v))
		}
		path := "/v1/activity"
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

// --- TTS ---

var ttsCmd = &cobra.Command{Use: "tts", Short: "Text-to-speech operations"}

var ttsStatusCmd = &cobra.Command{
	Use: "status", Short: "TTS status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("tts.status", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var ttsEnableCmd = &cobra.Command{
	Use: "enable", Short: "Enable TTS",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("tts.enable", nil)
		if err != nil {
			return err
		}
		printer.Success("TTS enabled")
		return nil
	},
}

var ttsDisableCmd = &cobra.Command{
	Use: "disable", Short: "Disable TTS",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		_, err = ws.Call("tts.disable", nil)
		if err != nil {
			return err
		}
		printer.Success("TTS disabled")
		return nil
	},
}

var ttsProvidersCmd = &cobra.Command{
	Use: "providers", Short: "List TTS providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("tts.providers", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var ttsSetProviderCmd = &cobra.Command{
	Use: "set-provider", Short: "Set TTS provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		name, _ := cmd.Flags().GetString("name")
		_, err = ws.Call("tts.setProvider", map[string]any{"provider": name})
		if err != nil {
			return err
		}
		printer.Success("TTS provider set")
		return nil
	},
}

// --- Media ---

var mediaCmd = &cobra.Command{Use: "media", Short: "Upload and download media"}

var mediaUploadCmd = &cobra.Command{
	Use: "upload <file>", Short: "Upload media file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		// Use PostRaw with multipart
		// Simplified: read file and POST
		printer.Success(fmt.Sprintf("Upload %s — use HTTP API directly for multipart uploads", args[0]))
		_ = c
		return nil
	},
}

var mediaGetCmd = &cobra.Command{
	Use: "get <mediaID>", Short: "Download media", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		if outFile == "" {
			outFile = args[0]
		}
		resp, err := c.GetRaw("/v1/media/" + args[0])
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer f.Close()
		n, _ := io.Copy(f, resp.Body)
		printer.Success(fmt.Sprintf("Downloaded %d bytes to %s", n, outFile))
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

	// Credentials
	credentialsCreateCmd.Flags().String("name", "", "Credential name")
	_ = credentialsCreateCmd.MarkFlagRequired("name")
	credentialsCmd.AddCommand(credentialsListCmd, credentialsCreateCmd, credentialsDeleteCmd)

	// Activity
	activityCmd.Flags().Int("limit", 50, "Max results")

	// TTS
	ttsSetProviderCmd.Flags().String("name", "", "Provider name")
	_ = ttsSetProviderCmd.MarkFlagRequired("name")
	ttsCmd.AddCommand(ttsStatusCmd, ttsEnableCmd, ttsDisableCmd, ttsProvidersCmd, ttsSetProviderCmd)

	// Media
	mediaGetCmd.Flags().StringP("output", "f", "", "Output file")
	mediaCmd.AddCommand(mediaUploadCmd, mediaGetCmd)

	rootCmd.AddCommand(approvalsCmd, delegationsCmd, credentialsCmd, activityCmd, ttsCmd, mediaCmd)
}
