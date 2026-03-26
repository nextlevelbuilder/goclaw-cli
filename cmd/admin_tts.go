package cmd

import (
	"github.com/spf13/cobra"
)

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

var ttsConvertCmd = &cobra.Command{
	Use: "convert", Short: "Convert text to speech",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		text, _ := cmd.Flags().GetString("text")
		provider, _ := cmd.Flags().GetString("provider")
		params := buildBody("text", text, "provider", provider)
		data, err := ws.Call("tts.convert", params)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	ttsSetProviderCmd.Flags().String("name", "", "Provider name")
	_ = ttsSetProviderCmd.MarkFlagRequired("name")

	ttsConvertCmd.Flags().String("text", "", "Text to convert")
	ttsConvertCmd.Flags().String("provider", "", "TTS provider name")
	_ = ttsConvertCmd.MarkFlagRequired("text")

	ttsCmd.AddCommand(ttsStatusCmd, ttsEnableCmd, ttsDisableCmd, ttsProvidersCmd,
		ttsSetProviderCmd, ttsConvertCmd)
	rootCmd.AddCommand(ttsCmd)
}
