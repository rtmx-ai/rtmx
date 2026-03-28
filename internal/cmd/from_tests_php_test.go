package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractPHPMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-020")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `<?php
class AuthTest extends TestCase {
    // rtmx:req REQ-AUTH-001
    public function testLogin() {
        $this->assertTrue(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTest.testLogin"},
			},
		},
		{
			name: "docblock @req annotation",
			content: `<?php
class AuthTest extends TestCase {
    /**
     * @req("REQ-AUTH-002")
     */
    public function testLogout() {
        $this->assertTrue(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "AuthTest.testLogout"},
			},
		},
		{
			name: "multiple markers",
			content: `<?php
class UserTest extends TestCase {
    // rtmx:req REQ-USER-001
    public function testCreate() {
        $this->assertTrue(true);
    }

    /**
     * @req("REQ-USER-002")
     */
    public function testDelete() {
        $this->assertTrue(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "UserTest.testCreate"},
				{ReqID: "REQ-USER-002", TestFunction: "UserTest.testDelete"},
			},
		},
		{
			name: "no markers",
			content: `<?php
class AuthTest extends TestCase {
    public function testLogin() {
        $this->assertTrue(true);
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-php-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "AuthTest.php")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractPHPMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractPHPMarkersFromFile failed: %v", err)
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

func TestExtractPHPMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-020")

	_, err := extractPHPMarkersFromFile("/nonexistent/AuthTest.php")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsPHPTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-020")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthTest.php", true},
		{"AuthTests.php", true},
		{"Auth.php", false},
		{"TestHelper.php", false},
		{"AuthTest.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isPHPTestFile(tt.path); got != tt.expected {
				t.Errorf("isPHPTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
