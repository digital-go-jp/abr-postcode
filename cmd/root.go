package cmd

import (
	"errors"
	"log/slog"
	"os"

	"abr-postcode/internal/config"
	"abr-postcode/internal/logging"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "abrp",
	Short:        "ABR Postcode converter",
	Long:         "A service to convert between postal codes and town/village IDs using ABR data",
	SilenceUsage: true,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		config.LoadDotEnv()
		cfg := config.Load()
		slog.SetDefault(logging.New(cfg.LogLevel, cfg.LogFormat))
	},
}

// exitCodeFor maps a command result to a process exit code. An error carrying
// an ExitCode reports that value. Anything else is 2 rather than 1, so that an
// unhandled failure - including a panic, which the runtime also exits 2 on -
// can never be read as a dry-run result.
func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	return 2
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(exitCodeFor(err))
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(versionCmd)
}
