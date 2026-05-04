package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/internal/version"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Manage release versions and gates",
	Long: `Plan, scope, and gate releases by assigning requirements to versions.

Examples:
    rtmx release assign v0.3.0 REQ-PLAN-001 REQ-PLAN-002
    rtmx release unassign REQ-PLAN-001
    rtmx release scope v0.3.0
    rtmx release gate v0.3.0
    rtmx release gate v0.3.0 --verify`,
}

var (
	releaseGateVerify   bool
	releaseGateJSON     bool
	releaseAllowBreak   bool
	releaseDryRun       bool
)

var releaseGateCmd = &cobra.Command{
	Use:   "gate <version>",
	Short: "Verify all requirements for a version are complete",
	Long: `Check that every requirement assigned to a version is COMPLETE.
Exits 0 on pass, 1 on failure. Use in CI to gate releases.

Examples:
    rtmx release gate v0.3.0
    rtmx release gate v0.3.0 --verify   # run tests first
    rtmx release gate v0.3.0 --json     # machine-readable`,
	Args: cobra.ExactArgs(1),
	RunE: runReleaseGate,
}

var releaseScopeCmd = &cobra.Command{
	Use:   "scope <version>",
	Short: "Show release planning summary",
	Long: `Display requirement count, completion status, effort estimate,
and blocking requirements for a release version.`,
	Args: cobra.ExactArgs(1),
	RunE: runReleaseScope,
}

var releaseAssignCmd = &cobra.Command{
	Use:   "assign <version> <req-id> [req-id...]",
	Short: "Assign requirements to a target version",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runReleaseAssign,
}

var releaseUnassignCmd = &cobra.Command{
	Use:   "unassign <req-id> [req-id...]",
	Short: "Remove version assignment from requirements",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runReleaseUnassign,
}

func init() {
	releaseGateCmd.Flags().BoolVar(&releaseGateVerify, "verify", false, "run test verification before gate check")
	releaseGateCmd.Flags().BoolVar(&releaseGateJSON, "json", false, "output as JSON")
	releaseGateCmd.Flags().BoolVar(&releaseAllowBreak, "allow-breaking", false, "allow major bump even when backward_compatible is true")
	releaseAssignCmd.Flags().BoolVar(&releaseDryRun, "dry-run", false, "preview changes without writing")
	releaseUnassignCmd.Flags().BoolVar(&releaseDryRun, "dry-run", false, "preview changes without writing")

	releaseCmd.AddCommand(releaseGateCmd)
	releaseCmd.AddCommand(releaseScopeCmd)
	releaseCmd.AddCommand(releaseAssignCmd)
	releaseCmd.AddCommand(releaseUnassignCmd)
	rootCmd.AddCommand(releaseCmd)
}

// GateResult holds the result of a release gate check.
type GateResult struct {
	Version       string             `json:"version"`
	Passed        bool               `json:"passed"`
	Total         int                `json:"total"`
	Complete      int                `json:"complete"`
	Partial       int                `json:"partial"`
	Missing       int                `json:"missing"`
	Incomplete    []GateDetail       `json:"incomplete,omitempty"`
	VersionPolicy *VersionPolicyResult `json:"version_policy,omitempty"`
}

// VersionPolicyResult holds the version policy check output.
type VersionPolicyResult struct {
	PreviousVersion string           `json:"previous_version"`
	RequiredBump    string           `json:"required_bump"`
	ActualBump      string           `json:"actual_bump"`
	Compliant       bool             `json:"compliant"`
	CategoryImpacts []CategoryImpact `json:"category_impacts"`
}

// CategoryImpact describes the version impact of a category in the release.
type CategoryImpact struct {
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Level       string `json:"level"`
	Count       int    `json:"count"`
}

// GateDetail describes an incomplete requirement in the gate report.
type GateDetail struct {
	ReqID    string `json:"req_id"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Text     string `json:"requirement_text"`
}

func loadDBWithConfig(cwd string) (*database.Database, string, *config.Config, error) {
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to load config: %w", err)
	}
	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to load database: %w", err)
	}
	return db, dbPath, cfg, nil
}

func loadDB(cwd string) (*database.Database, string, error) {
	db, dbPath, _, err := loadDBWithConfig(cwd)
	return db, dbPath, err
}

func runReleaseGate(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	ver := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Run verification first if requested
	if releaseGateVerify {
		verifyUpdate = true
		if err := runVerify(cmd, nil); err != nil {
			cmd.Printf("%s Verify completed with errors: %v\n\n", output.Color("!", output.Yellow), err)
		}
	}

	db, _, cfg, err := loadDBWithConfig(cwd)
	if err != nil {
		return err
	}

	// Filter to version
	versionReqs := db.Filter(database.FilterOptions{TargetVersion: ver})
	if len(versionReqs) == 0 {
		if releaseGateJSON {
			result := GateResult{Version: ver, Passed: false}
			data, _ := json.MarshalIndent(result, "", "  ")
			cmd.Println(string(data))
		} else {
			cmd.Printf("%s No requirements assigned to version %s\n", output.Color("FAIL", output.Red), ver)
		}
		return NewExitError(1, fmt.Sprintf("no requirements assigned to %s", ver))
	}

	// Completeness check
	result := GateResult{
		Version: ver,
		Total:   len(versionReqs),
	}

	for _, req := range versionReqs {
		switch req.Status {
		case database.StatusComplete:
			result.Complete++
		case database.StatusPartial:
			result.Partial++
			result.Incomplete = append(result.Incomplete, GateDetail{
				ReqID:    req.ReqID,
				Status:   string(req.Status),
				Priority: string(req.Priority),
				Text:     req.RequirementText,
			})
		default:
			result.Missing++
			result.Incomplete = append(result.Incomplete, GateDetail{
				ReqID:    req.ReqID,
				Status:   string(req.Status),
				Priority: string(req.Priority),
				Text:     req.RequirementText,
			})
		}
	}
	result.Passed = result.Complete == result.Total

	// Version policy check
	policy := &cfg.RTMX.VersionPolicy
	policyFailed := false
	if policy.IsEnabled() {
		policyResult := checkVersionPolicy(ver, versionReqs, policy)
		result.VersionPolicy = policyResult

		if !policyResult.Compliant {
			if policy.Enforcement == "enforce" {
				result.Passed = false
				policyFailed = true
			}
		}
	}

	// Output
	if releaseGateJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		cmd.Println(string(data))
	} else {
		width := 60
		cmd.Println(output.Header(fmt.Sprintf("Release Gate: %s", ver), width))
		cmd.Println()
		cmd.Printf("  Requirements: %d total, %d complete, %d partial, %d missing\n",
			result.Total, result.Complete, result.Partial, result.Missing)
		cmd.Println()

		if result.Complete == result.Total {
			cmd.Printf("  %s All requirements for %s are COMPLETE\n", output.Color("PASS", output.Green), ver)
		} else {
			cmd.Printf("  %s %d requirement(s) not complete:\n\n", output.Color("FAIL", output.Red), len(result.Incomplete))
			for _, inc := range result.Incomplete {
				cmd.Printf("    %s  %-16s  %-8s  %s\n",
					output.Color(inc.Status, output.Yellow),
					inc.ReqID, inc.Priority, truncate(inc.Text, 40))
			}
		}

		// Version policy output
		if result.VersionPolicy != nil {
			cmd.Println()
			cmd.Println("  Version policy check:")
			for _, ci := range result.VersionPolicy.CategoryImpacts {
				label := ci.Category
				if ci.Subcategory != "" {
					label += "/" + ci.Subcategory
				}
				cmd.Printf("    %-24s -> %s (%d req)\n", label, ci.Level, ci.Count)
			}
			cmd.Println()
			if result.VersionPolicy.PreviousVersion != "" {
				cmd.Printf("  Previous version:  %s\n", result.VersionPolicy.PreviousVersion)
			}
			cmd.Printf("  Required bump:     %s\n", result.VersionPolicy.RequiredBump)
			cmd.Printf("  Actual bump:       %s\n", result.VersionPolicy.ActualBump)

			if result.VersionPolicy.Compliant {
				cmd.Printf("  %s Version bump is sufficient\n", output.Color("PASS", output.Green))
			} else {
				label := "WARN"
				color := output.Yellow
				if policy.Enforcement == "enforce" {
					label = "FAIL"
					color = output.Red
				}
				cmd.Printf("  %s Version bump is insufficient (need %s, got %s)\n",
					output.Color(label, color),
					result.VersionPolicy.RequiredBump,
					result.VersionPolicy.ActualBump)
			}
		}
		cmd.Println()
	}

	if !result.Passed {
		if policyFailed {
			return NewExitError(1, fmt.Sprintf("version policy failed for %s", ver))
		}
		return NewExitError(1, fmt.Sprintf("release gate failed for %s", ver))
	}
	return nil
}

// checkVersionPolicy computes the version policy result for a release.
func checkVersionPolicy(ver string, reqs []*database.Requirement, policy *config.VersionPolicyConfig) *VersionPolicyResult {
	result := &VersionPolicyResult{}

	// Compute category impacts
	impactMap := make(map[string]*CategoryImpact) // key: "cat/subcat"
	var highestBump version.BumpLevel

	for _, req := range reqs {
		levelStr := policy.IncrementLevel(req.Category, req.Subcategory)
		level := version.ParseBumpLevel(levelStr)
		if level > highestBump {
			highestBump = level
		}

		key := req.Category
		if req.Subcategory != "" {
			key += "/" + req.Subcategory
		}
		if ci, ok := impactMap[key]; ok {
			ci.Count++
		} else {
			impactMap[key] = &CategoryImpact{
				Category:    req.Category,
				Subcategory: req.Subcategory,
				Level:       levelStr,
				Count:       1,
			}
		}
	}

	// Sort impacts by level descending, then by category
	for _, ci := range impactMap {
		result.CategoryImpacts = append(result.CategoryImpacts, *ci)
	}
	sort.Slice(result.CategoryImpacts, func(i, j int) bool {
		li := version.ParseBumpLevel(result.CategoryImpacts[i].Level)
		lj := version.ParseBumpLevel(result.CategoryImpacts[j].Level)
		if li != lj {
			return li > lj
		}
		return result.CategoryImpacts[i].Category < result.CategoryImpacts[j].Category
	})

	result.RequiredBump = highestBump.String()

	// Get previous version from git
	prevTag := getPreviousTag()
	result.PreviousVersion = prevTag

	// Parse versions and compare
	proposed, err := version.Parse(ver)
	if err != nil {
		result.ActualBump = "unknown"
		result.Compliant = true // can't check, don't block
		return result
	}

	if prevTag == "" {
		result.ActualBump = "initial"
		result.Compliant = true
		return result
	}

	prev, err := version.Parse(prevTag)
	if err != nil {
		result.ActualBump = "unknown"
		result.Compliant = true
		return result
	}

	actualBump := version.ActualBump(prev, proposed)
	result.ActualBump = actualBump.String()

	// Check compliance: actual bump must be >= required bump
	result.Compliant = actualBump >= highestBump

	// Backward compatibility check
	if policy.IsBackwardCompatible() && highestBump == version.BumpMajor && !releaseAllowBreak {
		if actualBump < version.BumpMajor {
			result.Compliant = false
		}
	}

	return result
}

// getPreviousTag returns the most recent version tag before HEAD.
func getPreviousTag() string {
	out, err := exec.Command("git", "describe", "--tags", "--abbrev=0", "HEAD~1").Output()
	if err != nil {
		// Try without HEAD~1 for repos with only one tag
		out, err = exec.Command("git", "tag", "-l", "v*", "--sort=-v:refname").Output()
		if err != nil || len(out) == 0 {
			return ""
		}
		tags := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(tags) >= 2 {
			return tags[1] // second most recent
		}
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runReleaseScope(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	version := args[0]
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	db, _, err := loadDB(cwd)
	if err != nil {
		return err
	}

	versionReqs := db.Filter(database.FilterOptions{TargetVersion: version})
	if len(versionReqs) == 0 {
		cmd.Printf("No requirements assigned to version %s\n", version)
		return nil
	}

	// Compute stats
	var complete, partial, missing int
	var totalEffort, remainingEffort float64
	for _, req := range versionReqs {
		switch req.Status {
		case database.StatusComplete:
			complete++
		case database.StatusPartial:
			partial++
		default:
			missing++
		}
		totalEffort += req.EffortWeeks
		if req.Status != database.StatusComplete {
			remainingEffort += req.EffortWeeks
		}
	}

	// Find external blockers (incomplete deps outside this version)
	var blockers []string
	versionSet := make(map[string]bool)
	for _, req := range versionReqs {
		versionSet[req.ReqID] = true
	}
	for _, req := range versionReqs {
		for dep := range req.Dependencies {
			if !versionSet[dep] {
				depReq := db.Get(dep)
				if depReq != nil && depReq.IsIncomplete() {
					blockers = append(blockers, dep)
				}
			}
		}
	}

	width := 60
	cmd.Println(output.Header(fmt.Sprintf("Release Scope: %s", version), width))
	cmd.Println()
	cmd.Printf("  Requirements:    %d total\n", len(versionReqs))
	cmd.Printf("  Complete:        %d\n", complete)
	cmd.Printf("  Partial:         %d\n", partial)
	cmd.Printf("  Missing:         %d\n", missing)
	cmd.Println()
	cmd.Printf("  Total effort:    %.1f weeks\n", totalEffort)
	cmd.Printf("  Remaining:       %.1f weeks\n", remainingEffort)

	if len(blockers) > 0 {
		cmd.Println()
		cmd.Printf("  %s External blockers (%d):\n", output.Color("!", output.Yellow), len(blockers))
		for _, b := range blockers {
			cmd.Printf("    - %s\n", b)
		}
	}
	cmd.Println()

	return nil
}

func runReleaseAssign(cmd *cobra.Command, args []string) error {
	version := args[0]
	reqIDs := args[1:]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	db, dbPath, err := loadDB(cwd)
	if err != nil {
		return err
	}

	var updated int
	for _, id := range reqIDs {
		req := db.Get(id)
		if req == nil {
			cmd.Printf("%s Unknown requirement: %s\n", output.Color("!", output.Yellow), id)
			continue
		}
		if req.TargetVersion() == version {
			cmd.Printf("  %s %s already assigned to %s\n", output.Color("[SKIP]", output.Dim), id, version)
			continue
		}
		old := req.TargetVersion()
		req.SetTargetVersion(version)
		updated++
		if old != "" {
			cmd.Printf("  %s %s: %s -> %s\n", output.Color("[UPDATE]", output.Green), id, old, version)
		} else {
			cmd.Printf("  %s %s -> %s\n", output.Color("[ASSIGN]", output.Green), id, version)
		}
	}

	if updated > 0 && !releaseDryRun {
		if err := db.Save(dbPath); err != nil {
			return fmt.Errorf("failed to save database: %w", err)
		}
		cmd.Printf("\nAssigned %d requirement(s) to %s\n", updated, version)
	} else if releaseDryRun {
		cmd.Printf("\nDry run: %d requirement(s) would be assigned to %s\n", updated, version)
	}

	return nil
}

func runReleaseUnassign(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	db, dbPath, err := loadDB(cwd)
	if err != nil {
		return err
	}

	var updated int
	for _, id := range args {
		req := db.Get(id)
		if req == nil {
			cmd.Printf("%s Unknown requirement: %s\n", output.Color("!", output.Yellow), id)
			continue
		}
		if req.TargetVersion() == "" {
			cmd.Printf("  %s %s has no version assignment\n", output.Color("[SKIP]", output.Dim), id)
			continue
		}
		old := req.TargetVersion()
		req.SetTargetVersion("")
		updated++
		cmd.Printf("  %s %s (was %s)\n", output.Color("[UNASSIGN]", output.Green), id, old)
	}

	if updated > 0 && !releaseDryRun {
		if err := db.Save(dbPath); err != nil {
			return fmt.Errorf("failed to save database: %w", err)
		}
		cmd.Printf("\nUnassigned %d requirement(s)\n", updated)
	} else if releaseDryRun {
		cmd.Printf("\nDry run: %d requirement(s) would be unassigned\n", updated)
	}

	return nil
}

// truncate is defined in context.go
