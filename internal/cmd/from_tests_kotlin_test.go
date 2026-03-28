package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractKotlinMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-024")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `class AuthTest {
    // rtmx:req REQ-AUTH-001
    fun testLogin() {
        assertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTest.testLogin"},
			},
		},
		{
			name: "annotation @Req",
			content: `class AuthTest {
    @Req("REQ-AUTH-002")
    fun testLogout() {
        assertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "AuthTest.testLogout"},
			},
		},
		{
			name: "multiple markers",
			content: `class UserTest {
    // rtmx:req REQ-USER-001
    fun testCreate() {
        assertTrue(true)
    }

    @Req("REQ-USER-002")
    fun testDelete() {
        assertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "UserTest.testCreate"},
				{ReqID: "REQ-USER-002", TestFunction: "UserTest.testDelete"},
			},
		},
		{
			name: "suspend fun",
			content: `class AsyncTest {
    // rtmx:req REQ-ASYNC-001
    suspend fun testAsync() {
        assertTrue(true)
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-ASYNC-001", TestFunction: "AsyncTest.testAsync"},
			},
		},
		{
			name: "no markers",
			content: `class AuthTest {
    fun testLogin() {
        assertTrue(true)
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
			tmpDir, err := os.MkdirTemp("", "rtmx-kotlin-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "AuthTest.kt")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractKotlinMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractKotlinMarkersFromFile failed: %v", err)
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

func TestExtractKotlinMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-024")

	_, err := extractKotlinMarkersFromFile("/nonexistent/AuthTest.kt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsKotlinTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-024")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthTest.kt", true},
		{"AuthTests.kt", true},
		{"Auth.kt", false},
		{"AuthTest.java", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isKotlinTestFile(tt.path); got != tt.expected {
				t.Errorf("isKotlinTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
