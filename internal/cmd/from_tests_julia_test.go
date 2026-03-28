package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractJuliaMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-023")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker # rtmx:req REQ-ID with testset",
			content: `using Test

# rtmx:req REQ-MATH-001
@testset "addition tests" begin
    @test 1 + 1 == 2
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-001", TestFunction: "addition tests"},
			},
		},
		{
			name: "macro @req annotation",
			content: `using Test

@req("REQ-MATH-002")
@testset "subtraction tests" begin
    @test 2 - 1 == 1
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-002", TestFunction: "subtraction tests"},
			},
		},
		{
			name: "comment marker with function",
			content: `# rtmx:req REQ-MATH-003
function test_multiply()
    @test 2 * 3 == 6
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-003", TestFunction: "test_multiply"},
			},
		},
		{
			name: "multiple markers",
			content: `using Test

# rtmx:req REQ-MATH-001
@testset "add" begin
    @test 1 + 1 == 2
end

@req("REQ-MATH-002")
@testset "sub" begin
    @test 2 - 1 == 1
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MATH-001", TestFunction: "add"},
				{ReqID: "REQ-MATH-002", TestFunction: "sub"},
			},
		},
		{
			name: "no markers",
			content: `using Test

@testset "no marker" begin
    @test true
end
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
			tmpDir, err := os.MkdirTemp("", "rtmx-julia-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "test_math.jl")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractJuliaMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractJuliaMarkersFromFile failed: %v", err)
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

func TestExtractJuliaMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-023")

	_, err := extractJuliaMarkersFromFile("/nonexistent/test_math.jl")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsJuliaTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-023")

	tests := []struct {
		path     string
		expected bool
	}{
		{"test_math.jl", true},
		{"math_test.jl", true},
		{"math.jl", false},
		{"test_math.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isJuliaTestFile(tt.path); got != tt.expected {
				t.Errorf("isJuliaTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
