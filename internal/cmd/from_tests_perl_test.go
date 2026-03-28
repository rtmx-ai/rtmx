package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractPerlMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-026")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker # rtmx:req REQ-ID with subtest",
			content: `use Test::More;

# rtmx:req REQ-AUTH-001
subtest "login test" => sub {
    ok(1, "login works");
};

done_testing();
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login test"},
			},
		},
		{
			name: "comment marker with sub",
			content: `use Test::More;

# rtmx:req REQ-AUTH-002
sub test_logout {
    ok(1, "logout works");
}

done_testing();
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "test_logout"},
			},
		},
		{
			name: "multiple markers",
			content: `use Test::More;

# rtmx:req REQ-AUTH-001
subtest "login" => sub {
    ok(1);
};

# rtmx:req REQ-AUTH-002
subtest "logout" => sub {
    ok(1);
};

done_testing();
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login"},
				{ReqID: "REQ-AUTH-002", TestFunction: "logout"},
			},
		},
		{
			name: "no markers",
			content: `use Test::More;

subtest "no marker" => sub {
    ok(1);
};

done_testing();
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
			tmpDir, err := os.MkdirTemp("", "rtmx-perl-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "auth_test.pl")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractPerlMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractPerlMarkersFromFile failed: %v", err)
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

func TestExtractPerlMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-026")

	_, err := extractPerlMarkersFromFile("/nonexistent/auth_test.pl")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsPerlTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-026")

	tests := []struct {
		path     string
		expected bool
	}{
		{"auth.t", true},
		{"auth_test.pl", true},
		{"test_auth.pl", true},
		{"auth.pl", false},
		{"auth.pm", false},
		{"auth_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isPerlTestFile(tt.path); got != tt.expected {
				t.Errorf("isPerlTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
