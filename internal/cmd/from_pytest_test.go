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

func TestBuildPytestRTMXResultsDimensionsAndSkip(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-004",
		rtmx.Scope("unit"), rtmx.Technique("nominal"), rtmx.Env("simulation"))

	markers := []TestRequirement{
		{ReqID: "REQ-PY-010", TestFile: "tests/test_x.py", TestFunction: "test_a", LineNumber: 5,
			Markers: []string{"scope_unit", "technique_nominal", "env_simulation"}},
		{ReqID: "REQ-PY-011", TestFile: "tests/test_x.py", TestFunction: "test_b", LineNumber: 9,
			Markers: []string{"scope_integration", "technique_monte_carlo", "env_static_field"}},
		{ReqID: "REQ-PY-012", TestFile: "tests/test_x.py", TestFunction: "test_skipped", LineNumber: 14,
			Markers: []string{"scope_unit", "technique_nominal", "env_simulation"}},
	}
	cases := []junitTestCase{
		{ClassName: "tests.test_x", Name: "test_a", Time: 0.1},
		{ClassName: "tests.test_x", Name: "test_b", Time: 0.2},
		{ClassName: "tests.test_x", Name: "test_skipped", Skipped: []interface{}{struct{}{}}},
	}

	got := buildPytestRTMXResults(markers, cases)

	// The skipped test is omitted entirely (neither pass nor fail).
	if len(got) != 2 {
		t.Fatalf("expected 2 results (skipped omitted), got %d: %#v", len(got), got)
	}
	byReq := map[string]results.Marker{}
	for _, r := range got {
		byReq[r.Marker.ReqID] = r.Marker
	}
	if _, ok := byReq["REQ-PY-012"]; ok {
		t.Error("skipped test must be omitted from results")
	}
	if m := byReq["REQ-PY-010"]; m.Scope != "unit" || m.Technique != "nominal" || m.Env != "simulation" {
		t.Errorf("dimensions not mapped for REQ-PY-010: %#v", m)
	}
	if m := byReq["REQ-PY-011"]; m.Scope != "integration" || m.Technique != "monte_carlo" || m.Env != "static_field" {
		t.Errorf("dimensions not mapped for REQ-PY-011: %#v", m)
	}
}

func TestBuildPytestRTMXResultsFileQualifiedJoin(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-004",
		rtmx.Scope("unit"), rtmx.Technique("stress"), rtmx.Env("simulation"))

	// Two different files define a module-level test with the SAME function name.
	// The passing one and the failing one must each join to their own file, not
	// collide on the bare name.
	markers := []TestRequirement{
		{ReqID: "REQ-A-001", TestFile: "packages/foo/tests/test_dup.py", TestFunction: "test_thing", LineNumber: 3},
		{ReqID: "REQ-B-001", TestFile: "tests/test_dup.py", TestFunction: "test_thing", LineNumber: 7},
	}
	cases := []junitTestCase{
		{ClassName: "packages.foo.tests.test_dup", Name: "test_thing",
			File: "packages/foo/tests/test_dup.py", Time: 0.1},
		{ClassName: "tests.test_dup", Name: "test_thing",
			File: "tests/test_dup.py", Failures: []interface{}{struct{}{}}},
	}

	got := buildPytestRTMXResults(markers, cases)
	res := map[string]bool{}
	for _, r := range got {
		res[r.Marker.ReqID] = r.Passed
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d: %#v", len(got), got)
	}
	if !res["REQ-A-001"] {
		t.Error("REQ-A-001 (packages/foo) should join its PASSING case, not the failing same-named one")
	}
	if res["REQ-B-001"] {
		t.Error("REQ-B-001 (tests/) should join its FAILING case, not the passing same-named one")
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
	fromPytestJUnit = []string{junitPath}
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

// TestScanPytestMarkersLowercaseSuffix is the regression for the Phoenix
// decomposition convention: a req-id with an optional trailing lowercase
// letter (REQ-HW-STRUCT-002c) must be CAPTURED by the pytest marker scanner.
// Previously the capture char-class was uppercase-only ([A-Z0-9-]+), so the
// whole @pytest.mark.req(...) match failed and the marker was silently dropped.
func TestScanPytestMarkersLowercaseSuffix(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test_lowercase.py")
	body := "import pytest\n\n" +
		"@pytest.mark.req(\"REQ-HW-STRUCT-002c\")\n" +
		"@pytest.mark.scope_unit\n" +
		"@pytest.mark.technique_nominal\n" +
		"@pytest.mark.env_simulation\n" +
		"def test_thing():\n    assert True\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	markers, err := scanPytestMarkers(p)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(markers) != 1 || markers[0].ReqID != "REQ-HW-STRUCT-002c" {
		t.Fatalf("expected to capture REQ-HW-STRUCT-002c, got %#v", markers)
	}
}

// --- PR-B: multi-file JUnit + JUnit-driven marker discovery ---

// TestClassNameToPath covers resolving a JUnit test case to its source file when
// the case carries no `file` attribute (REQ-LANG-031): a dotted class name maps
// to a path (dots -> slashes, ".py"), hyphenated package dirs are preserved, and a
// trailing ClassName segment is dropped. The `file` attribute takes precedence.
func TestClassNameToPath(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-031",
		rtmx.Scope("unit"), rtmx.Technique("nominal"), rtmx.Env("simulation"))
	cases := map[string]string{
		"tests.test_auth": "tests/test_auth.py",
		"packages.signal-processing.tests.test_range": "packages/signal-processing/tests/test_range.py",
		"packages.foo.tests.test_x.TestClass":         "packages/foo/tests/test_x.py",
		"":                                            "",
	}
	for in, want := range cases {
		if got := classNameToPath(in); got != want {
			t.Errorf("classNameToPath(%q) = %q, want %q", in, got, want)
		}
	}
	if got := caseSourceFile(junitTestCase{ClassName: "a.b.test_x", File: "pkg/test_x.py"}); got != "pkg/test_x.py" {
		t.Errorf("caseSourceFile file-attr precedence = %q, want pkg/test_x.py", got)
	}
}

// TestExpandJUnitPaths covers repeatable, glob-expanded --junitxml (REQ-LANG-030).
func TestExpandJUnitPaths(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-030",
		rtmx.Scope("unit"), rtmx.Technique("nominal"), rtmx.Env("simulation"))
	dir := t.TempDir()
	for _, n := range []string{"junit-a.xml", "junit-b.xml"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("<x/>"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// glob matches both; the repeated explicit path is de-duplicated.
	got, err := expandJUnitPaths([]string{filepath.Join(dir, "junit-*.xml"), filepath.Join(dir, "junit-a.xml")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 unique files, got %v", got)
	}
	// a literal, non-matching path is preserved verbatim (surfaces later as a read error).
	got2, err := expandJUnitPaths([]string{"nope.xml"})
	if err != nil || len(got2) != 1 || got2[0] != "nope.xml" {
		t.Errorf("literal passthrough = %v (err %v)", got2, err)
	}
}

func withFromPytestGlobals(t *testing.T, junit []string, output string) {
	t.Helper()
	oc, oj, oo, on := fromPytestCommand, fromPytestJUnit, fromPytestOutput, fromPytestNoRun
	fromPytestCommand, fromPytestJUnit, fromPytestOutput, fromPytestNoRun = "pytest", junit, output, true
	t.Cleanup(func() {
		fromPytestCommand, fromPytestJUnit, fromPytestOutput, fromPytestNoRun = oc, oj, oo, on
	})
}

func writePyMarker(t *testing.T, path, req, scope string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	body := "import pytest\n\n" +
		"@pytest.mark.req(\"" + req + "\")\n" +
		"@pytest.mark." + scope + "\n" +
		"@pytest.mark.technique_nominal\n" +
		"@pytest.mark.env_simulation\n" +
		"def test_fn():\n    pass\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestFromPytestNoRunDiscoversPackageMarkers is the Phoenix regression: with no
// path argument (the promote-from-junit invocation), from-pytest --no-run must
// still join a test under a hyphenated package dir by discovering its marker from
// the JUnit class name (REQ-LANG-031). The default "tests" path is absent here.
func TestFromPytestNoRunDiscoversPackageMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-031",
		rtmx.Scope("integration"), rtmx.Technique("nominal"), rtmx.Env("simulation"))
	dir := t.TempDir()
	writePyMarker(t, filepath.Join(dir, "packages", "sig-proc", "tests", "test_range.py"), "REQ-DSP-011", "scope_unit")
	junitPath := filepath.Join(dir, "j.xml")
	junit := `<?xml version="1.0"?><testsuite name="p"><testcase classname="packages.sig-proc.tests.test_range" name="test_fn" time="0.01"/></testsuite>`
	if err := os.WriteFile(junitPath, []byte(junit), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir) // discovery resolves class name -> path relative to cwd

	out := filepath.Join(dir, "results.json")
	withFromPytestGlobals(t, []string{junitPath}, out)

	if err := runFromPytest(newTestRootCmd(), nil); err != nil {
		t.Fatalf("runFromPytest: %v", err)
	}
	parsed := readResults(t, out)
	if len(parsed) != 1 || parsed[0].Marker.ReqID != "REQ-DSP-011" || parsed[0].Marker.Scope != "unit" {
		t.Fatalf("expected one REQ-DSP-011 (unit) result, got %+v", parsed)
	}
}

// TestFromPytestMultiFileJUnit covers ingesting several per-leg JUnit files in one
// call, each pointing at a different hyphenated package (REQ-LANG-030).
func TestFromPytestMultiFileJUnit(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-030",
		rtmx.Scope("integration"), rtmx.Technique("nominal"), rtmx.Env("simulation"))
	dir := t.TempDir()
	writePyMarker(t, filepath.Join(dir, "packages", "a-b", "tests", "test_1.py"), "REQ-X-001", "scope_unit")
	writePyMarker(t, filepath.Join(dir, "packages", "c-d", "tests", "test_2.py"), "REQ-Y-002", "scope_integration")
	j1 := filepath.Join(dir, "j1.xml")
	j2 := filepath.Join(dir, "j2.xml")
	if err := os.WriteFile(j1, []byte(`<?xml version="1.0"?><testsuite name="p"><testcase classname="packages.a-b.tests.test_1" name="test_fn"/></testsuite>`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(j2, []byte(`<?xml version="1.0"?><testsuite name="p"><testcase classname="packages.c-d.tests.test_2" name="test_fn"/></testsuite>`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	out := filepath.Join(dir, "results.json")
	withFromPytestGlobals(t, []string{j1, j2}, out)

	if err := runFromPytest(newTestRootCmd(), nil); err != nil {
		t.Fatalf("runFromPytest: %v", err)
	}
	parsed := readResults(t, out)
	if len(parsed) != 2 {
		t.Fatalf("expected 2 results from 2 junit files, got %d: %+v", len(parsed), parsed)
	}
	got := map[string]bool{}
	for _, r := range parsed {
		got[r.Marker.ReqID] = true
	}
	if !got["REQ-X-001"] || !got["REQ-Y-002"] {
		t.Fatalf("expected both REQ-X-001 and REQ-Y-002, got %v", got)
	}
}

func readResults(t *testing.T, path string) []results.Result {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read results: %v", err)
	}
	var parsed []results.Result
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid results JSON: %v", err)
	}
	return parsed
}
