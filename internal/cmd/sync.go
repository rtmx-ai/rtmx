package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rtmx-ai/rtmx-go/internal/adapters"
	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
)

var (
	syncService      string
	syncImport       bool
	syncExport       bool
	syncBidirect     bool
	syncDryRun       bool
	syncPreferLocal  bool
	syncPreferRemote bool
)

// SyncResult holds the results of a sync operation
type SyncResult struct {
	Created   []string
	Updated   []string
	Skipped   []string
	Conflicts []SyncConflict
	Errors    []SyncError
}

// SyncConflict represents a conflict during sync
type SyncConflict struct {
	ID     string
	Reason string
}

// SyncError represents an error during sync
type SyncError struct {
	ID    string
	Error string
}

// Summary returns a summary string of the sync result
func (r *SyncResult) Summary() string {
	parts := []string{}
	if len(r.Created) > 0 {
		parts = append(parts, fmt.Sprintf("%d created", len(r.Created)))
	}
	if len(r.Updated) > 0 {
		parts = append(parts, fmt.Sprintf("%d updated", len(r.Updated)))
	}
	if len(r.Skipped) > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", len(r.Skipped)))
	}
	if len(r.Conflicts) > 0 {
		parts = append(parts, fmt.Sprintf("%d conflicts", len(r.Conflicts)))
	}
	if len(r.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", len(r.Errors)))
	}
	if len(parts) == 0 {
		return "No changes"
	}
	return strings.Join(parts, ", ")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize RTM with external services",
	Long: `Synchronize requirements with GitHub Issues or Jira tickets.

Supports bidirectional sync with conflict resolution strategies.

Examples:
  # Import issues from GitHub
  rtmx sync --service github --import

  # Export requirements to Jira
  rtmx sync --service jira --export

  # Bidirectional sync with local preference
  rtmx sync --service github --bidirectional --prefer-local

  # Preview changes without writing
  rtmx sync --service github --import --dry-run`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().StringVarP(&syncService, "service", "s", "github", "service to sync with (github, jira)")
	syncCmd.Flags().BoolVarP(&syncImport, "import", "i", false, "pull from service into RTM")
	syncCmd.Flags().BoolVarP(&syncExport, "export", "e", false, "push RTM to service")
	syncCmd.Flags().BoolVarP(&syncBidirect, "bidirectional", "b", false, "two-way sync")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "preview changes without writing")
	syncCmd.Flags().BoolVar(&syncPreferLocal, "prefer-local", false, "RTM wins on conflicts")
	syncCmd.Flags().BoolVar(&syncPreferRemote, "prefer-remote", false, "service wins on conflicts")

	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Validate flags
	if !syncImport && !syncExport && !syncBidirect {
		fmt.Printf("%sNo sync direction specified. Use --import, --export, or --bidirectional%s\n",
			output.Yellow, output.Reset)
		return NewExitError(1, "no sync direction specified")
	}

	if syncPreferLocal && syncPreferRemote {
		fmt.Printf("%sCannot use both --prefer-local and --prefer-remote%s\n",
			output.Red, output.Reset)
		return NewExitError(1, "conflicting preferences")
	}

	// Header
	fmt.Printf("=== RTMX Sync: %s ===\n\n", strings.ToUpper(syncService))

	if syncDryRun {
		fmt.Printf("%sDRY RUN - no changes will be made%s\n\n", output.Yellow, output.Reset)
	}

	// Determine mode
	mode := "import"
	if syncBidirect || (syncImport && syncExport) {
		mode = "bidirectional"
	} else if syncExport {
		mode = "export"
	}

	// Determine conflict resolution
	conflictRes := "ask"
	if syncPreferLocal {
		conflictRes = "prefer-local"
	} else if syncPreferRemote {
		conflictRes = "prefer-remote"
	}

	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Conflict resolution: %s\n\n", conflictRes)

	// Load config
	cfg, err := config.LoadFromDir(".")
	if err != nil {
		fmt.Printf("%sWarning: Could not load config, using defaults%s\n", output.Yellow, output.Reset)
		cfg = config.DefaultConfig()
	}

	// Get adapter
	adapter, err := getAdapter(syncService, cfg)
	if err != nil {
		fmt.Printf("%s✗%s %v\n", output.Red, output.Reset, err)
		return NewExitError(1, err.Error())
	}

	// Test connection
	fmt.Printf("%sTesting connection...%s\n", output.Bold, output.Reset)
	success, message := adapter.TestConnection()
	if !success {
		fmt.Printf("  %s✗%s %s\n", output.Red, output.Reset, message)
		return NewExitError(1, "connection failed")
	}
	fmt.Printf("  %s✓%s %s\n\n", output.Green, output.Reset, message)

	// Run sync
	var result *SyncResult
	switch mode {
	case "import":
		result = runImport(adapter, cfg, syncDryRun)
	case "export":
		result = runExport(adapter, cfg, syncDryRun)
	default:
		result = runBidirectional(adapter, cfg, conflictRes, syncDryRun)
	}

	// Print summary
	printSyncSummary(result)

	if len(result.Errors) > 0 {
		return NewExitError(1, "sync completed with errors")
	}

	return nil
}

func getAdapter(service string, cfg *config.Config) (adapters.ServiceAdapter, error) {
	switch service {
	case "github":
		if !cfg.RTMX.Adapters.GitHub.Enabled {
			return nil, fmt.Errorf("GitHub adapter not enabled in rtmx.yaml")
		}
		return adapters.NewGitHubAdapter(&cfg.RTMX.Adapters.GitHub)

	case "jira":
		if !cfg.RTMX.Adapters.Jira.Enabled {
			return nil, fmt.Errorf("Jira adapter not enabled in rtmx.yaml")
		}
		return adapters.NewJiraAdapter(&cfg.RTMX.Adapters.Jira)

	default:
		return nil, fmt.Errorf("unknown service: %s", service)
	}
}

func runImport(adapter adapters.ServiceAdapter, cfg *config.Config, dryRun bool) *SyncResult {
	result := &SyncResult{}

	fmt.Printf("%sFetching items from %s...%s\n", output.Bold, adapter.Name(), output.Reset)

	// Load existing RTM
	dbPath := cfg.RTMX.Database
	if dbPath == "" {
		dbPath = ".rtmx/database.csv"
	}

	requirements := make(map[string]*database.Requirement)
	externalIDMap := make(map[string]string) // external_id -> req_id

	if _, err := os.Stat(dbPath); err == nil {
		db, err := database.Load(dbPath)
		if err == nil {
			for _, req := range db.All() {
				requirements[req.ReqID] = req
				if req.ExternalID != "" {
					externalIDMap[req.ExternalID] = req.ReqID
				}
			}
		}
	}

	// Fetch external items
	items, err := adapter.FetchItems(nil)
	if err != nil {
		result.Errors = append(result.Errors, SyncError{ID: "", Error: err.Error()})
		return result
	}

	for _, item := range items {
		// Check if already linked
		if reqID, ok := externalIDMap[item.ExternalID]; ok {
			req := requirements[reqID]

			// Update status from external
			newStatus := adapter.MapStatusToRTMX(item.Status)
			if newStatus != req.Status {
				if dryRun {
					fmt.Printf("  Would update %s status: %s → %s\n", reqID, req.Status, newStatus)
				} else {
					fmt.Printf("  %s↻%s %s: %s → %s\n", output.Blue, output.Reset, reqID, req.Status, newStatus)
				}
				result.Updated = append(result.Updated, reqID)
			} else {
				result.Skipped = append(result.Skipped, item.ExternalID)
			}

		} else if item.RequirementID != "" {
			// Item references a requirement we have
			if _, ok := requirements[item.RequirementID]; ok {
				if dryRun {
					fmt.Printf("  Would link %s to %s\n", item.RequirementID, item.ExternalID)
				} else {
					fmt.Printf("  %s⇄%s Linked %s ↔ %s\n", output.Green, output.Reset, item.RequirementID, item.ExternalID)
				}
				result.Updated = append(result.Updated, item.RequirementID)
			}
		} else {
			// New item - import candidate
			title := item.Title
			if len(title) > 50 {
				title = title[:50] + "..."
			}
			if dryRun {
				fmt.Printf("  Would import: [%s] %s\n", item.ExternalID, title)
			} else {
				fmt.Printf("  %s+%s [%s] %s\n", output.Green, output.Reset, item.ExternalID, title)
			}
			result.Created = append(result.Created, item.ExternalID)
		}
	}

	fmt.Printf("\nFound %d items in %s\n", len(items), adapter.Name())

	return result
}

func runExport(adapter adapters.ServiceAdapter, cfg *config.Config, dryRun bool) *SyncResult {
	result := &SyncResult{}

	fmt.Printf("%sExporting requirements to %s...%s\n", output.Bold, adapter.Name(), output.Reset)

	// Load RTM
	dbPath := cfg.RTMX.Database
	if dbPath == "" {
		dbPath = ".rtmx/database.csv"
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("%sRTM database not found: %s%s\n", output.Red, dbPath, output.Reset)
		result.Errors = append(result.Errors, SyncError{ID: "", Error: "database not found"})
		return result
	}

	db, err := database.Load(dbPath)
	if err != nil {
		result.Errors = append(result.Errors, SyncError{ID: "", Error: err.Error()})
		return result
	}

	for _, req := range db.All() {
		if req.ExternalID != "" {
			// Already exported - update
			if dryRun {
				fmt.Printf("  Would update: %s → %s\n", req.ReqID, req.ExternalID)
			} else {
				success := adapter.UpdateItem(req.ExternalID, req)
				if success {
					fmt.Printf("  %s↻%s Updated %s → %s\n", output.Blue, output.Reset, req.ReqID, req.ExternalID)
					result.Updated = append(result.Updated, req.ReqID)
				} else {
					fmt.Printf("  %s✗%s Failed to update %s\n", output.Red, output.Reset, req.ReqID)
					result.Errors = append(result.Errors, SyncError{ID: req.ReqID, Error: "update failed"})
				}
			}
		} else {
			// New export
			if dryRun {
				fmt.Printf("  Would export: %s\n", req.ReqID)
			} else {
				externalID, err := adapter.CreateItem(req)
				if err != nil {
					fmt.Printf("  %s✗%s Failed to export %s: %v\n", output.Red, output.Reset, req.ReqID, err)
					result.Errors = append(result.Errors, SyncError{ID: req.ReqID, Error: err.Error()})
				} else {
					fmt.Printf("  %s+%s Exported %s → %s\n", output.Green, output.Reset, req.ReqID, externalID)
					result.Created = append(result.Created, req.ReqID)
				}
			}
		}
	}

	return result
}

func runBidirectional(adapter adapters.ServiceAdapter, cfg *config.Config, conflictRes string, dryRun bool) *SyncResult {
	result := &SyncResult{}

	fmt.Printf("%sRunning bidirectional sync with %s...%s\n", output.Bold, adapter.Name(), output.Reset)

	// Load RTM
	dbPath := cfg.RTMX.Database
	if dbPath == "" {
		dbPath = ".rtmx/database.csv"
	}

	requirements := make(map[string]*database.Requirement)
	externalIDMap := make(map[string]string)

	if _, err := os.Stat(dbPath); err == nil {
		db, err := database.Load(dbPath)
		if err == nil {
			for _, req := range db.All() {
				requirements[req.ReqID] = req
				if req.ExternalID != "" {
					externalIDMap[req.ExternalID] = req.ReqID
				}
			}
		}
	}

	// Fetch external items
	fmt.Printf("\n%sFetching external items...%s\n", output.Dim, output.Reset)
	items, err := adapter.FetchItems(nil)
	if err != nil {
		result.Errors = append(result.Errors, SyncError{ID: "", Error: err.Error()})
		return result
	}

	externalItems := make(map[string]adapters.ExternalItem)
	for _, item := range items {
		externalItems[item.ExternalID] = item
	}

	fmt.Printf("Found %d external items\n", len(externalItems))
	fmt.Printf("Have %d local requirements\n\n", len(requirements))

	// Process linked items
	for externalID, reqID := range externalIDMap {
		if item, ok := externalItems[externalID]; ok {
			req := requirements[reqID]

			// Check for status conflict
			externalStatus := adapter.MapStatusToRTMX(item.Status)
			if externalStatus != req.Status {
				switch conflictRes {
				case "prefer-local":
					if dryRun {
						fmt.Printf("  Would update %s: %s → %s\n", externalID, item.Status, req.Status)
					} else {
						adapter.UpdateItem(externalID, req)
						fmt.Printf("  %s↻%s %s: Local wins (%s)\n", output.Blue, output.Reset, reqID, req.Status)
					}
					result.Updated = append(result.Updated, reqID)

				case "prefer-remote":
					if dryRun {
						fmt.Printf("  Would update %s: %s → %s\n", reqID, req.Status, externalStatus)
					} else {
						fmt.Printf("  %s↻%s %s: Remote wins (%s)\n", output.Blue, output.Reset, reqID, externalStatus)
					}
					result.Updated = append(result.Updated, reqID)

				default:
					fmt.Printf("  %s?%s Conflict: %s (local=%s, remote=%s)\n",
						output.Yellow, output.Reset, reqID, req.Status, externalStatus)
					result.Conflicts = append(result.Conflicts, SyncConflict{
						ID:     reqID,
						Reason: fmt.Sprintf("Status conflict: %s vs %s", req.Status, externalStatus),
					})
				}
			} else {
				result.Skipped = append(result.Skipped, reqID)
			}

			delete(externalItems, externalID)
		}
	}

	// Items only in external service (import candidates)
	for externalID, item := range externalItems {
		title := item.Title
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		if dryRun {
			fmt.Printf("  Would import: [%s] %s\n", externalID, title)
		} else {
			fmt.Printf("  %s←%s Import candidate: [%s] %s\n", output.Green, output.Reset, externalID, title)
		}
		result.Created = append(result.Created, externalID)
	}

	// Requirements not in external service (export candidates)
	exportedIDs := make(map[string]bool)
	for _, reqID := range externalIDMap {
		exportedIDs[reqID] = true
	}

	for reqID := range requirements {
		if !exportedIDs[reqID] {
			if dryRun {
				fmt.Printf("  Would export: %s\n", reqID)
			} else {
				fmt.Printf("  %s→%s Export candidate: %s\n", output.Green, output.Reset, reqID)
			}
		}
	}

	return result
}

func printSyncSummary(result *SyncResult) {
	fmt.Printf("\n%sSync Summary:%s\n", output.Bold, output.Reset)
	fmt.Printf("  %s\n", result.Summary())

	if len(result.Conflicts) > 0 {
		fmt.Printf("\n%sConflicts requiring attention:%s\n", output.Yellow, output.Reset)
		for _, c := range result.Conflicts {
			fmt.Printf("  • %s: %s\n", c.ID, c.Reason)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n%sErrors:%s\n", output.Red, output.Reset)
		for _, e := range result.Errors {
			fmt.Printf("  • %s: %s\n", e.ID, e.Error)
		}
	}
}
