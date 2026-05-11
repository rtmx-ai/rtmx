package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func newTestAssignCmd() *cobra.Command {
	var to string

	cmd := &cobra.Command{
		Use:  "assign",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assignTo = to
			return runAssign(cmd, args)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "user to assign")
	return cmd
}

func newTestUnassignCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "unassign",
		Args: cobra.ExactArgs(1),
		RunE: runUnassign,
	}
}

func newAssignTestRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newTestAssignCmd())
	root.AddCommand(newTestUnassignCmd())
	return root
}

// setupAssignTestDir creates a temp directory with rtmx config and database.
func setupAssignTestDir(t *testing.T, reqs []*database.Requirement) string {
	t.Helper()
	dir := t.TempDir()

	// Create .rtmx directory and database
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatal(err)
	}

	db := database.NewDatabase()
	for _, r := range reqs {
		if err := db.Add(r); err != nil {
			t.Fatal(err)
		}
	}

	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := db.Save(dbPath); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestAssignCommand(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-009")

	t.Run("assign_with_to_flag", func(t *testing.T) {
		req := &database.Requirement{
			ReqID:    "REQ-TEST-001",
			Category: "TEST",
			Status:   database.StatusMissing,
		}

		dir := setupAssignTestDir(t, []*database.Requirement{req})
		origDir, _ := os.Getwd()
		_ = os.Chdir(dir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := newAssignTestRoot()
		out, err := executeCommand(cmd, "assign", "REQ-TEST-001", "--to", "alice")
		if err != nil {
			t.Fatalf("assign failed: %v", err)
		}

		if !strings.Contains(out, "Assigned REQ-TEST-001 to alice") {
			t.Errorf("unexpected output: %s", out)
		}

		// Verify the database was updated
		db, err := database.Load(filepath.Join(dir, ".rtmx", "database.csv"))
		if err != nil {
			t.Fatal(err)
		}
		updated := db.Get("REQ-TEST-001")
		if updated.Assignee != "alice" {
			t.Errorf("Assignee = %q, want %q", updated.Assignee, "alice")
		}
		if updated.StartedDate == "" {
			t.Error("started_date should be set")
		}
	})

	t.Run("assign_without_to_and_no_auth", func(t *testing.T) {
		req := &database.Requirement{
			ReqID:    "REQ-TEST-001",
			Category: "TEST",
			Status:   database.StatusMissing,
		}

		dir := setupAssignTestDir(t, []*database.Requirement{req})
		origDir, _ := os.Getwd()
		_ = os.Chdir(dir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := newAssignTestRoot()
		_, err := executeCommand(cmd, "assign", "REQ-TEST-001")
		if err == nil {
			t.Fatal("expected error when no --to and no auth")
		}
		if !strings.Contains(err.Error(), "no --to specified") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("assign_nonexistent_req", func(t *testing.T) {
		dir := setupAssignTestDir(t, nil)
		origDir, _ := os.Getwd()
		_ = os.Chdir(dir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := newAssignTestRoot()
		_, err := executeCommand(cmd, "assign", "REQ-NOPE-999", "--to", "alice")
		if err == nil {
			t.Fatal("expected error for nonexistent requirement")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unassign_clears_assignee", func(t *testing.T) {
		req := &database.Requirement{
			ReqID:    "REQ-TEST-001",
			Category: "TEST",
			Status:   database.StatusMissing,
			Assignee: "alice",
		}

		dir := setupAssignTestDir(t, []*database.Requirement{req})
		origDir, _ := os.Getwd()
		_ = os.Chdir(dir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := newAssignTestRoot()
		out, err := executeCommand(cmd, "unassign", "REQ-TEST-001")
		if err != nil {
			t.Fatalf("unassign failed: %v", err)
		}

		if !strings.Contains(out, "Unassigned REQ-TEST-001") {
			t.Errorf("unexpected output: %s", out)
		}

		db, err := database.Load(filepath.Join(dir, ".rtmx", "database.csv"))
		if err != nil {
			t.Fatal(err)
		}
		updated := db.Get("REQ-TEST-001")
		if updated.Assignee != "" {
			t.Errorf("Assignee = %q, want empty", updated.Assignee)
		}
	})

	t.Run("assign_preserves_existing_started_date", func(t *testing.T) {
		req := &database.Requirement{
			ReqID:       "REQ-TEST-001",
			Category:    "TEST",
			Status:      database.StatusMissing,
			StartedDate: "2026-01-01",
		}

		dir := setupAssignTestDir(t, []*database.Requirement{req})
		origDir, _ := os.Getwd()
		_ = os.Chdir(dir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := newAssignTestRoot()
		_, err := executeCommand(cmd, "assign", "REQ-TEST-001", "--to", "bob")
		if err != nil {
			t.Fatalf("assign failed: %v", err)
		}

		db, err := database.Load(filepath.Join(dir, ".rtmx", "database.csv"))
		if err != nil {
			t.Fatal(err)
		}
		updated := db.Get("REQ-TEST-001")
		if updated.StartedDate != "2026-01-01" {
			t.Errorf("StartedDate = %q, want %q (should be preserved)", updated.StartedDate, "2026-01-01")
		}
	})
}
