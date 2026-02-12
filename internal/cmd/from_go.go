package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	fromGoUpdate  bool
	fromGoDryRun  bool
	fromGoVerbose bool
)

var fromGoCmd = &cobra.Command{
	Use:   "from-go <results.json>",
	Short: "Import Go test results into RTM database",
	Long: `Import test results from Go tests using the rtmx package.

The rtmx package captures test results with requirement markers and
writes them to a JSON file. This command imports those results into
the RTM database.

Usage in Go tests:
    import "github.com/rtmx-ai/rtmx-go/pkg/rtmx"

    func TestFeature(t *testing.T) {
        rtmx.Req(t, "REQ-FEAT-001", rtmx.Scope("unit"))
        // test
    }

    func TestMain(m *testing.M) {
        code := m.Run()
        rtmx.WriteResultsJSON("rtmx-results.json")
        os.Exit(code)
    }

Examples:
    rtmx from-go rtmx-results.json
    rtmx from-go rtmx-results.json --update
    rtmx from-go rtmx-results.json --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runFromGo,
}

func init() {
	fromGoCmd.Flags().BoolVar(&fromGoUpdate, "update", false, "update RTM database with results")
	fromGoCmd.Flags().BoolVar(&fromGoDryRun, "dry-run", false, "show changes without updating")
	fromGoCmd.Flags().BoolVarP(&fromGoVerbose, "verbose", "v", false, "show all imported results")

	rootCmd.AddCommand(fromGoCmd)
}

// GoTestResult matches the JSON structure from rtmx package.
type GoTestResult struct {
	Marker struct {
		ReqID     string `json:"req_id"`
		Scope     string `json:"scope"`
		Technique string `json:"technique"`
		Env       string `json:"env"`
		TestName  string `json:"test_name"`
		TestFile  string `json:"test_file"`
		Line      int    `json:"line"`
	} `json:"marker"`
	Passed    bool    `json:"passed"`
	Duration  float64 `json:"duration_ms"`
	Error     string  `json:"error"`
	Timestamp string  `json:"timestamp"`
}

func runFromGo(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	resultsPath := args[0]

	// Read results file
	data, err := os.ReadFile(resultsPath)
	if err != nil {
		return fmt.Errorf("failed to read results file: %w", err)
	}

	var results []GoTestResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("failed to parse results file: %w", err)
	}

	cmd.Printf("=== Import Go Test Results ===\n\n")
	cmd.Printf("Results file: %s\n", resultsPath)
	cmd.Printf("Total results: %d\n\n", len(results))

	if len(results) == 0 {
		cmd.Println("No results to import.")
		return nil
	}

	// Load config and database
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
	if err != nil && fromGoUpdate {
		return fmt.Errorf("failed to load database: %w", err)
	}

	// Group results by requirement
	byReq := make(map[string][]GoTestResult)
	for _, r := range results {
		byReq[r.Marker.ReqID] = append(byReq[r.Marker.ReqID], r)
	}

	// Process each requirement
	var updates []struct {
		reqID  string
		status database.Status
		reason string
	}

	for reqID, reqResults := range byReq {
		allPassed := true
		for _, r := range reqResults {
			if !r.Passed {
				allPassed = false
				break
			}
		}

		var newStatus database.Status
		if allPassed {
			newStatus = database.StatusComplete
		} else {
			newStatus = database.StatusMissing
		}

		// Check if requirement exists in database
		if db != nil {
			existing := db.Get(reqID)
			if existing != nil {
				if existing.Status != newStatus {
					updates = append(updates, struct {
						reqID  string
						status database.Status
						reason string
					}{
						reqID:  reqID,
						status: newStatus,
						reason: fmt.Sprintf("%d/%d tests passed", countPassed(reqResults), len(reqResults)),
					})
				}
			} else if fromGoVerbose {
				cmd.Printf("  %s %s: Not in database\n", output.Color("?", output.Yellow), reqID)
			}
		}

		if fromGoVerbose {
			statusIcon := output.Color("✓", output.Green)
			if !allPassed {
				statusIcon = output.Color("✗", output.Red)
			}
			cmd.Printf("  %s %s: %d tests (%d passed)\n",
				statusIcon, reqID, len(reqResults), countPassed(reqResults))
		}
	}

	// Summary
	cmd.Printf("\nSummary:\n")
	cmd.Printf("  Requirements found: %d\n", len(byReq))
	cmd.Printf("  Status changes: %d\n", len(updates))

	if len(updates) > 0 {
		cmd.Printf("\nChanges:\n")
		for _, u := range updates {
			statusStr := u.status.String()
			statusColor := output.StatusColor(statusStr)
			cmd.Printf("  %s → %s (%s)\n", u.reqID, output.Color(statusStr, statusColor), u.reason)
		}
	}

	// Apply updates
	if fromGoUpdate && !fromGoDryRun && len(updates) > 0 && db != nil {
		cmd.Println()
		for _, u := range updates {
			if err := db.Update(u.reqID, map[string]interface{}{"status": u.status.String()}); err != nil {
				cmd.Printf("  %s Failed to update %s: %v\n", output.Color("✗", output.Red), u.reqID, err)
			} else {
				cmd.Printf("  %s Updated %s\n", output.Color("✓", output.Green), u.reqID)
			}
		}

		if err := db.Save(dbPath); err != nil {
			return fmt.Errorf("failed to save database: %w", err)
		}
		cmd.Printf("\n%s Database saved\n", output.Color("✓", output.Green))
	} else if fromGoDryRun {
		cmd.Printf("\n%s\n", output.Color("DRY RUN - no changes made", output.Yellow))
	}

	return nil
}

func countPassed(results []GoTestResult) int {
	count := 0
	for _, r := range results {
		if r.Passed {
			count++
		}
	}
	return count
}
