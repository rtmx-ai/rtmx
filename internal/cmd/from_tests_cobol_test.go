package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractCobolMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-017")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "fixed-format column 7 comment marker",
			content: `       IDENTIFICATION DIVISION.
       PROGRAM-ID. TEST-AUTH.
      * rtmx:req REQ-AUTH-001
       PROCEDURE DIVISION.
       TEST-LOGIN.
           DISPLAY "Testing login".
           STOP RUN.
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "TEST-LOGIN"},
			},
		},
		{
			name: "free-format comment marker",
			content: `IDENTIFICATION DIVISION.
PROGRAM-ID. TEST-AUTH.
*> rtmx:req REQ-AUTH-002
PROCEDURE DIVISION.
TEST-LOGOUT.
    DISPLAY "Testing logout".
    STOP RUN.
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "TEST-LOGOUT"},
			},
		},
		{
			name: "multiple markers on different paragraphs",
			content: `       IDENTIFICATION DIVISION.
       PROGRAM-ID. TEST-SUITE.
      * rtmx:req REQ-DB-001
       PROCEDURE DIVISION.
       TEST-CONNECT.
           DISPLAY "Testing connect".
      * rtmx:req REQ-DB-002
       TEST-DISCONNECT.
           DISPLAY "Testing disconnect".
           STOP RUN.
`,
			expected: []TestRequirement{
				{ReqID: "REQ-DB-001", TestFunction: "TEST-CONNECT"},
				{ReqID: "REQ-DB-002", TestFunction: "TEST-DISCONNECT"},
			},
		},
		{
			name: "mixed fixed and free format markers",
			content: `       IDENTIFICATION DIVISION.
       PROGRAM-ID. TEST-MIX.
      * rtmx:req REQ-MIX-001
       PROCEDURE DIVISION.
       TEST-FIXED.
           DISPLAY "Fixed format test".
*> rtmx:req REQ-MIX-002
       TEST-FREE.
           DISPLAY "Free format test".
           STOP RUN.
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "TEST-FIXED"},
				{ReqID: "REQ-MIX-002", TestFunction: "TEST-FREE"},
			},
		},
		{
			name: "no markers",
			content: `       IDENTIFICATION DIVISION.
       PROGRAM-ID. TEST-EMPTY.
       PROCEDURE DIVISION.
       TEST-NOTHING.
           DISPLAY "No markers".
           STOP RUN.
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "marker with extra whitespace",
			content: `      *   rtmx:req   REQ-WS-001
       PROCEDURE DIVISION.
       TEST-WHITESPACE.
           DISPLAY "Whitespace test".
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "TEST-WHITESPACE"},
			},
		},
		{
			name: "free-format marker with extra whitespace",
			content: `*>   rtmx:req   REQ-WS-002
PROCEDURE DIVISION.
TEST-FREE-WS.
    DISPLAY "Free whitespace test".
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-002", TestFunction: "TEST-FREE-WS"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-cobol-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "test-example.cob")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractCobolMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractCobolMarkersFromFile failed: %v", err)
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

func TestExtractCobolMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-017")

	_, err := extractCobolMarkersFromFile("/nonexistent/test-example.cob")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsCobolTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-017")

	tests := []struct {
		path     string
		expected bool
	}{
		{"auth-test.cob", true},
		{"auth-test.cbl", true},
		{"test-auth.cob", true},
		{"test-auth.cbl", true},
		{"auth.cob", false},
		{"auth.cbl", false},
		{"auth-test.py", false},
		{"test-auth.rs", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isCobolTestFile(tt.path); got != tt.expected {
				t.Errorf("isCobolTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
