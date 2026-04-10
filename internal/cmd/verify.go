package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/internal/results"
	"github.com/spf13/cobra"
)

var (
	verifyUpdate  bool
	verifyDryRun  bool
	verifyVerbose bool
	verifyCommand string
	verifyResults string
)

var verifyCmd = &cobra.Command{
	Use:   "verify [test_path]",
	Short: "Verify requirements by running tests",
	Long: `Run tests and update requirement status based on results.

This is closed-loop verification: tests are run, and RTM status
is automatically updated based on pass/fail results.

The command runs "go test -json ./..." by default, but you can
specify a custom test command with --command.

For cross-language verification, use --results to provide a
language-agnostic RTMX results JSON file produced by any
test framework integration (Go, Python, Rust, etc.).

Status update rules:
  - All tests pass → COMPLETE
  - Any test fails → Downgrade COMPLETE to PARTIAL
  - No tests → Keep current status

Examples:
  rtmx verify                    # Run tests, show results
  rtmx verify --update           # Run tests and update RTM
  rtmx verify ./internal/... --update  # Verify specific package
  rtmx verify --dry-run          # Show what would change
  rtmx verify --command "pytest -v"    # Use custom test command
  rtmx verify --results results.json --update  # Cross-language results

Results file format (--results):
  A JSON array of result objects. Marker fields may be supplied
  nested under "marker" (canonical) or flat at the top level. Either
  a boolean "passed" or a string "status" ("pass"/"fail") is accepted.
  Unknown fields are rejected.

  Canonical:
    [{"marker":{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go"},"passed":true}]

  Flat (also accepted):
    [{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go","status":"pass"}]`,
	RunE: runVerify,
}

var verifyForce bool

func init() {
	verifyCmd.Flags().BoolVar(&verifyUpdate, "update", false, "update RTM database with results")
	verifyCmd.Flags().BoolVar(&verifyDryRun, "dry-run", false, "show changes without updating")
	verifyCmd.Flags().BoolVarP(&verifyVerbose, "verbose", "v", false, "verbose output")
	verifyCmd.Flags().BoolVar(&verifyForce, "force", false, "override fail threshold for this invocation")
	verifyCmd.Flags().StringVar(&verifyCommand, "command", "", "custom test command (default: go test -json)")
	verifyCmd.Flags().StringVar(&verifyResults, "results", "", "RTMX results JSON file (cross-language)")

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

	// Check for mutually exclusive flags
	if verifyResults != "" && len(args) > 0 {
		return fmt.Errorf("--results and test_path are mutually exclusive")
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

	var verifyResultsList []VerificationResult

	if verifyResults != "" {
		// Cross-language mode: read results file
		verifyResultsList, err = runVerifyFromResults(cmd, db)
		if err != nil {
			return err
		}
	} else {
		// Default mode: run go test
		verifyResultsList, err = runVerifyFromTests(cmd, db, args)
		if err != nil {
			return err
		}
	}

	// Print results
	printVerifyResults(cmd, verifyResultsList)

	// Update database if requested
	if verifyUpdate && !verifyDryRun {
		updateCount := 0
		for _, r := range verifyResultsList {
			if r.Updated {
				updateCount++
			}
		}

		// Enforce thresholds
		warnThreshold := cfg.RTMX.Verify.Thresholds.Warn
		failThreshold := cfg.RTMX.Verify.Thresholds.Fail
		if warnThreshold <= 0 {
			warnThreshold = 5
		}
		if failThreshold <= 0 {
			failThreshold = 15
		}

		if updateCount > failThreshold && !verifyForce {
			cmd.Printf("\n%s %d status changes exceed fail threshold (%d). Changes NOT written.\n",
				output.Color("ERROR:", output.Red), updateCount, failThreshold)
			cmd.Printf("  Run with --force to override, or adjust verify.thresholds.fail in config.\n")
			return fmt.Errorf("threshold exceeded: %d changes > fail threshold %d", updateCount, failThreshold)
		}

		if updateCount > warnThreshold {
			cmd.Printf("\n%s %d status changes exceed warn threshold (%d). Review changes carefully.\n",
				output.Color("WARNING:", output.Yellow), updateCount, warnThreshold)
		}

		// Apply the updates
		if updateCount > 0 {
			for _, r := range verifyResultsList {
				if r.Updated {
					req := db.Get(r.ReqID)
					if req != nil {
						req.Status = r.NewStatus
					}
				}
			}
			if err := db.Save(dbPath); err != nil {
				return fmt.Errorf("failed to save database: %w", err)
			}
			cmd.Printf("\n%s Updated %d requirement(s)\n", output.Color("✓", output.Green), updateCount)
		} else {
			cmd.Println("\nNo status changes needed")
		}
		// Write verify metadata
		rtmxDir := filepath.Dir(dbPath)
		if err := WriteVerifyMeta(rtmxDir); err != nil {
			cmd.Printf("%s Failed to write verify metadata: %v\n", output.Color("!", output.Yellow), err)
		}
	} else if verifyDryRun {
		cmd.Printf("\n%s\n", output.Color("Dry run - no changes made", output.Yellow))
	}

	// Return error if any tests failed
	for _, r := range verifyResultsList {
		if r.TestsFailed > 0 {
			return fmt.Errorf("verification failed: %d requirement(s) have failing tests", countFailingReqs(verifyResultsList))
		}
	}

	return nil
}

func countFailingReqs(results []VerificationResult) int {
	count := 0
	for _, r := range results {
		if r.TestsFailed > 0 {
			count++
		}
	}
	return count
}

// runVerifyFromResults processes an RTMX results JSON file (cross-language).
func runVerifyFromResults(cmd *cobra.Command, db *database.Database) ([]VerificationResult, error) {
	var r *os.File
	var err error

	if verifyResults == "-" {
		r = os.Stdin
	} else {
		r, err = os.Open(verifyResults)
		if err != nil {
			return nil, fmt.Errorf("failed to open results file: %w", err)
		}
		defer func() { _ = r.Close() }()
	}

	parsed, err := results.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results file: %w", err)
	}

	// Validate results. Any failure is fatal so structurally bad
	// payloads do not silently produce zero requirement matches
	// (REQ-VERIFY-004).
	if errs := results.Validate(parsed); len(errs) > 0 {
		for _, e := range errs {
			cmd.Printf("%s %v\n", output.Color("!", output.Red), e)
		}
		return nil, fmt.Errorf("results file failed validation: %d error(s)", len(errs))
	}

	cmd.Println("Processing RTMX results file...")
	cmd.Println()

	if len(parsed) == 0 {
		return nil, nil
	}

	// Group results by requirement
	grouped := results.GroupByRequirement(parsed)

	// Map to verification results
	var vResults []VerificationResult
	for reqID, reqResults := range grouped {
		req := db.Get(reqID)
		if req == nil {
			if verifyVerbose {
				cmd.Printf("  %s %s: not in database\n", output.Color("?", output.Yellow), reqID)
			}
			continue
		}

		passed := 0
		failed := 0
		for _, rr := range reqResults {
			if rr.Passed {
				passed++
			} else {
				failed++
			}
			if verifyVerbose {
				icon := output.Color("✓", output.Green)
				if !rr.Passed {
					icon = output.Color("✗", output.Red)
				}
				cmd.Printf("  %s %s\n", icon, rr.Marker.TestName)
			}
		}

		// Build a synthetic TestResult for status determination
		testResult := &TestResult{
			Test:   reqID,
			Passed: failed == 0 && passed > 0,
			Failed: failed > 0,
		}

		newStatus := determineNewStatus(testResult, req.Status)
		vResults = append(vResults, VerificationResult{
			ReqID:          reqID,
			TestsTotal:     len(reqResults),
			TestsPassed:    passed,
			TestsFailed:    failed,
			PreviousStatus: req.Status,
			NewStatus:      newStatus,
			Updated:        newStatus != req.Status,
		})
	}

	return vResults, nil
}

// runVerifyFromTests runs go test and processes results (original mode).
func runVerifyFromTests(cmd *cobra.Command, db *database.Database, args []string) ([]VerificationResult, error) {
	testPath := "./..."
	if len(args) > 0 {
		testPath = args[0]
	}

	cmd.Println("Running tests and collecting requirement coverage...")
	cmd.Println()

	testResults, err := runTests(cmd, testPath)
	if err != nil {
		cmd.Printf("%s Failed to run tests: %v\n", output.Color("!", output.Red), err)
	}

	return mapTestsToRequirements(db, testResults), nil
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

	// Parse test output (auto-detect format)
	scanner := bufio.NewScanner(stdout)
	// Patterns for cargo test / pytest / generic test runners
	cargoTestPattern := regexp.MustCompile(`^test\s+(\S+)\s+\.\.\.\s+(ok|FAILED|ignored)`)
	pytestPattern := regexp.MustCompile(`^(PASSED|FAILED|ERROR)\s+(\S+)`)
	pytestCollectPattern := regexp.MustCompile(`^(\S+\.py)::(\S+)\s+(PASSED|FAILED|SKIPPED)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Try Go test JSON format first
		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err == nil && event.Test != "" {
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
			continue
		}

		// Try cargo test format: "test tests::test_name ... ok/FAILED/ignored"
		if m := cargoTestPattern.FindStringSubmatch(line); len(m) > 2 {
			testName := m[1]
			status := m[2]
			key := "cargo/" + testName
			switch status {
			case "ok":
				results[key] = &TestResult{Package: "cargo", Test: testName, Passed: true}
				if verifyVerbose {
					cmd.Printf("  %s %s\n", output.Color("✓", output.Green), testName)
				}
			case "FAILED":
				results[key] = &TestResult{Package: "cargo", Test: testName, Failed: true}
				if verifyVerbose {
					cmd.Printf("  %s %s\n", output.Color("✗", output.Red), testName)
				}
			case "ignored":
				results[key] = &TestResult{Package: "cargo", Test: testName, Skipped: true}
				if verifyVerbose {
					cmd.Printf("  %s %s (ignored)\n", output.Color("-", output.Yellow), testName)
				}
			}
			continue
		}

		// Try pytest format: "path/test.py::test_name PASSED/FAILED/SKIPPED"
		if m := pytestCollectPattern.FindStringSubmatch(line); len(m) > 3 {
			testFile := m[1]
			testFunc := m[2]
			status := m[3]
			key := testFile + "/" + testFunc
			switch status {
			case "PASSED":
				results[key] = &TestResult{Package: testFile, Test: testFunc, Passed: true}
			case "FAILED":
				results[key] = &TestResult{Package: testFile, Test: testFunc, Failed: true}
			case "SKIPPED":
				results[key] = &TestResult{Package: testFile, Test: testFunc, Skipped: true}
			}
			continue
		}

		// Try pytest short format: "PASSED/FAILED path::func"
		if m := pytestPattern.FindStringSubmatch(line); len(m) > 2 {
			status := m[1]
			testPath := m[2]
			parts := strings.SplitN(testPath, "::", 2)
			testFunc := testPath
			testPkg := ""
			if len(parts) == 2 {
				testPkg = parts[0]
				testFunc = parts[1]
			}
			key := testPkg + "/" + testFunc
			switch status {
			case "PASSED":
				results[key] = &TestResult{Package: testPkg, Test: testFunc, Passed: true}
			case "FAILED", "ERROR":
				results[key] = &TestResult{Package: testPkg, Test: testFunc, Failed: true}
			}
			continue
		}

		// Unrecognized line
		if verifyVerbose {
			cmd.Println(line)
		}
	}

	_ = testCmd.Wait() // Ignore error - we already have results

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
				switch r.NewStatus {
				case database.StatusComplete:
					cmd.Printf("  %s %s: %s\n",
						output.Color("↑", output.Green),
						output.Color(r.ReqID, output.Cyan),
						output.Color(statusChange, output.Green))
				case database.StatusPartial:
					cmd.Printf("  %s %s: %s\n",
						output.Color("↓", output.Yellow),
						output.Color(r.ReqID, output.Cyan),
						output.Color(statusChange, output.Yellow))
				default:
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
