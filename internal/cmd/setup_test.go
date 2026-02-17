package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupDetectProject(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test empty directory
	detection := detectProject(tmpDir)

	if detection["is_git_repo"].(bool) {
		t.Error("Empty dir should not be git repo")
	}
	if detection["has_rtmx_config"].(bool) {
		t.Error("Empty dir should not have rtmx config")
	}
	if detection["has_rtm_database"].(bool) {
		t.Error("Empty dir should not have RTM database")
	}
	if detection["has_tests"].(bool) {
		t.Error("Empty dir should not have tests")
	}
	if detection["has_makefile"].(bool) {
		t.Error("Empty dir should not have Makefile")
	}
}

func TestSetupDetectProjectWithConfig(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create rtmx.yaml
	configPath := filepath.Join(tmpDir, "rtmx.yaml")
	if err := os.WriteFile(configPath, []byte("rtmx:\n  database: test.csv\n"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := detectProject(tmpDir)

	if !detection["has_rtmx_config"].(bool) {
		t.Error("Should detect rtmx.yaml")
	}
}

func TestSetupDetectProjectWithRTM(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create docs/rtm_database.csv
	docsDir := filepath.Join(tmpDir, "docs")
	_ = os.MkdirAll(docsDir, 0755)
	rtmPath := filepath.Join(docsDir, "rtm_database.csv")
	if err := os.WriteFile(rtmPath, []byte("req_id,status\nREQ-001,MISSING\n"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := detectProject(tmpDir)

	if !detection["has_rtm_database"].(bool) {
		t.Error("Should detect RTM database")
	}
}

func TestSetupDetectProjectWithTests(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create tests directory
	testsDir := filepath.Join(tmpDir, "tests")
	_ = os.MkdirAll(testsDir, 0755)

	detection := detectProject(tmpDir)

	if !detection["has_tests"].(bool) {
		t.Error("Should detect tests directory")
	}
}

func TestSetupDetectProjectWithMakefile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create Makefile
	makefilePath := filepath.Join(tmpDir, "Makefile")
	if err := os.WriteFile(makefilePath, []byte("all:\n\techo hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := detectProject(tmpDir)

	if !detection["has_makefile"].(bool) {
		t.Error("Should detect Makefile")
	}
}

func TestSetupDetectAgentConfigs(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CLAUDE.md with RTMX content
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte("# Claude\n\n## RTMX\nUse rtmx commands\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .cursorrules without RTMX
	cursorPath := filepath.Join(tmpDir, ".cursorrules")
	if err := os.WriteFile(cursorPath, []byte("# Cursor rules\nGeneric content only\n"), 0644); err != nil {
		t.Fatal(err)
	}

	detection := detectProject(tmpDir)
	agentConfigs := detection["agent_configs"].(map[string]map[string]interface{})

	// Check claude
	claudeInfo := agentConfigs["claude"]
	if !claudeInfo["exists"].(bool) {
		t.Error("Should detect CLAUDE.md")
	}
	if !claudeInfo["has_rtmx"].(bool) {
		t.Error("CLAUDE.md should have RTMX")
	}

	// Check cursor
	cursorInfo := agentConfigs["cursor"]
	if !cursorInfo["exists"].(bool) {
		t.Error("Should detect .cursorrules")
	}
	if cursorInfo["has_rtmx"].(bool) {
		t.Error(".cursorrules should not have RTMX")
	}
}

func TestSetupBackupFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file to backup
	origPath := filepath.Join(tmpDir, "test.txt")
	origContent := "original content"
	if err := os.WriteFile(origPath, []byte(origContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Backup the file
	backupPath := backupFile(origPath)

	if backupPath == "" {
		t.Fatal("Backup path should not be empty")
	}

	if !strings.Contains(backupPath, "rtmx-backup") {
		t.Error("Backup path should contain rtmx-backup")
	}

	// Check backup content matches
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backupContent) != origContent {
		t.Error("Backup content should match original")
	}
}

func TestSetupBackupNonexistent(t *testing.T) {
	// Try to backup non-existent file
	backupPath := backupFile("/nonexistent/file.txt")

	if backupPath != "" {
		t.Error("Backup of non-existent file should return empty string")
	}
}

func TestSetupHelpers(t *testing.T) {
	// Test boolToYesNo
	if boolToYesNo(true) != "Yes" {
		t.Error("boolToYesNo(true) should be 'Yes'")
	}
	if boolToYesNo(false) != "No" {
		t.Error("boolToYesNo(false) should be 'No'")
	}

	// Test cleanToStatus
	if cleanToStatus(true) != "Clean" {
		t.Error("cleanToStatus(true) should be 'Clean'")
	}
	if cleanToStatus(false) != "Has changes" {
		t.Error("cleanToStatus(false) should be 'Has changes'")
	}

	// Test foundToStatus
	if foundToStatus(true) != "Found" {
		t.Error("foundToStatus(true) should be 'Found'")
	}
	if foundToStatus(false) != "Not found" {
		t.Error("foundToStatus(false) should be 'Not found'")
	}
}

func TestSetupResultToJSON(t *testing.T) {
	result := &SetupResult{
		Success:        true,
		StepsCompleted: []string{"create_config", "create_rtm"},
		StepsSkipped:   []string{"agent_claude"},
		FilesCreated:   []string{"rtmx.yaml"},
		FilesModified:  []string{},
		FilesBackedUp:  []string{},
		Warnings:       []string{},
		Errors:         []string{},
	}

	jsonBytes, err := result.ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, `"success": true`) {
		t.Error("JSON should contain success field")
	}
	if !strings.Contains(jsonStr, "create_config") {
		t.Error("JSON should contain steps_completed")
	}
	if !strings.Contains(jsonStr, "rtmx.yaml") {
		t.Error("JSON should contain files_created")
	}
}

func TestSetupCommand(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current directory
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Reset flags
	setupDryRun = true  // Use dry-run to avoid file creation
	setupMinimal = true // Minimal mode for faster test
	setupBranch = false
	setupPR = false
	setupForce = false
	setupSkipAgents = true
	setupSkipMakefile = true

	// Run setup in dry-run mode
	// The command should complete without error
	err = setupCmd.RunE(setupCmd, []string{})
	if err != nil {
		t.Errorf("Setup command should not error in dry-run mode: %v", err)
	}
}
