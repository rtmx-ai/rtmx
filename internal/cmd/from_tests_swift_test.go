package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractSwiftMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-011")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `class AuthTests: XCTestCase {
    // rtmx:req REQ-AUTH-001
    func testLogin() {
        XCTAssertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTests.testLogin"},
			},
		},
		{
			name: "annotation @Req",
			content: `class AuthTests: XCTestCase {
    @Req("REQ-AUTH-002")
    func testLogout() {
        XCTAssertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "AuthTests.testLogout"},
			},
		},
		{
			name: "multiple markers",
			content: `class UserTests: XCTestCase {
    // rtmx:req REQ-USER-001
    func testCreate() {
        XCTAssertTrue(true)
    }

    @Req("REQ-USER-002")
    func testDelete() {
        XCTAssertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "UserTests.testCreate"},
				{ReqID: "REQ-USER-002", TestFunction: "UserTests.testDelete"},
			},
		},
		{
			name: "no markers",
			content: `class AuthTests: XCTestCase {
    func testLogin() {
        XCTAssertTrue(true)
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
			tmpDir, err := os.MkdirTemp("", "rtmx-swift-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "AuthTests.swift")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractSwiftMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractSwiftMarkersFromFile failed: %v", err)
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

func TestExtractSwiftMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-011")

	_, err := extractSwiftMarkersFromFile("/nonexistent/AuthTests.swift")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsSwiftTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-011")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthTests.swift", true},
		{"AuthTest.swift", true},
		{"Auth.swift", false},
		{"AuthTests.java", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isSwiftTestFile(tt.path); got != tt.expected {
				t.Errorf("isSwiftTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
