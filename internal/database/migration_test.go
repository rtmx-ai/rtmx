package database

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestSchemaMigration(t *testing.T) {
	rtmx.Req(t, "REQ-MIG-004")

	t.Run("detect_missing_columns", func(t *testing.T) {
		// Create a database with an old header missing some columns
		oldCSV := "req_id,category,requirement_text,status,priority\n" +
			"REQ-001,CLI,Feature one,MISSING,HIGH\n"
		db, err := ReadCSV(strings.NewReader(oldCSV))
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		result := db.DetectMigration()
		if !result.Migrated {
			t.Error("should detect migration needed")
		}
		if len(result.AddedColumns) == 0 {
			t.Error("should detect missing columns")
		}

		// Should be missing subcategory, target_value, test_module, etc.
		expectedMissing := []string{"subcategory", "target_value", "test_module",
			"test_function", "validation_method", "phase", "notes",
			"effort_weeks", "dependencies", "blocks", "assignee",
			"sprint", "started_date", "completed_date", "requirement_file", "external_id"}
		for _, col := range expectedMissing {
			found := false
			for _, added := range result.AddedColumns {
				if added == col {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %q in AddedColumns", col)
			}
		}
	})

	t.Run("no_migration_needed", func(t *testing.T) {
		db := NewDatabase()
		// NewDatabase returns standard header via Header()
		result := db.DetectMigration()
		if result.Migrated {
			t.Error("fresh database should not need migration")
		}
		if len(result.AddedColumns) > 0 {
			t.Errorf("should have no added columns, got %v", result.AddedColumns)
		}
	})

	t.Run("migrate_adds_columns", func(t *testing.T) {
		oldCSV := "req_id,category,requirement_text,status\n" +
			"REQ-001,CLI,Feature one,MISSING\n"
		db, err := ReadCSV(strings.NewReader(oldCSV))
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		result := db.Migrate()
		if !result.Migrated {
			t.Error("migrate should report migration")
		}

		// After migration, header should contain all standard columns
		header := db.Header()
		headerSet := make(map[string]bool)
		for _, h := range header {
			headerSet[h] = true
		}
		for _, std := range standardColumns {
			if !headerSet[std] {
				t.Errorf("after migration, header should contain %q", std)
			}
		}

		// Database should be dirty
		if !db.IsDirty() {
			t.Error("database should be marked dirty after migration")
		}
	})

	t.Run("preserves_extra_columns", func(t *testing.T) {
		oldCSV := "req_id,category,requirement_text,status,custom_field\n" +
			"REQ-001,CLI,Feature one,MISSING,custom_value\n"
		db, err := ReadCSV(strings.NewReader(oldCSV))
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		result := db.Migrate()
		if len(result.ExtraColumns) != 1 || result.ExtraColumns[0] != "custom_field" {
			t.Errorf("extra columns = %v, want [custom_field]", result.ExtraColumns)
		}

		// Extra column should be at the end of header
		header := db.Header()
		lastCol := header[len(header)-1]
		if lastCol != "custom_field" {
			t.Errorf("custom_field should be last in header, got %q", lastCol)
		}
	})

	t.Run("preserves_data", func(t *testing.T) {
		oldCSV := "req_id,category,requirement_text,status,priority\n" +
			"REQ-001,CLI,Feature one,MISSING,HIGH\n" +
			"REQ-002,DATA,Feature two,COMPLETE,MEDIUM\n"
		db, err := ReadCSV(strings.NewReader(oldCSV))
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		db.Migrate()

		// Data should still be accessible
		req := db.Get("REQ-001")
		if req == nil {
			t.Fatal("REQ-001 should still exist after migration")
		}
		if req.Category != "CLI" {
			t.Errorf("category = %q, want CLI", req.Category)
		}
		if req.RequirementText != "Feature one" {
			t.Errorf("requirement_text = %q, want 'Feature one'", req.RequirementText)
		}

		req2 := db.Get("REQ-002")
		if req2 == nil {
			t.Fatal("REQ-002 should still exist after migration")
		}
		if !req2.IsComplete() {
			t.Error("REQ-002 should still be COMPLETE after migration")
		}
	})

	t.Run("round_trip_after_migration", func(t *testing.T) {
		oldCSV := "req_id,category,requirement_text,status\n" +
			"REQ-001,CLI,Feature one,MISSING\n"
		db, err := ReadCSV(strings.NewReader(oldCSV))
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		db.Migrate()

		// Write and re-read
		var buf strings.Builder
		if err := db.WriteCSV(&buf); err != nil {
			t.Fatalf("WriteCSV failed: %v", err)
		}

		db2, err := ReadCSV(strings.NewReader(buf.String()))
		if err != nil {
			t.Fatalf("ReadCSV after migration failed: %v", err)
		}

		// Should not need migration anymore
		if db2.NeedsMigration() {
			t.Error("re-loaded database should not need migration")
		}

		// Data preserved
		req := db2.Get("REQ-001")
		if req == nil {
			t.Fatal("REQ-001 should exist after round-trip")
		}
		if req.Category != "CLI" {
			t.Errorf("category = %q, want CLI", req.Category)
		}
	})

	t.Run("schema_version", func(t *testing.T) {
		version := SchemaVersion()
		if version != 21 {
			t.Errorf("schema version = %d, want 21 (standard columns)", version)
		}

		db := NewDatabase()
		if db.DatabaseSchemaVersion() != 21 {
			t.Errorf("new database schema version = %d, want 21", db.DatabaseSchemaVersion())
		}
	})

	t.Run("needs_migration", func(t *testing.T) {
		oldCSV := "req_id,category,requirement_text,status\n" +
			"REQ-001,CLI,Feature,MISSING\n"
		db, err := ReadCSV(strings.NewReader(oldCSV))
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		if !db.NeedsMigration() {
			t.Error("old database should need migration")
		}

		db.Migrate()
		if db.NeedsMigration() {
			t.Error("migrated database should not need migration")
		}
	})
}
