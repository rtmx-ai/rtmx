package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
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
	Long: `Scan test files for requirement markers across all supported languages.

Supported languages and marker styles:
  Go:         rtmx.Req(t, "REQ-ID")
  Python:     @pytest.mark.req("REQ-ID")
  Rust:       #[req("REQ-ID")], // rtmx:req REQ-ID, rtmx::req("REQ-ID")
  JavaScript: req("REQ-ID"), // rtmx:req REQ-ID
  C#:         [Req("REQ-ID")], // rtmx:req REQ-ID
  C/C++:      RTMX_REQ("REQ-ID"), // rtmx:req REQ-ID
  Ruby:       it "...", req: "REQ-ID", # rtmx:req REQ-ID
  All others: // rtmx:req REQ-ID (comment marker, universal)

Examples:
  rtmx from-tests                 # Scan current directory
  rtmx from-tests src/            # Scan specific directory
  rtmx from-tests crates/         # Scan Rust workspace
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

// ConftestMarkerRegistration represents a marker registration found in conftest.py
type ConftestMarkerRegistration struct {
	FilePath   string
	MarkerName string
	MarkerArgs string
	MarkerHelp string
	LineNumber int
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

	// Scan for conftest.py marker registrations
	var conftestRegs []ConftestMarkerRegistration
	if info.IsDir() {
		conftestRegs, err = scanConftestFiles(testPath)
		if err != nil {
			return fmt.Errorf("failed to scan conftest files: %w", err)
		}
	} else if filepath.Base(testPath) == "conftest.py" {
		conftestRegs, err = extractConftestRegistrations(testPath)
		if err != nil {
			return fmt.Errorf("failed to parse conftest: %w", err)
		}
	}

	if len(conftestRegs) > 0 {
		cmd.Printf("conftest.py markers detected: %d registration(s)\n", len(conftestRegs))
		for _, reg := range conftestRegs {
			desc := reg.MarkerName
			if reg.MarkerArgs != "" {
				desc += "(" + reg.MarkerArgs + ")"
			}
			if reg.MarkerHelp != "" {
				desc += ": " + reg.MarkerHelp
			}
			cmd.Printf("  %s [%s:%d]\n", desc, reg.FilePath, reg.LineNumber)
		}
		cmd.Println()
	}

	// Scan for markers
	var markers []TestRequirement
	if info.IsDir() {
		markers, err = scanTestDirectory(testPath)
	} else {
		markers, err = extractMarkersFromSingleFile(testPath)
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

// extractMarkersFromSingleFile dispatches to the appropriate language-specific extractor.
func extractMarkersFromSingleFile(path string) ([]TestRequirement, error) {
	switch {
	case strings.HasSuffix(path, ".rs"):
		return extractRustMarkersFromFile(path)
	case strings.HasSuffix(path, "_test.go"):
		return extractGoMarkersFromFile(path)
	case strings.HasSuffix(path, ".py"):
		return extractMarkersFromFile(path)
	default:
		// Try Python parser as fallback (handles generic comment markers)
		return extractMarkersFromFile(path)
	}
}

// scanTestDirectory scans a directory for test files across all supported languages
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

		base := filepath.Base(path)

		// Match test_*.py pattern
		if strings.HasPrefix(base, "test_") && strings.HasSuffix(path, ".py") {
			markers, err := extractMarkersFromFile(path)
			if err != nil {
				// Skip files that can't be parsed
				return nil
			}
			results = append(results, markers...)
		}

		// Also scan conftest.py for requirement markers on fixtures
		if base == "conftest.py" {
			markers, err := extractMarkersFromFile(path)
			if err != nil {
				return nil
			}
			results = append(results, markers...)
		}

		// Scan Go test files for rtmx.Req() markers
		if strings.HasSuffix(base, "_test.go") {
			markers, err := extractGoMarkersFromFile(path)
			if err != nil {
				return nil
			}
			results = append(results, markers...)
		}

		// Scan Rust files for requirement markers
		if isRustTestFile(path) {
			markers, err := extractRustMarkersFromFile(path)
			if err != nil {
				return nil
			}
			results = append(results, markers...)
		}

		// Scan language-specific test files
		langScanners := []struct {
			check   func(string) bool
			extract func(string) ([]TestRequirement, error)
		}{
			{isJSTestFile, extractJSMarkersFromFile},
			{isCSharpTestFile, extractCSharpMarkersFromFile},
			{isCppTestFile, extractCppMarkersFromFile},
			{isTerraformTestFile, extractTerraformMarkersFromFile},
			{isRubyTestFile, extractRubyMarkersFromFile},
			{isCobolTestFile, extractCobolMarkersFromFile},
			{isMatlabTestFile, extractMatlabMarkersFromFile},
			{isAdaTestFile, extractAdaMarkersFromFile},
		}

		for _, scanner := range langScanners {
			if scanner.check(path) {
				markers, err := scanner.extract(path)
				if err != nil {
					break
				}
				results = append(results, markers...)
				break
			}
		}

		return nil
	})

	return results, err
}

// extractGoMarkersFromFile extracts rtmx.Req() markers from Go test files.
func extractGoMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")
	currentFunc := ""

	reqPattern := regexp.MustCompile(`rtmx\.Req\(t,\s*"(REQ-[^"]+)"`)
	funcPattern := regexp.MustCompile(`^func\s+(Test\w+)\s*\(`)

	for i, line := range lines {
		if m := funcPattern.FindStringSubmatch(line); len(m) > 1 {
			currentFunc = m[1]
		}
		if m := reqPattern.FindStringSubmatch(line); len(m) > 1 {
			results = append(results, TestRequirement{
				ReqID:        m[1],
				TestFile:     filePath,
				TestFunction: currentFunc,
				LineNumber:   i + 1,
			})
		}
	}

	return results, nil
}

// extractMarkersFromFile extracts requirement markers from a Python test file
func extractMarkersFromFile(filePath string) ([]TestRequirement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var results []TestRequirement

	isConftest := filepath.Base(filePath) == "conftest.py"

	// Regex patterns for pytest markers
	reqMarkerPattern := regexp.MustCompile(`@pytest\.mark\.req\(['"](REQ-[A-Z0-9-]+)['"]\)`)
	funcPattern := regexp.MustCompile(`^(?:async\s+)?def\s+(test_\w+)\s*\(`)
	classPattern := regexp.MustCompile(`^class\s+(Test\w+)\s*[:(]`)
	otherMarkerPattern := regexp.MustCompile(`@pytest\.mark\.(scope_\w+|technique_\w+|env_\w+)`)

	// For conftest.py, also match non-test functions (fixtures)
	fixtureFuncPattern := regexp.MustCompile(`^(?:async\s+)?def\s+(\w+)\s*\(`)

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

		// Check for function definition - in conftest.py also match fixture functions
		var funcMatch []string
		if isConftest && len(pendingReqIDs) > 0 {
			funcMatch = fixtureFuncPattern.FindStringSubmatch(trimmed)
		} else {
			// For non-conftest files, try the test function pattern first
			funcMatch = funcPattern.FindStringSubmatch(trimmed)

			// If a non-test function is found and there are pending markers, discard them
			if funcMatch == nil && len(pendingReqIDs) > 0 {
				if anyFunc := fixtureFuncPattern.FindStringSubmatch(trimmed); anyFunc != nil {
					pendingReqIDs = nil
					pendingMarkers = nil
					continue
				}
			}
		}

		if funcMatch != nil {
			funcName := funcMatch[1]
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

// extractConftestRegistrations parses conftest.py for marker registration patterns
// such as config.addinivalue_line("markers", "req(id, ...): ...").
func extractConftestRegistrations(filePath string) ([]ConftestMarkerRegistration, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var results []ConftestMarkerRegistration

	// Match patterns like:
	//   config.addinivalue_line("markers", "req(id, scope=None): Link test to requirement")
	//   config.addinivalue_line("markers", "scope_unit: Unit test scope")
	// Also handles multiline calls where arguments span multiple lines.
	addiniPattern := regexp.MustCompile(
		`addinivalue_line\s*\(\s*["']markers["']\s*,\s*["'](\w+)(?:\(([^)]*)\))?\s*(?::\s*(.+?))?["']\s*\)`,
	)
	// Detect start of multiline addinivalue_line call (line contains the call but no closing paren for markers arg)
	addiniStartPattern := regexp.MustCompile(`addinivalue_line\s*\(`)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var accumulator string
	accumulatorLine := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Handle multiline accumulation
		if accumulator != "" {
			accumulator += " " + trimmed
			// Try to match the accumulated lines
			if matches := addiniPattern.FindAllStringSubmatch(accumulator, -1); matches != nil {
				for _, m := range matches {
					reg := ConftestMarkerRegistration{
						FilePath:   filePath,
						MarkerName: m[1],
						LineNumber: accumulatorLine,
					}
					if len(m) > 2 {
						reg.MarkerArgs = m[2]
					}
					if len(m) > 3 {
						reg.MarkerHelp = strings.TrimSpace(m[3])
					}
					results = append(results, reg)
				}
				accumulator = ""
			} else if lineNum-accumulatorLine > 5 {
				// Give up after 5 lines of accumulation to avoid runaway
				accumulator = ""
			}
			continue
		}

		// Check if this is a single-line match
		if matches := addiniPattern.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				reg := ConftestMarkerRegistration{
					FilePath:   filePath,
					MarkerName: m[1],
					LineNumber: lineNum,
				}
				if len(m) > 2 {
					reg.MarkerArgs = m[2]
				}
				if len(m) > 3 {
					reg.MarkerHelp = strings.TrimSpace(m[3])
				}
				results = append(results, reg)
			}
			continue
		}

		// Check for start of multiline call
		if addiniStartPattern.MatchString(trimmed) {
			accumulator = trimmed
			accumulatorLine = lineNum
		}
	}

	return results, scanner.Err()
}

// extractRustMarkersFromFile extracts requirement markers from Rust test files.
// It recognizes four marker styles:
//   - #[req("REQ-ID")] attribute macros
//   - // rtmx:req REQ-ID comment markers
//   - rtmx::req("REQ-ID") function call markers
func extractRustMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	// Patterns for marker styles
	attrPattern := regexp.MustCompile(`#\[req\("(REQ-[A-Z0-9-]+)"`)
	commentPattern := regexp.MustCompile(`//\s*(?:rtmx:req|@req)\s+(REQ-[A-Z0-9-]+)`)
	callPattern := regexp.MustCompile(`rtmx::req\("(REQ-[A-Z0-9-]+)"`)

	// Pattern for Rust function definitions (fn, pub fn, async fn, pub async fn)
	funcPattern := regexp.MustCompile(`^\s*(?:pub\s+)?(?:async\s+)?fn\s+(\w+)\s*\(`)

	// Pattern for mod blocks
	modPattern := regexp.MustCompile(`^\s*(?:pub\s+)?mod\s+(\w+)\s*\{`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentMod := ""

	for i, line := range lines {
		lineNum := i + 1

		// Track mod blocks
		if m := modPattern.FindStringSubmatch(line); len(m) > 1 {
			currentMod = m[1]
		}

		// Check for attribute macro: #[req("REQ-ID")]
		if matches := attrPattern.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				pendingReqIDs = append(pendingReqIDs, struct {
					reqID  string
					lineNo int
				}{m[1], lineNum})
			}
			continue
		}

		// Check for comment marker: // rtmx:req REQ-ID
		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for function definition
		if funcMatch := funcPattern.FindStringSubmatch(line); funcMatch != nil {
			funcName := funcMatch[1]
			if currentMod != "" {
				funcName = currentMod + "::" + funcName
			}

			// Check for function call marker inside the function body
			// We look ahead a few lines for rtmx::req() calls
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				bodyLine := lines[j]
				// Stop at next function or closing brace at column 0
				if funcPattern.MatchString(bodyLine) {
					break
				}
				if cm := callPattern.FindStringSubmatch(bodyLine); len(cm) > 1 {
					results = append(results, TestRequirement{
						ReqID:        cm[1],
						TestFile:     filePath,
						TestFunction: funcName,
						LineNumber:   j + 1,
					})
				}
			}

			// Assign pending attribute/comment markers to this function
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: funcName,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil
		}
	}

	return results, nil
}

// isRustTestFile returns true if the file should be scanned for Rust requirement markers.
// It matches:
//   - ANY .rs file (Rust unit tests are inline in src/*.rs via #[cfg(test)] mod tests)
//   - *_test.rs files (unit test convention)
//   - any .rs file inside a tests/ directory (Rust integration test convention)
//
// We scan all .rs files because idiomatic Rust embeds tests alongside source code.
func isRustTestFile(path string) bool {
	return strings.HasSuffix(path, ".rs")
}

// scanConftestFiles finds and parses conftest.py marker registrations in a directory tree
func scanConftestFiles(dir string) ([]ConftestMarkerRegistration, error) {
	var results []ConftestMarkerRegistration

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "conftest.py" {
			regs, err := extractConftestRegistrations(path)
			if err != nil {
				return nil
			}
			results = append(results, regs...)
		}
		return nil
	})

	return results, err
}
