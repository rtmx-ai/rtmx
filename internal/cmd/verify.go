package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	verifyUpdate  bool
	verifyDryRun  bool
	verifyVerbose bool
	verifyCommand string
)

var verifyCmd = &cobra.Command{
	Use:   "verify [test_path]",
	Short: "Verify requirements by running tests",
	Long: `Run tests and update requirement status based on results.

This is closed-loop verification: tests are run, and RTM status
is automatically updated based on pass/fail results.

The command runs "go test -json ./..." by default, but you can
specify a custom test command with --command.

Status update rules:
  - All tests pass → COMPLETE
  - Any test fails → Downgrade COMPLETE to PARTIAL
  - No tests → Keep current status

Examples:
  rtmx verify                    # Run tests, show results
  rtmx verify --update           # Run tests and update RTM
  rtmx verify ./internal/... --update  # Verify specific package
  rtmx verify --dry-run          # Show what would change
  rtmx verify --command "pytest -v"    # Use custom test command`,
	RunE: runVerify,
}

func init() {
	verifyCmd.Flags().BoolVar(&verifyUpdate, "update", false, "update RTM database with results")
	verifyCmd.Flags().BoolVar(&verifyDryRun, "dry-run", false, "show changes without updating")
	verifyCmd.Flags().BoolVarP(&verifyVerbose, "verbose", "v", false, "verbose output")
	verifyCmd.Flags().StringVar(&verifyCommand, "command", "", "custom test command (default: go test -json)")

	rootCmd.AddCommand(verifyCmd)
}

// TestEvent represents a Go test JSON output event
type TestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

// TestResult aggregates results for a single test
type TestResult struct {
	Package string
	Test    string
	Passed  bool
	Failed  bool
	Skipped bool
}

// VerificationResult represents the verification outcome for a requirement
type VerificationResult struct {
	ReqID          string
	TestsTotal     int
	TestsPassed    int
	TestsFailed    int
	TestsSkipped   int
	PreviousStatus database.Status
	NewStatus      database.Status
	Updated        bool
}

func runVerify(cmd *cobra.Command, args []string) error {
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

	// Determine test path
	testPath := "./..."
	if len(args) > 0 {
		testPath = args[0]
	}

	cmd.Println("Running tests and collecting requirement coverage...")
	cmd.Println()

	// Run tests and get results
	testResults, err := runTests(cmd, testPath)
	if err != nil {
		cmd.Printf("%s Failed to run tests: %v\n", output.Color("!", output.Red), err)
		// Continue to show what we can
	}

	// Map tests to requirements
	verifyResults := mapTestsToRequirements(db, testResults)

	// Print results
	printVerifyResults(cmd, verifyResults)

	// Update database if requested
	if verifyUpdate && !verifyDryRun {
		updateCount := 0
		for _, r := range verifyResults {
			if r.Updated {
				req := db.Get(r.ReqID)
				if req != nil {
					req.Status = r.NewStatus
					updateCount++
				}
			}
		}
		if updateCount > 0 {
			if err := db.Save(dbPath); err != nil {
				return fmt.Errorf("failed to save database: %w", err)
			}
			cmd.Printf("\n%s Updated %d requirement(s)\n", output.Color("✓", output.Green), updateCount)
		} else {
			cmd.Println("\nNo status changes needed")
		}
	} else if verifyDryRun {
		cmd.Printf("\n%s\n", output.Color("Dry run - no changes made", output.Yellow))
	}

	// Exit with error if any tests failed
	for _, r := range verifyResults {
		if r.TestsFailed > 0 {
			os.Exit(1)
		}
	}

	return nil
}

func runTests(cmd *cobra.Command, testPath string) (map[string]*TestResult, error) {
	results := make(map[string]*TestResult)

	var testCmd *exec.Cmd
	if verifyCommand != "" {
		// Use custom command
		parts := strings.Fields(verifyCommand)
		if len(parts) == 0 {
			return nil, fmt.Errorf("empty test command")
		}
		testCmd = exec.Command(parts[0], parts[1:]...)
	} else {
		// Default: go test -json
		testCmd = exec.Command("go", "test", "-json", testPath)
	}

	testCmd.Dir, _ = os.Getwd()

	stdout, err := testCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := testCmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start test command: %w", err)
	}

	// Parse JSON output
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Not JSON, might be custom command output
			if verifyVerbose {
				cmd.Println(line)
			}
			continue
		}

		// Only care about test pass/fail/skip events
		if event.Test == "" {
			continue
		}

		key := event.Package + "/" + event.Test

		switch event.Action {
		case "pass":
			results[key] = &TestResult{
				Package: event.Package,
				Test:    event.Test,
				Passed:  true,
			}
			if verifyVerbose {
				cmd.Printf("  %s %s\n", output.Color("✓", output.Green), event.Test)
			}
		case "fail":
			results[key] = &TestResult{
				Package: event.Package,
				Test:    event.Test,
				Failed:  true,
			}
			if verifyVerbose {
				cmd.Printf("  %s %s\n", output.Color("✗", output.Red), event.Test)
			}
		case "skip":
			results[key] = &TestResult{
				Package: event.Package,
				Test:    event.Test,
				Skipped: true,
			}
			if verifyVerbose {
				cmd.Printf("  %s %s (skipped)\n", output.Color("-", output.Yellow), event.Test)
			}
		}
	}

	testCmd.Wait()

	return results, nil
}

func mapTestsToRequirements(db *database.Database, testResults map[string]*TestResult) []VerificationResult {
	var results []VerificationResult

	// Build a map of test function -> results
	testByFunction := make(map[string]*TestResult)
	for _, r := range testResults {
		// Index by just function name for matching
		testByFunction[r.Test] = r
	}

	// For each requirement with a test defined
	for _, req := range db.All() {
		if req.TestFunction == "" {
			continue
		}

		// Try to find matching test result
		testFunc := req.TestFunction
		result := testByFunction[testFunc]

		if result == nil {
			// No matching test found
			continue
		}

		// Determine new status
		newStatus := determineNewStatus(result, req.Status)

		results = append(results, VerificationResult{
			ReqID:          req.ReqID,
			TestsTotal:     1,
			TestsPassed:    boolToInt(result.Passed),
			TestsFailed:    boolToInt(result.Failed),
			TestsSkipped:   boolToInt(result.Skipped),
			PreviousStatus: req.Status,
			NewStatus:      newStatus,
			Updated:        newStatus != req.Status,
		})
	}

	return results
}

func determineNewStatus(result *TestResult, currentStatus database.Status) database.Status {
	if result.Failed {
		// Downgrade COMPLETE to PARTIAL on failure
		if currentStatus == database.StatusComplete {
			return database.StatusPartial
		}
		return currentStatus
	}

	if result.Passed {
		return database.StatusComplete
	}

	// Skipped - keep current status
	return currentStatus
}

func printVerifyResults(cmd *cobra.Command, results []VerificationResult) {
	if len(results) == 0 {
		cmd.Println("No requirements with linked tests found.")
		return
	}

	width := 60
	cmd.Println(output.Header("Verification Results", width))
	cmd.Println()

	var passing, failing, toUpdate int
	for _, r := range results {
		if r.TestsPassed > 0 && r.TestsFailed == 0 {
			passing++
		}
		if r.TestsFailed > 0 {
			failing++
		}
		if r.Updated {
			toUpdate++
		}
	}

	if passing > 0 {
		cmd.Printf("  %s PASSING: %d requirements\n", output.Color("✓", output.Green), passing)
	}
	if failing > 0 {
		cmd.Printf("  %s FAILING: %d requirements\n", output.Color("✗", output.Red), failing)
	}

	if toUpdate > 0 {
		cmd.Println()
		cmd.Println(output.SubHeader("Status Changes", width))
		for _, r := range results {
			if r.Updated {
				statusChange := fmt.Sprintf("%s → %s", r.PreviousStatus, r.NewStatus)
				if r.NewStatus == database.StatusComplete {
					cmd.Printf("  %s %s: %s\n",
						output.Color("↑", output.Green),
						output.Color(r.ReqID, output.Cyan),
						output.Color(statusChange, output.Green))
				} else if r.NewStatus == database.StatusPartial {
					cmd.Printf("  %s %s: %s\n",
						output.Color("↓", output.Yellow),
						output.Color(r.ReqID, output.Cyan),
						output.Color(statusChange, output.Yellow))
				} else {
					cmd.Printf("  %s: %s\n", r.ReqID, statusChange)
				}
			}
		}
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
