package rtmx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReqValidID(t *testing.T) {
	ClearRegistry()

	// This should not fail
	Req(t, "REQ-TEST-001")

	// Results are recorded on cleanup, so we need to check markers
	if len(globalRegistry.markers) == 0 {
		t.Error("Expected marker to be registered")
	}
}

func TestReqInvalidID(t *testing.T) {
	// Testing invalid ID requires using the regex directly since we can't
	// mock testing.TB (it has a private method)
	invalidIDs := []string{"INVALID-ID", "req-test-001", "REQ_TEST_001", ""}
	for _, id := range invalidIDs {
		if reqIDPattern.MatchString(id) {
			t.Errorf("Expected %q to be invalid", id)
		}
	}
}

func TestReqWithOptions(t *testing.T) {
	ClearRegistry()

	Req(t, "REQ-TEST-002",
		Scope("unit"),
		Technique("nominal"),
		Env("simulation"),
	)

	// Check the marker was registered with options
	for _, m := range globalRegistry.markers {
		if m.ReqID == "REQ-TEST-002" {
			if m.Scope != "unit" {
				t.Errorf("Expected scope 'unit', got '%s'", m.Scope)
			}
			if m.Technique != "nominal" {
				t.Errorf("Expected technique 'nominal', got '%s'", m.Technique)
			}
			if m.Env != "simulation" {
				t.Errorf("Expected env 'simulation', got '%s'", m.Env)
			}
			return
		}
	}
	t.Error("Marker not found in registry")
}

func TestReqIDPattern(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"REQ-TEST-001", true},
		{"REQ-AUTH-123", true},
		{"REQ-A-1", true},
		{"REQ-TEST-", false},
		{"REQ--001", false},
		{"REQ-test-001", false}, // lowercase
		{"INVALID", false},
		{"REQ_TEST_001", false}, // underscores
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			result := reqIDPattern.MatchString(tc.id)
			if result != tc.valid {
				t.Errorf("reqIDPattern.MatchString(%q) = %v, want %v", tc.id, result, tc.valid)
			}
		})
	}
}

func TestWriteResultsJSON(t *testing.T) {
	ClearRegistry()

	// Register a marker
	Req(t, "REQ-JSON-001", Scope("unit"))

	// Create temp file
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "results.json")

	// Manually add a result since cleanup hasn't run
	globalRegistry.mu.Lock()
	globalRegistry.results = append(globalRegistry.results, testResult{
		Marker: marker{ReqID: "REQ-JSON-001", Scope: "unit"},
		Passed: true,
	})
	globalRegistry.mu.Unlock()

	err := WriteResultsJSON(outPath)
	if err != nil {
		t.Fatalf("WriteResultsJSON failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read results file: %v", err)
	}

	if !strings.Contains(string(data), "REQ-JSON-001") {
		t.Error("Results file should contain requirement ID")
	}
}

func TestClearRegistry(t *testing.T) {
	// Register something first
	Req(t, "REQ-CLEAR-001")

	if len(globalRegistry.markers) == 0 {
		t.Skip("No markers registered, can't test clear")
	}

	ClearRegistry()

	if len(globalRegistry.markers) != 0 {
		t.Error("Expected markers to be cleared")
	}
	if len(globalRegistry.results) != 0 {
		t.Error("Expected results to be cleared")
	}
}

