package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	migrateFix   bool
	migrateCheck bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Check and fix project compatibility for Python-to-Go migration",
	Long: `Validate that an RTMX project is compatible with the Go CLI.

By default, runs in --check mode which reports issues without making changes.
Use --fix to automatically resolve compatibility issues.

Checks performed:
  - Config file exists (.rtmx/config.yaml or rtmx.yaml)
  - Database file exists at configured path
  - Legacy database path (docs/rtm_database.csv) detection
  - Database schema validates (21 standard columns)
  - Requirements directory exists
  - Git hooks reference correct binary

Examples:
  rtmx migrate              # Check mode (default)
  rtmx migrate --check      # Explicit check mode
  rtmx migrate --fix        # Auto-fix issues`,
	RunE: runMigrate,
}

func init() {
	migrateCmd.Flags().BoolVar(&migrateFix, "fix", false, "auto-fix compatibility issues")
	migrateCmd.Flags().BoolVar(&migrateCheck, "check", false, "validate compatibility without changes (default)")
	rootCmd.AddCommand(migrateCmd)
}

// migrateCheckResult represents the result of a single migration check.
type migrateCheckResult struct {
	Name   string
	Status string // PASS, WARN, FAIL
	Detail string
}

func runMigrate(cmd *cobra.Command, _ []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	results := []migrateCheckResult{}
	fixActions := []string{}

	// Check 1: Config file exists
	configResult := checkConfigFile(wd)
	results = append(results, configResult)

	// Check 2: Database file exists
	dbResult, legacyPath := checkDatabaseFile(wd)
	results = append(results, dbResult)

	// Check 3: Legacy database path
	legacyResult := checkLegacyDatabase(wd)
	results = append(results, legacyResult)

	// Check 4: Schema validates (21 columns)
	schemaResult := checkDatabaseSchema(wd, legacyPath)
	results = append(results, schemaResult)

	// Check 5: Requirements directory exists
	reqDirResult := checkRequirementsDir(wd)
	results = append(results, reqDirResult)

	// Check 6: Git hooks reference correct binary
	hookResult := checkGitHooks(wd)
	results = append(results, hookResult)

	// Apply fixes if --fix mode
	if migrateFix {
		fixActions = applyFixes(cmd, wd, results, legacyPath)
	}

	// Print results
	cmd.Println("Migration Compatibility Report")
	cmd.Println("==============================")
	cmd.Println("")

	passCount := 0
	warnCount := 0
	failCount := 0

	for _, r := range results {
		cmd.Printf("  [%s] %s", r.Status, r.Name)
		if r.Detail != "" {
			cmd.Printf(": %s", r.Detail)
		}
		cmd.Println("")

		switch r.Status {
		case "PASS":
			passCount++
		case "WARN":
			warnCount++
		case "FAIL":
			failCount++
		}
	}

	cmd.Println("")
	cmd.Printf("Summary: %d passed, %d warnings, %d failed\n", passCount, warnCount, failCount)

	if len(fixActions) > 0 {
		cmd.Println("")
		cmd.Println("Fix actions applied:")
		for _, action := range fixActions {
			cmd.Printf("  - %s\n", action)
		}
	}

	if failCount > 0 && !migrateFix {
		cmd.Println("")
		cmd.Println("Run 'rtmx migrate --fix' to auto-fix issues.")
	}

	if failCount > 0 && !migrateFix {
		return NewExitError(1, "migration check found failures")
	}

	return nil
}

func checkConfigFile(wd string) migrateCheckResult {
	candidates := []string{
		filepath.Join(wd, ".rtmx", "config.yaml"),
		filepath.Join(wd, "rtmx.yaml"),
		filepath.Join(wd, "rtmx.yml"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return migrateCheckResult{
				Name:   "Config file",
				Status: "PASS",
				Detail: filepath.Base(path),
			}
		}
	}

	return migrateCheckResult{
		Name:   "Config file",
		Status: "FAIL",
		Detail: "no config file found (.rtmx/config.yaml or rtmx.yaml)",
	}
}

func checkDatabaseFile(wd string) (migrateCheckResult, string) {
	// Check modern path first
	modernPath := filepath.Join(wd, ".rtmx", "database.csv")
	if _, err := os.Stat(modernPath); err == nil {
		return migrateCheckResult{
			Name:   "Database file",
			Status: "PASS",
			Detail: ".rtmx/database.csv",
		}, modernPath
	}

	// Check legacy path
	legacyPath := filepath.Join(wd, "docs", "rtm_database.csv")
	if _, err := os.Stat(legacyPath); err == nil {
		return migrateCheckResult{
			Name:   "Database file",
			Status: "WARN",
			Detail: "found at legacy path docs/rtm_database.csv",
		}, legacyPath
	}

	return migrateCheckResult{
		Name:   "Database file",
		Status: "FAIL",
		Detail: "no database file found",
	}, ""
}

func checkLegacyDatabase(wd string) migrateCheckResult {
	legacyPath := filepath.Join(wd, "docs", "rtm_database.csv")
	modernPath := filepath.Join(wd, ".rtmx", "database.csv")

	_, legacyErr := os.Stat(legacyPath)
	_, modernErr := os.Stat(modernPath)

	if legacyErr == nil && modernErr != nil {
		return migrateCheckResult{
			Name:   "Legacy database path",
			Status: "WARN",
			Detail: "docs/rtm_database.csv should be moved to .rtmx/database.csv",
		}
	}

	if legacyErr == nil && modernErr == nil {
		return migrateCheckResult{
			Name:   "Legacy database path",
			Status: "WARN",
			Detail: "both legacy and modern paths exist",
		}
	}

	return migrateCheckResult{
		Name:   "Legacy database path",
		Status: "PASS",
		Detail: "no legacy database found",
	}
}

func checkDatabaseSchema(wd string, dbPath string) migrateCheckResult {
	if dbPath == "" {
		// Try to find database
		candidates := []string{
			filepath.Join(wd, ".rtmx", "database.csv"),
			filepath.Join(wd, "docs", "rtm_database.csv"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				dbPath = c
				break
			}
		}
	}

	if dbPath == "" {
		return migrateCheckResult{
			Name:   "Database schema",
			Status: "FAIL",
			Detail: "no database file to validate",
		}
	}

	file, err := os.Open(dbPath)
	if err != nil {
		return migrateCheckResult{
			Name:   "Database schema",
			Status: "FAIL",
			Detail: fmt.Sprintf("cannot open database: %v", err),
		}
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return migrateCheckResult{
			Name:   "Database schema",
			Status: "FAIL",
			Detail: fmt.Sprintf("cannot read CSV header: %v", err),
		}
	}

	expectedColumns := 21
	if len(header) == expectedColumns {
		return migrateCheckResult{
			Name:   "Database schema",
			Status: "PASS",
			Detail: fmt.Sprintf("%d columns validated", len(header)),
		}
	}

	return migrateCheckResult{
		Name:   "Database schema",
		Status: "FAIL",
		Detail: fmt.Sprintf("expected %d columns, found %d", expectedColumns, len(header)),
	}
}

func checkRequirementsDir(wd string) migrateCheckResult {
	reqDir := filepath.Join(wd, ".rtmx", "requirements")
	if info, err := os.Stat(reqDir); err == nil && info.IsDir() {
		return migrateCheckResult{
			Name:   "Requirements directory",
			Status: "PASS",
			Detail: ".rtmx/requirements/",
		}
	}

	return migrateCheckResult{
		Name:   "Requirements directory",
		Status: "WARN",
		Detail: ".rtmx/requirements/ not found",
	}
}

func checkGitHooks(wd string) migrateCheckResult {
	hookPaths := []string{
		filepath.Join(wd, ".githooks", "pre-commit"),
		filepath.Join(wd, ".git", "hooks", "pre-commit"),
	}

	for _, hookPath := range hookPaths {
		data, err := os.ReadFile(hookPath)
		if err != nil {
			continue
		}

		content := string(data)
		if strings.Contains(content, "python") || strings.Contains(content, "pip") || strings.Contains(content, "pytest") {
			return migrateCheckResult{
				Name:   "Git hooks",
				Status: "WARN",
				Detail: "hooks reference Python tooling",
			}
		}

		if strings.Contains(content, "rtmx") || strings.Contains(content, "make") {
			return migrateCheckResult{
				Name:   "Git hooks",
				Status: "PASS",
				Detail: "hooks are compatible",
			}
		}

		return migrateCheckResult{
			Name:   "Git hooks",
			Status: "PASS",
			Detail: "hooks found",
		}
	}

	return migrateCheckResult{
		Name:   "Git hooks",
		Status: "PASS",
		Detail: "no pre-commit hooks found",
	}
}

func applyFixes(cmd *cobra.Command, wd string, results []migrateCheckResult, dbPath string) []string {
	actions := []string{}

	for _, r := range results {
		switch {
		case r.Name == "Legacy database path" && (r.Status == "WARN"):
			legacyPath := filepath.Join(wd, "docs", "rtm_database.csv")
			modernPath := filepath.Join(wd, ".rtmx", "database.csv")

			// Only move if modern path doesn't already exist
			if _, err := os.Stat(modernPath); err != nil {
				if _, err := os.Stat(legacyPath); err == nil {
					// Create backup
					backupPath := legacyPath + ".bak." + time.Now().Format("20060102-150405")
					if err := copyFile(legacyPath, backupPath); err != nil {
						cmd.Printf("Warning: failed to create backup: %v\n", err)
						continue
					}
					actions = append(actions, fmt.Sprintf("created backup: %s", filepath.Base(backupPath)))

					// Ensure target directory exists
					if err := os.MkdirAll(filepath.Join(wd, ".rtmx"), 0755); err != nil {
						cmd.Printf("Warning: failed to create .rtmx directory: %v\n", err)
						continue
					}

					// Move file
					if err := os.Rename(legacyPath, modernPath); err != nil {
						cmd.Printf("Warning: failed to move database: %v\n", err)
						continue
					}
					actions = append(actions, "moved docs/rtm_database.csv -> .rtmx/database.csv")
				}
			}

		case r.Name == "Git hooks" && r.Status == "WARN":
			hookPaths := []string{
				filepath.Join(wd, ".githooks", "pre-commit"),
				filepath.Join(wd, ".git", "hooks", "pre-commit"),
			}

			for _, hookPath := range hookPaths {
				data, err := os.ReadFile(hookPath)
				if err != nil {
					continue
				}

				content := string(data)
				if strings.Contains(content, "python") || strings.Contains(content, "pytest") {
					// Create backup
					backupPath := hookPath + ".bak." + time.Now().Format("20060102-150405")
					if err := copyFile(hookPath, backupPath); err != nil {
						cmd.Printf("Warning: failed to backup hook: %v\n", err)
						continue
					}
					actions = append(actions, fmt.Sprintf("created hook backup: %s", filepath.Base(backupPath)))

					// Replace python references with Go equivalents
					updated := strings.ReplaceAll(content, "python -m rtmx", "rtmx")
					updated = strings.ReplaceAll(updated, "pytest", "go test ./...")
					if err := os.WriteFile(hookPath, []byte(updated), 0755); err != nil {
						cmd.Printf("Warning: failed to update hook: %v\n", err)
						continue
					}
					actions = append(actions, fmt.Sprintf("updated %s", filepath.Base(hookPath)))
				}
			}

		case r.Name == "Requirements directory" && r.Status == "WARN":
			reqDir := filepath.Join(wd, ".rtmx", "requirements")
			if err := os.MkdirAll(reqDir, 0755); err != nil {
				cmd.Printf("Warning: failed to create requirements dir: %v\n", err)
				continue
			}
			actions = append(actions, "created .rtmx/requirements/")
		}
	}

	// Suppress unused variable warning for dbPath
	_ = dbPath

	return actions
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
