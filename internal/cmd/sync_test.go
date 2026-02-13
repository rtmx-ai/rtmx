package cmd

import (
	"strings"
	"testing"
)

func TestSyncResultSummary(t *testing.T) {
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
