package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{Use: "devices", Short: "Manage paired devices"}

var devicesListCmd = &cobra.Command{
	Use: "list", Short: "List paired devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("device.pair.list", map[string]any{})
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "USER", "STATUS", "PAIRED_AT", "LAST_SEEN")
		for _, d := range unmarshalList(data) {
			tbl.AddRow(str(d, "id"), str(d, "device_name"), str(d, "user_id"),
				str(d, "status"), str(d, "paired_at"), str(d, "last_seen_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var devicesRevokeCmd = &cobra.Command{
	Use: "revoke <id>", Short: "Revoke a paired device", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Revoke this device?", cfg.Yes) {
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
		_, err = ws.Call("device.pair.revoke", map[string]any{"id": args[0]})
		if err != nil {
			return err
		}
		printer.Success("Device revoked")
		return nil
	},
}

var devicesPairingStatusCmd = &cobra.Command{
	Use: "pairing-status", Short: "Check browser pairing status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("browser.pairing.status", map[string]any{})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	devicesCmd.AddCommand(devicesListCmd, devicesRevokeCmd, devicesPairingStatusCmd)
	rootCmd.AddCommand(devicesCmd)
}
