package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractScalaMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-025")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID with def",
			content: `class AuthTest extends AnyFunSuite {
    // rtmx:req REQ-AUTH-001
    def testLogin() = {
        assert(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTest.testLogin"},
			},
		},
		{
			name: "annotation @req with test string",
			content: `class AuthSpec extends AnyFunSuite {
    @req("REQ-AUTH-002")
    test("logout works") {
        assert(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "logout works"},
			},
		},
		{
			name: "comment marker with it block",
			content: `class AuthSpec extends AnyFlatSpec {
    // rtmx:req REQ-AUTH-003
    it("should authenticate") {
        assert(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-003", TestFunction: "should authenticate"},
			},
		},
		{
			name: "multiple markers",
			content: `class UserSpec extends AnyFunSuite {
    // rtmx:req REQ-USER-001
    test("creates user") {
        assert(true)
    }

    @req("REQ-USER-002")
    test("deletes user") {
        assert(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "creates user"},
				{ReqID: "REQ-USER-002", TestFunction: "deletes user"},
			},
		},
		{
			name: "no markers",
			content: `class AuthTest extends AnyFunSuite {
    def testLogin() = {
        assert(true)
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
			tmpDir, err := os.MkdirTemp("", "rtmx-scala-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "AuthSpec.scala")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractScalaMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractScalaMarkersFromFile failed: %v", err)
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

func TestExtractScalaMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-025")

	_, err := extractScalaMarkersFromFile("/nonexistent/AuthSpec.scala")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsScalaTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-025")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthTest.scala", true},
		{"AuthTests.scala", true},
		{"AuthSpec.scala", true},
		{"Auth.scala", false},
		{"AuthTest.java", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isScalaTestFile(tt.path); got != tt.expected {
				t.Errorf("isScalaTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
