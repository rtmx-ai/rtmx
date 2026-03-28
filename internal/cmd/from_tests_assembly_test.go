package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractAssemblyMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-029")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker ; rtmx:req REQ-ID with label",
			content: `section .text

; rtmx:req REQ-ALU-001
test_add:
    mov eax, 1
    add eax, 2
    ret
`,
			expected: []TestRequirement{
				{ReqID: "REQ-ALU-001", TestFunction: "test_add"},
			},
		},
		{
			name: "multiple markers",
			content: `section .text

; rtmx:req REQ-ALU-001
test_add:
    mov eax, 1
    add eax, 2
    ret

; rtmx:req REQ-ALU-002
test_sub:
    mov eax, 3
    sub eax, 1
    ret
`,
			expected: []TestRequirement{
				{ReqID: "REQ-ALU-001", TestFunction: "test_add"},
				{ReqID: "REQ-ALU-002", TestFunction: "test_sub"},
			},
		},
		{
			name: "no markers",
			content: `section .text

test_add:
    mov eax, 1
    add eax, 2
    ret
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
			tmpDir, err := os.MkdirTemp("", "rtmx-asm-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "alu_test.asm")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractAssemblyMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractAssemblyMarkersFromFile failed: %v", err)
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

func TestExtractAssemblyMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-029")

	_, err := extractAssemblyMarkersFromFile("/nonexistent/alu_test.asm")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsAssemblyTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-029")

	tests := []struct {
		path     string
		expected bool
	}{
		{"alu_test.asm", true},
		{"alu_test.s", true},
		{"alu.asm", false},
		{"alu.s", false},
		{"alu_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isAssemblyTestFile(tt.path); got != tt.expected {
				t.Errorf("isAssemblyTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
