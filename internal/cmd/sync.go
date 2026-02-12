package cmd

import (
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx-go/internal/adapters"
	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	syncImport        bool
	syncExport        bool
	syncBidirectional bool
	syncDryRun        bool
	syncPreferLocal   bool
	syncPreferRemote  bool
)

var syncCmd = &cobra.Command{
	Use:   "sync {github|jira}",
	Short: "Synchronize RTM with external services",
	Long: `Bi-directional sync with GitHub Issues or Jira.

Examples:
    rtmx sync github --import          # Pull from GitHub into RTM
    rtmx sync github --export          # Push RTM to GitHub
    rtmx sync github --bidirectional   # Two-way sync
    rtmx sync jira --import --dry-run  # Preview Jira import`,
	Args: cobra.ExactArgs(1),
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncImport, "import", false, "pull from service into RTM")
	syncCmd.Flags().BoolVar(&syncExport, "export", false, "push RTM to service")
	syncCmd.Flags().BoolVar(&syncBidirectional, "bidirectional", false, "two-way sync")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "preview changes")
	syncCmd.Flags().BoolVar(&syncPreferLocal, "prefer-local", false, "RTM wins on conflicts")
	syncCmd.Flags().BoolVar(&syncPreferRemote, "prefer-remote", false, "service wins on conflicts")

	rootCmd.AddCommand(syncCmd)
}

// SyncResult holds the results of a sync operation.
type SyncResult struct {
	Created   []string
	Updated   []string
	Skipped   []string
	Conflicts []SyncConflict
	Errors    []SyncError
}

type SyncConflict struct {
	ItemID string
	Reason string
}

type SyncError struct {
	ItemID string
	Error  string
}

func (r *SyncResult) Summary() string {
	return fmt.Sprintf("%d created, %d updated, %d skipped, %d conflicts, %d errors",
		len(r.Created), len(r.Updated), len(r.Skipped), len(r.Conflicts), len(r.Errors))
}

func runSync(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	service := args[0]
	if service != "github" && service != "jira" {
		return fmt.Errorf("unknown service: %s (use 'github' or 'jira')", service)
	}

	if !syncImport && !syncExport && !syncBidirectional {
		cmd.Printf("%s\n", output.Color("No sync direction specified. Use --import, --export, or --bidirectional", output.Yellow))
		return NewExitError(1, "no sync direction specified")
	}

	if syncPreferLocal && syncPreferRemote {
		cmd.Printf("%s\n", output.Color("Cannot use both --prefer-local and --prefer-remote", output.Red))
		return NewExitError(1, "conflicting options")
	}

	cmd.Printf("=== RTMX Sync: %s ===\n", output.Color(service, output.Bold))
	cmd.Println()

	if syncDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no changes will be made", output.Yellow))
		cmd.Println()
	}

	// Determine sync mode
	mode := "import"
	if syncBidirectional || (syncImport && syncExport) {
		mode = "bidirectional"
	} else if syncExport {
		mode = "export"
	}

	// Determine conflict resolution
	conflictResolution := "ask"
	if syncPreferLocal {
		conflictResolution = "prefer-local"
	} else if syncPreferRemote {
		conflictResolution = "prefer-remote"
	}

	cmd.Printf("Mode: %s\n", mode)
	cmd.Printf("Conflict resolution: %s\n", conflictResolution)
	cmd.Println()

	// Load config
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get adapter
	adapter, err := getAdapter(service, cfg)
	if err != nil {
		cmd.Printf("%s %s\n", output.Color("Error:", output.Red), err)
		return nil
	}

	// Test connection
	cmd.Printf("%s\n", output.Color("Testing connection...", output.Bold))
	success, message := adapter.TestConnection()
	if !success {
		cmd.Printf("  %s %s\n", output.Color("✗", output.Red), message)
		return nil
	}
	cmd.Printf("  %s %s\n", output.Color("✓", output.Green), message)
	cmd.Println()

	// Load RTM database
	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil && mode != "import" {
		return fmt.Errorf("failed to load database: %w", err)
	}

	// Run sync based on mode
	var result *SyncResult
	switch mode {
	case "import":
		result = runSyncImport(cmd, adapter, db, syncDryRun)
	case "export":
		result = runSyncExport(cmd, adapter, db, syncDryRun)
	case "bidirectional":
		result = runSyncBidirectional(cmd, adapter, db, conflictResolution, syncDryRun)
	}

	// Print summary
	cmd.Println()
	cmd.Printf("%s\n", output.Color("Sync Summary:", output.Bold))
	cmd.Printf("  %s\n", result.Summary())

	if len(result.Conflicts) > 0 {
		cmd.Println()
		cmd.Printf("%s\n", output.Color("Conflicts requiring attention:", output.Yellow))
		for _, c := range result.Conflicts {
			cmd.Printf("  • %s: %s\n", c.ItemID, c.Reason)
		}
	}

	if len(result.Errors) > 0 {
		cmd.Println()
		cmd.Printf("%s\n", output.Color("Errors:", output.Red))
		for _, e := range result.Errors {
			cmd.Printf("  • %s: %s\n", e.ItemID, e.Error)
		}
	}

	return nil
}

func getAdapter(service string, cfg *config.Config) (adapters.ServiceAdapter, error) {
	switch service {
	case "github":
		if !cfg.RTMX.Adapters.GitHub.Enabled {
			return nil, fmt.Errorf("GitHub adapter not enabled in rtmx.yaml")
		}
		if cfg.RTMX.Adapters.GitHub.Repo == "" {
			return nil, fmt.Errorf("GitHub repo not configured in rtmx.yaml")
		}
		return adapters.NewGitHubAdapter(&cfg.RTMX.Adapters.GitHub)

	case "jira":
		if !cfg.RTMX.Adapters.Jira.Enabled {
			return nil, fmt.Errorf("Jira adapter not enabled in rtmx.yaml")
		}
		if cfg.RTMX.Adapters.Jira.Project == "" {
			return nil, fmt.Errorf("Jira project not configured in rtmx.yaml")
		}
		return adapters.NewJiraAdapter(&cfg.RTMX.Adapters.Jira)

	default:
		return nil, fmt.Errorf("unknown service: %s", service)
	}
}

func runSyncImport(cmd *cobra.Command, adapter adapters.ServiceAdapter, db *database.Database, dryRun bool) *SyncResult {
	result := &SyncResult{}

	cmd.Printf("%s\n", output.Color("Fetching items from "+adapter.Name()+"...", output.Bold))

	// Build external ID map
	externalIDMap := make(map[string]string)
	if db != nil {
		for _, req := range db.All() {
			if req.ExternalID != "" {
				externalIDMap[req.ExternalID] = req.ReqID
			}
		}
	}

	// Fetch external items
	itemsFound := 0
	items, err := adapter.FetchItems(nil)
	if err != nil {
		result.Errors = append(result.Errors, SyncError{ItemID: "fetch", Error: err.Error()})
		return result
	}

	for _, item := range items {
		itemsFound++

		// Check if already linked
		if reqID, ok := externalIDMap[item.ExternalID]; ok {
			req := db.Get(reqID)
			if req != nil {
				newStatus := adapter.MapStatusToRTMX(item.Status)
				if string(newStatus) != req.Status.String() {
					if dryRun {
						cmd.Printf("  Would update %s status: %s → %s\n", reqID, req.Status, newStatus)
					} else {
						cmd.Printf("  %s %s: %s → %s\n", output.Color("↻", output.Blue), reqID, req.Status, newStatus)
					}
					result.Updated = append(result.Updated, reqID)
				} else {
					result.Skipped = append(result.Skipped, item.ExternalID)
				}
			}
		} else {
			// New item - would need to create requirement
			if dryRun {
				cmd.Printf("  Would import: [%s] %s\n", item.ExternalID, truncateString(item.Title, 50))
			} else {
				cmd.Printf("  %s [%s] %s\n", output.Color("+", output.Green), item.ExternalID, truncateString(item.Title, 50))
			}
			result.Created = append(result.Created, item.ExternalID)
		}
	}

	cmd.Println()
	cmd.Printf("Found %d items in %s\n", itemsFound, adapter.Name())

	return result
}

func runSyncExport(cmd *cobra.Command, adapter adapters.ServiceAdapter, db *database.Database, dryRun bool) *SyncResult {
	result := &SyncResult{}

	cmd.Printf("%s\n", output.Color("Exporting requirements to "+adapter.Name()+"...", output.Bold))

	if db == nil {
		result.Errors = append(result.Errors, SyncError{ItemID: "database", Error: "RTM database not found"})
		return result
	}

	for _, req := range db.All() {
		if req.ExternalID != "" {
			// Already exported - update
			if dryRun {
				cmd.Printf("  Would update: %s → %s\n", req.ReqID, req.ExternalID)
			} else {
				success := adapter.UpdateItem(req.ExternalID, req)
				if success {
					cmd.Printf("  %s Updated %s → %s\n", output.Color("↻", output.Blue), req.ReqID, req.ExternalID)
					result.Updated = append(result.Updated, req.ReqID)
				} else {
					cmd.Printf("  %s Failed to update %s\n", output.Color("✗", output.Red), req.ReqID)
					result.Errors = append(result.Errors, SyncError{ItemID: req.ReqID, Error: "Update failed"})
				}
			}
		} else {
			// New export
			if dryRun {
				cmd.Printf("  Would export: %s\n", req.ReqID)
			} else {
				externalID, err := adapter.CreateItem(req)
				if err != nil {
					cmd.Printf("  %s Failed to export %s: %v\n", output.Color("✗", output.Red), req.ReqID, err)
					result.Errors = append(result.Errors, SyncError{ItemID: req.ReqID, Error: err.Error()})
				} else {
					cmd.Printf("  %s Exported %s → %s\n", output.Color("+", output.Green), req.ReqID, externalID)
					result.Created = append(result.Created, req.ReqID)
				}
			}
		}
	}

	return result
}

func runSyncBidirectional(cmd *cobra.Command, adapter adapters.ServiceAdapter, db *database.Database, conflictResolution string, dryRun bool) *SyncResult {
	result := &SyncResult{}

	cmd.Printf("%s\n", output.Color("Running bidirectional sync with "+adapter.Name()+"...", output.Bold))

	// Build maps
	externalIDMap := make(map[string]string)
	if db != nil {
		for _, req := range db.All() {
			if req.ExternalID != "" {
				externalIDMap[req.ExternalID] = req.ReqID
			}
		}
	}

	// Fetch external items
	cmd.Printf("\n%s\n", output.Color("Fetching external items...", output.Dim))
	items, err := adapter.FetchItems(nil)
	if err != nil {
		result.Errors = append(result.Errors, SyncError{ItemID: "fetch", Error: err.Error()})
		return result
	}

	externalItems := make(map[string]adapters.ExternalItem)
	for _, item := range items {
		externalItems[item.ExternalID] = item
	}

	reqCount := 0
	if db != nil {
		reqCount = db.Len()
	}
	cmd.Printf("Found %d external items\n", len(externalItems))
	cmd.Printf("Have %d local requirements\n", reqCount)
	cmd.Println()

	// Process linked items (check for conflicts)
	for externalID, reqID := range externalIDMap {
		if item, ok := externalItems[externalID]; ok {
			req := db.Get(reqID)
			if req == nil {
				continue
			}

			// Check for status conflict
			externalStatus := adapter.MapStatusToRTMX(item.Status)
			if string(externalStatus) != req.Status.String() {
				switch conflictResolution {
				case "prefer-local":
					if dryRun {
						cmd.Printf("  Would update %s: %s → %s\n", externalID, item.Status, req.Status)
					} else {
						adapter.UpdateItem(externalID, req)
						cmd.Printf("  %s %s: Local wins (%s)\n", output.Color("↻", output.Blue), reqID, req.Status)
					}
					result.Updated = append(result.Updated, reqID)
				case "prefer-remote":
					if dryRun {
						cmd.Printf("  Would update %s: %s → %s\n", reqID, req.Status, externalStatus)
					} else {
						cmd.Printf("  %s %s: Remote wins (%s)\n", output.Color("↻", output.Blue), reqID, externalStatus)
					}
					result.Updated = append(result.Updated, reqID)
				default:
					cmd.Printf("  %s Conflict: %s (local=%s, remote=%s)\n",
						output.Color("?", output.Yellow), reqID, req.Status, externalStatus)
					result.Conflicts = append(result.Conflicts, SyncConflict{
						ItemID: reqID,
						Reason: fmt.Sprintf("Status conflict: %s vs %s", req.Status, externalStatus),
					})
				}
			} else {
				result.Skipped = append(result.Skipped, reqID)
			}

			// Remove from external items to track what's left
			delete(externalItems, externalID)
		}
	}

	// Items only in external service (import candidates)
	for externalID, item := range externalItems {
		if dryRun {
			cmd.Printf("  Would import: [%s] %s\n", externalID, truncateString(item.Title, 50))
		} else {
			cmd.Printf("  %s Import candidate: [%s] %s\n", output.Color("←", output.Green), externalID, truncateString(item.Title, 50))
		}
		result.Created = append(result.Created, externalID)
	}

	// Requirements not in external service (export candidates)
	if db != nil {
		for _, req := range db.All() {
			if req.ExternalID == "" {
				if dryRun {
					cmd.Printf("  Would export: %s\n", req.ReqID)
				} else {
					cmd.Printf("  %s Export candidate: %s\n", output.Color("→", output.Green), req.ReqID)
				}
			}
		}
	}

	return result
}
