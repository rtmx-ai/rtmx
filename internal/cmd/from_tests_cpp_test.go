package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractCppMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-012")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `#include <gtest/gtest.h>

// rtmx:req REQ-AUTH-001
TEST(AuthTest, LoginSuccess) {
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTest.LoginSuccess"},
			},
		},
		{
			name: "RTMX_REQ macro",
			content: `#include <gtest/gtest.h>

TEST(SecurityTest, Encryption) {
    RTMX_REQ("REQ-SEC-010");
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-SEC-010", TestFunction: "SecurityTest.Encryption"},
			},
		},
		{
			name: "TEST_F with comment marker on preceding line",
			content: `#include <gtest/gtest.h>

class DatabaseTest : public ::testing::Test {};

// rtmx:req REQ-DB-001
TEST_F(DatabaseTest, Connect) {
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-DB-001", TestFunction: "DatabaseTest.Connect"},
			},
		},
		{
			name: "multiple markers on different tests",
			content: `#include <gtest/gtest.h>

// rtmx:req REQ-AUTH-001
TEST(AuthTest, Login) {
    EXPECT_TRUE(true);
}

// rtmx:req REQ-AUTH-002
TEST(AuthTest, Logout) {
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTest.Login"},
				{ReqID: "REQ-AUTH-002", TestFunction: "AuthTest.Logout"},
			},
		},
		{
			name: "RTMX_REQ macro inside TEST_F",
			content: `#include <gtest/gtest.h>

class ApiTest : public ::testing::Test {};

TEST_F(ApiTest, GetEndpoint) {
    RTMX_REQ("REQ-API-001");
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-API-001", TestFunction: "ApiTest.GetEndpoint"},
			},
		},
		{
			name: "no markers",
			content: `#include <gtest/gtest.h>

TEST(PlainTest, Something) {
    EXPECT_TRUE(true);
}
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "mixed marker styles",
			content: `#include <gtest/gtest.h>

// rtmx:req REQ-MIX-001
TEST(MixTest, CommentMarker) {
    EXPECT_TRUE(true);
}

TEST(MixTest, MacroMarker) {
    RTMX_REQ("REQ-MIX-002");
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "MixTest.CommentMarker"},
				{ReqID: "REQ-MIX-002", TestFunction: "MixTest.MacroMarker"},
			},
		},
		{
			name: "TEST_P parameterized test",
			content: `#include <gtest/gtest.h>

// rtmx:req REQ-PARAM-001
TEST_P(ParamTest, Works) {
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-PARAM-001", TestFunction: "ParamTest.Works"},
			},
		},
		{
			name: "comment with extra whitespace",
			content: `#include <gtest/gtest.h>

//   rtmx:req   REQ-WS-001
TEST(WhitespaceTest, Works) {
    EXPECT_TRUE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "WhitespaceTest.Works"},
			},
		},
		{
			name: "Catch2 TEST_CASE with comment marker",
			content: `#include <catch2/catch.hpp>

// rtmx:req REQ-CATCH-001
TEST_CASE("feature works", "[feature]") {
    REQUIRE(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-CATCH-001", TestFunction: "feature works"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-cpp-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "example_test.cpp")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractCppMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractCppMarkersFromFile failed: %v", err)
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

func TestExtractCppMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-012")

	_, err := extractCppMarkersFromFile("/nonexistent/example_test.cpp")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsCppTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-012")

	tests := []struct {
		path     string
		expected bool
	}{
		{"feature_test.cpp", true},
		{"feature_test.cc", true},
		{"test_feature.cpp", true},
		{"feature_test.c", true},
		{"feature.cpp", false},
		{"feature.c", false},
		{"test.py", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isCppTestFile(tt.path)
			if got != tt.expected {
				t.Errorf("isCppTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestScanDirectoryFindsCppFiles(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-012")

	tmpDir, err := os.MkdirTemp("", "rtmx-cpp-scan")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testContent := `#include <gtest/gtest.h>

// rtmx:req REQ-CPP-001
TEST(FeatureTest, Works) {
    EXPECT_TRUE(true);
}
`

	libContent := `int add(int a, int b) { return a + b; }
`

	if err := os.WriteFile(filepath.Join(tmpDir, "feature_test.cpp"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "feature.cpp"), []byte(libContent), 0644); err != nil {
		t.Fatalf("Failed to write lib file: %v", err)
	}

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	foundIDs := make(map[string]bool)
	for _, m := range markers {
		foundIDs[m.ReqID] = true
	}

	if !foundIDs["REQ-CPP-001"] {
		t.Error("REQ-CPP-001 from feature_test.cpp not found")
	}
	if len(markers) != 1 {
		t.Errorf("Expected 1 marker, got %d", len(markers))
	}
}
