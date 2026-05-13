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
	verifyAudit   bool
	verifyCommand string
	verifyResults string
	verifyVersion string
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
  rtmx verify --audit                # Show audit diagnostics for stale refs

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

var (
	verifyForce         bool
	verifyWarnThreshold int
	verifyFailThreshold int
)

func init() {
	verifyCmd.Flags().BoolVar(&verifyUpdate, "update", false, "update RTM database with results")
	verifyCmd.Flags().BoolVar(&verifyDryRun, "dry-run", false, "show changes without updating")
	verifyCmd.Flags().BoolVarP(&verifyVerbose, "verbose", "v", false, "verbose output")
	verifyCmd.Flags().BoolVar(&verifyForce, "force", false, "override fail threshold for this invocation")
	verifyCmd.Flags().StringVar(&verifyCommand, "command", "", "custom test command (default: go test -json)")
	verifyCmd.Flags().StringVar(&verifyResults, "results", "", "RTMX results JSON file (cross-language)")
	verifyCmd.Flags().StringVar(&verifyVersion, "version", "", "verify only requirements targeting this version")
	verifyCmd.Flags().BoolVar(&verifyAudit, "audit", false, "show audit diagnostics for stale or unmatched test references")
	verifyCmd.Flags().IntVar(&verifyWarnThreshold, "warn-threshold", 0, "status change count that triggers a warning (0=use config, default: 5)")
	verifyCmd.Flags().IntVar(&verifyFailThreshold, "fail-threshold", 0, "status change count that blocks updates (0=use config, default: 15)")

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

	// Run audit diagnostics
	if verifyAudit {
		auditResult := runAudit(db, verifyResultsList)
		printAuditResults(cmd, auditResult)
	} else {
		// Even without --audit, warn if there are unmatched references (REQ-VERIFY-006)
		unmatchedCount := countUnmatchedRefs(db, verifyResultsList)
		if unmatchedCount > 0 {
			cmd.Printf("\n%s %d requirement(s) have test references that did not match any test result. Run with --audit for details.\n",
				output.Color("Warning:", output.Yellow), unmatchedCount)
		}
	}

	// Update database if requested
	if verifyUpdate && !verifyDryRun {
		updateCount := 0
		for _, r := range verifyResultsList {
			if r.Updated {
				updateCount++
			}
		}

		// Enforce thresholds: CLI flag > config > default
		// REQ-VERIFY-008
		warnThreshold := cfg.RTMX.Verify.Thresholds.Warn
		failThreshold := cfg.RTMX.Verify.Thresholds.Fail
		if warnThreshold <= 0 {
			warnThreshold = 5
		}
		if failThreshold <= 0 {
			failThreshold = 15
		}
		if verifyWarnThreshold > 0 {
			warnThreshold = verifyWarnThreshold
		}
		if verifyFailThreshold > 0 {
			failThreshold = verifyFailThreshold
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
						// Auto-set dates on status transitions
						if r.NewStatus == database.StatusPartial || r.NewStatus == database.StatusComplete {
							req.SetStartedDate()
						}
						if r.NewStatus == database.StatusComplete {
							req.SetCompletedDate()
							// REQ-PLAN-013: Auto-set assignee from git attribution
							if req.Assignee == "" && req.TestModule != "" {
								if author := GetGitAuthor(req.TestModule); author != "" {
									req.Assignee = author
								}
							}
						}
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
		// Auto-detect project type from build files
		cwd, _ := os.Getwd()
		cmdName, cmdArgs := DetectTestCommand(cwd)
		testCmd = exec.Command(cmdName, cmdArgs...)
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

// matchTestFunction checks if a test result name matches a database test_function
// value, supporting suffix matching at path separator boundaries (:: and .).
// This handles Rust module paths (embedding::tests::test_foo matches tests::test_foo)
// and Python package paths (pkg.tests.test_foo matches tests.test_foo).
// REQ-VERIFY-007
func matchTestFunction(testResultName, dbTestFunction string) bool {
	if testResultName == dbTestFunction {
		return true
	}
	// Suffix match at :: boundary (Rust module paths)
	if strings.HasSuffix(testResultName, "::"+dbTestFunction) {
		return true
	}
	// Suffix match at . boundary (Python package paths)
	if strings.HasSuffix(testResultName, "."+dbTestFunction) {
		return true
	}
	// Suffix match at / boundary (Go package paths)
	if strings.HasSuffix(testResultName, "/"+dbTestFunction) {
		return true
	}
	return false
}

func mapTestsToRequirements(db *database.Database, testResults map[string]*TestResult) []VerificationResult {
	var results []VerificationResult

	// Collect all test result names for suffix matching
	allTestResults := make([]*TestResult, 0, len(testResults))
	// Build a map of test function -> results for exact match (fast path)
	testByFunction := make(map[string]*TestResult)
	for _, r := range testResults {
		testByFunction[r.Test] = r
		allTestResults = append(allTestResults, r)
	}

	// For each requirement with a test defined
	for _, req := range db.All() {
		if req.TestFunction == "" {
			continue
		}

		// Try exact match first (fast path)
		testFunc := req.TestFunction
		result := testByFunction[testFunc]

		// Fall back to suffix matching at path separator boundaries
		if result == nil {
			for _, r := range allTestResults {
				if matchTestFunction(r.Test, testFunc) {
					result = r
					break
				}
			}
		}

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

// DetectTestCommand inspects the working directory for build files and returns
// the appropriate test command and arguments for the detected project type.
func DetectTestCommand(dir string) (string, []string) {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}

	switch {
	case exists("Cargo.toml"):
		return "cargo", []string{"test", "--workspace"}
	case exists("package.json"):
		return "npm", []string{"test"}
	case exists("pyproject.toml"), exists("setup.py"), exists("setup.cfg"):
		return "pytest", []string{"-v"}
	case exists("build.gradle"), exists("build.gradle.kts"):
		return "gradle", []string{"test"}
	case exists("pom.xml"):
		return "mvn", []string{"test"}
	case exists("mix.exs"):
		return "mix", []string{"test"}
	case exists("Gemfile"):
		return "bundle", []string{"exec", "rake", "test"}
	case exists("Package.swift"):
		return "swift", []string{"test"}
	case exists("pubspec.yaml"):
		return "dart", []string{"test"}
	default:
		return "go", []string{"test", "-json", "./..."}
	}
}

// AuditFinding represents a single audit diagnostic.
type AuditFinding struct {
	ReqID   string `json:"req_id"`
	Kind    string `json:"kind"` // "unmatched", "stale_path", "unverified_complete", "empty_ref"
	Field   string `json:"field,omitempty"`
	Value   string `json:"value,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// AuditResult holds all audit diagnostics.
type AuditResult struct {
	Unmatched          []AuditFinding `json:"unmatched_references"`
	StalePaths         []AuditFinding `json:"stale_test_modules"`
	UnverifiedComplete []AuditFinding `json:"unverified_complete"`
	EmptyRefs          []AuditFinding `json:"empty_references"`
}

// runAudit checks for stale or missing test references in the database.
func runAudit(db *database.Database, verifyResults []VerificationResult) AuditResult {
	// Build set of req IDs that had a test match in this run
	matched := make(map[string]bool)
	for _, r := range verifyResults {
		matched[r.ReqID] = true
	}

	var result AuditResult

	for _, req := range db.All() {
		hasModule := req.TestModule != ""
		hasFunction := req.TestFunction != ""

		// Check 1: Empty test references
		if !hasModule && !hasFunction {
			result.EmptyRefs = append(result.EmptyRefs, AuditFinding{
				ReqID:  req.ReqID,
				Kind:   "empty_ref",
				Detail: "no test_module or test_function set",
			})
			continue
		}

		// Check 2: Stale test_module path
		if hasModule {
			if _, err := os.Stat(req.TestModule); err != nil {
				result.StalePaths = append(result.StalePaths, AuditFinding{
					ReqID:  req.ReqID,
					Kind:   "stale_path",
					Field:  "test_module",
					Value:  req.TestModule,
					Detail: "file does not exist",
				})
			}
		}

		// Check 3: Unmatched test_function (has reference but no test result)
		if hasFunction && !matched[req.ReqID] {
			detail := "no matching test result"
			if hasModule {
				if _, err := os.Stat(req.TestModule); err != nil {
					detail = fmt.Sprintf("file missing: %s", req.TestModule)
				} else {
					detail = fmt.Sprintf("file exists (%s) but function not matched", req.TestModule)
				}
			}
			result.Unmatched = append(result.Unmatched, AuditFinding{
				ReqID:  req.ReqID,
				Kind:   "unmatched",
				Field:  "test_function",
				Value:  req.TestFunction,
				Detail: detail,
			})
		}

		// Check 4: COMPLETE but no test match (potential false positive)
		if req.Status == database.StatusComplete && !matched[req.ReqID] {
			result.UnverifiedComplete = append(result.UnverifiedComplete, AuditFinding{
				ReqID:  req.ReqID,
				Kind:   "unverified_complete",
				Field:  "test_function",
				Value:  req.TestFunction,
				Detail: "status is COMPLETE but no test matched in this run",
			})
		}
	}

	return result
}

// countUnmatchedRefs returns the number of requirements with test_function set
// but no matching test result. Used for the default warning (REQ-VERIFY-006).
func countUnmatchedRefs(db *database.Database, verifyResults []VerificationResult) int {
	matched := make(map[string]bool)
	for _, r := range verifyResults {
		matched[r.ReqID] = true
	}
	count := 0
	for _, req := range db.All() {
		if req.TestFunction != "" && !matched[req.ReqID] {
			count++
		}
	}
	return count
}

func printAuditResults(cmd *cobra.Command, audit AuditResult) {
	width := 60
	cmd.Println()
	cmd.Println(output.Header("Audit Diagnostics", width))

	if len(audit.Unmatched) > 0 {
		cmd.Println()
		cmd.Printf("  Unmatched test references (%d):\n", len(audit.Unmatched))
		for _, f := range audit.Unmatched {
			cmd.Printf("    %s  test_function=%s  (%s)\n",
				output.Color(f.ReqID, output.Cyan), f.Value, f.Detail)
		}
	}

	if len(audit.StalePaths) > 0 {
		cmd.Println()
		cmd.Printf("  Stale test_module paths (%d):\n", len(audit.StalePaths))
		for _, f := range audit.StalePaths {
			cmd.Printf("    %s  %s\n",
				output.Color(f.ReqID, output.Cyan), f.Value)
		}
	}

	if len(audit.UnverifiedComplete) > 0 {
		cmd.Println()
		cmd.Printf("  Unverified COMPLETE requirements (%d):\n", len(audit.UnverifiedComplete))
		for _, f := range audit.UnverifiedComplete {
			funcInfo := "(no test_function)"
			if f.Value != "" {
				funcInfo = fmt.Sprintf("test_function=%s", f.Value)
			}
			cmd.Printf("    %s  %s\n",
				output.Color(f.ReqID, output.Cyan), funcInfo)
		}
	}

	if len(audit.EmptyRefs) > 0 {
		cmd.Println()
		cmd.Printf("  Empty test references (%d):\n", len(audit.EmptyRefs))
		for _, f := range audit.EmptyRefs {
			cmd.Printf("    %s  (no test_function set)\n",
				output.Color(f.ReqID, output.Cyan))
		}
	}

	total := len(audit.Unmatched) + len(audit.StalePaths) + len(audit.UnverifiedComplete) + len(audit.EmptyRefs)
	if total == 0 {
		cmd.Println()
		cmd.Printf("  %s No audit findings.\n", output.Color("✓", output.Green))
	} else {
		cmd.Println()
		cmd.Printf("  Summary: %d unmatched, %d stale paths, %d unverified, %d empty\n",
			len(audit.Unmatched), len(audit.StalePaths),
			len(audit.UnverifiedComplete), len(audit.EmptyRefs))
	}
}
