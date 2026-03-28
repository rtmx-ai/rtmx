package sync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// CommandRunner abstracts command execution for testability.
type CommandRunner func(dir string, name string, args ...string) ([]byte, error)

// defaultCommandRunner runs a command in the given directory using os/exec.
func defaultCommandRunner(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// CrossRepoOptions configures a cross-repo move or clone operation.
type CrossRepoOptions struct {
	// SrcDir is the root directory of the source rtmx project.
	SrcDir string

	// DstDir is the root directory of the target rtmx project.
	DstDir string

	// TargetID overrides the requirement ID in the target database.
	// If empty, the original ID is preserved.
	TargetID string

	// DryRun previews changes without writing.
	DryRun bool

	// Branch, if set, creates a Git branch in the target repo before writing.
	Branch string

	// PR, if true, creates a pull request after writing to the branch.
	// Requires Branch to be set.
	PR bool

	// RunCmd is an optional command runner for testability.
	// If nil, defaultCommandRunner is used.
	RunCmd CommandRunner
}

// CrossRepoResult summarizes the outcome of a cross-repo operation.
type CrossRepoResult struct {
	// MovedID is the requirement ID transferred (for move operations).
	MovedID string

	// ClonedID is the requirement ID created (for clone operations).
	ClonedID string

	// SourceExternalID is the external_id set on the source requirement.
	SourceExternalID string

	// TargetExternalID is the external_id set on the target requirement.
	TargetExternalID string

	// DryRun indicates whether the operation was a dry run.
	DryRun bool

	// SpecFileCopied indicates whether a spec file was copied.
	SpecFileCopied bool

	// BranchCreated indicates whether a branch was created in the target repo.
	BranchCreated bool

	// PRCreated indicates whether a pull request was created.
	PRCreated bool

	// PRURL is the URL of the created pull request, if any.
	PRURL string
}

// isRtmxEnabled checks if a directory is an rtmx-enabled project.
func isRtmxEnabled(dir string) bool {
	// Check for .rtmx/ directory
	if _, err := os.Stat(filepath.Join(dir, ".rtmx")); err == nil {
		return true
	}
	// Check for docs/rtm_database.csv (legacy layout)
	if _, err := os.Stat(filepath.Join(dir, "docs", "rtm_database.csv")); err == nil {
		return true
	}
	return false
}

// repoName returns a short identifier for a project directory.
func repoName(dir string) string {
	return filepath.Base(dir)
}

// buildExternalID creates an external_id link string.
func buildExternalID(repoDir, reqID string) string {
	return repoName(repoDir) + "/" + reqID
}

// validateCrossRepo performs common validation for move/clone operations.
func validateCrossRepo(srcDB *database.Database, reqID string, opts CrossRepoOptions) (*database.Requirement, error) {
	if !isRtmxEnabled(opts.DstDir) {
		return nil, fmt.Errorf("target directory is not an rtmx-enabled project: %s", opts.DstDir)
	}

	srcReq := srcDB.Get(reqID)
	if srcReq == nil {
		return nil, fmt.Errorf("requirement %q not found in source database", reqID)
	}

	return srcReq, nil
}

// copySpecFile copies a requirement spec file from source to destination.
func copySpecFile(srcDir, dstDir string, srcReq *database.Requirement, targetID string) (bool, error) {
	if srcReq.RequirementFile == "" {
		return false, nil
	}

	srcSpecPath := filepath.Join(srcDir, ".rtmx", srcReq.RequirementFile)
	if _, err := os.Stat(srcSpecPath); os.IsNotExist(err) {
		return false, nil
	}

	// Determine destination spec path using same category structure
	dstSpecDir := filepath.Join(dstDir, ".rtmx", "requirements", srcReq.Category)
	if err := os.MkdirAll(dstSpecDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create spec directory: %w", err)
	}

	specFileName := targetID + ".md"
	dstSpecPath := filepath.Join(dstSpecDir, specFileName)

	content, err := os.ReadFile(srcSpecPath)
	if err != nil {
		return false, fmt.Errorf("failed to read spec file: %w", err)
	}

	if err := os.WriteFile(dstSpecPath, content, 0644); err != nil {
		return false, fmt.Errorf("failed to write spec file: %w", err)
	}

	return true, nil
}

// getRunner returns the command runner from opts, or the default.
func getRunner(opts CrossRepoOptions) CommandRunner {
	if opts.RunCmd != nil {
		return opts.RunCmd
	}
	return defaultCommandRunner
}

// createBranch creates a new Git branch in the target directory.
func createBranch(opts CrossRepoOptions) error {
	run := getRunner(opts)
	out, err := run(opts.DstDir, "git", "checkout", "-b", opts.Branch)
	if err != nil {
		return fmt.Errorf("failed to create branch %q: %s", opts.Branch, strings.TrimSpace(string(out)))
	}
	return nil
}

// commitAndPR commits changes in the target repo and creates a PR.
func commitAndPR(opts CrossRepoOptions, reqID, targetID string) (string, error) {
	run := getRunner(opts)

	// Stage all changes
	if out, err := run(opts.DstDir, "git", "add", "-A"); err != nil {
		return "", fmt.Errorf("failed to stage changes: %s", strings.TrimSpace(string(out)))
	}

	// Commit
	commitMsg := fmt.Sprintf("req: Accept %s from %s", targetID, repoName(opts.SrcDir))
	if out, err := run(opts.DstDir, "git", "commit", "-m", commitMsg); err != nil {
		return "", fmt.Errorf("failed to commit: %s", strings.TrimSpace(string(out)))
	}

	// Push branch
	if out, err := run(opts.DstDir, "git", "push", "-u", "origin", opts.Branch); err != nil {
		return "", fmt.Errorf("failed to push branch: %s", strings.TrimSpace(string(out)))
	}

	// Create PR
	prTitle := fmt.Sprintf("req: Accept %s from %s", targetID, repoName(opts.SrcDir))
	prBody := fmt.Sprintf("## Requirement Transfer\n\n- **Requirement**: %s\n- **Source**: %s\n- **Source ID**: %s\n", targetID, repoName(opts.SrcDir), reqID)
	out, err := run(opts.DstDir, "gh", "pr", "create",
		"--title", prTitle,
		"--body", prBody,
		"--label", "requirement,cross-repo",
	)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %s", strings.TrimSpace(string(out)))
	}

	return strings.TrimSpace(string(out)), nil
}

// MoveRequirement transfers a requirement from the source database to the destination database
// with bidirectional provenance links via external_id.
func MoveRequirement(srcDB, dstDB *database.Database, reqID string, opts CrossRepoOptions) (*CrossRepoResult, error) {
	srcReq, err := validateCrossRepo(srcDB, reqID, opts)
	if err != nil {
		return nil, err
	}

	targetID := reqID
	if opts.TargetID != "" {
		targetID = opts.TargetID
	}

	srcExtID := buildExternalID(opts.DstDir, targetID)
	dstExtID := buildExternalID(opts.SrcDir, reqID)

	result := &CrossRepoResult{
		MovedID:          targetID,
		SourceExternalID: srcExtID,
		TargetExternalID: dstExtID,
		DryRun:           opts.DryRun,
	}

	if opts.DryRun {
		// Annotate dry-run result with branch/PR intent
		if opts.Branch != "" {
			result.BranchCreated = true
		}
		if opts.PR {
			result.PRCreated = true
		}
		return result, nil
	}

	// Create branch in target repo if requested
	if opts.Branch != "" {
		if err := createBranch(opts); err != nil {
			return nil, err
		}
		result.BranchCreated = true
	}

	// Validate --pr requires --branch
	if opts.PR && opts.Branch == "" {
		return nil, fmt.Errorf("--pr requires --branch to be set")
	}

	// Clone the requirement for the destination
	dstReq := srcReq.Clone()
	dstReq.ReqID = targetID
	dstReq.ExternalID = dstExtID

	// Update requirement_file for the destination if it exists
	if srcReq.RequirementFile != "" {
		dstReq.RequirementFile = filepath.Join("requirements", srcReq.Category, targetID+".md")
	}

	// Add to destination
	if err := dstDB.Add(dstReq); err != nil {
		return nil, fmt.Errorf("failed to add requirement to destination: %w", err)
	}

	// Set provenance link on source
	srcReq.ExternalID = srcExtID

	// Copy spec file if it exists
	copied, err := copySpecFile(opts.SrcDir, opts.DstDir, srcReq, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to copy spec file: %w", err)
	}
	result.SpecFileCopied = copied

	// Create PR if requested (after database save happens in caller)
	if opts.PR {
		prURL, err := commitAndPR(opts, reqID, targetID)
		if err != nil {
			return nil, err
		}
		result.PRCreated = true
		result.PRURL = prURL
	}

	return result, nil
}

// CloneRequirement creates a copy of a requirement in the destination database
// while preserving the original in the source database. Both get bidirectional
// provenance links via external_id.
func CloneRequirement(srcDB, dstDB *database.Database, reqID string, opts CrossRepoOptions) (*CrossRepoResult, error) {
	srcReq, err := validateCrossRepo(srcDB, reqID, opts)
	if err != nil {
		return nil, err
	}

	targetID := reqID
	if opts.TargetID != "" {
		targetID = opts.TargetID
	}

	srcExtID := buildExternalID(opts.DstDir, targetID)
	dstExtID := buildExternalID(opts.SrcDir, reqID)

	result := &CrossRepoResult{
		ClonedID:         targetID,
		SourceExternalID: srcExtID,
		TargetExternalID: dstExtID,
		DryRun:           opts.DryRun,
	}

	if opts.DryRun {
		// Annotate dry-run result with branch/PR intent
		if opts.Branch != "" {
			result.BranchCreated = true
		}
		if opts.PR {
			result.PRCreated = true
		}
		return result, nil
	}

	// Create branch in target repo if requested
	if opts.Branch != "" {
		if err := createBranch(opts); err != nil {
			return nil, err
		}
		result.BranchCreated = true
	}

	// Validate --pr requires --branch
	if opts.PR && opts.Branch == "" {
		return nil, fmt.Errorf("--pr requires --branch to be set")
	}

	// Clone the requirement for the destination
	dstReq := srcReq.Clone()
	dstReq.ReqID = targetID
	dstReq.ExternalID = dstExtID

	// Update requirement_file for the destination if it exists
	if srcReq.RequirementFile != "" {
		dstReq.RequirementFile = filepath.Join("requirements", srcReq.Category, targetID+".md")
	}

	// Add to destination
	if err := dstDB.Add(dstReq); err != nil {
		return nil, fmt.Errorf("failed to add requirement to destination: %w", err)
	}

	// Set provenance link on source (original stays in place)
	srcReq.ExternalID = srcExtID

	// Copy spec file if it exists
	copied, err := copySpecFile(opts.SrcDir, opts.DstDir, srcReq, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to copy spec file: %w", err)
	}
	result.SpecFileCopied = copied

	// Create PR if requested (after database save happens in caller)
	if opts.PR {
		prURL, err := commitAndPR(opts, reqID, targetID)
		if err != nil {
			return nil, err
		}
		result.PRCreated = true
		result.PRURL = prURL
	}

	return result, nil
}
