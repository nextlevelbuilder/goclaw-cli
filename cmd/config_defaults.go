package cmd

import "github.com/spf13/cobra"

var configDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Show resolved server configuration defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("config.defaults", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configDefaultsCmd)
}
