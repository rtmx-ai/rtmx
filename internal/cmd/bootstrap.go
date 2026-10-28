package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/adapters"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	bootstrapFromTests  bool
	bootstrapFromGitHub bool
	bootstrapFromJira   bool
	bootstrapFromAsana  bool
	bootstrapFromMonday bool
	bootstrapFromGitLab bool
	bootstrapMerge      bool
	bootstrapDryRun     bool
	bootstrapPrefix     string
)

// bootstrapGetAdapter resolves a ServiceAdapter for bootstrap imports.
// Overridable in tests (same shape as sync getAdapter).
var bootstrapGetAdapter = getAdapter

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Generate initial RTM from project artifacts",
	Long: `Bootstrap requirements from existing tests or project-management adapters.

With no --from-* flag on a TTY, opens the same interactive provider picker as
rtmx link. Non-TTY / CI must pass an explicit --from-* flag.

Examples:
    rtmx bootstrap                     # TTY: interactive provider picker
    rtmx bootstrap --from-tests        # Generate from test markers
    rtmx bootstrap --from-github       # Import from GitHub issues
    rtmx bootstrap --from-jira         # Import from Jira tickets
    rtmx bootstrap --from-asana        # Import from Asana tasks
    rtmx bootstrap --from-monday       # Import from Monday.com boards
    rtmx bootstrap --from-gitlab       # Import from GitLab issues
    rtmx bootstrap --from-tests --merge  # Merge with existing RTM`,
	RunE: runBootstrap,
}

func init() {
	bootstrapCmd.Flags().BoolVar(&bootstrapFromTests, "from-tests", false, "generate requirements from test functions")
	bootstrapCmd.Flags().BoolVar(&bootstrapFromGitHub, "from-github", false, "import from GitHub issues")
	bootstrapCmd.Flags().BoolVar(&bootstrapFromJira, "from-jira", false, "import from Jira tickets")
	bootstrapCmd.Flags().BoolVar(&bootstrapFromAsana, "from-asana", false, "import from Asana tasks")
	bootstrapCmd.Flags().BoolVar(&bootstrapFromMonday, "from-monday", false, "import from Monday.com board items")
	bootstrapCmd.Flags().BoolVar(&bootstrapFromGitLab, "from-gitlab", false, "import from GitLab issues")
	bootstrapCmd.Flags().BoolVar(&bootstrapMerge, "merge", false, "merge with existing RTM (default: replace)")
	bootstrapCmd.Flags().BoolVar(&bootstrapDryRun, "dry-run", false, "preview without writing files")
	bootstrapCmd.Flags().StringVar(&bootstrapPrefix, "prefix", "REQ", "requirement ID prefix")

	rootCmd.AddCommand(bootstrapCmd)
}

// BootstrapRequirement represents a requirement discovered during bootstrap.
type BootstrapRequirement struct {
	ID          string
	Category    string
	Subcategory string
	Text        string
	TestModule  string
	TestFunc    string
	Source      string // "test", "github", "jira", "asana", "monday", "gitlab"
	ExternalID  string // GitHub issue number or Jira ticket key
}

func bootstrapSourceSelected() bool {
	return bootstrapFromTests ||
		bootstrapFromGitHub ||
		bootstrapFromJira ||
		bootstrapFromAsana ||
		bootstrapFromMonday ||
		bootstrapFromGitLab
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if !bootstrapSourceSelected() {
		if isTerminal() {
			if err := selectBootstrapSourceInteractive(cmd); err != nil {
				return err
			}
			if !bootstrapSourceSelected() {
				cmd.Println("Skipped linking an issue store.")
				return nil
			}
		} else {
			cmd.Printf("%s\n", output.Color("No source specified. Use --from-tests, --from-github, --from-jira, --from-asana, --from-monday, or --from-gitlab (or run rtmx link on a TTY)", output.Yellow))
			return NewExitError(1, "no source specified")
		}
	}

	cmd.Println("=== RTMX Bootstrap ===")
	cmd.Println()

	if bootstrapDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		// Use defaults if no config
		cfg = &config.Config{}
		cfg.RTMX.Database = ".rtmx/database.csv"
	}

	var requirements []BootstrapRequirement

	// Bootstrap from tests
	if bootstrapFromTests {
		cmd.Printf("%s\n", output.Color("Scanning tests...", output.Bold))
		testReqs := bootstrapFromTestFiles(cwd, bootstrapPrefix)
		requirements = append(requirements, testReqs...)
		cmd.Printf("  Found %d test functions with markers\n", len(testReqs))
		cmd.Println()
	}

	adapterSources := []struct {
		enabled bool
		service string
		label   string
	}{
		{bootstrapFromGitHub, "github", "GitHub issues"},
		{bootstrapFromJira, "jira", "Jira tickets"},
		{bootstrapFromAsana, "asana", "Asana tasks"},
		{bootstrapFromMonday, "monday", "Monday.com items"},
		{bootstrapFromGitLab, "gitlab", "GitLab issues"},
	}

	for _, src := range adapterSources {
		if !src.enabled {
			continue
		}
		cmd.Printf("%s\n", output.Color(fmt.Sprintf("Fetching %s...", src.label), output.Bold))
		adapterReqs, fetchErr := bootstrapFromService(cmd, src.service, cfg, bootstrapPrefix)
		if fetchErr != nil {
			return fetchErr
		}
		requirements = append(requirements, adapterReqs...)
		cmd.Printf("  Found %d items\n", len(adapterReqs))
		cmd.Println()
	}

	// Display discovered requirements
	if len(requirements) > 0 {
		cmd.Printf("%s\n", output.Color("Requirements to create:", output.Bold))
		for _, req := range requirements {
			cmd.Printf("  %s: %s\n", req.ID, truncateString(req.Text, 60))
		}
		cmd.Println()

		if !bootstrapDryRun {
			// Write requirements to RTM database
			err := writeBootstrapRequirements(cwd, cfg, requirements, bootstrapMerge)
			if err != nil {
				return fmt.Errorf("failed to write requirements: %w", err)
			}
			cmd.Printf("%s Created %d requirements\n", output.Color("✓", output.Green), len(requirements))
		}
	} else {
		cmd.Printf("%s\n", output.Color("No requirements to create", output.Dim))
	}

	return nil
}

// bootstrapFromService fetches ExternalItems via the adapter and maps them to BootstrapRequirement.
func bootstrapFromService(cmd *cobra.Command, service string, cfg *config.Config, prefix string) ([]BootstrapRequirement, error) {
	adapter, err := bootstrapGetAdapter(service, cfg)
	if err != nil {
		cmd.Printf("  %s\n", output.Color(err.Error(), output.Red))
		return nil, NewExitError(1, err.Error())
	}
	if !adapter.IsConfigured() {
		msg := fmt.Sprintf("%s adapter is enabled but not configured (check credentials / rtmx.yaml)", service)
		cmd.Printf("  %s\n", output.Color(msg, output.Red))
		return nil, NewExitError(1, msg)
	}

	items, err := adapter.FetchItems(nil)
	if err != nil {
		msg := fmt.Sprintf("failed to fetch from %s: %v", service, err)
		cmd.Printf("  %s\n", output.Color(msg, output.Red))
		return nil, NewExitError(1, msg)
	}

	return mapExternalItemsToBootstrap(items, service, prefix), nil
}

// mapExternalItemsToBootstrap converts adapter items into bootstrap rows.
func mapExternalItemsToBootstrap(items []adapters.ExternalItem, service, prefix string) []BootstrapRequirement {
	requirements := make([]BootstrapRequirement, 0, len(items))
	reqCounter := 0
	category := strings.ToUpper(service)

	for _, item := range items {
		reqID := item.RequirementID
		if reqID == "" {
			reqID = adapters.ExtractReqID(item.Title)
		}
		if reqID == "" {
			reqID = adapters.ExtractReqID(item.Description)
		}
		if reqID == "" {
			reqCounter++
			reqID = fmt.Sprintf("%s-%s-%03d", prefix, category, reqCounter)
		}

		text := strings.TrimSpace(item.Title)
		if text == "" {
			text = strings.TrimSpace(item.Description)
		}
		if text == "" {
			text = fmt.Sprintf("Imported from %s item %s", service, item.ExternalID)
		}

		requirements = append(requirements, BootstrapRequirement{
			ID:         reqID,
			Category:   category,
			Text:       text,
			Source:     service,
			ExternalID: item.ExternalID,
		})
	}
	return requirements
}

func bootstrapFromTestFiles(cwd string, prefix string) []BootstrapRequirement {
	var requirements []BootstrapRequirement

	testDirs := []string{
		filepath.Join(cwd, "tests"),
		filepath.Join(cwd, "test"),
	}

	// Regex to match @pytest.mark.req("REQ-XXX-NNN") or similar
	reqMarkerRe := regexp.MustCompile(`@pytest\.mark\.req\s*\(\s*["']([^"']+)["']\s*\)`)
	funcNameRe := regexp.MustCompile(`def\s+(test_\w+)\s*\(`)

	reqCounter := make(map[string]int)

	for _, testDir := range testDirs {
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			continue
		}

		_ = filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			if !strings.HasPrefix(filepath.Base(path), "test_") || !strings.HasSuffix(path, ".py") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			relPath, _ := filepath.Rel(cwd, path)
			lines := strings.Split(string(content), "\n")

			var currentMarkers []string
			for i, line := range lines {
				// Check for requirement marker
				if matches := reqMarkerRe.FindStringSubmatch(line); len(matches) > 1 {
					currentMarkers = append(currentMarkers, matches[1])
				}

				// Check for function definition
				if matches := funcNameRe.FindStringSubmatch(line); len(matches) > 1 {
					funcName := matches[1]

					if len(currentMarkers) == 0 {
						// Test without markers - create a new requirement
						category := inferCategoryFromPath(relPath)
						reqCounter[category]++
						reqID := fmt.Sprintf("%s-%s-%03d", prefix, category, reqCounter[category])

						// Try to extract docstring for requirement text
						text := inferRequirementText(lines, i, funcName)

						requirements = append(requirements, BootstrapRequirement{
							ID:          reqID,
							Category:    category,
							Subcategory: "",
							Text:        text,
							TestModule:  relPath,
							TestFunc:    funcName,
							Source:      "test",
						})
					}

					currentMarkers = nil // Reset markers for next function
				}
			}

			return nil
		})
	}

	return requirements
}

func inferCategoryFromPath(path string) string {
	// Extract category from test file path
	// e.g., tests/test_models.py -> MODELS
	// e.g., tests/cli/test_status.py -> CLI

	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// Remove "test_" prefix and ".py" suffix
	name := strings.TrimPrefix(base, "test_")
	name = strings.TrimSuffix(name, ".py")

	// Check if in subdirectory
	parts := strings.Split(dir, string(filepath.Separator))
	for _, part := range parts {
		if part != "tests" && part != "test" && part != "" {
			return strings.ToUpper(part)
		}
	}

	return strings.ToUpper(name)
}

func inferRequirementText(lines []string, funcLine int, funcName string) string {
	// Try to extract docstring from the function
	if funcLine+1 < len(lines) {
		nextLine := strings.TrimSpace(lines[funcLine+1])
		if strings.HasPrefix(nextLine, `"""`) || strings.HasPrefix(nextLine, `'''`) {
			// Found docstring
			doc := strings.TrimPrefix(nextLine, `"""`)
			doc = strings.TrimPrefix(doc, `'''`)
			doc = strings.TrimSuffix(doc, `"""`)
			doc = strings.TrimSuffix(doc, `'''`)
			if doc != "" {
				return doc
			}
		}
	}

	// Convert function name to requirement text
	// test_user_can_login -> User can login
	text := strings.TrimPrefix(funcName, "test_")
	text = strings.ReplaceAll(text, "_", " ")
	if len(text) > 0 {
		text = strings.ToUpper(text[:1]) + text[1:]
	}
	return text
}

func writeBootstrapRequirements(cwd string, cfg *config.Config, requirements []BootstrapRequirement, merge bool) error {
	dbPath := cfg.DatabasePath(cwd)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var existingContent string
	if merge {
		if content, err := os.ReadFile(dbPath); err == nil {
			existingContent = string(content)
		}
	}

	var sb strings.Builder

	// Write header if new file or not merging
	if existingContent == "" {
		sb.WriteString("req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n")
	} else {
		sb.WriteString(existingContent)
		if !strings.HasSuffix(existingContent, "\n") {
			sb.WriteString("\n")
		}
	}

	// Write new requirements
	for _, req := range requirements {
		// CSV escape values
		text := strings.ReplaceAll(req.Text, `"`, `""`)
		if strings.Contains(text, ",") || strings.Contains(text, `"`) || strings.Contains(text, "\n") {
			text = `"` + text + `"`
		}

		reqFile := fmt.Sprintf(".rtmx/requirements/%s/%s.md", req.Category, req.ID)

		fmt.Fprintf(&sb, "%s,%s,%s,%s,,%s,%s,Unit Test,MISSING,MEDIUM,1,Bootstrap generated,0.5,,,,,,,%s,%s\n",
			req.ID, req.Category, req.Subcategory, text, req.TestModule, req.TestFunc, reqFile, req.ExternalID)
	}

	return os.WriteFile(dbPath, []byte(sb.String()), 0644)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
