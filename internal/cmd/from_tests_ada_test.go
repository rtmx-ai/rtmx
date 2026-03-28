package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractAdaMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-019")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker -- rtmx:req REQ-ID",
			content: `with AUnit.Test_Cases;
-- rtmx:req REQ-AUTH-001
procedure Test_Login is
begin
   Assert(True);
end Test_Login;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "Test_Login"},
			},
		},
		{
			name: "pragma Req marker",
			content: `with AUnit.Test_Cases;
procedure Test_Logout is
   pragma Req("REQ-AUTH-002");
begin
   Assert(True);
end Test_Logout;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "Test_Logout"},
			},
		},
		{
			name: "multiple comment markers on different procedures",
			content: `-- rtmx:req REQ-DB-001
procedure Test_Connect is
begin
   Assert(True);
end Test_Connect;

-- rtmx:req REQ-DB-002
procedure Test_Disconnect is
begin
   Assert(True);
end Test_Disconnect;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-DB-001", TestFunction: "Test_Connect"},
				{ReqID: "REQ-DB-002", TestFunction: "Test_Disconnect"},
			},
		},
		{
			name: "mixed comment and pragma markers",
			content: `-- rtmx:req REQ-MIX-001
procedure Test_Comment is
begin
   Assert(True);
end Test_Comment;

procedure Test_Pragma is
   pragma Req("REQ-MIX-002");
begin
   Assert(True);
end Test_Pragma;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "Test_Comment"},
				{ReqID: "REQ-MIX-002", TestFunction: "Test_Pragma"},
			},
		},
		{
			name: "comment marker with extra whitespace",
			content: `--   rtmx:req   REQ-WS-001
procedure Test_Whitespace is
begin
   Assert(True);
end Test_Whitespace;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "Test_Whitespace"},
			},
		},
		{
			name: "no markers",
			content: `procedure Test_No_Markers is
begin
   Assert(True);
end Test_No_Markers;
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "function instead of procedure",
			content: `-- rtmx:req REQ-FUNC-001
function Test_Function return Boolean is
begin
   return True;
end Test_Function;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-FUNC-001", TestFunction: "Test_Function"},
			},
		},
		{
			name: "pragma with single quotes",
			content: `procedure Test_Single_Quote is
   pragma Req('REQ-SQ-001');
begin
   Assert(True);
end Test_Single_Quote;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-SQ-001", TestFunction: "Test_Single_Quote"},
			},
		},
		{
			name: "pragma inside procedure body lookahead",
			content: `procedure Test_Deep_Pragma is
   some_var : Integer := 0;
   pragma Req("REQ-DEEP-001");
begin
   Assert(True);
end Test_Deep_Pragma;
`,
			expected: []TestRequirement{
				{ReqID: "REQ-DEEP-001", TestFunction: "Test_Deep_Pragma"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-ada-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "test_example.adb")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractAdaMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractAdaMarkersFromFile failed: %v", err)
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

func TestExtractAdaMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-019")

	_, err := extractAdaMarkersFromFile("/nonexistent/test_example.adb")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsAdaTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-019")

	tests := []struct {
		path     string
		expected bool
	}{
		{"auth_test.adb", true},
		{"test_auth.adb", true},
		{"auth.adb", false},
		{"auth.ads", false},
		{"auth_test.py", false},
		{"test_auth.rs", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isAdaTestFile(tt.path); got != tt.expected {
				t.Errorf("isAdaTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
