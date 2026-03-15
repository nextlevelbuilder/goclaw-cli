package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var apiDocsCmd = &cobra.Command{
	Use:   "api-docs",
	Short: "View API documentation",
	Long:  "Open the Swagger UI in a browser or fetch the raw OpenAPI spec.",
}

var apiDocsOpenCmd = &cobra.Command{
	Use: "open", Short: "Open Swagger UI in default browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Server == "" {
			return client.ErrServerRequired
		}
		url := cfg.Server + "/docs"
		if err := openBrowser(url); err != nil {
			// Fallback: print URL for manual access
			fmt.Printf("Open in browser: %s\n", url)
			return nil
		}
		printer.Success(fmt.Sprintf("Opened %s", url))
		return nil
	},
}

var apiDocsSpecCmd = &cobra.Command{
	Use: "spec", Short: "Fetch the OpenAPI 3.0 spec (JSON)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/openapi.json")
		if err != nil {
			return err
		}
		// OpenAPI spec may be a complex object; unmarshal generically
		var spec any
		if err := json.Unmarshal(data, &spec); err != nil {
			return fmt.Errorf("parse openapi spec: %w", err)
		}
		printer.Print(spec)
		return nil
	},
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	// Reap process in background to avoid zombies
	go cmd.Wait() //nolint:errcheck
	return nil
}

func init() {
	apiDocsCmd.AddCommand(apiDocsOpenCmd, apiDocsSpecCmd)
	rootCmd.AddCommand(apiDocsCmd)
}
