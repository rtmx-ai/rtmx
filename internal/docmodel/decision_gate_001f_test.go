package docmodel

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestDecisionGate001fAcceptsADR0007(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001f")
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	adr, err := os.ReadFile(filepath.Join(root, "docs", "adr", "0007-requirement-document-model.md"))
	if err != nil {
		t.Fatal(err)
	}
	adrText := strings.ReplaceAll(string(adr), "\r\n", "\n")
	if !strings.Contains(adrText, "## Status\n\nAccepted") {
		t.Fatal("ADR-0007 Status must be Accepted")
	}
	for _, need := range []string{
		"JSONL",
		"Markdown companion",
		"atdd-required-all-v0",
		"AC-to-test ATDD",
		"REQ-DATA-002",
		"sync-protocol-v1",
	} {
		if !strings.Contains(adrText, need) {
			t.Errorf("ADR-0007 missing %q", need)
		}
	}

	gate, err := os.ReadFile(filepath.Join(root, "docs", "schemas", "DECISION_GATE_001f.md"))
	if err != nil {
		t.Fatal(err)
	}
	g := string(gate)
	for _, need := range []string{
		"**GO**",
		"not** uniformity",
		"REQ-DATA-002",
		"REQ-DATA-006",
		"managed-sync",
		"atdd-required-all-v0",
	} {
		if !strings.Contains(g, need) {
			t.Errorf("DECISION_GATE_001f missing %q", need)
		}
	}

	schema := filepath.Join(root, "docs", "schemas", "requirement-document-v0.schema.json")
	if _, err := os.Stat(schema); err != nil {
		t.Fatalf("published schema missing: %v", err)
	}
}

func TestParentDATA001ClosesWithGate(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001")
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	adr, err := os.ReadFile(filepath.Join(root, "docs", "adr", "0007-requirement-document-model.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ReplaceAll(string(adr), "\r\n", "\n"), "## Status\n\nAccepted") {
		t.Fatal("parent DATA-001 requires Accepted ADR-0007")
	}
	gate := filepath.Join(root, "docs", "schemas", "DECISION_GATE_001f.md")
	if _, err := os.Stat(gate); err != nil {
		t.Fatal(err)
	}
}
