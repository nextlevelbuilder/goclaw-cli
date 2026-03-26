package cmd

import (
	"github.com/spf13/cobra"
)

var configPermissionsCmd = &cobra.Command{Use: "permissions", Short: "Manage config permissions"}

var configPermissionsListCmd = &cobra.Command{
	Use: "list", Short: "List config permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("config.permissions.list", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var configPermissionsGrantCmd = &cobra.Command{
	Use: "grant", Short: "Grant a config permission",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		userID, _ := cmd.Flags().GetString("user-id")
		key, _ := cmd.Flags().GetString("key")
		_, err = ws.Call("config.permissions.grant", map[string]any{"user_id": userID, "key": key})
		if err != nil {
			return err
		}
		printer.Success("Permission granted")
		return nil
	},
}

var configPermissionsRevokeCmd = &cobra.Command{
	Use: "revoke", Short: "Revoke a config permission",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		userID, _ := cmd.Flags().GetString("user-id")
		key, _ := cmd.Flags().GetString("key")
		_, err = ws.Call("config.permissions.revoke", map[string]any{"user_id": userID, "key": key})
		if err != nil {
			return err
		}
		printer.Success("Permission revoked")
		return nil
	},
}

func init() {
	configPermissionsGrantCmd.Flags().String("user-id", "", "User ID")
	configPermissionsGrantCmd.Flags().String("key", "", "Config key")
	_ = configPermissionsGrantCmd.MarkFlagRequired("user-id")
	_ = configPermissionsGrantCmd.MarkFlagRequired("key")
	configPermissionsRevokeCmd.Flags().String("user-id", "", "User ID")
	configPermissionsRevokeCmd.Flags().String("key", "", "Config key")
	_ = configPermissionsRevokeCmd.MarkFlagRequired("user-id")
	_ = configPermissionsRevokeCmd.MarkFlagRequired("key")

	configPermissionsCmd.AddCommand(configPermissionsListCmd, configPermissionsGrantCmd, configPermissionsRevokeCmd)
	configCmd.AddCommand(configPermissionsCmd)
}
