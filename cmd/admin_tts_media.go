package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// admin_tts_media.go holds TTS and Media subcommands, extracted from admin.go
// to keep that file under 200 LoC.

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
	ttsSetProviderCmd.Flags().String("name", "", "Provider name")
	_ = ttsSetProviderCmd.MarkFlagRequired("name")
	ttsCmd.AddCommand(ttsStatusCmd, ttsEnableCmd, ttsDisableCmd, ttsProvidersCmd, ttsSetProviderCmd)

	mediaGetCmd.Flags().StringP("output", "f", "", "Output file")
	mediaCmd.AddCommand(mediaUploadCmd, mediaGetCmd)
}
