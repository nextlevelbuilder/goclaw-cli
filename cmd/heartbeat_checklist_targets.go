package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var heartbeatChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Manage heartbeat checklist",
}

var heartbeatChecklistGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get heartbeat checklist",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("heartbeat.checklist.get", map[string]any{})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatChecklistSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set heartbeat checklist",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		dataFlag, _ := cmd.Flags().GetString("data")
		content, err := readContent(dataFlag)
		if err != nil {
			return err
		}
		data, err := ws.Call("heartbeat.checklist.set", map[string]any{"data": content})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var heartbeatTargetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "List heartbeat targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("heartbeat.targets", map[string]any{})
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("NAME", "URL", "STATUS", "LAST_CHECK")
		for _, t := range unmarshalList(data) {
			tbl.AddRow(str(t, "name"), str(t, "url"), str(t, "status"), str(t, "last_check"))
		}
		printer.Print(tbl)
		return nil
	},
}

func init() {
	heartbeatChecklistSetCmd.Flags().String("data", "", "Checklist data (or @filepath)")
	_ = heartbeatChecklistSetCmd.MarkFlagRequired("data")

	heartbeatChecklistCmd.AddCommand(heartbeatChecklistGetCmd, heartbeatChecklistSetCmd)
	heartbeatCmd.AddCommand(heartbeatChecklistCmd, heartbeatTargetsCmd)
}
