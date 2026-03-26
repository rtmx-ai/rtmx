package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print the version, commit hash, build date, and Go version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Printf("rtmx version %s\n", Version)
		cmd.Printf("  commit:  %s\n", Commit)
		cmd.Printf("  built:   %s\n", Date)
		cmd.Printf("  go:      %s\n", runtime.Version())
		cmd.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return nil
	},
}
