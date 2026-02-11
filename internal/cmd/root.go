// Package cmd provides the CLI commands for RTMX.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "none"
	// Date is set at build time via ldflags.
	Date = "unknown"
)

var (
	cfgFile string
	noColor bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "rtmx",
	Short: "Requirements Traceability Matrix toolkit",
	Long: `RTMX is a CLI tool for managing requirements traceability in GenAI-driven development.

It provides commands to track requirements, run verification tests, manage dependencies,
and synchronize with external tools like GitHub and Jira.

Documentation: https://rtmx.ai/docs
Source: https://github.com/rtmx-ai/rtmx-go`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .rtmx/config.yaml or rtmx.yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(backlogCmd)
	rootCmd.AddCommand(healthCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		// TODO: Load config from specified file
	} else {
		// Search for config in standard locations
		// TODO: Implement config discovery
	}
}

// exitWithError prints an error message and exits with the given code.
func exitWithError(msg string, code int) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}
