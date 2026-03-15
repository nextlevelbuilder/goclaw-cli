package cmd

import (
	"fmt"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	printer *output.Printer
)

var rootCmd = &cobra.Command{
	Use:   "goclaw",
	Short: "CLI for managing GoClaw servers",
	Long:  "A production-ready CLI for managing GoClaw AI agent gateway servers.\nSupports interactive (human) and automation (AI agent / CI) modes.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config from file, env, and flags
		var err error
		cfg, err = config.Load(cmd)
		if err != nil {
			return fmt.Errorf("config load: %w", err)
		}
		printer = output.NewPrinter(cfg.OutputFormat)
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.String("server", "", "GoClaw server URL (env: GOCLAW_SERVER)")
	pf.String("token", "", "Auth token or API key (env: GOCLAW_TOKEN)")
	pf.StringP("output", "o", "table", "Output format: table, json, yaml")
	pf.BoolP("yes", "y", false, "Skip confirmation prompts (automation mode)")
	pf.Bool("insecure", false, "Skip TLS certificate verification")
	pf.BoolP("verbose", "v", false, "Enable verbose/debug output")
	pf.String("profile", "", "Config profile to use (default: active profile)")
}
