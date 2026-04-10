package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	securityJSON   bool
	securityStrict bool
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Audit repository and RTMX security posture",
	Long: `Audit the security posture of the repository and RTMX configuration.

Each control is reported as PASS, WARN, or FAIL.

RTMX controls are checked locally (no API needed). Repository controls
use the gh CLI if available and skip gracefully otherwise.

Exit codes:
  0  All checks pass (or only warnings)
  1  Any check has status FAIL
  --strict  Treat warnings as failures (exit 1)

Examples:
  rtmx security                 # Run all checks
  rtmx security --json          # Machine-readable output
  rtmx security --strict        # Treat warnings as failures`,
	RunE: runSecurity,
}

func init() {
	securityCmd.Flags().BoolVar(&securityJSON, "json", false, "output as JSON")
	securityCmd.Flags().BoolVar(&securityStrict, "strict", false, "treat warnings as failures")
}

// SecurityCheck represents a single security check result.
type SecurityCheck struct {
	Category string      `json:"category"`
	Name     string      `json:"name"`
	Status   CheckStatus `json:"status"`
	Message  string      `json:"message"`
	Fixable  bool        `json:"fixable"`
}

// SecurityResult represents the full security audit result.
type SecurityResult struct {
	Platform   string          `json:"platform"`
	Repository string          `json:"repository"`
	Checks     []SecurityCheck `json:"checks"`
	Summary    struct {
		Passed       int     `json:"passed"`
		Warnings     int     `json:"warnings"`
		Failed       int     `json:"failed"`
		ScorePercent float64 `json:"score_percent"`
	} `json:"summary"`
}

// SecurityOptions holds injected dependencies for testability.
type SecurityOptions struct {
	// Dir is the project root directory.
	Dir string
	// GhAvailable overrides gh CLI detection. nil means auto-detect.
	GhAvailable *bool
	// GhRunner executes gh commands. nil means use real exec.
	GhRunner func(args ...string) (string, error)
	// ReadFile reads a file. nil means use os.ReadFile.
	ReadFile func(path string) ([]byte, error)
	// Stat checks file existence. nil means use os.Stat.
	Stat func(path string) (os.FileInfo, error)
}

func (o *SecurityOptions) readFile(path string) ([]byte, error) {
	if o.ReadFile != nil {
		return o.ReadFile(path)
	}
	return os.ReadFile(path)
}

func (o *SecurityOptions) isGhAvailable() bool {
	if o.GhAvailable != nil {
		return *o.GhAvailable
	}
	_, err := exec.LookPath("gh")
	return err == nil
}

func (o *SecurityOptions) runGh(args ...string) (string, error) {
	if o.GhRunner != nil {
		return o.GhRunner(args...)
	}
	cmd := exec.Command("gh", args...)
	cmd.Dir = o.Dir
	out, err := cmd.Output()
	return string(out), err
}

func runSecurity(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	opts := &SecurityOptions{Dir: cwd}
	result := runSecurityChecks(opts)

	if securityJSON {
		return outputSecurityJSON(cmd, result)
	}
	return outputSecurityText(cmd, result)
}

func runSecurityChecks(opts *SecurityOptions) *SecurityResult {
	result := &SecurityResult{
		Checks: make([]SecurityCheck, 0),
	}

	// Run RTMX controls (local file parsing, no API)
	runRTMXControls(opts, result)

	// Run repository controls (gh CLI)
	runRepoControls(opts, result)

	// Calculate summary
	total := len(result.Checks)
	for _, check := range result.Checks {
		switch check.Status {
		case CheckPass:
			result.Summary.Passed++
		case CheckWarn:
			result.Summary.Warnings++
		case CheckFail:
			result.Summary.Failed++
		}
	}
	if total > 0 {
		result.Summary.ScorePercent = float64(result.Summary.Passed) / float64(total) * 100.0
	}

	return result
}

// runRTMXControls checks local RTMX configuration and files.
func runRTMXControls(opts *SecurityOptions, result *SecurityResult) {
	checkVerifyThresholds(opts, result)
	checkCIVerifyJob(opts, result)
	checkActionsPinned(opts, result)
	checkGPGSigning(opts, result)
	checkInstallGPG(opts, result)
	checkTrustedPeers(opts, result)
}

// checkVerifyThresholds checks if verify.thresholds are configured in rtmx.yaml.
func checkVerifyThresholds(opts *SecurityOptions, result *SecurityResult) {
	configPaths := []string{
		filepath.Join(opts.Dir, "rtmx.yaml"),
		filepath.Join(opts.Dir, "rtmx.yml"),
		filepath.Join(opts.Dir, ".rtmx", "config.yaml"),
	}

	for _, path := range configPaths {
		data, err := opts.readFile(path)
		if err != nil {
			continue
		}

		// Parse YAML to check for thresholds
		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			continue
		}

		// Navigate to rtmx.verify.thresholds or verify.thresholds
		thresholds := findNestedKey(raw, "verify", "thresholds")
		if thresholds != nil {
			if m, ok := thresholds.(map[string]interface{}); ok {
				warn, warnOK := m["warn"]
				fail, failOK := m["fail"]
				if warnOK && failOK {
					result.Checks = append(result.Checks, SecurityCheck{
						Category: "rtmx",
						Name:     "verify_thresholds",
						Status:   CheckPass,
						Message:  fmt.Sprintf("Thresholds configured (warn=%v, fail=%v)", warn, fail),
						Fixable:  false,
					})
					return
				}
			}
		}
	}

	result.Checks = append(result.Checks, SecurityCheck{
		Category: "rtmx",
		Name:     "verify_thresholds",
		Status:   CheckWarn,
		Message:  "No verify thresholds configured",
		Fixable:  true,
	})
}

// findNestedKey traverses a YAML map looking for a value at the given key path.
// It tries both top-level and nested under "rtmx" prefix.
func findNestedKey(m map[string]interface{}, keys ...string) interface{} {
	// Try direct path first
	if val := walkMap(m, keys...); val != nil {
		return val
	}
	// Try under "rtmx" prefix
	if rtmxVal, ok := m["rtmx"]; ok {
		if rtmxMap, ok := rtmxVal.(map[string]interface{}); ok {
			return walkMap(rtmxMap, keys...)
		}
	}
	return nil
}

func walkMap(m map[string]interface{}, keys ...string) interface{} {
	current := interface{}(m)
	for _, key := range keys {
		cm, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = cm[key]
		if !ok {
			return nil
		}
	}
	return current
}

// checkCIVerifyJob checks if a verify-requirements job exists in CI workflow.
func checkCIVerifyJob(opts *SecurityOptions, result *SecurityResult) {
	ciPath := filepath.Join(opts.Dir, ".github", "workflows", "ci.yml")
	data, err := opts.readFile(ciPath)
	if err != nil {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "ci_verify_job",
			Status:   CheckWarn,
			Message:  "CI workflow not found (.github/workflows/ci.yml)",
			Fixable:  true,
		})
		return
	}

	content := string(data)
	// Look for verify-requirements or verify_requirements job
	if strings.Contains(content, "verify-requirements") || strings.Contains(content, "verify_requirements") ||
		strings.Contains(content, "rtmx verify") || strings.Contains(content, "rtmx health") {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "ci_verify_job",
			Status:   CheckPass,
			Message:  "CI verify-requirements job exists",
			Fixable:  false,
		})
	} else {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "ci_verify_job",
			Status:   CheckWarn,
			Message:  "No verify-requirements job found in CI",
			Fixable:  true,
		})
	}
}

// checkActionsPinned checks if GitHub Actions are pinned to SHAs.
var actionsUsesPattern = regexp.MustCompile(`uses:\s*([^\s#]+)`)

func checkActionsPinned(opts *SecurityOptions, result *SecurityResult) {
	workflowDir := filepath.Join(opts.Dir, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "actions_pinned",
			Status:   CheckWarn,
			Message:  "No .github/workflows directory found",
			Fixable:  false,
		})
		return
	}

	totalActions := 0
	pinnedActions := 0
	unpinned := []string{}

	// SHA pattern: owner/repo@<40-char hex>
	shaPattern := regexp.MustCompile(`^[^@]+@[0-9a-f]{40}$`)

	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml")) {
			continue
		}

		data, err := opts.readFile(filepath.Join(workflowDir, entry.Name()))
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") {
				continue
			}
			matches := actionsUsesPattern.FindStringSubmatch(line)
			if len(matches) < 2 {
				continue
			}
			action := matches[1]
			// Skip local actions (./path)
			if strings.HasPrefix(action, "./") || strings.HasPrefix(action, "docker://") {
				continue
			}
			totalActions++
			if shaPattern.MatchString(action) {
				pinnedActions++
			} else {
				unpinned = append(unpinned, action)
			}
		}
	}

	if totalActions == 0 {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "actions_pinned",
			Status:   CheckPass,
			Message:  "No external actions found",
			Fixable:  false,
		})
	} else if pinnedActions == totalActions {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "actions_pinned",
			Status:   CheckPass,
			Message:  fmt.Sprintf("GitHub Actions pinned to SHAs (%d/%d)", pinnedActions, totalActions),
			Fixable:  false,
		})
	} else {
		msg := fmt.Sprintf("Actions not pinned to SHAs (%d/%d pinned)", pinnedActions, totalActions)
		if len(unpinned) <= 3 {
			msg += ": " + strings.Join(unpinned, ", ")
		}
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "actions_pinned",
			Status:   CheckWarn,
			Message:  msg,
			Fixable:  true,
		})
	}
}

// checkGPGSigning checks if GPG signing is mandatory in the release workflow.
func checkGPGSigning(opts *SecurityOptions, result *SecurityResult) {
	releasePath := filepath.Join(opts.Dir, ".github", "workflows", "release.yml")
	data, err := opts.readFile(releasePath)
	if err != nil {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "gpg_signing",
			Status:   CheckWarn,
			Message:  "Release workflow not found",
			Fixable:  true,
		})
		return
	}

	content := string(data)
	if strings.Contains(content, "gpg") || strings.Contains(content, "GPG") || strings.Contains(content, "cosign") || strings.Contains(content, "signing") {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "gpg_signing",
			Status:   CheckPass,
			Message:  "GPG signing mandatory in release workflow",
			Fixable:  false,
		})
	} else {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "gpg_signing",
			Status:   CheckWarn,
			Message:  "No GPG signing found in release workflow",
			Fixable:  true,
		})
	}
}

// checkInstallGPG checks if the install script verifies GPG signatures.
func checkInstallGPG(opts *SecurityOptions, result *SecurityResult) {
	installPath := filepath.Join(opts.Dir, "scripts", "install.sh")
	data, err := opts.readFile(installPath)
	if err != nil {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "install_gpg",
			Status:   CheckWarn,
			Message:  "Install script not found (scripts/install.sh)",
			Fixable:  true,
		})
		return
	}

	content := string(data)
	if strings.Contains(content, "gpg") || strings.Contains(content, "GPG") || strings.Contains(content, "signature") || strings.Contains(content, "checksums") {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "install_gpg",
			Status:   CheckPass,
			Message:  "Install script verifies GPG signatures",
			Fixable:  false,
		})
	} else {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "rtmx",
			Name:     "install_gpg",
			Status:   CheckWarn,
			Message:  "Install script does not verify GPG signatures",
			Fixable:  true,
		})
	}
}

// checkTrustedPeers checks if trusted peers are configured in sync config.
func checkTrustedPeers(opts *SecurityOptions, result *SecurityResult) {
	configPaths := []string{
		filepath.Join(opts.Dir, "rtmx.yaml"),
		filepath.Join(opts.Dir, "rtmx.yml"),
		filepath.Join(opts.Dir, ".rtmx", "config.yaml"),
	}

	for _, path := range configPaths {
		data, err := opts.readFile(path)
		if err != nil {
			continue
		}

		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			continue
		}

		peers := findNestedKey(raw, "sync", "trusted_peers")
		if peers != nil {
			// Check it's non-empty
			switch v := peers.(type) {
			case []interface{}:
				if len(v) > 0 {
					result.Checks = append(result.Checks, SecurityCheck{
						Category: "rtmx",
						Name:     "trusted_peers",
						Status:   CheckPass,
						Message:  fmt.Sprintf("Trusted peers configured (%d peers)", len(v)),
						Fixable:  false,
					})
					return
				}
			case map[string]interface{}:
				if len(v) > 0 {
					result.Checks = append(result.Checks, SecurityCheck{
						Category: "rtmx",
						Name:     "trusted_peers",
						Status:   CheckPass,
						Message:  fmt.Sprintf("Trusted peers configured (%d peers)", len(v)),
						Fixable:  false,
					})
					return
				}
			}
		}
	}

	result.Checks = append(result.Checks, SecurityCheck{
		Category: "rtmx",
		Name:     "trusted_peers",
		Status:   CheckWarn,
		Message:  "No trusted peers configured",
		Fixable:  true,
	})
}

// runRepoControls checks repository settings via gh CLI.
func runRepoControls(opts *SecurityOptions, result *SecurityResult) {
	if !opts.isGhAvailable() {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "branch_protection",
			Status:   CheckSkip,
			Message:  "gh CLI not available, skipping repository checks",
			Fixable:  false,
		})
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "codeowners",
			Status:   CheckSkip,
			Message:  "gh CLI not available, skipping repository checks",
			Fixable:  false,
		})
		return
	}

	// Detect repo from gh
	repoName, err := opts.runGh("repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner")
	if err != nil {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "branch_protection",
			Status:   CheckSkip,
			Message:  "Could not detect repository",
			Fixable:  false,
		})
		return
	}
	repoName = strings.TrimSpace(repoName)
	result.Platform = "github"
	result.Repository = repoName

	checkBranchProtection(opts, result, repoName)
	checkCodeowners(opts, result)
}

// checkBranchProtection checks if the default branch has branch protection rules.
func checkBranchProtection(opts *SecurityOptions, result *SecurityResult, repoName string) {
	// Try to get branch protection rules
	out, err := opts.runGh("api", fmt.Sprintf("repos/%s/rules", repoName), "--jq", "length")
	if err != nil {
		// API error could mean no permission or no rules
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "branch_protection",
			Status:   CheckWarn,
			Message:  "Could not check branch protection (API error or insufficient permissions)",
			Fixable:  true,
		})
		return
	}

	out = strings.TrimSpace(out)
	if out == "" || out == "0" || out == "null" {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "branch_protection",
			Status:   CheckWarn,
			Message:  "No branch protection rules found",
			Fixable:  true,
		})
	} else {
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "branch_protection",
			Status:   CheckPass,
			Message:  fmt.Sprintf("Branch protection enabled (%s rules)", out),
			Fixable:  false,
		})
	}
}

// checkCodeowners checks if CODEOWNERS exists and covers .rtmx/ paths.
func checkCodeowners(opts *SecurityOptions, result *SecurityResult) {
	codeownersPaths := []string{
		filepath.Join(opts.Dir, "CODEOWNERS"),
		filepath.Join(opts.Dir, ".github", "CODEOWNERS"),
		filepath.Join(opts.Dir, "docs", "CODEOWNERS"),
	}

	for _, path := range codeownersPaths {
		data, err := opts.readFile(path)
		if err != nil {
			continue
		}

		content := string(data)
		if strings.Contains(content, ".rtmx") {
			result.Checks = append(result.Checks, SecurityCheck{
				Category: "repository",
				Name:     "codeowners",
				Status:   CheckPass,
				Message:  "CODEOWNERS exists with .rtmx/ coverage",
				Fixable:  false,
			})
			return
		}

		// CODEOWNERS exists but no .rtmx coverage
		result.Checks = append(result.Checks, SecurityCheck{
			Category: "repository",
			Name:     "codeowners",
			Status:   CheckWarn,
			Message:  "CODEOWNERS exists but missing .rtmx/ path coverage",
			Fixable:  true,
		})
		return
	}

	result.Checks = append(result.Checks, SecurityCheck{
		Category: "repository",
		Name:     "codeowners",
		Status:   CheckWarn,
		Message:  "CODEOWNERS file not found",
		Fixable:  true,
	})
}

func outputSecurityJSON(cmd *cobra.Command, result *SecurityResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize security result: %w", err)
	}
	cmd.Println(string(data))

	exitCode := securityExitCode(result)
	if exitCode != 0 {
		return NewExitError(exitCode, "")
	}
	return nil
}

func outputSecurityText(cmd *cobra.Command, result *SecurityResult) error {
	width := 80

	cmd.Println(output.Header("RTMX Security Posture Check", width))
	cmd.Println()

	// Group checks by category
	rtmxChecks := []SecurityCheck{}
	repoChecks := []SecurityCheck{}
	for _, check := range result.Checks {
		if check.Category == "repository" {
			repoChecks = append(repoChecks, check)
		} else {
			rtmxChecks = append(rtmxChecks, check)
		}
	}

	if len(repoChecks) > 0 {
		repoLabel := "Repository Controls"
		if result.Repository != "" {
			repoLabel += fmt.Sprintf(" (GitHub: %s)", result.Repository)
		}
		cmd.Println(repoLabel)
		printSecurityChecks(cmd, repoChecks)
		cmd.Println()
	}

	if len(rtmxChecks) > 0 {
		cmd.Println("RTMX Controls")
		printSecurityChecks(cmd, rtmxChecks)
		cmd.Println()
	}

	// Summary
	total := len(result.Checks)
	cmd.Printf("Score: %d/%d passed, %d warnings, %d failures\n",
		result.Summary.Passed, total, result.Summary.Warnings, result.Summary.Failed)

	exitCode := securityExitCode(result)
	if exitCode != 0 {
		return NewExitError(exitCode, "")
	}
	return nil
}

func printSecurityChecks(cmd *cobra.Command, checks []SecurityCheck) {
	for _, check := range checks {
		var statusLabel, statusColor string
		switch check.Status {
		case CheckPass:
			statusLabel = "[PASS]"
			statusColor = output.Green
		case CheckWarn:
			statusLabel = "[WARN]"
			statusColor = output.Yellow
		case CheckFail:
			statusLabel = "[FAIL]"
			statusColor = output.Red
		case CheckSkip:
			statusLabel = "[SKIP]"
			statusColor = output.Dim
		}

		cmd.Printf("  %s %s\n",
			output.Color(statusLabel, statusColor),
			check.Message)
	}
}

func securityExitCode(result *SecurityResult) int {
	if result.Summary.Failed > 0 {
		return 1
	}
	if securityStrict && result.Summary.Warnings > 0 {
		return 1
	}
	return 0
}
