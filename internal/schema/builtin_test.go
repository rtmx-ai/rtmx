package schema

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func copyMap(m map[string]string) map[string]string {
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func TestBuiltinSchemas(t *testing.T) {
	rtmx.Req(t, "REQ-PLUGIN-006")

	t.Run("all_builtins_registered", func(t *testing.T) {
		names := Names()
		expected := []string{"core", "do178c", "iso26262", "phoenix"}
		for _, want := range expected {
			found := false
			for _, got := range names {
				if got == want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected schema %q in registry, got %v", want, names)
			}
		}
	})

	t.Run("do178c_schema", func(t *testing.T) {
		s := Get("do178c")
		if s == nil {
			t.Fatal("do178c schema not registered")
		}
		if s.Description == "" {
			t.Error("do178c schema should have a description")
		}

		// Must have core columns plus DO-178C-specific columns
		requiredCols := []string{"req_id", "category", "dal_level", "sw_level",
			"objective_id", "structural_coverage", "trace_to_srs", "evidence_artifact"}
		for _, col := range requiredCols {
			if !s.HasColumn(col) {
				t.Errorf("do178c schema missing column %q", col)
			}
		}

		// Verify DAL enum values
		for _, c := range s.Columns {
			if c.Name == "dal_level" {
				if len(c.EnumValues) != 5 {
					t.Errorf("dal_level should have 5 values, got %d", len(c.EnumValues))
				}
				break
			}
		}
	})

	t.Run("iso26262_schema", func(t *testing.T) {
		s := Get("iso26262")
		if s == nil {
			t.Fatal("iso26262 schema not registered")
		}
		if s.Description == "" {
			t.Error("iso26262 schema should have a description")
		}

		// Must have core columns plus ISO 26262-specific columns
		requiredCols := []string{"req_id", "category", "asil_level", "safety_goal_id",
			"severity", "exposure", "controllability", "fault_tolerance", "evidence_artifact"}
		for _, col := range requiredCols {
			if !s.HasColumn(col) {
				t.Errorf("iso26262 schema missing column %q", col)
			}
		}

		// Verify ASIL enum values
		for _, c := range s.Columns {
			if c.Name == "asil_level" {
				want := []string{"QM", "A", "B", "C", "D"}
				if len(c.EnumValues) != len(want) {
					t.Errorf("asil_level should have %d values, got %d", len(want), len(c.EnumValues))
				}
				break
			}
		}
	})

	t.Run("phoenix_schema", func(t *testing.T) {
		s := Get("phoenix")
		if s == nil {
			t.Fatal("phoenix schema not registered")
		}
		// Already tested in detail elsewhere; just verify registration
		if !s.HasColumn("dal_level") {
			t.Error("phoenix schema should have dal_level column")
		}
	})

	t.Run("schemas_extend_core", func(t *testing.T) {
		core := Get("core")
		if core == nil {
			t.Fatal("core schema not registered")
		}
		coreCount := len(core.Columns)

		for _, name := range []string{"do178c", "iso26262", "phoenix"} {
			s := Get(name)
			if s == nil {
				t.Fatalf("%s schema not registered", name)
			}
			if len(s.Columns) <= coreCount {
				t.Errorf("%s should have more columns than core (%d), got %d", name, coreCount, len(s.Columns))
			}
			// First N columns should match core
			for i := 0; i < coreCount; i++ {
				if s.Columns[i].Name != core.Columns[i].Name {
					t.Errorf("%s column %d: got %q, want %q (core)", name, i, s.Columns[i].Name, core.Columns[i].Name)
				}
			}
		}
	})

	t.Run("validation_works", func(t *testing.T) {
		// Include required core columns to avoid required-field errors
		base := map[string]string{
			"req_id":           "REQ-TEST-001",
			"category":         "TEST",
			"requirement_text": "Test requirement",
		}

		do178c := Get("do178c")
		// Valid DAL level
		row := copyMap(base)
		row["dal_level"] = "A"
		errs := do178c.Validate(row)
		if len(errs) > 0 {
			t.Errorf("valid DAL 'A' produced errors: %v", errs)
		}

		// Invalid DAL level
		row = copyMap(base)
		row["dal_level"] = "X"
		errs = do178c.Validate(row)
		dalErr := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "dal_level") {
				dalErr = true
			}
		}
		if !dalErr {
			t.Error("invalid DAL 'X' should produce dal_level error")
		}

		iso := Get("iso26262")
		// Valid ASIL
		row = copyMap(base)
		row["asil_level"] = "D"
		errs = iso.Validate(row)
		if len(errs) > 0 {
			t.Errorf("valid ASIL 'D' produced errors: %v", errs)
		}

		// Invalid severity
		row = copyMap(base)
		row["severity"] = "S9"
		errs = iso.Validate(row)
		sevErr := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "severity") {
				sevErr = true
			}
		}
		if !sevErr {
			t.Error("invalid severity 'S9' should produce severity error")
		}
	})
}
