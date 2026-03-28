package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractFortranMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-018")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker ! rtmx:req REQ-ID with subroutine",
			content: `! rtmx:req REQ-MATH-001
subroutine test_addition
  implicit none
  integer :: result
  result = 1 + 1
  if (result /= 2) stop 1
end subroutine
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-001", TestFunction: "test_addition"},
			},
		},
		{
			name: "comment marker with function",
			content: `! rtmx:req REQ-MATH-002
function test_multiply() result(res)
  implicit none
  integer :: res
  res = 2 * 3
end function
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-002", TestFunction: "test_multiply"},
			},
		},
		{
			name: "multiple markers",
			content: `! rtmx:req REQ-MATH-001
subroutine test_add
  implicit none
end subroutine

! rtmx:req REQ-MATH-002
subroutine test_sub
  implicit none
end subroutine
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-001", TestFunction: "test_add"},
				{ReqID: "REQ-MATH-002", TestFunction: "test_sub"},
			},
		},
		{
			name: "no markers",
			content: `subroutine test_add
  implicit none
end subroutine
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
			tmpDir, err := os.MkdirTemp("", "rtmx-fortran-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "math_test.f90")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractFortranMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractFortranMarkersFromFile failed: %v", err)
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

func TestExtractFortranMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-018")

	_, err := extractFortranMarkersFromFile("/nonexistent/math_test.f90")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsFortranTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-018")

	tests := []struct {
		path     string
		expected bool
	}{
		{"math_test.f90", true},
		{"math_test.f95", true},
		{"test_math.f90", true},
		{"math.f90", false},
		{"math_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isFortranTestFile(tt.path); got != tt.expected {
				t.Errorf("isFortranTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
