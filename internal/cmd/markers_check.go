package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/markers"
	"github.com/spf13/cobra"
)

var markersCheckRoot string

var markersCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Fail on dangling or malformed requirement markers",
	Long: `Scan source for requirement markers and exit non-zero when a marker is
dangling (ID not in the RTM) or malformed (grammar mismatch).

This invokes the markers library as a CLI. It is not an IDE plugin or a
forge plugin (REQ-MONO-068c).`,
	RunE: runMarkersCheck,
}

func init() {
	markersCheckCmd.Flags().StringVar(&markersCheckRoot, "root", ".", "directory to scan")
	markersCmd.AddCommand(markersCheckCmd)
}

func runMarkersCheck(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	db, err := database.Load(cfg.DatabasePath(cwd))
	if err != nil {
		return fmt.Errorf("load database: %w", err)
	}
	known := map[string]struct{}{}
	for _, req := range db.All() {
		known[req.ReqID] = struct{}{}
	}
	root := markersCheckRoot
	if !filepath.IsAbs(root) {
		root = filepath.Join(cwd, root)
	}
	files, err := markers.CollectSourceFiles(root)
	if err != nil {
		return err
	}
	issues := markers.CheckSources(files, known)
	for _, iss := range issues {
		cmd.PrintErrln(iss.Error())
	}
	if code := markers.ExitCode(issues); code != 0 {
		return NewExitError(code, fmt.Sprintf("%d marker issue(s)", len(issues)))
	}
	cmd.Println("No dangling or malformed markers.")
	return nil
}
