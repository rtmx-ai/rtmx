package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/results"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestBuildPytestRTMXResults(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-004")

	markers := []TestRequirement{
		{
			ReqID:        "REQ-PY-001",
			TestFile:     "tests/test_auth.py",
			TestFunction: "TestAuth::test_login",
			LineNumber:   12,
		},
		{
			ReqID:        "REQ-PY-002",
			TestFile:     "tests/test_auth.py",
			TestFunction: "test_logout",
			LineNumber:   20,
		},
	}
	cases := []junitTestCase{
		{ClassName: "tests.test_auth.TestAuth", Name: "test_login", Time: 0.25},
		{ClassName: "tests.test_auth", Name: "test_logout", Failures: []interface{}{struct{}{}}},
	}

	got := buildPytestRTMXResults(markers, cases)
	if len(got) != 2 {
		t.Fatalf("expected 2 RTMX results, got %d", len(got))
	}
	if got[0].Marker.ReqID != "REQ-PY-001" || !got[0].Passed {
		t.Fatalf("expected passing REQ-PY-001 result, got %#v", got[0])
	}
	if got[1].Marker.ReqID != "REQ-PY-002" || got[1].Passed {
		t.Fatalf("expected failing REQ-PY-002 result, got %#v", got[1])
	}
	if got[0].Duration != 250 {
		t.Fatalf("expected duration in ms, got %v", got[0].Duration)
	}
}

func TestParsePytestJUnit(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-004")

	tmpDir := t.TempDir()
	junitPath := filepath.Join(tmpDir, "pytest.xml")
	xml := `<?xml version="1.0" encoding="utf-8"?>
<testsuites>
  <testsuite name="pytest">
    <testcase classname="tests.test_auth.TestAuth" name="test_login" time="0.1"></testcase>
    <testcase classname="tests.test_auth" name="test_logout" time="0.2"><failure>boom</failure></testcase>
  </testsuite>
</testsuites>`
	if err := os.WriteFile(junitPath, []byte(xml), 0644); err != nil {
		t.Fatalf("failed to write junit fixture: %v", err)
	}

	cases, err := parsePytestJUnit(junitPath)
	if err != nil {
		t.Fatalf("parsePytestJUnit failed: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}
	if cases[1].Name != "test_logout" || len(cases[1].Failures) != 1 {
		t.Fatalf("expected failed logout case, got %#v", cases[1])
	}
}

func TestFromPytestNoRunWritesResults(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-004")

	tmpDir := t.TempDir()
	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("failed to create tests dir: %v", err)
	}

	testFile := filepath.Join(testsDir, "test_auth.py")
	testContent := `import pytest

class TestAuth:
    @pytest.mark.req("REQ-PY-001")
    def test_login(self):
        pass
`
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write pytest fixture: %v", err)
	}

	junitPath := filepath.Join(tmpDir, "pytest.xml")
	junit := `<?xml version="1.0" encoding="utf-8"?>
<testsuite name="pytest">
  <testcase classname="tests.test_auth.TestAuth" name="test_login" time="0.1"></testcase>
</testsuite>`
	if err := os.WriteFile(junitPath, []byte(junit), 0644); err != nil {
		t.Fatalf("failed to write junit fixture: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "results.json")
	origCommand, origJUnit, origOutput, origNoRun := fromPytestCommand, fromPytestJUnit, fromPytestOutput, fromPytestNoRun
	fromPytestCommand = "pytest"
	fromPytestJUnit = junitPath
	fromPytestOutput = outputPath
	fromPytestNoRun = true
	defer func() {
		fromPytestCommand, fromPytestJUnit, fromPytestOutput, fromPytestNoRun = origCommand, origJUnit, origOutput, origNoRun
	}()

	cmd := newTestRootCmd()
	if err := runFromPytest(cmd, []string{testsDir}); err != nil {
		t.Fatalf("runFromPytest failed: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	var parsed []results.Result
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid RTMX results JSON: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 result, got %d", len(parsed))
	}
	if parsed[0].Marker.ReqID != "REQ-PY-001" || !parsed[0].Passed {
		t.Fatalf("unexpected result: %s", strings.TrimSpace(string(data)))
	}
}
