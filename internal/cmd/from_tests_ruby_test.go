package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractRubyMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-010")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker # rtmx:req REQ-ID",
			content: `# rtmx:req REQ-AUTH-001
def test_login_success
  assert true
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "test_login_success"},
			},
		},
		{
			name: "RSpec it block with req metadata",
			content: `describe "Authentication" do
  it "logs in successfully", req: "REQ-AUTH-002" do
    expect(true).to be true
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "logs in successfully"},
			},
		},
		{
			name: "RSpec it block with double-quoted req metadata",
			content: `describe "Auth" do
  it "validates token", req: "REQ-AUTH-003" do
    expect(token).to be_valid
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-003", TestFunction: "validates token"},
			},
		},
		{
			name: "RSpec it block with single-quoted req metadata",
			content: `describe "Auth" do
  it 'checks password', req: 'REQ-AUTH-004' do
    expect(password).to be_strong
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-004", TestFunction: "checks password"},
			},
		},
		{
			name: "multiple markers on different functions",
			content: `# rtmx:req REQ-AUTH-001
def test_login
  assert true
end

# rtmx:req REQ-AUTH-002
def test_logout
  assert true
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "test_login"},
				{ReqID: "REQ-AUTH-002", TestFunction: "test_logout"},
			},
		},
		{
			name: "multiple RSpec it blocks",
			content: `describe "User" do
  it "creates account", req: "REQ-USER-001" do
    expect(user).to be_persisted
  end

  it "deletes account", req: "REQ-USER-002" do
    expect(user).to be_destroyed
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-USER-001", TestFunction: "creates account"},
				{ReqID: "REQ-USER-002", TestFunction: "deletes account"},
			},
		},
		{
			name: "comment marker with extra whitespace",
			content: `#   rtmx:req   REQ-WS-001
def test_whitespace
  assert true
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "test_whitespace"},
			},
		},
		{
			name: "no markers",
			content: `def test_no_markers
  assert true
end
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "mixed comment and RSpec markers",
			content: `# rtmx:req REQ-MIX-001
def test_unit_feature
  assert true
end

describe "Feature" do
  it "works in integration", req: "REQ-MIX-002" do
    expect(true).to be true
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "test_unit_feature"},
				{ReqID: "REQ-MIX-002", TestFunction: "works in integration"},
			},
		},
		{
			name: "comment marker before class method",
			content: `class TestAuth < Minitest::Test
  # rtmx:req REQ-CLASS-001
  def test_class_method
    assert true
  end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-CLASS-001", TestFunction: "test_class_method"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-ruby-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "test_example_spec.rb")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractRubyMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractRubyMarkersFromFile failed: %v", err)
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

func TestExtractRubyMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-010")

	_, err := extractRubyMarkersFromFile("/nonexistent/test_example_spec.rb")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsRubyTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-010")

	tests := []struct {
		path     string
		expected bool
	}{
		{"auth_spec.rb", true},
		{"auth_test.rb", true},
		{"test_auth.rb", true},
		{"spec_response.rb", true},
		{"spec_request.rb", true},
		{"test/spec_utils.rb", true},
		{"auth.rb", false},
		{"spec_helper.rb", false},
		{"auth_spec.py", false},
		{"test_auth.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isRubyTestFile(tt.path); got != tt.expected {
				t.Errorf("isRubyTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
