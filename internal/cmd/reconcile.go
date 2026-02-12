package cmd

import (
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var reconcileExecute bool

var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Check and fix dependency reciprocity",
	Long: `Ensure dependency/blocks relationships are consistent.

When A depends on B, B should block A. This command finds and
optionally fixes missing reciprocal relationships.

Examples:
    rtmx reconcile              # Dry-run, show what would be fixed
    rtmx reconcile --execute    # Actually fix the issues`,
	RunE: runReconcile,
}

func init() {
	reconcileCmd.Flags().BoolVar(&reconcileExecute, "execute", false, "execute fixes (default is dry-run)")

	rootCmd.AddCommand(reconcileCmd)
}

func runReconcile(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

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

	width := 80
	cmd.Println(output.Header("Dependency Reconciliation", width))
	cmd.Println()

	// Find missing reciprocity
	type fix struct {
		reqID     string
		dep       string
		fixType   string // "add_block" or "add_dep"
	}

	var fixes []fix

	// Check: if A depends on B, then B should block A
	for _, req := range db.All() {
		for dep := range req.Dependencies {
			depReq := db.Get(dep)
			if depReq == nil {
				continue // Skip non-existent deps
			}
			if !depReq.Blocks.Contains(req.ReqID) {
				fixes = append(fixes, fix{
					reqID:   dep,
					dep:     req.ReqID,
					fixType: "add_block",
				})
			}
		}
	}

	// Check: if A blocks B, then B should depend on A
	for _, req := range db.All() {
		for blocked := range req.Blocks {
			blockedReq := db.Get(blocked)
			if blockedReq == nil {
				continue
			}
			if !blockedReq.Dependencies.Contains(req.ReqID) {
				fixes = append(fixes, fix{
					reqID:   blocked,
					dep:     req.ReqID,
					fixType: "add_dep",
				})
			}
		}
	}

	if len(fixes) == 0 {
		cmd.Printf("%s All dependencies are reciprocal!\n", output.Color("✓", output.Green))
		cmd.Println()
		cmd.Println("No fixes needed.")
		return nil
	}

	// Display fixes
	cmd.Printf("Found %d reciprocity issue(s):\n\n", len(fixes))

	addBlockFixes := 0
	addDepFixes := 0

	for _, f := range fixes {
		if f.fixType == "add_block" {
			cmd.Printf("  %s %s should block %s\n",
				output.Color("→", output.Yellow), f.reqID, f.dep)
			addBlockFixes++
		} else {
			cmd.Printf("  %s %s should depend on %s\n",
				output.Color("→", output.Yellow), f.reqID, f.dep)
			addDepFixes++
		}
	}

	cmd.Println()
	cmd.Printf("Summary: %d missing blocks, %d missing dependencies\n\n", addBlockFixes, addDepFixes)

	if !reconcileExecute {
		cmd.Println(output.Color("Dry-run mode - no changes made.", output.Cyan))
		cmd.Println("Run with --execute to apply fixes.")
		return nil
	}

	// Execute fixes
	cmd.Println("Applying fixes...")

	for _, f := range fixes {
		req := db.Get(f.reqID)
		if req == nil {
			continue
		}

		if f.fixType == "add_block" {
			req.Blocks.Add(f.dep)
			cmd.Printf("  %s Added block: %s -> %s\n",
				output.Color("✓", output.Green), f.reqID, f.dep)
		} else {
			req.Dependencies.Add(f.dep)
			cmd.Printf("  %s Added dependency: %s -> %s\n",
				output.Color("✓", output.Green), f.reqID, f.dep)
		}
	}

	// Save database
	if err := db.Save(dbPath); err != nil {
		return fmt.Errorf("failed to save database: %w", err)
	}

	cmd.Println()
	cmd.Printf("%s Applied %d fixes and saved database.\n",
		output.Color("✓", output.Green), len(fixes))

	return nil
}
