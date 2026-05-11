package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/spf13/cobra"
)

var assignTo string

var assignCmd = &cobra.Command{
	Use:   "assign <req-id>",
	Short: "Assign a requirement to a user",
	Long: `Set the assignee field on a requirement.

If --to is omitted, uses the current authenticated user from the stored
OIDC token. Also sets started_date if not already set.

Examples:
    rtmx assign REQ-PLAN-001 --to alice
    rtmx assign REQ-PLAN-001          # uses current user`,
	Args: cobra.ExactArgs(1),
	RunE: runAssign,
}

var unassignCmd = &cobra.Command{
	Use:   "unassign <req-id>",
	Short: "Clear the assignee from a requirement",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnassign,
}

func init() {
	assignCmd.Flags().StringVar(&assignTo, "to", "", "user to assign (defaults to current authenticated user)")
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(unassignCmd)
}

func runAssign(cmd *cobra.Command, args []string) error {
	reqID := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	req := db.Get(reqID)
	if req == nil {
		return fmt.Errorf("requirement not found: %s", reqID)
	}

	assignee := assignTo
	if assignee == "" {
		home, _ := os.UserHomeDir()
		tokenPath := filepath.Join(home, ".rtmx", "auth", "tokens.json")
		identity, err := auth.CurrentUser(tokenPath)
		if err != nil || identity.Email == "" {
			return fmt.Errorf("no --to specified and no authenticated user found; use --to <user> or run rtmx auth login")
		}
		assignee = identity.Email
	}

	updates := map[string]interface{}{
		"assignee": assignee,
	}
	if req.StartedDate == "" {
		updates["started_date"] = time.Now().Format("2006-01-02")
	}

	if err := db.Update(reqID, updates); err != nil {
		return fmt.Errorf("failed to update requirement: %w", err)
	}

	if err := db.Save(dbPath); err != nil {
		return fmt.Errorf("failed to save database: %w", err)
	}

	cmd.Printf("Assigned %s to %s\n", reqID, assignee)
	return nil
}

func runUnassign(cmd *cobra.Command, args []string) error {
	reqID := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	if !db.Exists(reqID) {
		return fmt.Errorf("requirement not found: %s", reqID)
	}

	if err := db.Update(reqID, map[string]interface{}{
		"assignee": "",
	}); err != nil {
		return fmt.Errorf("failed to update requirement: %w", err)
	}

	if err := db.Save(dbPath); err != nil {
		return fmt.Errorf("failed to save database: %w", err)
	}

	cmd.Printf("Unassigned %s\n", reqID)
	return nil
}
