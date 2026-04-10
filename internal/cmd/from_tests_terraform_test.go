package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractTerraformMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-016")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker # rtmx:req REQ-ID",
			content: `# rtmx:req REQ-INFRA-001
run "test_vpc_creation" {
  command = plan

  assert {
    condition     = aws_vpc.main.cidr_block == "10.0.0.0/16"
    error_message = "VPC CIDR block is incorrect"
  }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-INFRA-001", TestFunction: "test_vpc_creation"},
			},
		},
		{
			name: "labels with req in run block",
			content: `run "test_security_group" {
  command = plan

  labels = { req = "REQ-SEC-001" }

  assert {
    condition     = length(aws_security_group.main.ingress) > 0
    error_message = "No ingress rules"
  }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-SEC-001", TestFunction: "test_security_group"},
			},
		},
		{
			name: "multiple run blocks with markers",
			content: `# rtmx:req REQ-INFRA-001
run "test_vpc" {
  command = plan
  assert {
    condition     = true
    error_message = "fail"
  }
}

# rtmx:req REQ-INFRA-002
run "test_subnet" {
  command = plan
  assert {
    condition     = true
    error_message = "fail"
  }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-INFRA-001", TestFunction: "test_vpc"},
				{ReqID: "REQ-INFRA-002", TestFunction: "test_subnet"},
			},
		},
		{
			name: "no markers",
			content: `run "test_plain" {
  command = plan
  assert {
    condition     = true
    error_message = "fail"
  }
}
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "labels on same line as run",
			content: `run "test_iam_policy" {
  command = apply
  labels = { req = "REQ-IAM-001" }

  assert {
    condition     = true
    error_message = "fail"
  }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-IAM-001", TestFunction: "test_iam_policy"},
			},
		},
		{
			name: "mixed marker styles",
			content: `# rtmx:req REQ-MIX-001
run "test_comment_marker" {
  command = plan
  assert {
    condition     = true
    error_message = "fail"
  }
}

run "test_label_marker" {
  command = plan
  labels = { req = "REQ-MIX-002" }
  assert {
    condition     = true
    error_message = "fail"
  }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "test_comment_marker"},
				{ReqID: "REQ-MIX-002", TestFunction: "test_label_marker"},
			},
		},
		{
			name: "comment with extra whitespace",
			content: `#   rtmx:req   REQ-WS-001
run "test_whitespace" {
  command = plan
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "test_whitespace"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-terraform-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "example.tftest.hcl")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractTerraformMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractTerraformMarkersFromFile failed: %v", err)
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

func TestExtractTerraformMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-016")

	_, err := extractTerraformMarkersFromFile("/nonexistent/example.tftest.hcl")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsTerraformTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-016")

	tests := []struct {
		path     string
		expected bool
	}{
		{"main.tftest.hcl", true},
		{"vpc.tftest.hcl", true},
		{filepath.Join("tests", "main.tf"), true},
		{filepath.Join("tests", "sub", "main.tf"), true},
		{"main.tf", false},
		{"main.hcl", false},
		{"test.py", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTerraformTestFile(tt.path)
			if got != tt.expected {
				t.Errorf("isTerraformTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestScanDirectoryFindsTerraformFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory scanning hangs on Windows CI")
	}
	rtmx.Req(t, "REQ-LANG-016")

	tmpDir, err := os.MkdirTemp("", "rtmx-terraform-scan")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a .tftest.hcl file
	tfTestContent := `# rtmx:req REQ-TF-001
run "test_vpc" {
  command = plan
  assert {
    condition     = true
    error_message = "fail"
  }
}
`

	// Create a tests/ directory with a .tf file
	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests dir: %v", err)
	}

	tfContent := `# rtmx:req REQ-TF-002
run "test_subnet" {
  command = plan
}
`

	// Create a non-test .tf file (should be ignored)
	mainTfContent := `resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "vpc.tftest.hcl"), []byte(tfTestContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "subnet.tf"), []byte(tfContent), 0644); err != nil {
		t.Fatalf("Failed to write tests dir file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainTfContent), 0644); err != nil {
		t.Fatalf("Failed to write main.tf file: %v", err)
	}

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	foundIDs := make(map[string]bool)
	for _, m := range markers {
		foundIDs[m.ReqID] = true
	}

	if !foundIDs["REQ-TF-001"] {
		t.Error("REQ-TF-001 from vpc.tftest.hcl not found")
	}
	if !foundIDs["REQ-TF-002"] {
		t.Error("REQ-TF-002 from tests/subnet.tf not found")
	}
	if len(markers) != 2 {
		t.Errorf("Expected 2 markers, got %d", len(markers))
	}
}
