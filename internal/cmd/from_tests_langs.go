package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// extractJSMarkersFromFile extracts requirement markers from JavaScript/TypeScript test files.
// It recognizes three marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - req("REQ-ID") or rtmx.req("REQ-ID") function calls
//   - describe.rtmx("REQ-ID", ...) for Jest/Vitest
func extractJSMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	// Patterns for the three marker styles
	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	reqCallPattern := regexp.MustCompile(`(?:rtmx\.)?req\(["'](REQ-[A-Z0-9-]+)["']\)`)
	describeRtmxPattern := regexp.MustCompile(`describe\.rtmx\(["'](REQ-[A-Z0-9-]+)["']\s*,\s*["']([^"']+)["']`)

	// Pattern for test/it function definitions: test("name", ...) or it("name", ...)
	testFuncPattern := regexp.MustCompile(`^\s*(?:test|it)\s*\(\s*["']([^"']+)["']`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		// Check for describe.rtmx("REQ-ID", "description", ...)
		if m := describeRtmxPattern.FindStringSubmatch(line); len(m) > 2 {
			results = append(results, TestRequirement{
				ReqID:        m[1],
				TestFile:     filePath,
				TestFunction: m[2],
				LineNumber:   lineNum,
			})
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

		// Check for test/it function definition
		if testMatch := testFuncPattern.FindStringSubmatch(line); testMatch != nil {
			funcName := testMatch[1]

			// Assign pending comment markers to this test
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: funcName,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil

			// Look ahead for req() calls inside the test body
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				bodyLine := lines[j]
				// Stop at next test/it definition
				if testFuncPattern.MatchString(bodyLine) {
					break
				}
				if cm := reqCallPattern.FindStringSubmatch(bodyLine); len(cm) > 1 {
					results = append(results, TestRequirement{
						ReqID:        cm[1],
						TestFile:     filePath,
						TestFunction: funcName,
						LineNumber:   j + 1,
					})
				}
			}
		}
	}

	return results, nil
}

// isJSTestFile returns true if the file should be scanned for JavaScript/TypeScript requirement markers.
// It matches:
//   - *.test.js, *.test.ts files
//   - *.spec.js, *.spec.ts files
//   - files inside a __tests__/ directory with .js or .ts extension
func isJSTestFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(base)

	// Check for .test.js, .test.ts, .spec.js, .spec.ts
	if ext == ".js" || ext == ".ts" {
		nameWithoutExt := base[:len(base)-len(ext)]
		if strings.HasSuffix(nameWithoutExt, ".test") || strings.HasSuffix(nameWithoutExt, ".spec") {
			return true
		}
	}

	// Check if inside a __tests__/ directory
	if (ext == ".js" || ext == ".ts") && strings.Contains(filepath.ToSlash(path), "__tests__/") {
		return true
	}

	return false
}

// extractCSharpMarkersFromFile extracts requirement markers from C# test files.
// It recognizes two marker styles:
//   - [Req("REQ-ID")] attribute
//   - // rtmx:req REQ-ID comment markers
func extractCSharpMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	// Patterns
	reqAttrPattern := regexp.MustCompile(`\[Req\("(REQ-[A-Z0-9-]+)"\)\]`)
	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	classPattern := regexp.MustCompile(`^\s*(?:public\s+)?class\s+(\w+)`)
	methodPattern := regexp.MustCompile(`^\s*(?:public\s+)?(?:async\s+)?(?:\w+\s+)+(\w+)\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentClass := ""

	for i, line := range lines {
		lineNum := i + 1

		// Track class context
		if m := classPattern.FindStringSubmatch(line); len(m) > 1 {
			currentClass = m[1]
		}

		// Check for [Req("REQ-ID")] attribute
		if matches := reqAttrPattern.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				pendingReqIDs = append(pendingReqIDs, struct {
					reqID  string
					lineNo int
				}{m[1], lineNum})
			}
			continue
		}

		// Check for // rtmx:req REQ-ID comment
		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for method definition
		if len(pendingReqIDs) > 0 {
			if methodMatch := methodPattern.FindStringSubmatch(line); methodMatch != nil {
				funcName := methodMatch[1]
				if currentClass != "" {
					funcName = currentClass + "." + funcName
				}

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
	}

	return results, nil
}

// isCSharpTestFile returns true if the file should be scanned for C# requirement markers.
// It matches:
//   - *Test.cs files
//   - *Tests.cs files
//   - *_test.cs files
func isCSharpTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".cs") {
		return false
	}
	name := base[:len(base)-3]
	if strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "Tests") || strings.HasSuffix(name, "_test") {
		return true
	}
	return false
}

// extractCppMarkersFromFile extracts requirement markers from C/C++ test files.
// It recognizes three marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - RTMX_REQ("REQ-ID") macro
//   - TEST_F/TEST/TEST_P with // rtmx:req on preceding line
func extractCppMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	// Patterns
	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	macroPattern := regexp.MustCompile(`RTMX_REQ\("(REQ-[A-Z0-9-]+)"\)`)

	// GoogleTest patterns: TEST(Suite, Name), TEST_F(Suite, Name), TEST_P(Suite, Name)
	gtestPattern := regexp.MustCompile(`^\s*TEST(?:_F|_P)?\s*\(\s*(\w+)\s*,\s*(\w+)\s*\)`)

	// Catch2 pattern: TEST_CASE("description", ...)
	catch2Pattern := regexp.MustCompile(`^\s*TEST_CASE\s*\(\s*["']([^"']+)["']`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		// Check for comment marker: // rtmx:req REQ-ID
		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for GoogleTest TEST/TEST_F/TEST_P
		if gtestMatch := gtestPattern.FindStringSubmatch(line); gtestMatch != nil {
			funcName := gtestMatch[1] + "." + gtestMatch[2]

			// Assign pending comment markers to this test
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: funcName,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil

			// Look ahead for RTMX_REQ macro inside the test body
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				bodyLine := lines[j]
				if gtestPattern.MatchString(bodyLine) || catch2Pattern.MatchString(bodyLine) {
					break
				}
				if cm := macroPattern.FindStringSubmatch(bodyLine); len(cm) > 1 {
					results = append(results, TestRequirement{
						ReqID:        cm[1],
						TestFile:     filePath,
						TestFunction: funcName,
						LineNumber:   j + 1,
					})
				}
			}
			continue
		}

		// Check for Catch2 TEST_CASE
		if catch2Match := catch2Pattern.FindStringSubmatch(line); catch2Match != nil {
			funcName := catch2Match[1]

			// Assign pending comment markers
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: funcName,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil

			// Look ahead for RTMX_REQ macro
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				bodyLine := lines[j]
				if gtestPattern.MatchString(bodyLine) || catch2Pattern.MatchString(bodyLine) {
					break
				}
				if cm := macroPattern.FindStringSubmatch(bodyLine); len(cm) > 1 {
					results = append(results, TestRequirement{
						ReqID:        cm[1],
						TestFile:     filePath,
						TestFunction: funcName,
						LineNumber:   j + 1,
					})
				}
			}
			continue
		}
	}

	return results, nil
}

// isCppTestFile returns true if the file should be scanned for C/C++ requirement markers.
// It matches:
//   - *_test.cpp, *_test.cc, *_test.c files
//   - test_*.cpp files
func isCppTestFile(path string) bool {
	base := filepath.Base(path)

	if strings.HasSuffix(base, "_test.cpp") || strings.HasSuffix(base, "_test.cc") ||
		strings.HasSuffix(base, "_test.c") {
		return true
	}
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".cpp") {
		return true
	}
	return false
}

// extractTerraformMarkersFromFile extracts requirement markers from Terraform test files.
// It recognizes two marker styles:
//   - # rtmx:req REQ-ID comment markers in .tftest.hcl files
//   - labels = { req = "REQ-ID" } in run blocks
func extractTerraformMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	// Patterns
	commentPattern := regexp.MustCompile(`#\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	runBlockPattern := regexp.MustCompile(`^\s*run\s+"([^"]+)"\s*\{`)
	labelsPattern := regexp.MustCompile(`labels\s*=\s*\{[^}]*req\s*=\s*"(REQ-[A-Z0-9-]+)"`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentRun := ""

	for i, line := range lines {
		lineNum := i + 1

		// Check for comment marker: # rtmx:req REQ-ID
		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for run block
		if runMatch := runBlockPattern.FindStringSubmatch(line); runMatch != nil {
			currentRun = runMatch[1]

			// Assign pending comment markers
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: currentRun,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil
		}

		// Check for labels with req inside a run block
		if currentRun != "" {
			if m := labelsPattern.FindStringSubmatch(line); len(m) > 1 {
				results = append(results, TestRequirement{
					ReqID:        m[1],
					TestFile:     filePath,
					TestFunction: currentRun,
					LineNumber:   lineNum,
				})
			}
		}
	}

	return results, nil
}

// isTerraformTestFile returns true if the file should be scanned for Terraform requirement markers.
// It matches:
//   - *.tftest.hcl files
//   - *.tf files inside a tests/ directory
func isTerraformTestFile(path string) bool {
	base := filepath.Base(path)

	if strings.HasSuffix(base, ".tftest.hcl") {
		return true
	}

	// Check if .tf file is inside a tests/ directory
	if strings.HasSuffix(base, ".tf") {
		dir := filepath.Dir(path)
		for dir != "." && dir != "/" {
			if filepath.Base(dir) == "tests" {
				return true
			}
			dir = filepath.Dir(dir)
		}
	}

	return false
}

// scanConftestFiles finds and parses conftest.py marker registrations in a directory tree



// extractRubyMarkersFromFile extracts requirement markers from Ruby test files.
// It recognizes two marker styles:
//   - # rtmx:req REQ-ID comment markers
//   - it "...", req: "REQ-ID" RSpec metadata tags
func extractRubyMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`#\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	rspecPattern := regexp.MustCompile(`it\s+["']([^"']+)["']\s*,\s*req:\s*["'](REQ-[A-Z0-9-]+)["']`)
	funcPattern := regexp.MustCompile(`^\s*def\s+(test_\w+)`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		// Check for RSpec metadata tag (self-contained, no pending needed)
		if m := rspecPattern.FindStringSubmatch(line); len(m) > 2 {
			results = append(results, TestRequirement{
				ReqID:        m[2],
				TestFile:     filePath,
				TestFunction: m[1],
				LineNumber:   lineNum,
			})
			continue
		}

		// Check for comment marker
		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for Ruby def method
		if funcMatch := funcPattern.FindStringSubmatch(line); funcMatch != nil {
			funcName := funcMatch[1]
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

// isRubyTestFile returns true if the file should be scanned for Ruby requirement markers.
// It matches: *_spec.rb, *_test.rb, test_*.rb
func isRubyTestFile(path string) bool {
	if !strings.HasSuffix(path, ".rb") {
		return false
	}
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_spec.rb") || strings.HasSuffix(base, "_test.rb") {
		return true
	}
	if strings.HasPrefix(base, "test_") {
		return true
	}
	return false
}

// extractCobolMarkersFromFile extracts requirement markers from COBOL test files.
// It recognizes two marker styles:
//   - * rtmx:req REQ-ID in column 7 comment (COBOL fixed-format)
//   - *> rtmx:req REQ-ID free-format comment
func extractCobolMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	// Fixed-format: column 7 is '*', rest contains rtmx:req
	fixedPattern := regexp.MustCompile(`^\s{0,6}\*\s+rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	// Free-format: *> comment
	freePattern := regexp.MustCompile(`^\*>\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	// COBOL paragraph name: identifier followed by a period at the start of a line
	// Paragraph names are typically uppercase with hyphens
	paragraphPattern := regexp.MustCompile(`^\s{0,7}([A-Z][A-Z0-9-]*)\.\s*$`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		// Check for fixed-format comment marker
		if m := fixedPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for free-format comment marker
		if m := freePattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for paragraph name
		if paragraphMatch := paragraphPattern.FindStringSubmatch(line); paragraphMatch != nil {
			paragraphName := paragraphMatch[1]
			// Skip COBOL division/section keywords
			if paragraphName == "PROCEDURE" || paragraphName == "IDENTIFICATION" ||
				paragraphName == "DATA" || paragraphName == "ENVIRONMENT" ||
				paragraphName == "WORKING-STORAGE" || paragraphName == "FILE" {
				continue
			}
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: paragraphName,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil
		}
	}

	return results, nil
}

// isCobolTestFile returns true if the file should be scanned for COBOL requirement markers.
// It matches: *-test.cob, *-test.cbl, test-*.cob, test-*.cbl
func isCobolTestFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, "-test.cob") || strings.HasSuffix(base, "-test.cbl") {
		return true
	}
	if (strings.HasPrefix(base, "test-") && strings.HasSuffix(base, ".cob")) ||
		(strings.HasPrefix(base, "test-") && strings.HasSuffix(base, ".cbl")) {
		return true
	}
	return false
}

// extractMatlabMarkersFromFile extracts requirement markers from MATLAB test files.
// It recognizes:
//   - % rtmx:req REQ-ID comment markers
//   - Markers before function definitions (including test methods in TestCase classes)
func extractMatlabMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`%\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	funcPattern := regexp.MustCompile(`^\s*function\s+(?:\w+\s*=\s*)?(\w+)\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		// Check for comment marker
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

// isMatlabTestFile returns true if the file should be scanned for MATLAB requirement markers.
// It matches: *Test.m, test*.m
func isMatlabTestFile(path string) bool {
	if !strings.HasSuffix(path, ".m") {
		return false
	}
	base := filepath.Base(path)
	nameWithoutExt := strings.TrimSuffix(base, ".m")
	if strings.HasSuffix(nameWithoutExt, "Test") {
		return true
	}
	if strings.HasPrefix(base, "test") {
		return true
	}
	return false
}

// extractAdaMarkersFromFile extracts requirement markers from Ada test files.
// It recognizes two marker styles:
//   - -- rtmx:req REQ-ID comment markers
//   - pragma Req("REQ-ID") pragma markers
func extractAdaMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`--\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	pragmaPattern := regexp.MustCompile(`pragma\s+Req\s*\(\s*["'](REQ-[A-Z0-9-]+)["']\s*\)`)
	// Match procedure or function declarations
	procPattern := regexp.MustCompile(`(?i)^\s*(?:overriding\s+)?(?:procedure|function)\s+(\w+)`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		// Check for comment marker
		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		// Check for procedure/function definition
		if procMatch := procPattern.FindStringSubmatch(line); procMatch != nil {
			funcName := procMatch[1]

			// Assign pending comment markers to this procedure/function
			for _, pending := range pendingReqIDs {
				results = append(results, TestRequirement{
					ReqID:        pending.reqID,
					TestFile:     filePath,
					TestFunction: funcName,
					LineNumber:   pending.lineNo,
				})
			}
			pendingReqIDs = nil

			// Look ahead for pragma Req inside the procedure/function body
			for j := i + 1; j < len(lines) && j < i+20; j++ {
				bodyLine := lines[j]
				// Stop at next procedure/function or "begin" keyword (end of declarations)
				trimmed := strings.TrimSpace(bodyLine)
				if strings.HasPrefix(strings.ToLower(trimmed), "begin") {
					break
				}
				if procPattern.MatchString(bodyLine) {
					break
				}
				if pm := pragmaPattern.FindStringSubmatch(bodyLine); len(pm) > 1 {
					results = append(results, TestRequirement{
						ReqID:        pm[1],
						TestFile:     filePath,
						TestFunction: funcName,
						LineNumber:   j + 1,
					})
				}
			}
		}
	}

	return results, nil
}

// isAdaTestFile returns true if the file should be scanned for Ada requirement markers.
// It matches: *_test.adb, test_*.adb
func isAdaTestFile(path string) bool {
	if !strings.HasSuffix(path, ".adb") {
		return false
	}
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_test.adb") || strings.HasPrefix(base, "test_") {
		return true
	}
	return false
}

// scanConftestFiles finds and parses conftest.py marker registrations in a directory tree

