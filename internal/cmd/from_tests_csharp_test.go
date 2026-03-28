package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractCSharpMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-009")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "Req attribute",
			content: `using Xunit;

public class LoginTests
{
    [Fact]
    [Req("REQ-AUTH-001")]
    public void TestLoginSuccess()
    {
        Assert.True(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "LoginTests.TestLoginSuccess"},
			},
		},
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `using Xunit;

public class SecurityTests
{
    [Fact]
    // rtmx:req REQ-SEC-010
    public void TestEncryption()
    {
        Assert.True(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-SEC-010", TestFunction: "SecurityTests.TestEncryption"},
			},
		},
		{
			name: "multiple markers on different methods",
			content: `using NUnit.Framework;

[TestFixture]
public class AuthTests
{
    [Test]
    [Req("REQ-AUTH-001")]
    public void TestLogin()
    {
        Assert.Pass();
    }

    [Test]
    [Req("REQ-AUTH-002")]
    public void TestLogout()
    {
        Assert.Pass();
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuthTests.TestLogin"},
				{ReqID: "REQ-AUTH-002", TestFunction: "AuthTests.TestLogout"},
			},
		},
		{
			name: "multiple attributes on same method",
			content: `using Xunit;

public class AuditTests
{
    [Fact]
    [Req("REQ-AUTH-001")]
    [Req("REQ-AUDIT-001")]
    public void TestLoginAudited()
    {
        Assert.True(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "AuditTests.TestLoginAudited"},
				{ReqID: "REQ-AUDIT-001", TestFunction: "AuditTests.TestLoginAudited"},
			},
		},
		{
			name: "no markers",
			content: `using Xunit;

public class PlainTests
{
    [Fact]
    public void TestSomething()
    {
        Assert.True(true);
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
		{
			name: "async test method",
			content: `using Xunit;

public class AsyncTests
{
    [Fact]
    [Req("REQ-ASYNC-001")]
    public async Task TestAsyncOperation()
    {
        await Task.CompletedTask;
        Assert.True(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-ASYNC-001", TestFunction: "AsyncTests.TestAsyncOperation"},
			},
		},
		{
			name: "method without class context",
			content: `[Fact]
[Req("REQ-NOCLASS-001")]
public void TestWithoutClass()
{
    Assert.True(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-NOCLASS-001", TestFunction: "TestWithoutClass"},
			},
		},
		{
			name: "mixed marker styles",
			content: `using Xunit;

public class MixedTests
{
    [Fact]
    [Req("REQ-MIX-001")]
    public void TestAttribute()
    {
        Assert.True(true);
    }

    [Fact]
    // rtmx:req REQ-MIX-002
    public void TestComment()
    {
        Assert.True(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "MixedTests.TestAttribute"},
				{ReqID: "REQ-MIX-002", TestFunction: "MixedTests.TestComment"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-csharp-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "ExampleTests.cs")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractCSharpMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractCSharpMarkersFromFile failed: %v", err)
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

func TestExtractCSharpMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-009")

	_, err := extractCSharpMarkersFromFile("/nonexistent/ExampleTests.cs")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsCSharpTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-009")

	tests := []struct {
		path     string
		expected bool
	}{
		{"LoginTest.cs", true},
		{"LoginTests.cs", true},
		{"login_test.cs", true},
		{"Login.cs", false},
		{"LoginTestHelper.cs", false},
		{"test.py", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isCSharpTestFile(tt.path)
			if got != tt.expected {
				t.Errorf("isCSharpTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestScanDirectoryFindsCSharpFiles(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-009")

	tmpDir, err := os.MkdirTemp("", "rtmx-csharp-scan")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testContent := `using Xunit;

public class FeatureTests
{
    [Fact]
    [Req("REQ-CS-001")]
    public void TestFeature()
    {
        Assert.True(true);
    }
}
`

	libContent := `public class Feature
{
    public bool IsEnabled() => true;
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "FeatureTests.cs"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "Feature.cs"), []byte(libContent), 0644); err != nil {
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

	if !foundIDs["REQ-CS-001"] {
		t.Error("REQ-CS-001 from FeatureTests.cs not found")
	}
	if len(markers) != 1 {
		t.Errorf("Expected 1 marker, got %d", len(markers))
	}
}
