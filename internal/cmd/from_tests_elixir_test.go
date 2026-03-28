package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractElixirMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-021")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker # rtmx:req REQ-ID",
			content: `defmodule AuthTest do
  use ExUnit.Case

  # rtmx:req REQ-AUTH-001
  test "login succeeds" do
    assert true
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "login succeeds"},
			},
		},
		{
			name: "tag annotation @tag req:",
			content: `defmodule AuthTest do
  use ExUnit.Case

  @tag req: "REQ-AUTH-002"
  test "logout works" do
    assert true
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "logout works"},
			},
		},
		{
			name: "multiple markers",
			content: `defmodule UserTest do
  use ExUnit.Case

  # rtmx:req REQ-USER-001
  test "creates user" do
    assert true
  end

  @tag req: "REQ-USER-002"
  test "deletes user" do
    assert true
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "creates user"},
				{ReqID: "REQ-USER-002", TestFunction: "deletes user"},
			},
		},
		{
			name: "no markers",
			content: `defmodule AuthTest do
  use ExUnit.Case

  test "login succeeds" do
    assert true
  end
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
			tmpDir, err := os.MkdirTemp("", "rtmx-elixir-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "auth_test.exs")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractElixirMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractElixirMarkersFromFile failed: %v", err)
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

func TestExtractElixirMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-021")

	_, err := extractElixirMarkersFromFile("/nonexistent/auth_test.exs")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsElixirTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-021")

	tests := []struct {
		path     string
		expected bool
	}{
		{"auth_test.exs", true},
		{"user_test.exs", true},
		{"auth.exs", false},
		{"auth_test.ex", false},
		{"auth_test.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isElixirTestFile(tt.path); got != tt.expected {
				t.Errorf("isElixirTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
