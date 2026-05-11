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
	Use:   "from-pytest [test_path]",
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
	testPath := "tests"
	if len(args) > 0 {
		testPath = args[0]
	}

	markers, err := scanPytestMarkers(testPath)
	if err != nil {
		return err
	}
	if len(markers) == 0 {
		return fmt.Errorf("no pytest requirement markers found under %s", testPath)
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
		if err := runPytestForJUnit(testPath, junitPath); err != nil {
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

func runPytestForJUnit(testPath, junitPath string) error {
	parts := strings.Fields(fromPytestCommand)
	if len(parts) == 0 {
		return fmt.Errorf("empty pytest command")
	}
	args := append([]string{}, parts[1:]...)
	args = append(args, testPath, "--junitxml", junitPath)
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
		tc, ok := caseIndex[marker.TestFunction]
		if !ok {
			continue
		}
		out = append(out, results.Result{
			Marker: results.Marker{
				ReqID:    marker.ReqID,
				TestName: marker.TestFunction,
				TestFile: filepath.ToSlash(marker.TestFile),
				Line:     marker.LineNumber,
			},
			Passed:   len(tc.Failures) == 0 && len(tc.Errors) == 0 && len(tc.Skipped) == 0,
			Duration: tc.Time * 1000,
		})
	}
	return out
}

func pytestCaseKeys(tc junitTestCase) []string {
	keys := []string{tc.Name}
	if tc.ClassName != "" {
		parts := strings.Split(tc.ClassName, ".")
		className := parts[len(parts)-1]
		if className != "" {
			keys = append(keys, className+"::"+tc.Name)
		}
	}
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
