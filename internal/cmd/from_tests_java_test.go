package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractJavaMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-008")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `public class AuthTest {
    // rtmx:req REQ-AUTH-001
    public void testLogin() {
        assertTrue(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTest.testLogin"},
			},
		},
		{
			name: "annotation @Req",
			content: `public class AuthTest {
    @Req("REQ-AUTH-002")
    public void testLogout() {
        assertTrue(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "AuthTest.testLogout"},
			},
		},
		{
			name: "multiple markers",
			content: `public class UserTest {
    // rtmx:req REQ-USER-001
    public void testCreate() {
        assertTrue(true);
    }

    @Req("REQ-USER-002")
    public void testDelete() {
        assertTrue(true);
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
			content: `public class AuthTest {
    public void testLogin() {
        assertTrue(true);
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
			tmpDir, err := os.MkdirTemp("", "rtmx-java-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "AuthTest.java")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractJavaMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractJavaMarkersFromFile failed: %v", err)
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

func TestExtractJavaMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-008")

	_, err := extractJavaMarkersFromFile("/nonexistent/AuthTest.java")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsJavaTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-008")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthTest.java", true},
		{"AuthTests.java", true},
		{"Auth.java", false},
		{"AuthTest.py", false},
		{"TestHelper.java", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isJavaTestFile(tt.path); got != tt.expected {
				t.Errorf("isJavaTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
