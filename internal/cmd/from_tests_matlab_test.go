package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractMatlabMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-014")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "percent comment marker",
			content: `% rtmx:req REQ-AUTH-001
function testLogin(testCase)
    verifyTrue(testCase, true);
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "testLogin"},
			},
		},
		{
			name: "marker inside test method of TestCase class",
			content: `classdef AuthTest < matlab.unittest.TestCase
    methods (Test)
        % rtmx:req REQ-AUTH-002
        function testLogout(testCase)
            verifyTrue(testCase, true);
        end
    end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "testLogout"},
			},
		},
		{
			name: "multiple markers on different functions",
			content: `% rtmx:req REQ-DB-001
function testConnect(testCase)
    verifyTrue(testCase, true);
end

% rtmx:req REQ-DB-002
function testDisconnect(testCase)
    verifyTrue(testCase, true);
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-DB-001", TestFunction: "testConnect"},
				{ReqID: "REQ-DB-002", TestFunction: "testDisconnect"},
			},
		},
		{
			name: "marker with extra whitespace",
			content: `%   rtmx:req   REQ-WS-001
function testWhitespace(testCase)
    verifyTrue(testCase, true);
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "testWhitespace"},
			},
		},
		{
			name: "no markers",
			content: `function testNoMarkers(testCase)
    verifyTrue(testCase, true);
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
			name: "multiple markers in classdef",
			content: `classdef FeatureTest < matlab.unittest.TestCase
    methods (Test)
        % rtmx:req REQ-FEAT-001
        function testFeatureA(testCase)
            verifyEqual(testCase, 1, 1);
        end

        % rtmx:req REQ-FEAT-002
        function testFeatureB(testCase)
            verifyEqual(testCase, 2, 2);
        end
    end
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-FEAT-001", TestFunction: "testFeatureA"},
				{ReqID: "REQ-FEAT-002", TestFunction: "testFeatureB"},
			},
		},
		{
			name: "function without test prefix is still captured if marker present",
			content: `% rtmx:req REQ-HELPER-001
function helperSetup(testCase)
    % setup code
end
`,
			expected: []TestRequirement{
				{ReqID: "REQ-HELPER-001", TestFunction: "helperSetup"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-matlab-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "testExample.m")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractMatlabMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractMatlabMarkersFromFile failed: %v", err)
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

func TestExtractMatlabMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-014")

	_, err := extractMatlabMarkersFromFile("/nonexistent/testExample.m")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsMatlabTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-014")

	tests := []struct {
		path     string
		expected bool
	}{
		{"AuthTest.m", true},
		{"testAuth.m", true},
		{"auth.m", false},
		{"AuthTest.py", false},
		{"testAuth.rb", false},
		{"MyTestHelper.m", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isMatlabTestFile(tt.path); got != tt.expected {
				t.Errorf("isMatlabTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
