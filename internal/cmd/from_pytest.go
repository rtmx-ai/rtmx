package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/results"
	"github.com/spf13/cobra"
)

var (
	fromPytestCommand string
	fromPytestJUnit   string
	fromPytestOutput  string
	fromPytestNoRun   bool
)

var fromPytestCmd = &cobra.Command{
	Use:   "from-pytest [test_path...]",
	Short: "Generate RTMX results JSON from pytest markers and JUnit XML",
	Long: `Generate RTMX results JSON from pytest tests marked with
@pytest.mark.req("REQ-...").

By default this command runs pytest with a temporary --junitxml file, scans the
test tree for RTMX requirement markers, joins pytest results back to those
markers, and writes language-agnostic RTMX results JSON. The output can then be
used directly with:

  rtmx verify --results .rtmx/cache/pytest-results.json --update

Use --junitxml with --no-run to convert an existing pytest JUnit XML file.`,
	RunE: runFromPytest,
}

func init() {
	fromPytestCmd.Flags().StringVar(&fromPytestCommand, "command", "pytest", "pytest command to run")
	fromPytestCmd.Flags().StringVar(&fromPytestJUnit, "junitxml", "", "existing or generated pytest JUnit XML path")
	fromPytestCmd.Flags().StringVarP(&fromPytestOutput, "output", "o", ".rtmx/cache/pytest-results.json", "RTMX results JSON output path")
	fromPytestCmd.Flags().BoolVar(&fromPytestNoRun, "no-run", false, "do not run pytest; read --junitxml instead")

	rootCmd.AddCommand(fromPytestCmd)
}

type junitTestSuites struct {
	TestCases []junitTestCase  `xml:"testcase"`
	Suites    []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	TestCases []junitTestCase  `xml:"testcase"`
	Suites    []junitTestSuite `xml:"testsuite"`
}

type junitTestCase struct {
	ClassName string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	File      string        `xml:"file,attr"`
	Line      int           `xml:"line,attr"`
	Time      float64       `xml:"time,attr"`
	Failures  []interface{} `xml:"failure"`
	Errors    []interface{} `xml:"error"`
	Skipped   []interface{} `xml:"skipped"`
}

func runFromPytest(cmd *cobra.Command, args []string) error {
	// Accept one or more test paths so multi-package layouts (e.g. a top-level
	// tests/ plus packages/*/tests/) are scanned and run together; default to
	// "tests" when none are given.
	testPaths := []string{"tests"}
	if len(args) > 0 {
		testPaths = args
	}

	var markers []TestRequirement
	for _, p := range testPaths {
		m, err := scanPytestMarkers(p)
		if err != nil {
			return err
		}
		markers = append(markers, m...)
	}
	if len(markers) == 0 {
		return fmt.Errorf("no pytest requirement markers found under %s", strings.Join(testPaths, ", "))
	}

	junitPath := fromPytestJUnit
	cleanup := func() {}
	if junitPath == "" {
		tmp, err := os.CreateTemp("", "rtmx-pytest-*.xml")
		if err != nil {
			return fmt.Errorf("failed to create temporary JUnit file: %w", err)
		}
		junitPath = tmp.Name()
		_ = tmp.Close()
		cleanup = func() { _ = os.Remove(junitPath) }
	}
	defer cleanup()

	if !fromPytestNoRun {
		if err := runPytestForJUnit(testPaths, junitPath); err != nil {
			cmd.Printf("! pytest exited with error: %v\n", err)
		}
	}

	cases, err := parsePytestJUnit(junitPath)
	if err != nil {
		return err
	}

	rtmxResults := buildPytestRTMXResults(markers, cases)
	if len(rtmxResults) == 0 {
		return fmt.Errorf("no pytest results matched RTMX requirement markers")
	}

	if err := writeRTMXResults(fromPytestOutput, rtmxResults); err != nil {
		return err
	}

	cmd.Printf("Wrote %d RTMX pytest result(s) to %s\n", len(rtmxResults), fromPytestOutput)
	return nil
}

func scanPytestMarkers(testPath string) ([]TestRequirement, error) {
	info, err := os.Stat(testPath)
	if err != nil {
		return nil, fmt.Errorf("test path does not exist: %s", testPath)
	}
	if info.IsDir() {
		return scanTestDirectory(testPath)
	}
	return extractMarkersFromSingleFile(testPath)
}

func runPytestForJUnit(testPaths []string, junitPath string) error {
	parts := strings.Fields(fromPytestCommand)
	if len(parts) == 0 {
		return fmt.Errorf("empty pytest command")
	}
	args := append([]string{}, parts[1:]...)
	args = append(args, testPaths...)
	args = append(args, "--junitxml", junitPath)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir, _ = os.Getwd()
	return cmd.Run()
}

func parsePytestJUnit(path string) ([]junitTestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pytest JUnit XML: %w", err)
	}

	var suites junitTestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		var suite junitTestSuite
		if err2 := xml.Unmarshal(data, &suite); err2 != nil {
			return nil, fmt.Errorf("failed to parse pytest JUnit XML: %w", err)
		}
		return flattenJUnitSuite(suite), nil
	}

	var cases []junitTestCase
	cases = append(cases, suites.TestCases...)
	for _, suite := range suites.Suites {
		cases = append(cases, flattenJUnitSuite(suite)...)
	}
	return cases, nil
}

func flattenJUnitSuite(suite junitTestSuite) []junitTestCase {
	cases := append([]junitTestCase{}, suite.TestCases...)
	for _, nested := range suite.Suites {
		cases = append(cases, flattenJUnitSuite(nested)...)
	}
	return cases
}

func buildPytestRTMXResults(markers []TestRequirement, cases []junitTestCase) []results.Result {
	caseIndex := make(map[string]junitTestCase)
	for _, tc := range cases {
		for _, key := range pytestCaseKeys(tc) {
			caseIndex[key] = tc
		}
	}

	var out []results.Result
	for _, marker := range markers {
		var tc junitTestCase
		ok := false
		for _, k := range markerJoinKeys(marker) {
			if c, found := caseIndex[k]; found {
				tc, ok = c, true
				break
			}
		}
		if !ok {
			continue
		}
		// A skipped (or xfail, which pytest reports as skipped in JUnit) test is
		// not evidence: it must neither promote nor downgrade a requirement, so
		// omit it entirely rather than recording it as a failure.
		if len(tc.Skipped) > 0 {
			continue
		}
		scope, technique, env := pytestMarkerDimensions(marker.Markers)
		out = append(out, results.Result{
			Marker: results.Marker{
				ReqID:     marker.ReqID,
				Scope:     scope,
				Technique: technique,
				Env:       env,
				TestName:  marker.TestFunction,
				TestFile:  filepath.ToSlash(marker.TestFile),
				Line:      marker.LineNumber,
			},
			Passed:   len(tc.Failures) == 0 && len(tc.Errors) == 0,
			Duration: tc.Time * 1000,
		})
	}
	return out
}

// pytestMarkerDimensions maps the scope_*/technique_*/env_* pytest markers
// collected for a test into the results schema's scope/technique/env dimensions
// (e.g. "scope_unit" -> "unit", "env_static_field" -> "static_field"). Without
// this, the JUnit-based pytest path produced dimensionless results, so projects
// using a multi-dimensional completeness policy could never reach COMPLETE.
func pytestMarkerDimensions(markers []string) (scope, technique, env string) {
	for _, m := range markers {
		switch {
		case strings.HasPrefix(m, "scope_"):
			scope = strings.TrimPrefix(m, "scope_")
		case strings.HasPrefix(m, "technique_"):
			technique = strings.TrimPrefix(m, "technique_")
		case strings.HasPrefix(m, "env_"):
			env = strings.TrimPrefix(m, "env_")
		}
	}
	return scope, technique, env
}

func pytestCaseKeys(tc junitTestCase) []string {
	keys := []string{tc.Name}
	var className string
	if tc.ClassName != "" {
		parts := strings.Split(tc.ClassName, ".")
		className = parts[len(parts)-1]
		if className != "" {
			keys = append(keys, className+"::"+tc.Name)
		}
	}
	// Path-qualified keys disambiguate tests that share a function name across
	// files (common once a multi-package tree is scanned), preventing a passing
	// test in one file from joining a same-named test in another. The JUnit
	// `file` attribute and the scanner's TestFile are both repo-relative paths.
	if p := pyPathKey(tc.File); p != "" {
		keys = append(keys, p+"::"+tc.Name)
		if className != "" {
			keys = append(keys, p+"::"+className+"::"+tc.Name)
		}
	}
	return keys
}

// pyPathKey normalizes a test file path (slash-separated, cleaned) for use in
// join keys so the JUnit `file` attribute and the scanner's TestFile agree.
func pyPathKey(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

// markerJoinKeys returns the candidate caseIndex keys for a scanned marker,
// most-specific (path-qualified) first, so a join prefers an exact file match
// and only falls back to the bare function name when JUnit carries no file.
func markerJoinKeys(m TestRequirement) []string {
	keys := []string{}
	if p := pyPathKey(m.TestFile); p != "" {
		keys = append(keys, p+"::"+m.TestFunction)
	}
	keys = append(keys, m.TestFunction)
	return keys
}

func writeRTMXResults(path string, rtmxResults []results.Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}
	data, err := json.MarshalIndent(rtmxResults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode RTMX results: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write RTMX results: %w", err)
	}
	return nil
}
