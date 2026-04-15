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

		// Resolve output format: flag > GOCLAW_OUTPUT env > TTY detect
		// config.Load already applied env + flag precedence into cfg.OutputFormat,
		// but only when the flag was explicitly Changed. Re-resolve here so that
		// the TTY fallback kicks in when neither flag nor env is set.
		flagVal := ""
		if cmd.Flags().Changed("output") {
			flagVal, _ = cmd.Flags().GetString("output")
		} else if v := os.Getenv("GOCLAW_OUTPUT"); v != "" {
			flagVal = v
		}
		cfg.OutputFormat = output.ResolveFormat(flagVal)

		printer = output.NewPrinter(cfg.OutputFormat)
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and handles errors centrally.
// Errors are printed via output.PrintError and the process exits with a mapped code.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Determine output format for error printing.
		// At this point cfg may be nil if PersistentPreRunE never ran.
		format := "table"
		if cfg != nil {
			format = cfg.OutputFormat
		} else {
			// Best-effort: check flag and env without full config load
			if v := os.Getenv("GOCLAW_OUTPUT"); v != "" {
				format = v
			} else if output.IsTTY(int(os.Stdout.Fd())) {
				format = "table"
			} else {
				format = "json"
			}
		}

		output.PrintError(err, format)
		output.Exit(output.FromError(err))
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.String("server", "", "GoClaw server URL (env: GOCLAW_SERVER)")
	pf.String("token", "", "Auth token or API key (env: GOCLAW_TOKEN)")
	pf.StringP("output", "o", "", "Output format: table, json, yaml (default: auto-detect TTY)")
	pf.BoolP("yes", "y", false, "Skip confirmation prompts (automation mode)")
	pf.Bool("insecure", false, "Skip TLS certificate verification")
	pf.BoolP("verbose", "v", false, "Enable verbose/debug output")
	pf.String("profile", "", "Config profile to use (default: active profile)")
	pf.Bool("quiet", false, "Suppress banners and informational messages")
}
