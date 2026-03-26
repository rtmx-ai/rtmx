package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestRemoteCommands(t *testing.T) {
	rtmx.Req(t, "REQ-GO-033")

	// Create a temp project directory with config
	tmpDir, err := os.MkdirTemp("", "rtmx-remote-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .rtmx directory and config
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	cfg := config.DefaultConfig()
	configPath := filepath.Join(rtmxDir, "config.yaml")
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Save original working directory and change to temp dir
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	t.Run("list_empty", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := remoteListCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("remote list failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "No remotes configured") {
			t.Errorf("Expected empty remotes message, got: %s", out)
		}
	})

	t.Run("add_remote", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := remoteAddCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Set flags
		remoteRepo = "rtmx-ai/rtmx"
		remoteDatabase = ".rtmx/database.csv"
		remotePath = ""

		if err := cmd.RunE(cmd, []string{"upstream"}); err != nil {
			t.Fatalf("remote add failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "Added remote") {
			t.Errorf("Expected add confirmation, got: %s", out)
		}
		if !strings.Contains(out, "upstream") {
			t.Errorf("Expected alias in output, got: %s", out)
		}

		// Verify config was saved
		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		remote, exists := reloaded.RTMX.Sync.Remotes["upstream"]
		if !exists {
			t.Fatal("Remote 'upstream' not found in saved config")
		}
		if remote.Repo != "rtmx-ai/rtmx" {
			t.Errorf("Expected repo 'rtmx-ai/rtmx', got %q", remote.Repo)
		}
	})

	t.Run("add_remote_with_path", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := remoteAddCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		remoteRepo = "rtmx-ai/rtmx-sync"
		remoteDatabase = ".rtmx/database.csv"
		remotePath = "../rtmx-sync"

		if err := cmd.RunE(cmd, []string{"sync-repo"}); err != nil {
			t.Fatalf("remote add with path failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "Local path: ../rtmx-sync") {
			t.Errorf("Expected path in output, got: %s", out)
		}
	})

	t.Run("add_duplicate_fails", func(t *testing.T) {
		remoteRepo = "rtmx-ai/other"
		remoteDatabase = ".rtmx/database.csv"
		remotePath = ""

		err := remoteAddCmd.RunE(remoteAddCmd, []string{"upstream"})
		if err == nil {
			t.Fatal("Expected error for duplicate alias")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("list_with_remotes", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := remoteListCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		if err := cmd.RunE(cmd, nil); err != nil {
			t.Fatalf("remote list failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "upstream") {
			t.Errorf("Expected 'upstream' in list, got: %s", out)
		}
		if !strings.Contains(out, "rtmx-ai/rtmx") {
			t.Errorf("Expected repo in list, got: %s", out)
		}
		if !strings.Contains(out, "ALIAS") {
			t.Errorf("Expected table headers, got: %s", out)
		}
	})

	t.Run("remove_remote", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := remoteRemoveCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		if err := cmd.RunE(cmd, []string{"upstream"}); err != nil {
			t.Fatalf("remote remove failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "Removed remote") {
			t.Errorf("Expected remove confirmation, got: %s", out)
		}

		// Verify removed from config
		reloaded, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		if _, exists := reloaded.RTMX.Sync.Remotes["upstream"]; exists {
			t.Error("Remote 'upstream' should have been removed")
		}
	})

	t.Run("remove_nonexistent_fails", func(t *testing.T) {
		err := remoteRemoveCmd.RunE(remoteRemoveCmd, []string{"nonexistent"})
		if err == nil {
			t.Fatal("Expected error for nonexistent alias")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}
