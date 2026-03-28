package cmd

import (
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/sync"
	"github.com/spf13/cobra"
)

var (
	moveTo     string
	moveID     string
	moveDryRun bool
	moveBranch string
	movePR     bool
)

var moveCmd = &cobra.Command{
	Use:   "move REQ-ID",
	Short: "Transfer a requirement to another rtmx-enabled repo",
	Long: `Move transfers a requirement from the current project to a target rtmx-enabled
repository. Both repos get bidirectional provenance links via external_id.

The source requirement is kept as a reference with an external_id pointer.
The target repo receives a full copy of the requirement row and spec file.

Examples:
  rtmx move REQ-WEB-005 --to /path/to/other-repo
  rtmx move REQ-WEB-005 --to /path/to/other-repo --id REQ-CLI-001
  rtmx move REQ-WEB-005 --to /path/to/other-repo --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runMove,
}

var cloneCmd = &cobra.Command{
	Use:   "clone REQ-ID",
	Short: "Fork a requirement into another rtmx-enabled repo",
	Long: `Clone creates a copy of a requirement in a target rtmx-enabled repository
while preserving the original in the current project. Both repos get
bidirectional provenance links via external_id.

Both requirements maintain independent status tracking after the clone.

Examples:
  rtmx clone REQ-WEB-005 --to /path/to/other-repo
  rtmx clone REQ-WEB-005 --to /path/to/other-repo --id REQ-CLI-001
  rtmx clone REQ-WEB-005 --to /path/to/other-repo --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runClone,
}

func init() {
	moveCmd.Flags().StringVar(&moveTo, "to", "", "target repo path (required)")
	moveCmd.Flags().StringVar(&moveID, "id", "", "override target requirement ID")
	moveCmd.Flags().BoolVar(&moveDryRun, "dry-run", false, "preview changes without writing")
	moveCmd.Flags().StringVar(&moveBranch, "branch", "", "create branch in target repo")
	moveCmd.Flags().BoolVar(&movePR, "pr", false, "create pull request after writing")
	_ = moveCmd.MarkFlagRequired("to")

	cloneCmd.Flags().StringVar(&moveTo, "to", "", "target repo path (required)")
	cloneCmd.Flags().StringVar(&moveID, "id", "", "override target requirement ID")
	cloneCmd.Flags().BoolVar(&moveDryRun, "dry-run", false, "preview changes without writing")
	cloneCmd.Flags().StringVar(&moveBranch, "branch", "", "create branch in target repo")
	cloneCmd.Flags().BoolVar(&movePR, "pr", false, "create pull request after writing")
	_ = cloneCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(moveCmd)
	rootCmd.AddCommand(cloneCmd)
}

func loadSourceDB(cmd *cobra.Command) (*database.Database, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load source database: %w", err)
	}

	return db, cwd, nil
}

func loadTargetDB(targetDir string) (*database.Database, error) {
	dbPath, err := database.FindDatabase(targetDir)
	if err != nil {
		return nil, fmt.Errorf("target is not an rtmx-enabled project: %w", err)
	}

	db, err := database.Load(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load target database: %w", err)
	}

	return db, nil
}

func runMove(cmd *cobra.Command, args []string) error {
	reqID := args[0]

	srcDB, srcDir, err := loadSourceDB(cmd)
	if err != nil {
		return err
	}

	dstDB, err := loadTargetDB(moveTo)
	if err != nil {
		return err
	}

	opts := sync.CrossRepoOptions{
		SrcDir:   srcDir,
		DstDir:   moveTo,
		TargetID: moveID,
		DryRun:   moveDryRun,
		Branch:   moveBranch,
		PR:       movePR,
	}

	result, err := sync.MoveRequirement(srcDB, dstDB, reqID, opts)
	if err != nil {
		return err
	}

	if result.DryRun {
		cmd.Println("[dry-run] Move preview:")
		cmd.Printf("  Source: %s -> external_id: %s\n", reqID, result.SourceExternalID)
		cmd.Printf("  Target: %s -> external_id: %s\n", result.MovedID, result.TargetExternalID)
		if result.BranchCreated {
			cmd.Printf("  Branch: %s (would be created)\n", moveBranch)
		}
		if result.PRCreated {
			cmd.Println("  PR: would be created")
		}
		return nil
	}

	// Save both databases
	if err := srcDB.Save(""); err != nil {
		return fmt.Errorf("failed to save source database: %w", err)
	}
	if err := dstDB.Save(""); err != nil {
		return fmt.Errorf("failed to save target database: %w", err)
	}

	cmd.Printf("Moved %s -> %s\n", reqID, result.MovedID)
	cmd.Printf("  Source external_id: %s\n", result.SourceExternalID)
	cmd.Printf("  Target external_id: %s\n", result.TargetExternalID)
	if result.SpecFileCopied {
		cmd.Println("  Spec file copied")
	}
	if result.BranchCreated {
		cmd.Printf("  Branch: %s\n", moveBranch)
	}
	if result.PRCreated {
		cmd.Printf("  PR: %s\n", result.PRURL)
	}

	return nil
}

func runClone(cmd *cobra.Command, args []string) error {
	reqID := args[0]

	srcDB, srcDir, err := loadSourceDB(cmd)
	if err != nil {
		return err
	}

	dstDB, err := loadTargetDB(moveTo)
	if err != nil {
		return err
	}

	opts := sync.CrossRepoOptions{
		SrcDir:   srcDir,
		DstDir:   moveTo,
		TargetID: moveID,
		DryRun:   moveDryRun,
		Branch:   moveBranch,
		PR:       movePR,
	}

	result, err := sync.CloneRequirement(srcDB, dstDB, reqID, opts)
	if err != nil {
		return err
	}

	if result.DryRun {
		cmd.Println("[dry-run] Clone preview:")
		cmd.Printf("  Source: %s -> external_id: %s\n", reqID, result.SourceExternalID)
		cmd.Printf("  Target: %s -> external_id: %s\n", result.ClonedID, result.TargetExternalID)
		if result.BranchCreated {
			cmd.Printf("  Branch: %s (would be created)\n", moveBranch)
		}
		if result.PRCreated {
			cmd.Println("  PR: would be created")
		}
		return nil
	}

	// Save both databases
	if err := srcDB.Save(""); err != nil {
		return fmt.Errorf("failed to save source database: %w", err)
	}
	if err := dstDB.Save(""); err != nil {
		return fmt.Errorf("failed to save target database: %w", err)
	}

	cmd.Printf("Cloned %s -> %s\n", reqID, result.ClonedID)
	cmd.Printf("  Source external_id: %s\n", result.SourceExternalID)
	cmd.Printf("  Target external_id: %s\n", result.TargetExternalID)
	if result.SpecFileCopied {
		cmd.Println("  Spec file copied")
	}
	if result.BranchCreated {
		cmd.Printf("  Branch: %s\n", moveBranch)
	}
	if result.PRCreated {
		cmd.Printf("  PR: %s\n", result.PRURL)
	}

	return nil
}
