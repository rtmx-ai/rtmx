package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestExtractMarkersFromFile(t *testing.T) {
	rtmx.Req(t, "REQ-GO-017")

	// Create a temporary test file
	tmpDir, err := os.MkdirTemp("", "rtmx-from-tests")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	defer func() { _ = os.RemoveAll(tmpDir) }()

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

func TestExtractConftestRegistrations(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-005")

	tests := []struct {
		name     string
		content  string
		expected []ConftestMarkerRegistration
	}{
		{
			name: "standard req marker registration",
			content: `import pytest

def pytest_configure(config):
    config.addinivalue_line(
        "markers", "req(id, scope=None, technique=None, env=None): Link test to requirement"
    )
`,
			expected: []ConftestMarkerRegistration{
				{MarkerName: "req", MarkerArgs: "id, scope=None, technique=None, env=None", MarkerHelp: "Link test to requirement"},
			},
		},
		{
			name: "multiple marker registrations",
			content: `def pytest_configure(config):
    config.addinivalue_line("markers", "req(id): Link test to requirement")
    config.addinivalue_line("markers", "scope_unit: Unit test scope")
    config.addinivalue_line("markers", "scope_integration: Integration test scope")
`,
			expected: []ConftestMarkerRegistration{
				{MarkerName: "req", MarkerArgs: "id", MarkerHelp: "Link test to requirement"},
				{MarkerName: "scope_unit", MarkerHelp: "Unit test scope"},
				{MarkerName: "scope_integration", MarkerHelp: "Integration test scope"},
			},
		},
		{
			name: "single-quoted strings",
			content: `def pytest_configure(config):
    config.addinivalue_line('markers', 'req(id): Requirement marker')
`,
			expected: []ConftestMarkerRegistration{
				{MarkerName: "req", MarkerArgs: "id", MarkerHelp: "Requirement marker"},
			},
		},
		{
			name: "marker with no help text",
			content: `def pytest_configure(config):
    config.addinivalue_line("markers", "req(id)")
`,
			expected: []ConftestMarkerRegistration{
				{MarkerName: "req", MarkerArgs: "id"},
			},
		},
		{
			name: "marker with no args and no help",
			content: `def pytest_configure(config):
    config.addinivalue_line("markers", "scope_unit")
`,
			expected: []ConftestMarkerRegistration{
				{MarkerName: "scope_unit"},
			},
		},
		{
			name: "no marker registrations",
			content: `import pytest

def pytest_configure(config):
    config.addinivalue_line("disable", "something_else")
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
		{
			name: "technique and env markers",
			content: `def pytest_configure(config):
    config.addinivalue_line("markers", "technique_boundary: Boundary value testing")
    config.addinivalue_line("markers", "env_ci: CI environment marker")
`,
			expected: []ConftestMarkerRegistration{
				{MarkerName: "technique_boundary", MarkerHelp: "Boundary value testing"},
				{MarkerName: "env_ci", MarkerHelp: "CI environment marker"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-conftest-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			confFile := filepath.Join(tmpDir, "conftest.py")
			if err := os.WriteFile(confFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write conftest.py: %v", err)
			}

			regs, err := extractConftestRegistrations(confFile)
			if err != nil {
				t.Fatalf("extractConftestRegistrations failed: %v", err)
			}

			if len(regs) != len(tt.expected) {
				t.Fatalf("Expected %d registrations, got %d", len(tt.expected), len(regs))
			}

			for i, exp := range tt.expected {
				got := regs[i]
				if got.MarkerName != exp.MarkerName {
					t.Errorf("Registration %d: expected marker name %q, got %q", i, exp.MarkerName, got.MarkerName)
				}
				if got.MarkerArgs != exp.MarkerArgs {
					t.Errorf("Registration %d: expected marker args %q, got %q", i, exp.MarkerArgs, got.MarkerArgs)
				}
				if got.MarkerHelp != exp.MarkerHelp {
					t.Errorf("Registration %d: expected marker help %q, got %q", i, exp.MarkerHelp, got.MarkerHelp)
				}
				if got.FilePath != confFile {
					t.Errorf("Registration %d: expected file path %q, got %q", i, confFile, got.FilePath)
				}
				if got.LineNumber == 0 {
					t.Errorf("Registration %d: expected non-zero line number", i)
				}
			}
		})
	}
}

func TestExtractConftestRegistrationsFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-005")

	_, err := extractConftestRegistrations("/nonexistent/conftest.py")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestExtractMarkersFromConftest(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-005")

	tmpDir, err := os.MkdirTemp("", "rtmx-conftest-markers")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// conftest.py with markers on fixtures (not test_ functions)
	conftestContent := `import pytest

@pytest.fixture
@pytest.mark.req("REQ-FIX-001")
def authenticated_user():
    return {"user": "admin"}

@pytest.mark.req("REQ-FIX-002")
@pytest.mark.scope_integration
def database_connection():
    return "db://localhost"
`

	confFile := filepath.Join(tmpDir, "conftest.py")
	if err := os.WriteFile(confFile, []byte(conftestContent), 0644); err != nil {
		t.Fatalf("Failed to write conftest.py: %v", err)
	}

	markers, err := extractMarkersFromFile(confFile)
	if err != nil {
		t.Fatalf("extractMarkersFromFile for conftest.py failed: %v", err)
	}

	if len(markers) != 2 {
		t.Fatalf("Expected 2 markers from conftest.py fixtures, got %d", len(markers))
	}

	foundIDs := make(map[string]string)
	for _, m := range markers {
		foundIDs[m.ReqID] = m.TestFunction
	}

	if fn, ok := foundIDs["REQ-FIX-001"]; !ok {
		t.Error("REQ-FIX-001 not found in conftest.py markers")
	} else if fn != "authenticated_user" {
		t.Errorf("Expected function authenticated_user, got %s", fn)
	}

	if fn, ok := foundIDs["REQ-FIX-002"]; !ok {
		t.Error("REQ-FIX-002 not found in conftest.py markers")
	} else if fn != "database_connection" {
		t.Errorf("Expected function database_connection, got %s", fn)
	}

	// Verify scope marker is attached to REQ-FIX-002
	for _, m := range markers {
		if m.ReqID == "REQ-FIX-002" {
			if len(m.Markers) != 1 || m.Markers[0] != "scope_integration" {
				t.Errorf("Expected scope_integration marker on REQ-FIX-002, got %v", m.Markers)
			}
		}
	}
}

func TestScanTestDirectoryIncludesConftest(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-005")

	tmpDir, err := os.MkdirTemp("", "rtmx-scan-conftest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Regular test file
	testContent := `import pytest

@pytest.mark.req("REQ-TEST-001")
def test_something():
    pass
`
	// conftest.py with fixture markers
	conftestContent := `import pytest

@pytest.mark.req("REQ-FIX-001")
def setup_fixture():
    pass
`

	_ = os.WriteFile(filepath.Join(tmpDir, "test_main.py"), []byte(testContent), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "conftest.py"), []byte(conftestContent), 0644)

	markers, err := scanTestDirectory(tmpDir)
	if err != nil {
		t.Fatalf("scanTestDirectory failed: %v", err)
	}

	if len(markers) != 2 {
		t.Fatalf("Expected 2 markers (1 test + 1 conftest fixture), got %d", len(markers))
	}

	foundIDs := make(map[string]bool)
	for _, m := range markers {
		foundIDs[m.ReqID] = true
	}

	if !foundIDs["REQ-TEST-001"] {
		t.Error("REQ-TEST-001 from test file not found")
	}
	if !foundIDs["REQ-FIX-001"] {
		t.Error("REQ-FIX-001 from conftest.py not found")
	}
}

func TestScanConftestFiles(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-005")

	tmpDir, err := os.MkdirTemp("", "rtmx-scan-conftest-files")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Root conftest.py
	rootConftest := `def pytest_configure(config):
    config.addinivalue_line("markers", "req(id): Link test to requirement")
    config.addinivalue_line("markers", "scope_unit: Unit scope")
`
	// Subdirectory conftest.py
	subDir := filepath.Join(tmpDir, "integration")
	_ = os.MkdirAll(subDir, 0755)
	subConftest := `def pytest_configure(config):
    config.addinivalue_line("markers", "env_ci: CI environment")
`

	_ = os.WriteFile(filepath.Join(tmpDir, "conftest.py"), []byte(rootConftest), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "conftest.py"), []byte(subConftest), 0644)

	regs, err := scanConftestFiles(tmpDir)
	if err != nil {
		t.Fatalf("scanConftestFiles failed: %v", err)
	}

	if len(regs) != 3 {
		t.Fatalf("Expected 3 marker registrations across conftest files, got %d", len(regs))
	}

	markerNames := make(map[string]bool)
	for _, reg := range regs {
		markerNames[reg.MarkerName] = true
	}

	expected := []string{"req", "scope_unit", "env_ci"}
	for _, name := range expected {
		if !markerNames[name] {
			t.Errorf("Expected marker registration %q not found", name)
		}
	}
}

func TestExtractMarkersNonConftestSkipsFixtures(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-005")

	// Verify that non-conftest files still only match test_ functions
	tmpDir, err := os.MkdirTemp("", "rtmx-nonconftest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// A regular test file with a marker on a non-test function should not be picked up
	testContent := `import pytest

@pytest.mark.req("REQ-HELPER-001")
def helper_function():
    pass

@pytest.mark.req("REQ-TEST-001")
def test_real_test():
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

	// Only test_real_test should be found; helper_function should be skipped
	if len(markers) != 1 {
		t.Fatalf("Expected 1 marker (only test_ functions), got %d", len(markers))
	}

	if markers[0].ReqID != "REQ-TEST-001" {
		t.Errorf("Expected REQ-TEST-001, got %s", markers[0].ReqID)
	}
}

// TestPytestPlugin validates that the Go CLI provides full Python pytest
// integration: marker extraction from @pytest.mark.req decorators,
// scope/technique/env marker recognition, conftest.py fixture scanning,
// and verify --results consumption of pytest output.
// REQ-LANG-004: Python pytest integration with requirement markers
func TestPytestPlugin(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-004")

	t.Run("extracts_pytest_mark_req_decorators", func(t *testing.T) {
		tmpDir := t.TempDir()
		testContent := `import pytest

@pytest.mark.req("REQ-AUTH-001")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_login():
    pass

@pytest.mark.req("REQ-AUTH-002")
def test_logout():
    pass
`
		testFile := filepath.Join(tmpDir, "test_auth.py")
		_ = os.WriteFile(testFile, []byte(testContent), 0644)

		markers, err := extractMarkersFromFile(testFile)
		if err != nil {
			t.Fatalf("extraction failed: %v", err)
		}

		if len(markers) != 2 {
			t.Fatalf("expected 2 markers, got %d", len(markers))
		}

		// Find REQ-AUTH-001 and verify scope/technique/env
		for _, m := range markers {
			if m.ReqID == "REQ-AUTH-001" {
				if m.TestFunction != "test_login" {
					t.Errorf("test_function = %q, want test_login", m.TestFunction)
				}
				hasScope := false
				hasTechnique := false
				hasEnv := false
				for _, mk := range m.Markers {
					if mk == "scope_integration" {
						hasScope = true
					}
					if mk == "technique_nominal" {
						hasTechnique = true
					}
					if mk == "env_simulation" {
						hasEnv = true
					}
				}
				if !hasScope {
					t.Error("expected scope_integration marker")
				}
				if !hasTechnique {
					t.Error("expected technique_nominal marker")
				}
				if !hasEnv {
					t.Error("expected env_simulation marker")
				}
			}
		}
	})

	t.Run("extracts_class_method_markers", func(t *testing.T) {
		tmpDir := t.TempDir()
		testContent := `import pytest

class TestUserAPI:
    @pytest.mark.req("REQ-API-001")
    def test_create_user(self):
        pass

    @pytest.mark.req("REQ-API-002")
    def test_delete_user(self):
        pass
`
		testFile := filepath.Join(tmpDir, "test_api.py")
		_ = os.WriteFile(testFile, []byte(testContent), 0644)

		markers, err := extractMarkersFromFile(testFile)
		if err != nil {
			t.Fatalf("extraction failed: %v", err)
		}

		if len(markers) != 2 {
			t.Fatalf("expected 2 markers, got %d", len(markers))
		}

		// Class method should include class name
		for _, m := range markers {
			if m.ReqID == "REQ-API-001" {
				if !strings.Contains(m.TestFunction, "TestUserAPI") {
					t.Errorf("test_function should include class name, got %q", m.TestFunction)
				}
			}
		}
	})

	t.Run("conftest_fixture_markers", func(t *testing.T) {
		tmpDir := t.TempDir()
		conftestContent := `import pytest

@pytest.mark.req("REQ-FIX-001")
@pytest.fixture
def auth_client():
    return Client()
`
		confFile := filepath.Join(tmpDir, "conftest.py")
		_ = os.WriteFile(confFile, []byte(conftestContent), 0644)

		markers, err := extractMarkersFromFile(confFile)
		if err != nil {
			t.Fatalf("conftest extraction failed: %v", err)
		}

		if len(markers) != 1 {
			t.Fatalf("expected 1 marker from conftest fixture, got %d", len(markers))
		}
		if markers[0].ReqID != "REQ-FIX-001" {
			t.Errorf("expected REQ-FIX-001, got %s", markers[0].ReqID)
		}
	})

	t.Run("verify_consumes_pytest_results_json", func(t *testing.T) {
		// Verify --results accepts the RTMX JSON format that a pytest
		// plugin would produce
		tmpDir := setupTestProject(t, testDBHeader+
			"REQ-PY-001,LANG,Python,Python req,Pass,test_auth.py,test_login,Unit Test,MISSING,HIGH,1,,1,,,,,,,\n"+
			"REQ-PY-002,LANG,Python,Another req,Pass,test_auth.py,test_logout,Unit Test,MISSING,HIGH,1,,1,,,,,,,\n")

		resultsJSON := `[
  {"marker":{"req_id":"REQ-PY-001","test_name":"test_login","test_file":"test_auth.py"},"passed":true},
  {"marker":{"req_id":"REQ-PY-002","test_name":"test_logout","test_file":"test_auth.py"},"passed":false}
]`
		resultsPath := filepath.Join(tmpDir, "results.json")
		_ = os.WriteFile(resultsPath, []byte(resultsJSON), 0644)

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := newPytestVerifyCmd()
		buf := new(strings.Builder)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"verify", "--results", resultsPath, "--dry-run"})

		_ = cmd.Execute() // May return error due to failing test (REQ-PY-002)

		out := buf.String()
		// Verify the pipeline processed both results
		if !strings.Contains(out, "REQ-PY-001") {
			t.Errorf("expected REQ-PY-001 in output, got:\n%s", out)
		}
		// REQ-PY-001 should be promoted to COMPLETE
		if !strings.Contains(out, "COMPLETE") {
			t.Errorf("expected COMPLETE status change, got:\n%s", out)
		}
		// Should show both passing and failing
		if !strings.Contains(out, "PASSING") {
			t.Errorf("expected PASSING count, got:\n%s", out)
		}
	})

	t.Run("async_test_functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		testContent := `import pytest

@pytest.mark.req("REQ-ASYNC-001")
async def test_async_handler():
    await some_async_call()
`
		testFile := filepath.Join(tmpDir, "test_async.py")
		_ = os.WriteFile(testFile, []byte(testContent), 0644)

		markers, err := extractMarkersFromFile(testFile)
		if err != nil {
			t.Fatalf("extraction failed: %v", err)
		}

		if len(markers) != 1 {
			t.Fatalf("expected 1 marker for async test, got %d", len(markers))
		}
		if markers[0].TestFunction != "test_async_handler" {
			t.Errorf("expected test_async_handler, got %s", markers[0].TestFunction)
		}
	})
}

func newPytestVerifyCmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	var results string
	var dryRun bool
	verify := &cobra.Command{
		Use:  "verify",
		RunE: func(cmd *cobra.Command, args []string) error {
			verifyResults = results
			verifyDryRun = dryRun
			return runVerify(cmd, args)
		},
	}
	verify.Flags().StringVar(&results, "results", "", "")
	verify.Flags().BoolVar(&dryRun, "dry-run", false, "")
	root.AddCommand(verify)
	return root
}
