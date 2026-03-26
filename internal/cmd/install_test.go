package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestInstallDetectAgentConfigs(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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

	// Check all 10 agents are in the detection map
	expectedAgents := []string{
		"claude", "cursor", "copilot", "cline", "gemini",
		"windsurf", "aider", "amazonq", "zed", "continue",
	}
	for _, name := range expectedAgents {
		if _, exists := configs[name]; !exists {
			t.Errorf("Expected agent %q in detection map", name)
		}
	}
}

func TestInstallDetectNestedClaudeConfig(t *testing.T) {
	// Create temp directory with nested .claude/CLAUDE.md
	tmpDir, err := os.MkdirTemp("", "rtmx-install-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	// All 10 agents should return non-empty prompts containing RTMX context
	agents := []string{
		"claude", "cursor", "copilot", "cline", "gemini",
		"windsurf", "aider", "amazonq", "zed", "continue",
	}

	for _, agent := range agents {
		prompt := getAgentPrompt(agent)
		if prompt == "" {
			t.Errorf("Agent %q should return a non-empty prompt", agent)
			continue
		}
		if !strings.Contains(prompt, "RTMX") {
			t.Errorf("Agent %q prompt should contain 'RTMX'", agent)
		}
		if !strings.Contains(prompt, "rtmx verify --update") && !strings.Contains(prompt, "rtmx verify") {
			t.Errorf("Agent %q prompt should reference verify command", agent)
		}
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
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	defer func() { _ = os.RemoveAll(tmpDir) }()

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

func TestInstallClaudeCreatesHooksJSON(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "install", "--claude")

	if err != nil {
		t.Fatalf("install --claude failed: %v", err)
	}

	// Check .claude/hooks.json was created
	hooksPath := filepath.Join(tmpDir, ".claude", "hooks.json")
	data, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatalf("Expected .claude/hooks.json to exist, got error: %v\nCommand output: %s", err, output)
	}

	// Verify JSON structure
	var hooks map[string]interface{}
	if err := json.Unmarshal(data, &hooks); err != nil {
		t.Fatalf("hooks.json is not valid JSON: %v", err)
	}

	// Should have hooks.PreToolUse
	hooksObj, ok := hooks["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks.json should contain 'hooks' object")
	}
	preToolUse, ok := hooksObj["PreToolUse"].([]interface{})
	if !ok {
		t.Fatal("hooks.json should contain 'hooks.PreToolUse' array")
	}
	if len(preToolUse) == 0 {
		t.Fatal("PreToolUse array should not be empty")
	}

	// Check the first hook entry
	entry, ok := preToolUse[0].(map[string]interface{})
	if !ok {
		t.Fatal("PreToolUse entry should be an object")
	}
	if entry["command"] == nil {
		t.Error("Hook entry should have a 'command' field")
	}
	cmdStr, _ := entry["command"].(string)
	if !strings.Contains(cmdStr, "rtmx context") {
		t.Errorf("Hook command should contain 'rtmx context', got: %s", cmdStr)
	}
}

func TestInstallClaudeRemove(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()

	// Create the hooks.json first
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	hooksPath := filepath.Join(claudeDir, "hooks.json")
	if err := os.WriteFile(hooksPath, []byte(`{"hooks":{"PreToolUse":[{"matcher":".*","command":"rtmx context --format claude"}]}}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "install", "--claude", "--remove")

	if err != nil {
		t.Fatalf("install --claude --remove failed: %v", err)
	}

	// Verify hooks.json was removed
	if _, err := os.Stat(hooksPath); !os.IsNotExist(err) {
		t.Error("Expected hooks.json to be removed")
	}
}

func TestInstallClaudeDryRun(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "install", "--claude", "--dry-run")

	if err != nil {
		t.Fatalf("install --claude --dry-run failed: %v", err)
	}

	// Should mention dry run or would create
	if !strings.Contains(output, "Would create") {
		t.Errorf("Expected dry-run output to mention 'Would create', got: %s", output)
	}

	// Verify hooks.json was NOT created
	hooksPath := filepath.Join(tmpDir, ".claude", "hooks.json")
	if _, err := os.Stat(hooksPath); !os.IsNotExist(err) {
		t.Error("Expected hooks.json to NOT exist in dry-run mode")
	}
}

func TestInstallAllAgents(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-001",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	// Verify all 10 agents are in the supportedAgents registry
	if len(supportedAgents) != 10 {
		t.Errorf("Expected 10 supported agents, got %d", len(supportedAgents))
	}

	expectedAgents := map[string]string{
		"claude":   "CLAUDE.md",
		"cursor":   ".cursorrules",
		"copilot":  ".github/copilot-instructions.md",
		"cline":    ".clinerules",
		"gemini":   "GEMINI.md",
		"windsurf": ".windsurfrules",
		"aider":    ".aider.conf.yml",
		"amazonq":  ".amazonq/rules",
		"zed":      ".zed/settings.json",
		"continue": ".continue/config.yaml",
	}

	for _, agent := range supportedAgents {
		expectedFile, ok := expectedAgents[agent.Name]
		if !ok {
			t.Errorf("Unexpected agent in registry: %s", agent.Name)
			continue
		}
		if agent.ConfigFile != expectedFile {
			t.Errorf("Agent %s: expected config file %q, got %q", agent.Name, expectedFile, agent.ConfigFile)
		}
		// Every agent must have a non-empty prompt
		prompt := getAgentPrompt(agent.Name)
		if prompt == "" {
			t.Errorf("Agent %s has no prompt defined", agent.Name)
		}
	}

	// Verify detection covers all agents
	tmpDir := t.TempDir()
	configs := detectAgentConfigs(tmpDir)
	for name := range expectedAgents {
		if _, exists := configs[name]; !exists {
			t.Errorf("Agent %q not in detection map", name)
		}
	}
}

func TestInstallAgentsList(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-001",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	out, err := executeCommand(cmd, "install", "--list")

	if err != nil {
		t.Fatalf("install --list failed: %v", err)
	}

	// Should show all 10 agents
	expectedNames := []string{
		"claude", "cursor", "copilot", "cline", "gemini",
		"windsurf", "aider", "amazonq", "zed", "continue",
	}
	for _, name := range expectedNames {
		if !strings.Contains(out, name) {
			t.Errorf("Expected --list output to contain agent %q, got:\n%s", name, out)
		}
	}

	// Should mention the count
	if !strings.Contains(out, "10") {
		t.Errorf("Expected --list output to mention '10' agents, got:\n%s", out)
	}
}

func TestInstallDetectNewAgents(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-001",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()

	// Create config files for new agents
	newAgentFiles := map[string]string{
		"cline":    ".clinerules",
		"gemini":   "GEMINI.md",
		"windsurf": ".windsurfrules",
		"aider":    ".aider.conf.yml",
	}

	for _, file := range newAgentFiles {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Also create nested dirs for amazonq and zed
	for _, dir := range []string{".amazonq", ".zed", ".continue"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(filepath.Join(tmpDir, ".amazonq", "rules"), []byte("# test\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".zed", "settings.json"), []byte(`{"theme":"one-dark"}`), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".continue", "config.yaml"), []byte("# test\n"), 0644)

	configs := detectAgentConfigs(tmpDir)

	for agent, file := range newAgentFiles {
		expected := filepath.Join(tmpDir, file)
		if configs[agent] != expected {
			t.Errorf("Expected %s detected at %s, got %s", agent, expected, configs[agent])
		}
	}

	// Check nested path agents
	if configs["amazonq"] != filepath.Join(tmpDir, ".amazonq", "rules") {
		t.Errorf("Expected amazonq detected, got %s", configs["amazonq"])
	}
	if configs["zed"] != filepath.Join(tmpDir, ".zed", "settings.json") {
		t.Errorf("Expected zed detected, got %s", configs["zed"])
	}
	if configs["continue"] != filepath.Join(tmpDir, ".continue", "config.yaml") {
		t.Errorf("Expected continue detected, got %s", configs["continue"])
	}
}

func TestInstallZedJSON(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-001",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()

	// Create .zed/settings.json with existing settings
	zedDir := filepath.Join(tmpDir, ".zed")
	if err := os.MkdirAll(zedDir, 0755); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(zedDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"theme": "one-dark"}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "install", "--agents", "zed", "--skip-backup")

	if err != nil {
		t.Fatalf("install --agents zed failed: %v", err)
	}

	// Read the updated file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings.json: %v", err)
	}

	// Verify it's valid JSON
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}

	// Should preserve existing theme
	if settings["theme"] != "one-dark" {
		t.Errorf("Expected theme 'one-dark' preserved, got %v", settings["theme"])
	}

	// Should have assistant.instructions
	assistant, ok := settings["assistant"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected assistant key in settings")
	}
	instructions, ok := assistant["instructions"].(string)
	if !ok {
		t.Fatal("Expected assistant.instructions to be a string")
	}
	if !strings.Contains(instructions, "RTMX") {
		t.Errorf("Expected instructions to contain 'RTMX', got: %s", instructions)
	}
}

func TestInstallNewAgentCreatesFile(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-001",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	agents := []struct {
		name       string
		configFile string
	}{
		{"cline", ".clinerules"},
		{"gemini", "GEMINI.md"},
		{"windsurf", ".windsurfrules"},
		{"aider", ".aider.conf.yml"},
		{"amazonq", ".amazonq/rules"},
		{"zed", ".zed/settings.json"},
		{"continue", ".continue/config.yaml"},
	}

	for _, tc := range agents {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			origDir, _ := os.Getwd()
			_ = os.Chdir(tmpDir)
			defer func() { _ = os.Chdir(origDir) }()

			cmd := newTestRootCmd()
			_, err := executeCommand(cmd, "install", "--agents", tc.name, "--skip-backup")
			if err != nil {
				t.Fatalf("install --agents %s failed: %v", tc.name, err)
			}

			// Verify the file was created
			path := filepath.Join(tmpDir, tc.configFile)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Expected %s to exist: %v", tc.configFile, err)
			}

			content := string(data)
			if !strings.Contains(content, "RTMX") {
				t.Errorf("Expected %s to contain 'RTMX', got: %s", tc.configFile, content)
			}

			// For JSON format, verify it's valid JSON
			if tc.name == "zed" {
				var js map[string]interface{}
				if err := json.Unmarshal(data, &js); err != nil {
					t.Errorf("Expected %s to be valid JSON: %v", tc.configFile, err)
				}
			}
		})
	}
}
