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

// extractJavaMarkersFromFile extracts requirement markers from Java test files.
// It recognizes two marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - @Req("REQ-ID") annotation
func extractJavaMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	annotationPattern := regexp.MustCompile(`@Req\("(REQ-[A-Z0-9-]+)"\)`)
	classPattern := regexp.MustCompile(`^\s*(?:public\s+)?class\s+(\w+)`)
	methodPattern := regexp.MustCompile(`^\s*(?:public\s+|private\s+|protected\s+)?(?:static\s+)?(?:void|boolean|int|String|[A-Z]\w*)\s+(\w+)\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentClass := ""

	for i, line := range lines {
		lineNum := i + 1

		if m := classPattern.FindStringSubmatch(line); len(m) > 1 {
			currentClass = m[1]
		}

		if m := annotationPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

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

// isJavaTestFile returns true if the file should be scanned for Java requirement markers.
func isJavaTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".java") {
		return false
	}
	name := base[:len(base)-5]
	return strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "Tests")
}

// extractSwiftMarkersFromFile extracts requirement markers from Swift test files.
// It recognizes two marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - @Req("REQ-ID") annotation
func extractSwiftMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	annotationPattern := regexp.MustCompile(`@Req\("(REQ-[A-Z0-9-]+)"\)`)
	classPattern := regexp.MustCompile(`^\s*(?:final\s+)?class\s+(\w+)`)
	funcPattern := regexp.MustCompile(`^\s*func\s+(test\w+)\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentClass := ""

	for i, line := range lines {
		lineNum := i + 1

		if m := classPattern.FindStringSubmatch(line); len(m) > 1 {
			currentClass = m[1]
		}

		if m := annotationPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if funcMatch := funcPattern.FindStringSubmatch(line); funcMatch != nil {
			funcName := funcMatch[1]
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

	return results, nil
}

// isSwiftTestFile returns true if the file should be scanned for Swift requirement markers.
func isSwiftTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".swift") {
		return false
	}
	name := base[:len(base)-6]
	return strings.HasSuffix(name, "Tests") || strings.HasSuffix(name, "Test")
}

// extractDartMarkersFromFile extracts requirement markers from Dart test files.
// It recognizes two marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - req("REQ-ID") function call
func extractDartMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	reqCallPattern := regexp.MustCompile(`req\(["'](REQ-[A-Z0-9-]+)["']\)`)
	testFuncPattern := regexp.MustCompile(`^\s*(?:test|group)\s*\(\s*["']([^"']+)["']`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if testMatch := testFuncPattern.FindStringSubmatch(line); testMatch != nil {
			funcName := testMatch[1]

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

// isDartTestFile returns true if the file should be scanned for Dart requirement markers.
func isDartTestFile(path string) bool {
	return strings.HasSuffix(filepath.Base(path), "_test.dart")
}

// extractVerilogMarkersFromFile extracts requirement markers from Verilog/SystemVerilog test files.
// It recognizes:
//   - // rtmx:req REQ-ID comment markers
func extractVerilogMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	modulePattern := regexp.MustCompile(`^\s*module\s+(\w+)`)
	taskPattern := regexp.MustCompile(`^\s*task\s+(\w+)`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentModule := ""

	for i, line := range lines {
		lineNum := i + 1

		if m := modulePattern.FindStringSubmatch(line); len(m) > 1 {
			currentModule = m[1]
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if taskMatch := taskPattern.FindStringSubmatch(line); taskMatch != nil {
			funcName := taskMatch[1]
			if currentModule != "" {
				funcName = currentModule + "." + funcName
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

	return results, nil
}

// isVerilogTestFile returns true if the file should be scanned for Verilog requirement markers.
func isVerilogTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.sv") || strings.HasSuffix(base, "_test.v") ||
		strings.HasSuffix(base, "_tb.sv") || strings.HasSuffix(base, "_tb.v")
}

// extractFortranMarkersFromFile extracts requirement markers from Fortran test files.
// It recognizes:
//   - ! rtmx:req REQ-ID comment markers
func extractFortranMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`!\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	subroutinePattern := regexp.MustCompile(`(?i)^\s*subroutine\s+(\w+)`)
	funcPatternFortran := regexp.MustCompile(`(?i)^\s*(?:integer|real|logical|character|double\s+precision)?\s*function\s+(\w+)`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := subroutinePattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := funcPatternFortran.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		}

		if funcName != "" {
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

// isFortranTestFile returns true if the file should be scanned for Fortran requirement markers.
func isFortranTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.f90") || strings.HasSuffix(base, "_test.f95") ||
		strings.HasSuffix(base, "_test.f03") || strings.HasPrefix(base, "test_") &&
		(strings.HasSuffix(base, ".f90") || strings.HasSuffix(base, ".f95") || strings.HasSuffix(base, ".f03"))
}

// extractPHPMarkersFromFile extracts requirement markers from PHP test files.
// It recognizes two marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - @req("REQ-ID") docblock annotation
func extractPHPMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	docblockPattern := regexp.MustCompile(`@req\("(REQ-[A-Z0-9-]+)"\)`)
	classPattern := regexp.MustCompile(`^\s*class\s+(\w+)`)
	methodPattern := regexp.MustCompile(`^\s*(?:public\s+|protected\s+|private\s+)?(?:static\s+)?function\s+(\w+)\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentClass := ""

	for i, line := range lines {
		lineNum := i + 1

		if m := classPattern.FindStringSubmatch(line); len(m) > 1 {
			currentClass = m[1]
		}

		if m := docblockPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

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

// isPHPTestFile returns true if the file should be scanned for PHP requirement markers.
func isPHPTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".php") {
		return false
	}
	name := base[:len(base)-4]
	return strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "Tests")
}

// extractElixirMarkersFromFile extracts requirement markers from Elixir test files.
// It recognizes two marker styles:
//   - # rtmx:req REQ-ID comment markers
//   - @tag req: "REQ-ID" module attribute
func extractElixirMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`#\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	tagPattern := regexp.MustCompile(`@tag\s+req:\s*"(REQ-[A-Z0-9-]+)"`)
	testPattern := regexp.MustCompile(`^\s*test\s+"([^"]+)"`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := tagPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if testMatch := testPattern.FindStringSubmatch(line); testMatch != nil {
			funcName := testMatch[1]
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

// isElixirTestFile returns true if the file should be scanned for Elixir requirement markers.
func isElixirTestFile(path string) bool {
	return strings.HasSuffix(filepath.Base(path), "_test.exs")
}

// extractRMarkersFromFile extracts requirement markers from R test files.
// It recognizes:
//   - # rtmx:req REQ-ID comment markers
func extractRMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`#\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	testThatPattern := regexp.MustCompile(`test_that\s*\(\s*["']([^"']+)["']`)
	funcPattern := regexp.MustCompile(`^\s*(\w+)\s*<-\s*function\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := testThatPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := funcPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		}

		if funcName != "" {
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

// isRTestFile returns true if the file should be scanned for R requirement markers.
func isRTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".R") && !strings.HasSuffix(base, ".r") {
		return false
	}
	name := base[:len(base)-2]
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(name, "_test") || strings.HasPrefix(base, "test-")
}

// extractJuliaMarkersFromFile extracts requirement markers from Julia test files.
// It recognizes two marker styles:
//   - # rtmx:req REQ-ID comment markers
//   - @req("REQ-ID") macro
func extractJuliaMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`#\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	macroPattern := regexp.MustCompile(`@req\("(REQ-[A-Z0-9-]+)"\)`)
	testsetPattern := regexp.MustCompile(`@testset\s+"([^"]+)"`)
	funcPattern := regexp.MustCompile(`^\s*function\s+(\w+)`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := macroPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := testsetPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := funcPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		}

		if funcName != "" {
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

// isJuliaTestFile returns true if the file should be scanned for Julia requirement markers.
func isJuliaTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".jl") {
		return false
	}
	return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.jl")
}

// extractKotlinMarkersFromFile extracts requirement markers from Kotlin test files.
// It recognizes two marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - @Req("REQ-ID") annotation
func extractKotlinMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	annotationPattern := regexp.MustCompile(`@Req\("(REQ-[A-Z0-9-]+)"\)`)
	classPattern := regexp.MustCompile(`^\s*class\s+(\w+)`)
	funPattern := regexp.MustCompile(`^\s*(?:fun|suspend\s+fun)\s+(\w+)\s*\(`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentClass := ""

	for i, line := range lines {
		lineNum := i + 1

		if m := classPattern.FindStringSubmatch(line); len(m) > 1 {
			currentClass = m[1]
		}

		if m := annotationPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if len(pendingReqIDs) > 0 {
			if funMatch := funPattern.FindStringSubmatch(line); funMatch != nil {
				funcName := funMatch[1]
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

// isKotlinTestFile returns true if the file should be scanned for Kotlin requirement markers.
func isKotlinTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".kt") {
		return false
	}
	name := base[:len(base)-3]
	return strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "Tests")
}

// extractScalaMarkersFromFile extracts requirement markers from Scala test files.
// It recognizes two marker styles:
//   - // rtmx:req REQ-ID comment markers
//   - @req("REQ-ID") annotation
func extractScalaMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`//\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	annotationPattern := regexp.MustCompile(`@req\("(REQ-[A-Z0-9-]+)"\)`)
	classPattern := regexp.MustCompile(`^\s*class\s+(\w+)`)
	defPattern := regexp.MustCompile(`^\s*def\s+(\w+)`)
	testStringPattern := regexp.MustCompile(`^\s*(?:test|it)\s*\(\s*"([^"]+)"`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}
	currentClass := ""

	for i, line := range lines {
		lineNum := i + 1

		if m := classPattern.FindStringSubmatch(line); len(m) > 1 {
			currentClass = m[1]
		}

		if m := annotationPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := testStringPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := defPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
			if currentClass != "" {
				funcName = currentClass + "." + funcName
			}
		}

		if funcName != "" && len(pendingReqIDs) > 0 {
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

// isScalaTestFile returns true if the file should be scanned for Scala requirement markers.
func isScalaTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".scala") {
		return false
	}
	name := base[:len(base)-6]
	return strings.HasSuffix(name, "Test") || strings.HasSuffix(name, "Spec") || strings.HasSuffix(name, "Tests")
}

// extractPerlMarkersFromFile extracts requirement markers from Perl test files.
// It recognizes:
//   - # rtmx:req REQ-ID comment markers
func extractPerlMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`#\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	subPattern := regexp.MustCompile(`^\s*sub\s+(\w+)`)
	subtestPattern := regexp.MustCompile(`subtest\s+["']([^"']+)["']`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := subtestPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := subPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		}

		if funcName != "" {
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

// isPerlTestFile returns true if the file should be scanned for Perl requirement markers.
func isPerlTestFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, ".t") {
		return true
	}
	if strings.HasSuffix(base, "_test.pl") || strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".pl") {
		return true
	}
	return false
}

// extractLuaMarkersFromFile extracts requirement markers from Lua test files.
// It recognizes:
//   - -- rtmx:req REQ-ID comment markers
func extractLuaMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`--\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	funcPattern := regexp.MustCompile(`^\s*(?:local\s+)?function\s+(\w+)`)
	itPattern := regexp.MustCompile(`^\s*it\s*\(\s*["']([^"']+)["']`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := itPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := funcPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		}

		if funcName != "" {
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

// isLuaTestFile returns true if the file should be scanned for Lua requirement markers.
func isLuaTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".lua") {
		return false
	}
	return strings.HasSuffix(base, "_test.lua") || strings.HasPrefix(base, "test_")
}

// extractHaskellMarkersFromFile extracts requirement markers from Haskell test files.
// It recognizes:
//   - -- rtmx:req REQ-ID comment markers
func extractHaskellMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`--\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	itPattern := regexp.MustCompile(`^\s*it\s+"([^"]+)"`)
	describePattern := regexp.MustCompile(`^\s*describe\s+"([^"]+)"`)
	funcPattern := regexp.MustCompile(`^(\w+)\s+::`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		var funcName string
		if m := itPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := describePattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		} else if m := funcPattern.FindStringSubmatch(line); len(m) > 1 {
			funcName = m[1]
		}

		if funcName != "" {
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

// isHaskellTestFile returns true if the file should be scanned for Haskell requirement markers.
func isHaskellTestFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".hs") {
		return false
	}
	name := base[:len(base)-3]
	return strings.HasSuffix(name, "Spec") || strings.HasSuffix(name, "Test")
}

// extractAssemblyMarkersFromFile extracts requirement markers from Assembly test files.
// It recognizes:
//   - ; rtmx:req REQ-ID comment markers
func extractAssemblyMarkersFromFile(filePath string) ([]TestRequirement, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var results []TestRequirement
	lines := strings.Split(string(data), "\n")

	commentPattern := regexp.MustCompile(`;\s*rtmx:req\s+(REQ-[A-Z0-9-]+)`)
	labelPattern := regexp.MustCompile(`^(\w+):`)

	var pendingReqIDs []struct {
		reqID  string
		lineNo int
	}

	for i, line := range lines {
		lineNum := i + 1

		if m := commentPattern.FindStringSubmatch(line); len(m) > 1 {
			pendingReqIDs = append(pendingReqIDs, struct {
				reqID  string
				lineNo int
			}{m[1], lineNum})
			continue
		}

		if labelMatch := labelPattern.FindStringSubmatch(line); labelMatch != nil {
			funcName := labelMatch[1]
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

// isAssemblyTestFile returns true if the file should be scanned for Assembly requirement markers.
func isAssemblyTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.asm") || strings.HasSuffix(base, "_test.s")
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

