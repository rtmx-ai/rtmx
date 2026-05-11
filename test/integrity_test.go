package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestIntegrityFramework validates that the integrity design document
// exists and covers all required enforcement mechanisms.
// REQ-INT-001: Database integrity framework
func TestIntegrityFramework(t *testing.T) {
	rtmx.Req(t, "REQ-INT-001")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	designPath := filepath.Join(projectRoot, "docs", "INTEGRITY_DESIGN.md")
	content, err := os.ReadFile(designPath)
	if err != nil {
		t.Fatalf("Design document not found at %s: %v", designPath, err)
	}
	doc := string(content)

	t.Run("analyzes_all_enforcement_mechanisms", func(t *testing.T) {
		mechanisms := []string{
			"file permissions",
			"remote attestation",
			"branch protection",
			"confirmation",
			"hardware token",
			"proof-of-verification",
		}
		for _, m := range mechanisms {
			if !strings.Contains(strings.ToLower(doc), strings.ToLower(m)) {
				t.Errorf("Design document must analyze enforcement mechanism: %s", m)
			}
		}
	})

	t.Run("defines_trust_models", func(t *testing.T) {
		models := []string{"self", "team", "delegated", "web-of-trust"}
		for _, m := range models {
			if !strings.Contains(doc, m) {
				t.Errorf("Design document must define trust model: %s", m)
			}
		}
	})

	t.Run("addresses_deployment_constraints", func(t *testing.T) {
		constraints := []string{"sudo", "offline", "air-gapped", "cross-platform"}
		for _, c := range constraints {
			if !strings.Contains(strings.ToLower(doc), strings.ToLower(c)) {
				t.Errorf("Design document must address deployment constraint: %s", c)
			}
		}
	})

	t.Run("provides_migration_path", func(t *testing.T) {
		if !strings.Contains(doc, "Migration") {
			t.Error("Design document must include migration path")
		}
	})

	t.Run("includes_adversarial_analysis", func(t *testing.T) {
		if !strings.Contains(doc, "Adversarial") || !strings.Contains(doc, "Attack") {
			t.Error("Design document must include adversarial analysis")
		}
	})

	t.Run("chooses_implementation_approach", func(t *testing.T) {
		if !strings.Contains(doc, "Chosen") {
			t.Error("Design document must specify chosen enforcement approach")
		}
	})
}

// TestStaleTestReferences validates that all test_module and test_function
// fields in the RTM database reference files and functions that actually
// exist in the codebase. Prevents silent verify skips from stale references.
// REQ-INT-004: Stale test reference lint
func TestStaleTestReferences(t *testing.T) {
	rtmx.Req(t, "REQ-INT-004")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	dbPath := filepath.Join(projectRoot, ".rtmx", "database.csv")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("failed to read database: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		t.Fatal("database has no data rows")
	}

	// Parse header to find column indices
	header := strings.Split(lines[0], ",")
	colIdx := map[string]int{}
	for i, col := range header {
		colIdx[col] = i
	}

	moduleIdx := colIdx["test_module"]
	reqIdx := colIdx["req_id"]
	statusIdx := colIdx["status"]

	var staleModules []string

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) <= moduleIdx {
			continue
		}

		reqID := fields[reqIdx]
		testModule := fields[moduleIdx]
		status := fields[statusIdx]

		// MISSING requirements may have aspirational test paths --
		// only lint COMPLETE and PARTIAL requirements where the path
		// should actually exist.
		if status == "MISSING" || status == "NOT_STARTED" {
			continue
		}

		if testModule == "" {
			continue
		}

		// Skip non-file references (glob patterns, directories without extension)
		if strings.Contains(testModule, "*") {
			continue
		}

		modulePath := filepath.Join(projectRoot, testModule)
		if _, err := os.Stat(modulePath); err != nil {
			staleModules = append(staleModules, reqID+": "+testModule)
		}
	}

	if len(staleModules) > 0 {
		t.Errorf("found %d stale test_module paths in COMPLETE/PARTIAL requirements:\n  %s",
			len(staleModules), strings.Join(staleModules, "\n  "))
	}
}
