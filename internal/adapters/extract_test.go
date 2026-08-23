package adapters

import (
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestExtractReqID validates the unified requirement ID extraction function.
// REQ-ADAPT-013: Unified requirement ID extraction across all adapters.
func TestExtractReqID(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-013")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Bracketed title format (used by CreateItem in all adapters)
		{"bracketed_title", "[REQ-CLI-001] Build CLI", "REQ-CLI-001"},
		{"bracketed_with_text", "[REQ-MCP-010] MCP server implementation", "REQ-MCP-010"},

		// RTMX: prefix format (used in Asana notes, GitHub bodies)
		{"rtmx_prefix", "RTMX: REQ-CLI-001", "REQ-CLI-001"},
		{"rtmx_prefix_in_text", "Build CLI framework\n\nRTMX: REQ-MCP-001", "REQ-MCP-001"},

		// Inline mention (bare ID in text)
		{"inline_mention", "See REQ-E2E-010 for details", "REQ-E2E-010"},
		{"inline_start", "REQ-VERIFY-003 is the target", "REQ-VERIFY-003"},

		// Alphanumeric categories (the bug this fixes)
		{"alphanumeric_e2e", "Description: REQ-E2E-010", "REQ-E2E-010"},
		{"alphanumeric_v2", "Tracking REQ-V2-001 migration", "REQ-V2-001"},
		{"alphanumeric_k8s", "[REQ-K8S-042] Deploy config", "REQ-K8S-042"},
		{"alphanumeric_ci2", "RTMX: REQ-CI2-005", "REQ-CI2-005"},

		// Multi-segment category prefixes (REQ-VERIFY-011)
		{"multi_segment_infra_dt", "Implements REQ-INFRA-DT-002 now", "REQ-INFRA-DT-002"},
		{"multi_segment_mode_s", "[REQ-MODE-S-006] search mode", "REQ-MODE-S-006"},
		{"multi_segment_sw_dsp", "RTMX: REQ-SW-DSP-015", "REQ-SW-DSP-015"},
		{"multi_segment_three", "See REQ-A-B-C-001 mention", "REQ-A-B-C-001"},

		// First match wins when multiple IDs present
		{"multiple_ids", "REQ-CLI-001 depends on REQ-MCP-002", "REQ-CLI-001"},

		// Edge cases
		{"empty_string", "", ""},
		{"no_match", "This has no requirement ID", ""},
		{"partial_match", "REQ- is not valid", ""},
		{"lowercase_rejected", "req-cli-001 is lowercase", ""},
		{"no_number", "REQ-CLI- is incomplete", ""},
		{"digit_only_category", "REQ-123-001 starts with digit", ""},

		// Real-world adapter formats
		{"github_body", "Implements the feature.\n\nRTMX: REQ-ADAPT-007\n\nSee also REQ-ADAPT-008.", "REQ-ADAPT-007"},
		{"jira_description", "As a user I want REQ-SEC-003 to be done", "REQ-SEC-003"},
		{"gitlab_description", "RTMX: REQ-INT-005\n\nIntegration tests", "REQ-INT-005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractReqID(tt.input)
			if got != tt.want {
				t.Errorf("ExtractReqID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestMondayRequirementIDRegression verifies that Monday adapter now
// correctly extracts requirement IDs from item names (previously always empty).
// REQ-ADAPT-013: Unified requirement ID extraction.
func TestMondayRequirementIDRegression(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-013")

	server := newTestMondayServer(t)
	defer server.Close()

	cfg := &config.MondayAdapterConfig{
		Enabled:  true,
		BoardID:  "board-1",
		TokenEnv: "MONDAY_TOKEN",
	}
	m := mustMondayAdapter(t, server, cfg)

	items, err := m.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}

	// Monday items have names like "[REQ-CLI-001] Build CLI"
	// Before the fix, RequirementID was always "" because extractReqID
	// only matched "RTMX: " prefix, not bracketed format.
	if items[0].RequirementID != "REQ-CLI-001" {
		t.Errorf("first item RequirementID = %q, want REQ-CLI-001 (was empty before fix)", items[0].RequirementID)
	}
	if items[1].RequirementID != "REQ-MCP-001" {
		t.Errorf("second item RequirementID = %q, want REQ-MCP-001 (was empty before fix)", items[1].RequirementID)
	}

	// GetItem should also work
	item, err := m.GetItem("1001")
	if err != nil {
		t.Fatalf("GetItem error: %v", err)
	}
	if item.RequirementID != "REQ-CLI-001" {
		t.Errorf("GetItem RequirementID = %q, want REQ-CLI-001", item.RequirementID)
	}
}
