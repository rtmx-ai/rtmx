package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractDartMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-013")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `// rtmx:req REQ-AUTH-001
test('login succeeds', () {
  expect(true, isTrue);
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login succeeds"},
			},
		},
		{
			name: "req function call inside test",
			content: `test('logout works', () {
  req("REQ-AUTH-002");
  expect(true, isTrue);
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "logout works"},
			},
		},
		{
			name: "group with comment marker",
			content: `// rtmx:req REQ-USER-001
group('user management', () {
  test('creates user', () {
    expect(true, isTrue);
  });
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "user management"},
			},
		},
		{
			name: "multiple markers",
			content: `// rtmx:req REQ-AUTH-001
test('login', () {
  expect(true, isTrue);
});

// rtmx:req REQ-AUTH-002
test('logout', () {
  expect(true, isTrue);
});
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login"},
				{ReqID: "REQ-AUTH-002", TestFunction: "logout"},
			},
		},
		{
			name: "no markers",
			content: `test('no marker', () {
  expect(true, isTrue);
});
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-dart-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "auth_test.dart")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractDartMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractDartMarkersFromFile failed: %v", err)
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

func TestExtractDartMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-013")

	_, err := extractDartMarkersFromFile("/nonexistent/auth_test.dart")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsDartTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-013")

	tests := []struct {
		path     string
		expected bool
	}{
		{"auth_test.dart", true},
		{"widget_test.dart", true},
		{"auth.dart", false},
		{"auth_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isDartTestFile(tt.path); got != tt.expected {
				t.Errorf("isDartTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
