// Package cmd provides the CLI commands for RTMX.
package cmd

import (
	"fmt"

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

// ExitError is an error that carries an exit code.
// Commands should return this instead of calling os.Exit() directly.
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

// NewExitError creates a new ExitError with the given code.
func NewExitError(code int, message string) *ExitError {
	return &ExitError{Code: code, Message: message}
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "rtmx",
	Short:   "Requirements Traceability Matrix toolkit",
	Version: Version,
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
	// Config loading is handled by individual commands via config.LoadFromDir()
	// The --config flag is reserved for future use
	_ = cfgFile // Suppress unused warning until implemented
}
