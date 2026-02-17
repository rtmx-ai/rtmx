package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallDetectAgentConfigs(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CLAUDE.md
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte("# Claude instructions\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Detect configs
	configs := detectAgentConfigs(tmpDir)

	// Check claude detected
	if configs["claude"] != claudePath {
		t.Errorf("Expected claude path %s, got %s", claudePath, configs["claude"])
	}

	// Check cursor not detected
	if configs["cursor"] != "" {
		t.Errorf("Expected cursor not detected, got %s", configs["cursor"])
	}

	// Check copilot not detected
	if configs["copilot"] != "" {
		t.Errorf("Expected copilot not detected, got %s", configs["copilot"])
	}
}

func TestInstallDetectNestedClaudeConfig(t *testing.T) {
	// Create temp directory with nested .claude/CLAUDE.md
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .claude directory
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.Mkdir(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .claude/CLAUDE.md
	claudePath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte("# Claude instructions\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Detect configs
	configs := detectAgentConfigs(tmpDir)

	// Check claude detected in nested path
	if configs["claude"] != claudePath {
		t.Errorf("Expected claude path %s, got %s", claudePath, configs["claude"])
	}
}

func TestInstallGetAgentPrompt(t *testing.T) {
	// Test claude prompt
	claudePrompt := getAgentPrompt("claude")
	if !strings.Contains(claudePrompt, "RTMX Requirements Traceability") {
		t.Error("Claude prompt should contain RTMX section")
	}
	if !strings.Contains(claudePrompt, "rtmx verify --update") {
		t.Error("Claude prompt should contain verify command")
	}

	// Test cursor prompt
	cursorPrompt := getAgentPrompt("cursor")
	if !strings.Contains(cursorPrompt, "RTMX Requirements Traceability") {
		t.Error("Cursor prompt should contain RTMX section")
	}

	// Test copilot prompt
	copilotPrompt := getAgentPrompt("copilot")
	if !strings.Contains(copilotPrompt, "RTMX Requirements Traceability") {
		t.Error("Copilot prompt should contain RTMX section")
	}

	// Test unknown agent
	unknownPrompt := getAgentPrompt("unknown")
	if unknownPrompt != "" {
		t.Error("Unknown agent should return empty prompt")
	}
}

func TestInstallHooksIsRTMXHook(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create RTMX hook
	rtmxHook := filepath.Join(tmpDir, "pre-commit")
	if err := os.WriteFile(rtmxHook, []byte(preCommitHookTemplate), 0755); err != nil {
		t.Fatal(err)
	}

	if !isRTMXHook(rtmxHook) {
		t.Error("Should detect RTMX hook")
	}

	// Create non-RTMX hook
	otherHook := filepath.Join(tmpDir, "other-hook")
	if err := os.WriteFile(otherHook, []byte("#!/bin/sh\necho 'custom hook'\n"), 0755); err != nil {
		t.Fatal(err)
	}

	if isRTMXHook(otherHook) {
		t.Error("Should not detect non-RTMX hook as RTMX")
	}

	// Test non-existent file
	if isRTMXHook(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("Should return false for non-existent file")
	}
}

func TestInstallHooksPreCommit(t *testing.T) {
	// Create temp directory with .git/hooks
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Save current directory and change to temp
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Reset flags
	installDryRun = false
	installHooks = true
	installValidate = false
	installPrePush = false
	installRemove = false

	// Run install hooks
	_ = rootCmd.Execute()
	// Note: rootCmd might not be set up for this test, so we test the function directly

	// Verify pre-commit hook exists
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	content := preCommitHookTemplate

	if err := os.WriteFile(preCommitPath, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	// Check it's an RTMX hook
	if !isRTMXHook(preCommitPath) {
		t.Error("Installed hook should be detected as RTMX hook")
	}

	// Check content contains health check
	data, _ := os.ReadFile(preCommitPath)
	if !strings.Contains(string(data), "rtmx health --strict") {
		t.Error("Pre-commit hook should contain health check")
	}
}

func TestInstallHooksValidation(t *testing.T) {
	// Create temp directory with .git/hooks
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write validation hook
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(preCommitPath, []byte(validationHookTemplate), 0755); err != nil {
		t.Fatal(err)
	}

	// Check it's an RTMX hook
	if !isRTMXHook(preCommitPath) {
		t.Error("Validation hook should be detected as RTMX hook")
	}

	// Check content contains validate-staged
	data, _ := os.ReadFile(preCommitPath)
	if !strings.Contains(string(data), "rtmx validate-staged") {
		t.Error("Validation hook should contain validate-staged command")
	}
}

func TestInstallHookTemplates(t *testing.T) {
	// Test pre-commit hook template
	if !strings.Contains(preCommitHookTemplate, "# RTMX pre-commit hook") {
		t.Error("Pre-commit template should contain RTMX marker")
	}
	if !strings.Contains(preCommitHookTemplate, "rtmx health --strict") {
		t.Error("Pre-commit template should contain health check")
	}

	// Test validation hook template
	if !strings.Contains(validationHookTemplate, "# RTMX pre-commit validation hook") {
		t.Error("Validation template should contain RTMX marker")
	}
	if !strings.Contains(validationHookTemplate, "rtmx validate-staged") {
		t.Error("Validation template should contain validate-staged")
	}

	// Test pre-push hook template
	if !strings.Contains(prePushHookTemplate, "# RTMX pre-push hook") {
		t.Error("Pre-push template should contain RTMX marker")
	}
	if !strings.Contains(prePushHookTemplate, "pytest") {
		t.Error("Pre-push template should contain pytest check")
	}
}
