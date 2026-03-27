package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var markersShowMissing bool

var markersCmd = &cobra.Command{
	Use:   "markers",
	Short: "Display requirement markers found in test files",
	Long: `Scan test files for requirement markers (rtmx.Req, @pytest.mark.req)
and display which requirements have markers and which don't.

Examples:
  rtmx markers             # List all markers by requirement
  rtmx markers --missing   # Show requirements with no markers`,
	RunE: runMarkers,
}

func init() {
	markersCmd.Flags().BoolVar(&markersShowMissing, "missing", false, "show only requirements with no test markers")
	rootCmd.AddCommand(markersCmd)
}

func runMarkers(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	// Scan for markers in test files
	markers, err := scanTestDirectory(cwd)
	if err != nil {
		cmd.Printf("%s Failed to scan test files: %v\n", output.Color("!", output.Yellow), err)
	}

	// Build a map of reqID -> markers
	markersByReq := make(map[string][]TestRequirement)
	for _, m := range markers {
		markersByReq[m.ReqID] = append(markersByReq[m.ReqID], m)
	}

	if markersShowMissing {
		// Show requirements with no markers
		cmd.Println("Requirements with no test markers:")
		cmd.Println()

		count := 0
		for _, req := range db.All() {
			if _, found := markersByReq[req.ReqID]; !found {
				cmd.Printf("  %s %s - %s\n",
					output.Color("!", output.Yellow),
					req.ReqID,
					truncateStr(req.RequirementText, 60))
				count++
			}
		}

		if count == 0 {
			cmd.Printf("  %s All requirements have test markers\n", output.Color("OK", output.Green))
		} else {
			cmd.Printf("\n%d requirement(s) without markers\n", count)
		}
		return nil
	}

	// Show all markers grouped by requirement
	reqIDs := make([]string, 0, len(markersByReq))
	for id := range markersByReq {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	total := len(db.All())
	covered := 0

	for _, reqID := range reqIDs {
		markers := markersByReq[reqID]
		req := db.Get(reqID)
		status := "unknown"
		if req != nil {
			status = string(req.Status)
			covered++
		}

		cmd.Printf("%s [%s]\n", reqID, status)
		for _, m := range markers {
			cmd.Printf("  %s %s :: %s\n",
				output.Color("*", output.Green),
				m.TestFile, m.TestFunction)
		}
	}

	// Count requirements in DB that have markers
	for _, req := range db.All() {
		if _, found := markersByReq[req.ReqID]; found {
			// already counted above only if in markersByReq AND in db
			_ = req
		}
	}

	cmd.Printf("\n%d markers found across %d requirements (%d/%d in database)\n",
		len(markers), len(markersByReq), covered, total)

	return nil
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
