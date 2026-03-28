package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractJSMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-006")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `// rtmx:req REQ-AUTH-001
test("login succeeds", () => {
    expect(login()).toBeTruthy();
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login succeeds"},
			},
		},
		{
			name: "req function call",
			content: `test("login succeeds", () => {
    req("REQ-AUTH-001");
    expect(login()).toBeTruthy();
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login succeeds"},
			},
		},
		{
			name: "rtmx.req function call",
			content: `test("login succeeds", () => {
    rtmx.req("REQ-AUTH-002");
    expect(login()).toBeTruthy();
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "login succeeds"},
			},
		},
		{
			name: "describe.rtmx for Jest/Vitest",
			content: `describe.rtmx("REQ-FEAT-001", "feature tests", () => {
    test("does something", () => {
        expect(true).toBe(true);
    });
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-FEAT-001", TestFunction: "feature tests"},
			},
		},
		{
			name: "it() function instead of test()",
			content: `// rtmx:req REQ-UI-001
it("should render correctly", () => {
    expect(render()).toBeTruthy();
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-UI-001", TestFunction: "should render correctly"},
			},
		},
		{
			name: "multiple markers on different tests",
			content: `// rtmx:req REQ-AUTH-001
test("login succeeds", () => {
    expect(true).toBe(true);
});

// rtmx:req REQ-AUTH-002
test("logout succeeds", () => {
    expect(true).toBe(true);
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login succeeds"},
				{ReqID: "REQ-AUTH-002", TestFunction: "logout succeeds"},
			},
		},
		{
			name: "req call with single quotes",
			content: `test('login test', () => {
    req('REQ-AUTH-003');
    expect(true).toBe(true);
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-003", TestFunction: "login test"},
			},
		},
		{
			name: "no markers",
			content: `test("plain test", () => {
    expect(true).toBe(true);
});
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "describe.rtmx with single quotes",
			content: `describe.rtmx('REQ-FEAT-002', 'another feature', () => {
    it('works', () => {});
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-FEAT-002", TestFunction: "another feature"},
			},
		},
		{
			name: "mixed marker styles",
			content: `// rtmx:req REQ-MIX-001
test("comment marker test", () => {
    expect(true).toBe(true);
});

test("function call test", () => {
    req("REQ-MIX-002");
    expect(true).toBe(true);
});

describe.rtmx("REQ-MIX-003", "describe block", () => {
    test("inner test", () => {});
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "comment marker test"},
				{ReqID: "REQ-MIX-002", TestFunction: "function call test"},
				{ReqID: "REQ-MIX-003", TestFunction: "describe block"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-js-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "example.test.js")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractJSMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractJSMarkersFromFile failed: %v", err)
			}

			if len(markers) != len(tt.expected) {
				t.Fatalf("Expected %d markers, got %d", len(tt.expected), len(markers))
			}

			for i, exp := range tt.expected {
				got := markers[i]
				if got.ReqID != exp.ReqID {
					t.Errorf("Marker %d: expected ReqID %q, got %q", i, exp.ReqID, got.ReqID)
				}
				if got.TestFunction != exp.TestFunction {
					t.Errorf("Marker %d: expected TestFunction %q, got %q", i, exp.TestFunction, got.TestFunction)
				}
				if got.TestFile != testFile {
					t.Errorf("Marker %d: expected TestFile %q, got %q", i, testFile, got.TestFile)
				}
				if got.LineNumber == 0 {
					t.Errorf("Marker %d: expected non-zero line number", i)
				}
			}
		})
	}
}

func TestExtractJSMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-006")

	_, err := extractJSMarkersFromFile("/nonexistent/example.test.js")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsJSTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-006")

	tests := []struct {
		path     string
		expected bool
	}{
		{"app.test.js", true},
		{"app.test.ts", true},
		{"app.spec.js", true},
		{"app.spec.ts", true},
		{"__tests__/app.js", true},
		{"src/__tests__/feature.js", true},
		{"app.js", false},
		{"app.ts", false},
		{"test.py", false},
		{"README.md", false},
		{"__tests__/app.ts", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isJSTestFile(tt.path)
			if got != tt.expected {
				t.Errorf("isJSTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestScanDirectoryFindsJSFiles(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-006")

	tmpDir, err := os.MkdirTemp("", "rtmx-js-scan")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a .test.js file
	testJSContent := `// rtmx:req REQ-JS-001
test("js feature", () => {
    expect(true).toBe(true);
});
`
	// Create a .spec.ts file
	specTSContent := `// rtmx:req REQ-TS-001
it("ts feature", () => {
    expect(true).toBe(true);
});
`
	// Create a __tests__/ directory with a .js file
	testsDir := filepath.Join(tmpDir, "__tests__")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create __tests__ dir: %v", err)
	}

	jestContent := `test("jest feature", () => {
    req("REQ-JEST-001");
    expect(true).toBe(true);
});
`

	// Create a non-test JS file (should be ignored)
	libContent := `export function add(a, b) { return a + b; }
`

	if err := os.WriteFile(filepath.Join(tmpDir, "feature.test.js"), []byte(testJSContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "feature.spec.ts"), []byte(specTSContent), 0644); err != nil {
		t.Fatalf("Failed to write spec file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "jest.js"), []byte(jestContent), 0644); err != nil {
		t.Fatalf("Failed to write jest file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "lib.js"), []byte(libContent), 0644); err != nil {
		t.Fatalf("Failed to write lib file: %v", err)
	}

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	foundIDs := make(map[string]bool)
	for _, m := range markers {
		foundIDs[m.ReqID] = true
	}

	if !foundIDs["REQ-JS-001"] {
		t.Error("REQ-JS-001 from feature.test.js not found")
	}
	if !foundIDs["REQ-TS-001"] {
		t.Error("REQ-TS-001 from feature.spec.ts not found")
	}
	if !foundIDs["REQ-JEST-001"] {
		t.Error("REQ-JEST-001 from __tests__/jest.js not found")
	}
	if len(markers) != 3 {
		t.Errorf("Expected 3 markers, got %d", len(markers))
	}
}
