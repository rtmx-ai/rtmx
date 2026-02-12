package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	fromTestsShowAll     bool
	fromTestsShowMissing bool
	fromTestsUpdate      bool
)

var fromTestsCmd = &cobra.Command{
	Use:   "from-tests [test_path]",
	Short: "Scan test files for requirement markers",
	Long: `Scan test files for @pytest.mark.req() markers and report coverage.

This command parses Python test files to find requirement markers and
shows which requirements have tests linked to them.

Examples:
  rtmx from-tests                 # Scan tests/ directory
  rtmx from-tests tests/unit/     # Scan specific directory
  rtmx from-tests --show-all      # Show all markers found
  rtmx from-tests --update        # Update RTM with test info`,
	RunE: runFromTests,
}

func init() {
	fromTestsCmd.Flags().BoolVar(&fromTestsShowAll, "show-all", false, "show all markers found")
	fromTestsCmd.Flags().BoolVar(&fromTestsShowMissing, "show-missing", false, "show requirements not in database")
	fromTestsCmd.Flags().BoolVar(&fromTestsUpdate, "update", false, "update RTM database with test information")

	rootCmd.AddCommand(fromTestsCmd)
}

// TestRequirement represents a requirement marker found in a test file
type TestRequirement struct {
	ReqID        string
	TestFile     string
	TestFunction string
	LineNumber   int
	Markers      []string
}

func runFromTests(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	// Determine test path
	testPath := "tests"
	if len(args) > 0 {
		testPath = args[0]
	}

	// Check if path exists
	info, err := os.Stat(testPath)
	if err != nil {
		return fmt.Errorf("test path does not exist: %s", testPath)
	}

	cmd.Printf("Scanning %s for requirement markers...\n\n", testPath)

	// Scan for markers
	var markers []TestRequirement
	if info.IsDir() {
		markers, err = scanTestDirectory(testPath)
	} else {
		markers, err = extractMarkersFromFile(testPath)
	}
	if err != nil {
		return fmt.Errorf("failed to scan tests: %w", err)
	}

	if len(markers) == 0 {
		cmd.Printf("%s No requirement markers found.\n", output.Color("!", output.Yellow))
		return nil
	}

	// Group by requirement
	byReq := make(map[string][]TestRequirement)
	for _, m := range markers {
		byReq[m.ReqID] = append(byReq[m.ReqID], m)
	}

	cmd.Printf("Found %d test(s) linked to %d requirement(s)\n\n", len(markers), len(byReq))

	// Load RTM database
	cwd, _ := os.Getwd()
	cfg, err := config.LoadFromDir(cwd)
	var db *database.Database
	var dbReqs map[string]bool
	dbPath := ""

	if err == nil {
		dbPath = cfg.DatabasePath(cwd)
		db, err = database.Load(dbPath)
		if err == nil {
			dbReqs = make(map[string]bool)
			for _, req := range db.All() {
				dbReqs[req.ReqID] = true
			}
			cmd.Printf("RTM database: %s (%d requirements)\n", dbPath, len(dbReqs))
		}
	}

	if db == nil {
		cmd.Printf("%s No RTM database found\n", output.Color("!", output.Yellow))
	}
	cmd.Println()

	// Show markers
	if fromTestsShowAll {
		cmd.Println(output.Color("All Requirements with Tests:", output.Bold))
		cmd.Println(strings.Repeat("-", 60))

		reqIDs := make([]string, 0, len(byReq))
		for id := range byReq {
			reqIDs = append(reqIDs, id)
		}
		sort.Strings(reqIDs)

		for _, reqID := range reqIDs {
			tests := byReq[reqID]
			inDB := dbReqs[reqID]
			icon := "✓"
			color := output.Green
			if !inDB {
				icon = "✗"
				color = output.Yellow
			}

			cmd.Printf("%s %s (%d test(s))\n",
				output.Color(icon, color),
				output.Color(reqID, output.Bold),
				len(tests))

			for _, t := range tests {
				markerStr := ""
				if len(t.Markers) > 0 {
					markerStr = fmt.Sprintf(" [%s]", strings.Join(t.Markers, ", "))
				}
				cmd.Printf("    %s::%s%s\n", t.TestFile, t.TestFunction, markerStr)
			}
		}
		cmd.Println()
	}

	// Show requirements not in database
	if fromTestsShowMissing || !fromTestsShowAll {
		var notInDB []string
		for reqID := range byReq {
			if !dbReqs[reqID] {
				notInDB = append(notInDB, reqID)
			}
		}
		sort.Strings(notInDB)

		if len(notInDB) > 0 {
			cmd.Printf("%s\n", output.Color("Requirements in tests but not in RTM database:", output.Yellow))
			for _, reqID := range notInDB {
				tests := byReq[reqID]
				cmd.Printf("  %s (%d test(s))\n", output.Color(reqID, output.Bold), len(tests))
			}
			cmd.Println()
		}
	}

	// Show requirements in database without tests
	if db != nil && (fromTestsShowMissing || !fromTestsShowAll) {
		var noTests []string
		for reqID := range dbReqs {
			if _, hasTest := byReq[reqID]; !hasTest {
				noTests = append(noTests, reqID)
			}
		}
		sort.Strings(noTests)

		if len(noTests) > 0 {
			cmd.Printf("%s\n", output.Color("Requirements in RTM database without tests:", output.Yellow))
			for _, reqID := range noTests {
				cmd.Printf("  %s\n", output.Color(reqID, output.Dim))
			}
			cmd.Println()
		}
	}

	// Summary
	cmd.Println(output.Color("Summary:", output.Bold))
	tested := 0
	for reqID := range byReq {
		if dbReqs[reqID] {
			tested++
		}
	}
	dbTotal := "?"
	if db != nil {
		dbTotal = fmt.Sprintf("%d", len(dbReqs))
	}
	cmd.Printf("  Requirements with tests: %d/%s\n", tested, dbTotal)
	cmd.Printf("  Tests linked to requirements: %d\n", len(markers))

	// Update database if requested
	if fromTestsUpdate && db != nil {
		updated := 0
		for reqID, tests := range byReq {
			if db.Exists(reqID) && len(tests) > 0 {
				req := db.Get(reqID)
				relPath := tests[0].TestFile
				if rel, err := filepath.Rel(cwd, tests[0].TestFile); err == nil {
					relPath = rel
				}
				req.TestModule = relPath
				req.TestFunction = tests[0].TestFunction
				updated++
			}
		}

		if updated > 0 {
			if err := db.Save(dbPath); err != nil {
				return fmt.Errorf("failed to save database: %w", err)
			}
			cmd.Printf("\n%s Updated %d requirement(s) in RTM database\n",
				output.Color("✓", output.Green), updated)
		}
	}

	return nil
}

// scanTestDirectory scans a directory for Python test files
func scanTestDirectory(dir string) ([]TestRequirement, error) {
	var results []TestRequirement

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Python files
		if info.IsDir() {
			return nil
		}

		// Match test_*.py pattern
		if strings.HasPrefix(filepath.Base(path), "test_") && strings.HasSuffix(path, ".py") {
			markers, err := extractMarkersFromFile(path)
			if err != nil {
				// Skip files that can't be parsed
				return nil
			}
			results = append(results, markers...)
		}

		return nil
	})

	return results, err
}

// extractMarkersFromFile extracts requirement markers from a Python test file
func extractMarkersFromFile(filePath string) ([]TestRequirement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []TestRequirement

	// Regex patterns for pytest markers
	reqMarkerPattern := regexp.MustCompile(`@pytest\.mark\.req\(['"](REQ-[A-Z0-9-]+)['"]\)`)
	funcPattern := regexp.MustCompile(`^(?:async\s+)?def\s+(test_\w+)\s*\(`)
	classPattern := regexp.MustCompile(`^class\s+(Test\w+)\s*[:(]`)
	otherMarkerPattern := regexp.MustCompile(`@pytest\.mark\.(scope_\w+|technique_\w+|env_\w+)`)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var pendingReqIDs []string
	var pendingMarkers []string
	var currentClass string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for class definition
		if match := classPattern.FindStringSubmatch(trimmed); match != nil {
			currentClass = match[1]
			continue
		}

		// Check for requirement marker
		if matches := reqMarkerPattern.FindAllStringSubmatch(trimmed, -1); matches != nil {
			for _, m := range matches {
				pendingReqIDs = append(pendingReqIDs, m[1])
			}
			continue
		}

		// Check for other RTM markers
		if matches := otherMarkerPattern.FindAllStringSubmatch(trimmed, -1); matches != nil {
			for _, m := range matches {
				pendingMarkers = append(pendingMarkers, m[1])
			}
			continue
		}

		// Check for function definition
		if match := funcPattern.FindStringSubmatch(trimmed); match != nil {
			funcName := match[1]
			if currentClass != "" {
				funcName = currentClass + "::" + funcName
			}

			// Create TestRequirement for each pending req ID
			for _, reqID := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        reqID,
					TestFile:     filePath,
					TestFunction: funcName,
					LineNumber:   lineNum,
					Markers:      append([]string{}, pendingMarkers...),
				})
			}

			// Reset pending markers
			pendingReqIDs = nil
			pendingMarkers = nil
		}
	}

	return results, scanner.Err()
}
