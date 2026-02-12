package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	analyzeOutput string
	analyzeFormat string
	analyzeDeep   bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [PATH]",
	Short: "Analyze project for requirements artifacts",
	Long: `Discover tests, issues, documentation that could become requirements.

Examples:
    rtmx analyze                        # Analyze current directory
    rtmx analyze /path/to/project       # Analyze specific path
    rtmx analyze --format json          # JSON output
    rtmx analyze -o report.md --format markdown`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "", "output file for report")
	analyzeCmd.Flags().StringVar(&analyzeFormat, "format", "terminal", "output format: terminal, json, markdown")
	analyzeCmd.Flags().BoolVar(&analyzeDeep, "deep", false, "include source code analysis")

	rootCmd.AddCommand(analyzeCmd)
}

// AnalysisReport holds the analysis results.
type AnalysisReport struct {
	Path            string          `json:"path"`
	TestFiles       []TestFileInfo  `json:"test_files"`
	TotalTests      int             `json:"total_tests"`
	TestsWithMarker int             `json:"tests_with_marker"`
	UnmarkedTests   int             `json:"unmarked_tests"`
	GitHubConfigured bool           `json:"github_configured"`
	GitHubRepo      string          `json:"github_repo,omitempty"`
	JiraConfigured  bool            `json:"jira_configured"`
	JiraProject     string          `json:"jira_project,omitempty"`
	RTMExists       bool            `json:"rtm_exists"`
	RTMPath         string          `json:"rtm_path,omitempty"`
	Recommendations []string        `json:"recommendations"`
}

type TestFileInfo struct {
	Path       string `json:"path"`
	HasMarkers bool   `json:"has_markers"`
	MarkerCount int   `json:"marker_count"`
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Load config
	cfg, _ := config.LoadFromDir(absPath)

	// Run analysis
	report := analyzeProject(absPath, cfg)

	// Output
	var content string
	switch analyzeFormat {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		content = string(data)
	case "markdown":
		content = formatAnalysisMarkdown(report)
	default:
		return formatAnalysisTerminal(cmd, report)
	}

	if analyzeOutput != "" {
		if err := os.WriteFile(analyzeOutput, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		cmd.Printf("Report written to %s\n", analyzeOutput)
	} else {
		cmd.Print(content)
	}

	return nil
}

func analyzeProject(path string, cfg *config.Config) *AnalysisReport {
	report := &AnalysisReport{
		Path:            path,
		TestFiles:       []TestFileInfo{},
		Recommendations: []string{},
	}

	// Find test files
	testDirs := []string{
		filepath.Join(path, "tests"),
		filepath.Join(path, "test"),
	}

	for _, testDir := range testDirs {
		if info, err := os.Stat(testDir); err == nil && info.IsDir() {
			filepath.Walk(testDir, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && strings.HasPrefix(filepath.Base(p), "test_") && strings.HasSuffix(p, ".py") {
					// Analyze test file for markers
					markers := countMarkersInFile(p)
					relPath, _ := filepath.Rel(path, p)
					report.TestFiles = append(report.TestFiles, TestFileInfo{
						Path:        relPath,
						HasMarkers:  markers > 0,
						MarkerCount: markers,
					})
					report.TotalTests++
					if markers > 0 {
						report.TestsWithMarker++
					} else {
						report.UnmarkedTests++
					}
				}
				return nil
			})
		}
	}

	// Check GitHub adapter config
	if cfg != nil && cfg.RTMX.Adapters.GitHub.Enabled {
		report.GitHubConfigured = true
		report.GitHubRepo = cfg.RTMX.Adapters.GitHub.Repo
	}

	// Check Jira adapter config
	if cfg != nil && cfg.RTMX.Adapters.Jira.Enabled {
		report.JiraConfigured = true
		report.JiraProject = cfg.RTMX.Adapters.Jira.Project
	}

	// Check for existing RTM
	if cfg != nil {
		rtmPath := cfg.DatabasePath(path)
		if _, err := os.Stat(rtmPath); err == nil {
			report.RTMExists = true
			report.RTMPath = rtmPath
		}
	}

	// Generate recommendations
	if !report.RTMExists {
		report.Recommendations = append(report.Recommendations, "Run 'rtmx init' to create RTM structure")
	}
	if report.UnmarkedTests > 0 {
		report.Recommendations = append(report.Recommendations, "Run 'rtmx bootstrap --from-tests' to generate requirements from tests")
	}
	if report.GitHubConfigured && report.GitHubRepo != "" {
		report.Recommendations = append(report.Recommendations, "Run 'rtmx sync github --import' to import GitHub issues")
	}
	if report.JiraConfigured && report.JiraProject != "" {
		report.Recommendations = append(report.Recommendations, "Run 'rtmx sync jira --import' to import Jira tickets")
	}
	report.Recommendations = append(report.Recommendations, "Run 'rtmx install' to add RTM prompts to AI agent configs")

	return report
}

func countMarkersInFile(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	// Count @pytest.mark.req markers
	count := 0
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "@pytest.mark.req") || strings.Contains(line, "pytest.mark.req(") {
			count++
		}
	}
	return count
}

func formatAnalysisTerminal(cmd *cobra.Command, report *AnalysisReport) error {
	cmd.Println("=== RTMX Project Analysis ===")
	cmd.Println()
	cmd.Printf("Analyzing: %s\n", report.Path)
	cmd.Println()

	// Test files
	cmd.Printf("%s\n", output.Color("Test Files:", output.Bold))
	if len(report.TestFiles) > 0 {
		shown := 0
		for _, tf := range report.TestFiles {
			if shown >= 10 {
				break
			}
			if tf.HasMarkers {
				cmd.Printf("  %s %s (%d req markers)\n", output.Color("✓", output.Green), tf.Path, tf.MarkerCount)
			} else {
				cmd.Printf("  %s %s\n", output.Color("○", output.Yellow), tf.Path)
			}
			shown++
		}
		if len(report.TestFiles) > 10 {
			cmd.Printf("  ... and %d more test files\n", len(report.TestFiles)-10)
		}
		cmd.Println()
		cmd.Printf("  Tests without req markers: %d\n", report.UnmarkedTests)
	} else {
		cmd.Printf("  %s\n", output.Color("No test files found", output.Dim))
	}
	cmd.Println()

	// GitHub
	cmd.Printf("%s\n", output.Color("GitHub Issues:", output.Bold))
	if report.GitHubConfigured && report.GitHubRepo != "" {
		cmd.Printf("  Repository: %s\n", report.GitHubRepo)
		cmd.Printf("  %s\n", output.Color("(Run 'rtmx bootstrap --from-github' to import)", output.Yellow))
	} else {
		cmd.Printf("  %s\n", output.Color("Not configured (add adapters.github to rtmx.yaml)", output.Dim))
	}
	cmd.Println()

	// Jira
	cmd.Printf("%s\n", output.Color("Jira Tickets:", output.Bold))
	if report.JiraConfigured && report.JiraProject != "" {
		cmd.Printf("  Project: %s\n", report.JiraProject)
		cmd.Printf("  %s\n", output.Color("(Run 'rtmx bootstrap --from-jira' to import)", output.Yellow))
	} else {
		cmd.Printf("  %s\n", output.Color("Not configured (add adapters.jira to rtmx.yaml)", output.Dim))
	}
	cmd.Println()

	// Existing RTM
	cmd.Printf("%s\n", output.Color("Existing RTM:", output.Bold))
	if report.RTMExists {
		cmd.Printf("  %s Found at %s\n", output.Color("✓", output.Green), report.RTMPath)
	} else {
		cmd.Printf("  %s Not found (run 'rtmx init' to create)\n", output.Color("○", output.Yellow))
	}
	cmd.Println()

	// Recommendations
	cmd.Printf("%s\n", output.Color("Recommendations:", output.Bold))
	for i, rec := range report.Recommendations {
		cmd.Printf("  %d. %s\n", i+1, rec)
	}

	return nil
}

func formatAnalysisMarkdown(report *AnalysisReport) string {
	var sb strings.Builder

	sb.WriteString("# RTMX Project Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Path:** %s\n\n", report.Path))

	sb.WriteString("## Test Files\n\n")
	if len(report.TestFiles) > 0 {
		sb.WriteString("| File | Has Markers | Marker Count |\n")
		sb.WriteString("|------|-------------|-------------|\n")
		for _, tf := range report.TestFiles {
			hasMarkers := "No"
			if tf.HasMarkers {
				hasMarkers = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", tf.Path, hasMarkers, tf.MarkerCount))
		}
		sb.WriteString(fmt.Sprintf("\n**Tests without markers:** %d\n\n", report.UnmarkedTests))
	} else {
		sb.WriteString("No test files found.\n\n")
	}

	sb.WriteString("## Integrations\n\n")
	if report.GitHubConfigured {
		sb.WriteString(fmt.Sprintf("- **GitHub:** Configured (%s)\n", report.GitHubRepo))
	} else {
		sb.WriteString("- **GitHub:** Not configured\n")
	}
	if report.JiraConfigured {
		sb.WriteString(fmt.Sprintf("- **Jira:** Configured (%s)\n", report.JiraProject))
	} else {
		sb.WriteString("- **Jira:** Not configured\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## RTM Status\n\n")
	if report.RTMExists {
		sb.WriteString(fmt.Sprintf("RTM database found at `%s`\n\n", report.RTMPath))
	} else {
		sb.WriteString("RTM database not found.\n\n")
	}

	sb.WriteString("## Recommendations\n\n")
	for i, rec := range report.Recommendations {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}

	return sb.String()
}
