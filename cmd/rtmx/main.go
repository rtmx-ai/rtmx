// Package main provides the entry point for the RTMX CLI.
package main

import (
	"errors"
	"os"

	"github.com/rtmx-ai/rtmx-go/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Print error (SilenceErrors suppresses Cobra output)
		os.Stderr.WriteString("Error: " + err.Error() + "\n")
		// Check for ExitError to get specific exit code
		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
