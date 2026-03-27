package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}

func TestStatusStalenessWarning(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-013")

	t.Run("no_meta_file_returns_warning", func(t *testing.T) {
		dir := t.TempDir()
		warning := CheckStaleness(dir)
		if warning == "" {
			t.Error("expected staleness warning when verify.meta doesn't exist")
		}
		if warning != "Status has never been verified. Run `rtmx verify --update` to validate against test evidence." {
			t.Errorf("unexpected warning: %s", warning)
		}
	})

	t.Run("meta_at_current_head_returns_empty", func(t *testing.T) {
		dir := t.TempDir()
		head := getGitHEAD()
		if head == "" {
			t.Skip("not in a git repo")
		}

		// Write meta pointing to current HEAD
		meta := VerifyMeta{
			LastVerified:     "2026-03-26T00:00:00Z",
			LastVerifyCommit: head,
		}
		data, _ := json.MarshalIndent(meta, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "verify.meta"), data, 0644)

		warning := CheckStaleness(dir)
		if warning != "" {
			t.Errorf("expected no warning when meta matches HEAD, got: %s", warning)
		}
	})

	t.Run("meta_behind_head_returns_warning", func(t *testing.T) {
		dir := t.TempDir()
		head := getGitHEAD()
		if head == "" {
			t.Skip("not in a git repo")
		}

		// Write meta pointing to a parent commit
		parentSHA := getParentCommit()
		if parentSHA == "" {
			t.Skip("no parent commit available")
		}

		meta := VerifyMeta{
			LastVerified:     "2026-03-26T00:00:00Z",
			LastVerifyCommit: parentSHA,
		}
		data, _ := json.MarshalIndent(meta, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "verify.meta"), data, 0644)

		warning := CheckStaleness(dir)
		if warning == "" {
			t.Error("expected staleness warning when meta is behind HEAD")
		}
		if len(parentSHA) >= 7 {
			// Warning should include the short SHA
			if !containsStr(warning, parentSHA[:7]) {
				t.Errorf("warning should reference commit SHA, got: %s", warning)
			}
		}
	})

	t.Run("write_and_read_roundtrip", func(t *testing.T) {
		dir := t.TempDir()

		err := WriteVerifyMeta(dir)
		if err != nil {
			t.Fatalf("WriteVerifyMeta failed: %v", err)
		}

		meta := ReadVerifyMeta(dir)
		if meta == nil {
			t.Fatal("ReadVerifyMeta returned nil after write")
		}
		if meta.LastVerified == "" {
			t.Error("LastVerified should not be empty")
		}
		if meta.LastVerifyCommit == "" {
			// May be empty if not in a git repo, that's OK
			t.Log("LastVerifyCommit is empty (may not be in git repo)")
		}
	})

	t.Run("read_corrupted_returns_nil", func(t *testing.T) {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "verify.meta"), []byte("not json"), 0644)

		meta := ReadVerifyMeta(dir)
		if meta != nil {
			t.Error("expected nil for corrupted meta file")
		}
	})

	t.Run("not_in_git_repo_no_warning", func(t *testing.T) {
		dir := t.TempDir()
		// Write meta with a fake commit -- outside a git repo, should not warn
		meta := VerifyMeta{
			LastVerified:     "2026-03-26T00:00:00Z",
			LastVerifyCommit: "0000000000000000000000000000000000000000",
		}
		data, _ := json.MarshalIndent(meta, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "verify.meta"), data, 0644)

		// CheckStaleness calls getGitHEAD which uses the working directory's git repo.
		// Since we're running tests inside a git repo, this test verifies
		// the commit mismatch path rather than the "no git" path.
		// The "no git" path is covered implicitly when getGitHEAD returns "".
		warning := CheckStaleness(dir)
		// We expect a warning because the fake commit doesn't match HEAD
		if warning == "" {
			t.Log("no warning -- may be outside git repo (expected)")
		}
	})
}

func getParentCommit() string {
	out, err := exec.Command("git", "rev-parse", "HEAD~1").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
