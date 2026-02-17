package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExtractMarkersFromFile(t *testing.T) {
	// Create a temporary test file
	tmpDir, err := os.MkdirTemp("", "rtmx-from-tests")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testContent := `import pytest

@pytest.mark.req("REQ-TEST-001")
@pytest.mark.scope_unit
def test_first_feature():
    pass

@pytest.mark.req("REQ-TEST-002")
@pytest.mark.technique_nominal
def test_second_feature():
    pass

class TestClass:
    @pytest.mark.req("REQ-TEST-003")
    def test_method(self):
        pass
`

	testFile := filepath.Join(tmpDir, "test_example.py")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	markers, err := extractMarkersFromFile(testFile)
	if err != nil {
		t.Fatalf("extractMarkersFromFile failed: %v", err)
	}

	if len(markers) != 3 {
		t.Errorf("Expected 3 markers, got %d", len(markers))
	}

	// Check first marker
	found := false
	for _, m := range markers {
		if m.ReqID == "REQ-TEST-001" {
			found = true
			if m.TestFunction != "test_first_feature" {
				t.Errorf("Expected test_first_feature, got %s", m.TestFunction)
			}
			if len(m.Markers) != 1 || m.Markers[0] != "scope_unit" {
				t.Errorf("Expected scope_unit marker, got %v", m.Markers)
			}
		}
	}
	if !found {
		t.Error("REQ-TEST-001 not found")
	}

	// Check class method marker
	found = false
	for _, m := range markers {
		if m.ReqID == "REQ-TEST-003" {
			found = true
			if !strings.Contains(m.TestFunction, "TestClass") {
				t.Errorf("Expected TestClass in function name, got %s", m.TestFunction)
			}
		}
	}
	if !found {
		t.Error("REQ-TEST-003 not found")
	}
}

func TestScanTestDirectory(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "rtmx-scan-tests")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testContent1 := `import pytest

@pytest.mark.req("REQ-SCAN-001")
def test_one():
    pass
`
	testContent2 := `import pytest

@pytest.mark.req("REQ-SCAN-002")
def test_two():
    pass
`
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.MkdirAll(subDir, 0755)

	_ = os.WriteFile(filepath.Join(tmpDir, "test_a.py"), []byte(testContent1), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "test_b.py"), []byte(testContent2), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "helper.py"), []byte("# not a test"), 0644)

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	if len(markers) != 2 {
		t.Errorf("Expected 2 markers, got %d", len(markers))
	}

	foundIDs := make(map[string]bool)
	for _, m := range markers {
		foundIDs[m.ReqID] = true
	}

	if !foundIDs["REQ-SCAN-001"] || !foundIDs["REQ-SCAN-002"] {
		t.Errorf("Missing expected requirement IDs: %v", foundIDs)
	}
}

func TestFromTestsCommandHelp(t *testing.T) {
	rootCmd := newTestRootCmd()
	rootCmd.AddCommand(newTestFromTestsCmd())

	output, err := executeCommand(rootCmd, "from-tests", "--help")
	if err != nil {
		t.Fatalf("from-tests --help failed: %v", err)
	}

	expectedPhrases := []string{
		"from-tests",
		"--show-all",
		"--show-missing",
		"--update",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected help to contain %q", phrase)
		}
	}
}

// newTestFromTestsCmd creates a fresh from-tests command for testing
func newTestFromTestsCmd() *cobra.Command {
	var showAll, showMissing, update bool

	cmd := &cobra.Command{
		Use:   "from-tests [test_path]",
		Short: "Scan test files for requirement markers",
		RunE: func(cmd *cobra.Command, args []string) error {
			fromTestsShowAll = showAll
			fromTestsShowMissing = showMissing
			fromTestsUpdate = update
			return runFromTests(cmd, args)
		},
	}
	cmd.Flags().BoolVar(&showAll, "show-all", false, "show all markers found")
	cmd.Flags().BoolVar(&showMissing, "show-missing", false, "show requirements not in database")
	cmd.Flags().BoolVar(&update, "update", false, "update RTM database")
	return cmd
}
