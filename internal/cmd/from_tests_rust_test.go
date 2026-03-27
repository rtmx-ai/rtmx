package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractRustMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-005")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "attribute macro #[req(REQ-ID)]",
			content: `use rtmx::req;

#[test]
#[req("REQ-AUTH-001")]
fn test_login_success() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "test_login_success"},
			},
		},
		{
			name: "comment marker // rtmx:req REQ-ID",
			content: `#[test]
// rtmx:req REQ-SEC-010
fn test_encryption() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-SEC-010", TestFunction: "test_encryption"},
			},
		},
		{
			name: "function call rtmx::req(REQ-ID)",
			content: `#[test]
fn test_database_connect() {
    rtmx::req("REQ-DB-001");
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-DB-001", TestFunction: "test_database_connect"},
			},
		},
		{
			name: "multiple markers on different functions",
			content: `use rtmx::req;

#[test]
#[req("REQ-AUTH-001")]
fn test_login() {
    assert!(true);
}

#[test]
#[req("REQ-AUTH-002")]
fn test_logout() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "test_login"},
				{ReqID: "REQ-AUTH-002", TestFunction: "test_logout"},
			},
		},
		{
			name: "multiple markers on same function",
			content: `#[test]
#[req("REQ-AUTH-001")]
#[req("REQ-AUDIT-001")]
fn test_login_audited() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-001", TestFunction: "test_login_audited"},
				{ReqID: "REQ-AUDIT-001", TestFunction: "test_login_audited"},
			},
		},
		{
			name: "attribute macro with options",
			content: `#[test]
#[req("REQ-AUTH-002", scope = "integration", technique = "boundary")]
fn test_login_invalid() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-AUTH-002", TestFunction: "test_login_invalid"},
			},
		},
		{
			name: "mixed marker styles",
			content: `use rtmx::req;

#[test]
#[req("REQ-MIX-001")]
fn test_attr_macro() {
    assert!(true);
}

#[test]
// rtmx:req REQ-MIX-002
fn test_comment_marker() {
    assert!(true);
}

#[test]
fn test_func_call() {
    rtmx::req("REQ-MIX-003");
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MIX-001", TestFunction: "test_attr_macro"},
				{ReqID: "REQ-MIX-002", TestFunction: "test_comment_marker"},
				{ReqID: "REQ-MIX-003", TestFunction: "test_func_call"},
			},
		},
		{
			name: "no markers",
			content: `#[test]
fn test_no_markers() {
    assert!(true);
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
			name: "non-test function with marker is still captured",
			content: `#[req("REQ-UTIL-001")]
fn helper_function() {
    // helper
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-UTIL-001", TestFunction: "helper_function"},
			},
		},
		{
			name: "function inside mod tests block",
			content: `#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    #[req("REQ-MOD-001")]
    fn test_inside_mod() {
        assert!(true);
    }
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-MOD-001", TestFunction: "tests::test_inside_mod"},
			},
		},
		{
			name: "comment marker with extra whitespace",
			content: `#[test]
//   rtmx:req   REQ-WS-001
fn test_whitespace() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-WS-001", TestFunction: "test_whitespace"},
			},
		},
		{
			name: "async test function",
			content: `#[tokio::test]
#[req("REQ-ASYNC-001")]
async fn test_async_operation() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-ASYNC-001", TestFunction: "test_async_operation"},
			},
		},
		{
			name: "pub fn test function",
			content: `#[test]
#[req("REQ-PUB-001")]
pub fn test_public() {
    assert!(true);
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-PUB-001", TestFunction: "test_public"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-rust-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "test_example.rs")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractRustMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractRustMarkersFromFile failed: %v", err)
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

func TestExtractRustMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-005")

	_, err := extractRustMarkersFromFile("/nonexistent/test_example.rs")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestScanDirectoryFindsRustFiles(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-005")

	tmpDir, err := os.MkdirTemp("", "rtmx-rust-scan")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a Rust file matching *_test.rs pattern
	unitTestContent := `use rtmx::req;

#[test]
#[req("REQ-RUST-001")]
fn test_unit_feature() {
    assert!(true);
}
`
	// Create a Rust file in tests/ directory
	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests dir: %v", err)
	}

	integrationTestContent := `// rtmx:req REQ-RUST-002
#[test]
fn test_integration() {
    assert!(true);
}
`

	// Create a non-test Rust file (should be ignored)
	libContent := `pub fn add(a: i32, b: i32) -> i32 {
    a + b
}
`

	// Create a Python test file to confirm both languages work together
	pyTestContent := `import pytest

@pytest.mark.req("REQ-PY-001")
def test_python():
    pass
`

	if err := os.WriteFile(filepath.Join(tmpDir, "feature_test.rs"), []byte(unitTestContent), 0644); err != nil {
		t.Fatalf("Failed to write unit test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testsDir, "integration.rs"), []byte(integrationTestContent), 0644); err != nil {
		t.Fatalf("Failed to write integration test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "lib.rs"), []byte(libContent), 0644); err != nil {
		t.Fatalf("Failed to write lib file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test_python.py"), []byte(pyTestContent), 0644); err != nil {
		t.Fatalf("Failed to write Python test file: %v", err)
	}

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	foundIDs := make(map[string]bool)
	for _, m := range markers {
		foundIDs[m.ReqID] = true
	}

	// Should find Rust markers
	if !foundIDs["REQ-RUST-001"] {
		t.Error("REQ-RUST-001 from feature_test.rs not found")
	}
	if !foundIDs["REQ-RUST-002"] {
		t.Error("REQ-RUST-002 from tests/integration.rs not found")
	}
	// Should also find Python markers
	if !foundIDs["REQ-PY-001"] {
		t.Error("REQ-PY-001 from test_python.py not found")
	}
	// lib.rs should not produce any markers
	if len(markers) != 3 {
		t.Errorf("Expected 3 markers total (2 Rust + 1 Python), got %d", len(markers))
	}
}

func TestScanDirectoryFindsRustTestsSubdir(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-005")

	tmpDir, err := os.MkdirTemp("", "rtmx-rust-testsdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create tests/ subdirectory (standard Rust integration test location)
	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests dir: %v", err)
	}

	// .rs file in tests/ directory (not *_test.rs but should be scanned)
	content := `#[test]
fn test_api() {
    rtmx::req("REQ-API-001");
    assert!(true);
}
`
	if err := os.WriteFile(filepath.Join(testsDir, "api.rs"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	if len(markers) != 1 {
		t.Fatalf("Expected 1 marker, got %d", len(markers))
	}

	if markers[0].ReqID != "REQ-API-001" {
		t.Errorf("Expected REQ-API-001, got %s", markers[0].ReqID)
	}
}
