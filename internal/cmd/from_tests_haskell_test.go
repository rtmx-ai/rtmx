package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractHaskellMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-028")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker -- rtmx:req REQ-ID with it block",
			content: `module AuthSpec where

import Test.Hspec

spec :: Spec
spec = do
  -- rtmx:req REQ-AUTH-001
  it "authenticates user" $ do
    True ` + "`shouldBe`" + ` True
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "authenticates user"},
			},
		},
		{
			name: "comment marker with describe block",
			content: `module AuthSpec where

-- rtmx:req REQ-AUTH-002
describe "authentication" $ do
  it "works" $ do
    True ` + "`shouldBe`" + ` True
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "authentication"},
			},
		},
		{
			name: "comment marker with type signature",
			content: `module AuthSpec where

-- rtmx:req REQ-AUTH-003
testLogin :: IO ()
testLogin = do
  return ()
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-003", TestFunction: "testLogin"},
			},
		},
		{
			name: "multiple markers",
			content: `module UserSpec where

-- rtmx:req REQ-USER-001
  it "creates user" $ do
    True ` + "`shouldBe`" + ` True

-- rtmx:req REQ-USER-002
  it "deletes user" $ do
    True ` + "`shouldBe`" + ` True
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "creates user"},
				{ReqID: "REQ-USER-002", TestFunction: "deletes user"},
			},
		},
		{
			name: "no markers",
			content: `module AuthSpec where

spec :: Spec
spec = do
  it "no marker" $ do
    True ` + "`shouldBe`" + ` True
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
			tmpDir, err := os.MkdirTemp("", "rtmx-haskell-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "AuthSpec.hs")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractHaskellMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractHaskellMarkersFromFile failed: %v", err)
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

func TestExtractHaskellMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-028")

	_, err := extractHaskellMarkersFromFile("/nonexistent/AuthSpec.hs")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsHaskellTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-028")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthSpec.hs", true},
		{"AuthTest.hs", true},
		{"Auth.hs", false},
		{"AuthSpec.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isHaskellTestFile(tt.path); got != tt.expected {
				t.Errorf("isHaskellTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
