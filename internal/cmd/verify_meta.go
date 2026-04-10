package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VerifyMeta tracks when verify last ran.
type VerifyMeta struct {
	LastVerified     string `json:"last_verified"`
	LastVerifyCommit string `json:"last_verify_commit"`
}

// WriteVerifyMeta writes verify metadata after a successful verify --update.
func WriteVerifyMeta(rtmxDir string) error {
	meta := VerifyMeta{
		LastVerified:     time.Now().UTC().Format(time.RFC3339),
		LastVerifyCommit: getGitHEAD(),
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal verify meta: %w", err)
	}

	metaPath := filepath.Join(rtmxDir, "verify.meta")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write verify meta: %w", err)
	}

	return nil
}

// ReadVerifyMeta reads verify metadata. Returns nil if not found.
func ReadVerifyMeta(rtmxDir string) *VerifyMeta {
	metaPath := filepath.Join(rtmxDir, "verify.meta")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil
	}

	var meta VerifyMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil
	}

	return &meta
}

// CheckStaleness returns a warning message if status is stale, or empty string if fresh.
func CheckStaleness(rtmxDir string) string {
	meta := ReadVerifyMeta(rtmxDir)
	if meta == nil {
		return "Status has never been verified. Run `rtmx verify --update` to validate against test evidence."
	}

	currentHEAD := getGitHEAD()
	if currentHEAD == "" {
		// Not in a git repo -- can't check staleness
		return ""
	}

	if meta.LastVerifyCommit == currentHEAD {
		return "" // Fresh
	}

	// Count commits since last verify
	distance := getCommitDistance(meta.LastVerifyCommit, currentHEAD)
	if distance > 0 {
		return fmt.Sprintf("Status not verified since %s (%d commit(s) behind HEAD). Run `rtmx verify --update` or `rtmx status --verify` to refresh.",
			meta.LastVerifyCommit[:7], distance)
	}

	return fmt.Sprintf("Status not verified at current HEAD. Run `rtmx verify --update` to refresh.")
}

func getGitHEAD() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getCommitDistance(from, to string) int {
	cmd := exec.Command("git", "rev-list", "--count", from+".."+to)
	out, err := cmd.Output()
	if err != nil {
		return -1
	}
	var count int
	_, _ = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &count)
	return count
}
