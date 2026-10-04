package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/spf13/cobra"
)

var (
	loopOnce bool
	loopJSON bool

	// loopEdgesOverride, when non-nil, replaces falling-edge detection for --once.
	// Tests inject edges. The foreground watcher uses loopMergedSource instead.
	loopEdgesOverride *[]LoopEdge

	// loopMergedSource lists merged PRs for the watcher. Nil uses gh.
	loopMergedSource PRSource

	// loopPollLimit stops the foreground watcher after N polls. 0 watches until interrupted.
	loopPollLimit int

	// loopWait pauses between polls. Tests replace it so the watcher does not sleep.
	loopWait = func() { time.Sleep(30 * time.Second) }
)

var loopCmd = &cobra.Command{
	Use:   "loop",
	Short: "Watch merge edges and plan the next requirement",
	Long: `Run the delivery loop in the foreground.

rtmx loop blocks, watches for falling edges (a requirement merged), and
runs one tick per edge. On first start, with no stored cursor and work
still unblocked, it runs one bootstrap tick and then waits.

rtmx loop --once runs at most one tick and exits 0. When nothing is
unblocked the plan is {"idle":true}.

rtmx loop install writes a LaunchAgent or systemd user unit. It does
not load or enable that unit.

The tick does not create commits, branches, or pull requests.`,
	RunE: runLoop,
}

func init() {
	loopCmd.Flags().BoolVar(&loopOnce, "once", false, "run at most one tick and exit")
	loopCmd.Flags().BoolVar(&loopJSON, "json", false, "print the JSON plan (always on for a tick)")
	rootCmd.AddCommand(loopCmd)
}

func runLoop(cmd *cobra.Command, args []string) error {
	_ = loopJSON
	if loopOnce {
		return runLoopOnce(cmd)
	}
	return runLoopWatch(cmd)
}

func runLoopWatch(cmd *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(LoopCursorPath(cwd)); os.IsNotExist(err) {
		if hasUnblockedWork(cmd, cwd) {
			if err := runDeliveryTick(cmd); err != nil {
				return err
			}
		}
		if err := seedLoopCursor(cmd, cwd); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	polls := 0
	for {
		if err := pollLoop(cmd, cwd); err != nil {
			return err
		}
		loopWait()
		polls++
		if loopPollLimit > 0 && polls >= loopPollLimit {
			return nil
		}
	}
}

func pollLoop(cmd *cobra.Command, cwd string) error {
	cursor, err := LoadLoopCursor(LoopCursorPath(cwd))
	if err != nil {
		return err
	}
	merged, err := listMergedPRs()
	if err != nil {
		cmd.PrintErrf("warning: merged PR lookup unavailable (%v); waiting\n", err)
		return nil
	}
	db, _, err := loadLoopDB(cwd)
	if err != nil {
		return err
	}
	edges, next := DetectFallingEdges(cursor, merged, completeIDs(db))
	for range edges {
		if err := runDeliveryTick(cmd); err != nil {
			return err
		}
	}
	return SaveLoopCursor(LoopCursorPath(cwd), next)
}

func seedLoopCursor(cmd *cobra.Command, cwd string) error {
	merged, err := listMergedPRs()
	if err != nil {
		cmd.PrintErrf("warning: merged PR lookup unavailable (%v); seeding cursor without PRs\n", err)
		merged = nil
	}
	db, _, err := loadLoopDB(cwd)
	if err != nil {
		return err
	}
	_ = EnsureLoopGitignore(filepath.Join(cwd, ".rtmx"))
	return SaveLoopCursor(LoopCursorPath(cwd), seedCursor(merged, completeIDs(db)))
}

func seedCursor(merged []PullRequest, complete []string) LoopCursor {
	c := LoopCursor{CompleteIDs: append([]string{}, complete...)}
	sort.Strings(c.CompleteIDs)
	seen := map[int]bool{}
	for _, pr := range merged {
		if !pr.Merged || seen[pr.Number] {
			continue
		}
		seen[pr.Number] = true
		c.MergedPRNumbers = append(c.MergedPRNumbers, pr.Number)
		if pr.SHA != "" {
			c.MergedSHAs = append(c.MergedSHAs, pr.SHA)
		}
	}
	return c
}

func hasUnblockedWork(cmd *cobra.Command, cwd string) bool {
	db, _, err := loadLoopDB(cwd)
	if err != nil {
		return false
	}
	return selectNextUnclaimed(cmd, db, cwd) != nil
}

func loadLoopDB(cwd string) (*database.Database, string, error) {
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}
	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load database: %w", err)
	}
	return db, dbPath, nil
}

func completeIDs(db *database.Database) []string {
	var ids []string
	for _, req := range db.All() {
		if req.Status == database.StatusComplete {
			ids = append(ids, req.ReqID)
		}
	}
	sort.Strings(ids)
	return ids
}

func listMergedPRs() ([]PullRequest, error) {
	src := loopMergedSource
	if src == nil {
		src = ghMergedPRSource{}
	}
	return src.ListMerged()
}

type ghMergedPRSource struct{}

func (ghMergedPRSource) ListMerged() ([]PullRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, "gh", "pr", "list", "--state", "merged", "--json", "number,title,body,state,mergeCommit", "--limit", "50")
	out, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}
	var rows []struct {
		Number      int    `json:"number"`
		Title       string `json:"title"`
		Body        string `json:"body"`
		State       string `json:"state"`
		MergeCommit struct {
			OID string `json:"oid"`
		} `json:"mergeCommit"`
	}
	if err := json.Unmarshal(out, &rows); err != nil {
		return nil, fmt.Errorf("gh pr list json: %w", err)
	}
	prs := make([]PullRequest, 0, len(rows))
	for _, row := range rows {
		prs = append(prs, PullRequest{
			Number: row.Number,
			Title:  row.Title,
			Body:   row.Body,
			Merged: row.State == "MERGED",
			SHA:    row.MergeCommit.OID,
		})
	}
	return prs, nil
}
