package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/adapters"
	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// mockAdapter implements adapters.ServiceAdapter for testing sync functions.
type mockAdapter struct {
	name            string
	connected       bool
	connMsg         string
	items           []adapters.ExternalItem
	fetchErr        error
	createResult    string
	createErr       error
	updateResult    bool
	statusMapping   map[string]database.Status
	statusReverse   map[database.Status]string
	createCalls     int
	updateCalls     int
	updateCallIDs   []string
}

func (m *mockAdapter) Name() string { return m.name }
func (m *mockAdapter) IsConfigured() bool { return true }
func (m *mockAdapter) TestConnection() (bool, string) { return m.connected, m.connMsg }
func (m *mockAdapter) FetchItems(_ map[string]interface{}) ([]adapters.ExternalItem, error) {
	return m.items, m.fetchErr
}
func (m *mockAdapter) GetItem(externalID string) (*adapters.ExternalItem, error) {
	for _, item := range m.items {
		if item.ExternalID == externalID {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockAdapter) CreateItem(_ *database.Requirement) (string, error) {
	m.createCalls++
	return m.createResult, m.createErr
}
func (m *mockAdapter) UpdateItem(externalID string, _ *database.Requirement) bool {
	m.updateCalls++
	m.updateCallIDs = append(m.updateCallIDs, externalID)
	return m.updateResult
}
func (m *mockAdapter) MapStatusToRTMX(externalStatus string) database.Status {
	if s, ok := m.statusMapping[externalStatus]; ok {
		return s
	}
	return database.StatusMissing
}
func (m *mockAdapter) MapStatusFromRTMX(status database.Status) string {
	if s, ok := m.statusReverse[status]; ok {
		return s
	}
	return "open"
}

// createTestDatabase creates a temp CSV database with given requirements and returns its path.
func createTestDatabase(t *testing.T, reqs []*database.Requirement) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "database.csv")
	db := database.NewDatabase()
	for _, r := range reqs {
		if err := db.Add(r); err != nil {
			t.Fatalf("failed to add requirement: %v", err)
		}
	}
	if err := db.Save(dbPath); err != nil {
		t.Fatalf("failed to save database: %v", err)
	}
	return dbPath
}

// createTestConfig returns a config pointing to the given database path.
func createTestConfig(dbPath string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.RTMX.Database = dbPath
	return cfg
}

func TestSyncResultSummary(t *testing.T) {
	rtmx.Req(t, "REQ-GO-028")

	tests := []struct {
		name     string
		result   *SyncResult
		expected string
	}{
		{
			name:     "empty result",
			result:   &SyncResult{},
			expected: "No changes",
		},
		{
			name: "only created",
			result: &SyncResult{
				Created: []string{"REQ-001", "REQ-002"},
			},
			expected: "2 created",
		},
		{
			name: "only updated",
			result: &SyncResult{
				Updated: []string{"REQ-003"},
			},
			expected: "1 updated",
		},
		{
			name: "mixed results",
			result: &SyncResult{
				Created: []string{"REQ-001"},
				Updated: []string{"REQ-002", "REQ-003"},
				Skipped: []string{"REQ-004"},
			},
			expected: "1 created, 2 updated, 1 skipped",
		},
		{
			name: "with conflicts",
			result: &SyncResult{
				Updated:   []string{"REQ-001"},
				Conflicts: []SyncConflict{{ID: "REQ-002", Reason: "test conflict"}},
			},
			expected: "1 updated, 1 conflicts",
		},
		{
			name: "with errors",
			result: &SyncResult{
				Errors: []SyncError{{ID: "REQ-001", Error: "test error"}},
			},
			expected: "1 errors",
		},
		{
			name: "all types",
			result: &SyncResult{
				Created:   []string{"1"},
				Updated:   []string{"2", "3"},
				Skipped:   []string{"4", "5", "6"},
				Conflicts: []SyncConflict{{ID: "7", Reason: "conflict"}},
				Errors:    []SyncError{{ID: "8", Error: "error"}},
			},
			expected: "1 created, 2 updated, 3 skipped, 1 conflicts, 1 errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.Summary()
			if got != tt.expected {
				t.Errorf("Summary() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSyncConflictStruct(t *testing.T) {
	conflict := SyncConflict{
		ID:     "REQ-TEST-001",
		Reason: "Status conflict: COMPLETE vs MISSING",
	}

	if conflict.ID != "REQ-TEST-001" {
		t.Errorf("Expected ID 'REQ-TEST-001', got %s", conflict.ID)
	}

	if !strings.Contains(conflict.Reason, "Status conflict") {
		t.Errorf("Expected reason to contain 'Status conflict', got %s", conflict.Reason)
	}
}

func TestSyncErrorStruct(t *testing.T) {
	syncErr := SyncError{
		ID:    "REQ-TEST-002",
		Error: "Connection failed",
	}

	if syncErr.ID != "REQ-TEST-002" {
		t.Errorf("Expected ID 'REQ-TEST-002', got %s", syncErr.ID)
	}

	if syncErr.Error != "Connection failed" {
		t.Errorf("Expected error 'Connection failed', got %s", syncErr.Error)
	}
}

func TestSyncCommandNoDirection(t *testing.T) {
	// Reset flags
	syncImport = false
	syncExport = false
	syncBidirect = false
	syncDryRun = true
	syncPreferLocal = false
	syncPreferRemote = false

	// Run sync without direction
	err := syncCmd.RunE(syncCmd, []string{})
	if err == nil {
		t.Error("Expected error when no direction specified")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Errorf("Expected ExitError, got %T", err)
	}
	if exitErr.Code != 1 {
		t.Errorf("Expected exit code 1, got %d", exitErr.Code)
	}
}

func TestSyncCommandConflictingPreferences(t *testing.T) {
	// Reset flags
	syncImport = true
	syncExport = false
	syncBidirect = false
	syncDryRun = true
	syncPreferLocal = true
	syncPreferRemote = true

	// Run sync with conflicting preferences
	err := syncCmd.RunE(syncCmd, []string{})
	if err == nil {
		t.Error("Expected error with conflicting preferences")
	}

	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Errorf("Expected ExitError, got %T", err)
	}
	if exitErr.Code != 1 {
		t.Errorf("Expected exit code 1, got %d", exitErr.Code)
	}
}

func TestSyncCommandFlags(t *testing.T) {
	// Test that command has all expected flags
	cmd := syncCmd

	// Verify flags exist
	if cmd.Flags().Lookup("service") == nil {
		t.Error("Expected 'service' flag to exist")
	}
	if cmd.Flags().Lookup("import") == nil {
		t.Error("Expected 'import' flag to exist")
	}
	if cmd.Flags().Lookup("export") == nil {
		t.Error("Expected 'export' flag to exist")
	}
	if cmd.Flags().Lookup("bidirectional") == nil {
		t.Error("Expected 'bidirectional' flag to exist")
	}
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Error("Expected 'dry-run' flag to exist")
	}
	if cmd.Flags().Lookup("prefer-local") == nil {
		t.Error("Expected 'prefer-local' flag to exist")
	}
	if cmd.Flags().Lookup("prefer-remote") == nil {
		t.Error("Expected 'prefer-remote' flag to exist")
	}

	// Verify short flags
	if cmd.Flags().ShorthandLookup("s") == nil {
		t.Error("Expected 's' short flag for service")
	}
	if cmd.Flags().ShorthandLookup("i") == nil {
		t.Error("Expected 'i' short flag for import")
	}
	if cmd.Flags().ShorthandLookup("e") == nil {
		t.Error("Expected 'e' short flag for export")
	}
	if cmd.Flags().ShorthandLookup("b") == nil {
		t.Error("Expected 'b' short flag for bidirectional")
	}
}

func TestSyncServiceUnknown(t *testing.T) {
	// Reset flags
	syncService = "unknown"
	syncImport = true
	syncExport = false
	syncBidirect = false
	syncDryRun = true
	syncPreferLocal = false
	syncPreferRemote = false

	// Run sync with unknown service
	err := syncCmd.RunE(syncCmd, []string{})
	if err == nil {
		t.Error("Expected error with unknown service")
	}

	// Reset service
	syncService = "github"
}

func TestSyncResultEmpty(t *testing.T) {
	result := &SyncResult{}

	if len(result.Created) != 0 {
		t.Error("Expected empty Created slice")
	}
	if len(result.Updated) != 0 {
		t.Error("Expected empty Updated slice")
	}
	if len(result.Skipped) != 0 {
		t.Error("Expected empty Skipped slice")
	}
	if len(result.Conflicts) != 0 {
		t.Error("Expected empty Conflicts slice")
	}
	if len(result.Errors) != 0 {
		t.Error("Expected empty Errors slice")
	}
}

// --- Tests for getAdapter ---

func TestGetAdapterUnknownService(t *testing.T) {
	cfg := config.DefaultConfig()
	_, err := getAdapter("unknown-service", cfg)
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
	if !strings.Contains(err.Error(), "unknown service") {
		t.Errorf("expected 'unknown service' in error, got: %v", err)
	}
}

func TestGetAdapterGitHubDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RTMX.Adapters.GitHub.Enabled = false
	_, err := getAdapter("github", cfg)
	if err == nil {
		t.Fatal("expected error when github adapter is disabled")
	}
	if !strings.Contains(err.Error(), "not enabled") {
		t.Errorf("expected 'not enabled' in error, got: %v", err)
	}
}

func TestGetAdapterJiraDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RTMX.Adapters.Jira.Enabled = false
	_, err := getAdapter("jira", cfg)
	if err == nil {
		t.Fatal("expected error when jira adapter is disabled")
	}
	if !strings.Contains(err.Error(), "not enabled") {
		t.Errorf("expected 'not enabled' in error, got: %v", err)
	}
}

// --- Tests for printSyncSummary ---

func TestPrintSyncSummaryNoChanges(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := &SyncResult{}
	printSyncSummary(result)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "No changes") {
		t.Errorf("expected 'No changes' in output, got: %s", output)
	}
	if !strings.Contains(output, "Sync Summary") {
		t.Errorf("expected 'Sync Summary' in output, got: %s", output)
	}
}

func TestPrintSyncSummaryWithConflicts(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := &SyncResult{
		Updated: []string{"REQ-001"},
		Conflicts: []SyncConflict{
			{ID: "REQ-002", Reason: "Status conflict: COMPLETE vs MISSING"},
		},
	}
	printSyncSummary(result)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Conflicts requiring attention") {
		t.Errorf("expected conflicts section in output, got: %s", output)
	}
	if !strings.Contains(output, "REQ-002") {
		t.Errorf("expected conflict ID in output, got: %s", output)
	}
}

func TestPrintSyncSummaryWithErrors(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := &SyncResult{
		Errors: []SyncError{
			{ID: "REQ-003", Error: "connection timeout"},
		},
	}
	printSyncSummary(result)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Errors") {
		t.Errorf("expected errors section in output, got: %s", output)
	}
	if !strings.Contains(output, "connection timeout") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

// --- Tests for runImport ---

func TestRunImportDryRunNewItems(t *testing.T) {
	// Create an empty database
	dbPath := createTestDatabase(t, nil)
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "New item from service", Status: "open"},
			{ExternalID: "EXT-2", Title: "Another new item", Status: "open"},
		},
		statusMapping: map[string]database.Status{
			"open":   database.StatusMissing,
			"closed": database.StatusComplete,
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runImport(adapter, cfg, true)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// New items without existing linkage should be created
	if len(result.Created) != 2 {
		t.Errorf("expected 2 created items, got %d", len(result.Created))
	}
	if !strings.Contains(output, "Would import") {
		t.Errorf("expected dry-run import message, got: %s", output)
	}
}

func TestRunImportDryRunLinkedItemStatusChange(t *testing.T) {
	// Create a database with a requirement linked to an external item
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "closed"},
		},
		statusMapping: map[string]database.Status{
			"open":   database.StatusMissing,
			"closed": database.StatusComplete,
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runImport(adapter, cfg, true)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Status changed from MISSING to COMPLETE, should be updated
	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated item, got %d", len(result.Updated))
	}
	if !strings.Contains(output, "Would update") {
		t.Errorf("expected dry-run update message, got: %s", output)
	}
}

func TestRunImportLinkedItemNoChange(t *testing.T) {
	// Create a database with a requirement already at matching status
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "open"},
		},
		statusMapping: map[string]database.Status{
			"open":   database.StatusMissing,
			"closed": database.StatusComplete,
		},
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runImport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	// No change, should be skipped
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped item, got %d", len(result.Skipped))
	}
	if len(result.Updated) != 0 {
		t.Errorf("expected 0 updated items, got %d", len(result.Updated))
	}
}

func TestRunImportItemWithRequirementID(t *testing.T) {
	// External item references a known requirement by RequirementID field
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Links to requirement", Status: "open", RequirementID: "REQ-TEST-001"},
		},
		statusMapping: map[string]database.Status{
			"open": database.StatusMissing,
		},
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runImport(adapter, cfg, true)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated (linked) item, got %d", len(result.Updated))
	}
}

func TestRunImportFetchError(t *testing.T) {
	dbPath := createTestDatabase(t, nil)
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		fetchErr:  fmt.Errorf("API rate limit exceeded"),
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runImport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error, "rate limit") {
		t.Errorf("expected rate limit error, got: %s", result.Errors[0].Error)
	}
}

func TestRunImportLongTitle(t *testing.T) {
	// Test that long titles are truncated
	dbPath := createTestDatabase(t, nil)
	cfg := createTestConfig(dbPath)

	longTitle := strings.Repeat("A", 100)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: longTitle, Status: "open"},
		},
		statusMapping: map[string]database.Status{
			"open": database.StatusMissing,
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runImport(adapter, cfg, true)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if len(result.Created) != 1 {
		t.Errorf("expected 1 created item, got %d", len(result.Created))
	}
	// Title should be truncated with "..."
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated title with '...' in output, got: %s", output)
	}
}

// --- Tests for runExport ---

func TestRunExportDryRunNewExport(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, true)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Would export: REQ-TEST-001") {
		t.Errorf("expected dry-run export message, got: %s", output)
	}
	if len(result.Created) != 0 && len(result.Updated) != 0 {
		// Dry run should not actually create/update
	}
}

func TestRunExportDryRunExistingItem(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusComplete
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, true)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Would update") {
		t.Errorf("expected dry-run update message, got: %s", output)
	}
	_ = result
}

func TestRunExportActualCreateSuccess(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:         "test-service",
		connected:    true,
		createResult: "EXT-NEW-1",
		createErr:    nil,
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Created) != 1 {
		t.Errorf("expected 1 created, got %d", len(result.Created))
	}
	if adapter.createCalls != 1 {
		t.Errorf("expected 1 create call, got %d", adapter.createCalls)
	}
}

func TestRunExportActualCreateFailure(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		createErr: fmt.Errorf("API error: HTTP 500"),
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error, "500") {
		t.Errorf("expected API error in result, got: %s", result.Errors[0].Error)
	}
}

func TestRunExportActualUpdateSuccess(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusComplete
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:         "test-service",
		connected:    true,
		updateResult: true,
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Updated))
	}
	if adapter.updateCalls != 1 {
		t.Errorf("expected 1 update call, got %d", adapter.updateCalls)
	}
}

func TestRunExportActualUpdateFailure(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusComplete
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:         "test-service",
		connected:    true,
		updateResult: false,
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error, "update failed") {
		t.Errorf("expected 'update failed' error, got: %s", result.Errors[0].Error)
	}
}

func TestRunExportDatabaseNotFound(t *testing.T) {
	cfg := createTestConfig("/nonexistent/path/database.csv")

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runExport(adapter, cfg, false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error, "database not found") {
		t.Errorf("expected 'database not found' error, got: %s", result.Errors[0].Error)
	}
}

// --- Tests for runBidirectional ---

func TestRunBidirectionalDryRunPreferLocal(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusPartial
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "closed"},
		},
		statusMapping: map[string]database.Status{
			"closed": database.StatusComplete,
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "prefer-local", true)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Status conflict: local=PARTIAL, remote=COMPLETE, prefer-local should update external
	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Updated))
	}
	if !strings.Contains(output, "Would update") {
		t.Errorf("expected dry-run update message, got: %s", output)
	}
}

func TestRunBidirectionalPreferRemote(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "closed"},
		},
		statusMapping: map[string]database.Status{
			"closed": database.StatusComplete,
		},
		updateResult: true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "prefer-remote", false)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Updated))
	}
	if !strings.Contains(output, "Remote wins") {
		t.Errorf("expected 'Remote wins' message, got: %s", output)
	}
}

func TestRunBidirectionalConflictAsk(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "closed"},
		},
		statusMapping: map[string]database.Status{
			"closed": database.StatusComplete,
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "ask", false)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Default "ask" mode should report a conflict
	if len(result.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(result.Conflicts))
	}
	if !strings.Contains(output, "Conflict") {
		t.Errorf("expected 'Conflict' in output, got: %s", output)
	}
}

func TestRunBidirectionalNoConflict(t *testing.T) {
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "open"},
		},
		statusMapping: map[string]database.Status{
			"open": database.StatusMissing,
		},
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "ask", false)

	_ = w.Close()
	os.Stdout = oldStdout

	// Statuses match, should be skipped
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(result.Skipped))
	}
}

func TestRunBidirectionalNewExternalItems(t *testing.T) {
	// No existing requirements, external items should be import candidates
	dbPath := createTestDatabase(t, nil)
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "New external item", Status: "open"},
			{ExternalID: "EXT-2", Title: strings.Repeat("X", 100), Status: "open"},
		},
		statusMapping: map[string]database.Status{
			"open": database.StatusMissing,
		},
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "ask", true)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Created) != 2 {
		t.Errorf("expected 2 created (import candidates), got %d", len(result.Created))
	}
}

func TestRunBidirectionalFetchError(t *testing.T) {
	dbPath := createTestDatabase(t, nil)
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:     "test-service",
		fetchErr: fmt.Errorf("network timeout"),
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "ask", false)

	_ = w.Close()
	os.Stdout = oldStdout

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if !strings.Contains(result.Errors[0].Error, "network timeout") {
		t.Errorf("expected network timeout error, got: %s", result.Errors[0].Error)
	}
}

func TestRunBidirectionalPreferLocalActual(t *testing.T) {
	// Test actual (non-dry-run) prefer-local updates the adapter
	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Test requirement"
	req.Status = database.StatusPartial
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)

	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked item", Status: "closed"},
		},
		statusMapping: map[string]database.Status{
			"closed": database.StatusComplete,
		},
		updateResult: true,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := runBidirectional(adapter, cfg, "prefer-local", false)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Updated))
	}
	if !strings.Contains(output, "Local wins") {
		t.Errorf("expected 'Local wins' message, got: %s", output)
	}
	if adapter.updateCalls != 1 {
		t.Errorf("expected 1 update call on adapter, got %d", adapter.updateCalls)
	}
}
