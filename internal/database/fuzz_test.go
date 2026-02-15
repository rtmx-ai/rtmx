package database

import (
	"bytes"
	"strings"
	"testing"
)

// FuzzCSVParse tests CSV parsing with arbitrary input.
// It ensures that malformed CSV data doesn't cause panics.
// REQ-GO-070: Go CLI shall include fuzz tests for CSV and YAML parsing
func FuzzCSVParse(f *testing.F) {
	// Seed corpus with valid CSV data
	f.Add(`req_id,category,requirement_text,status,priority,phase,dependencies
REQ-001,CLI,First requirement,COMPLETE,HIGH,1,
`)
	f.Add(`req_id,category,requirement_text
REQ-001,TEST,Simple requirement
`)
	f.Add(`req_id,category,requirement_text,status,priority,phase,dependencies,blocks,effort_weeks
REQ-001,DATA,Multi-field requirement,PARTIAL,MEDIUM,2,REQ-A|REQ-B,REQ-C,1.5
`)

	// Edge cases
	f.Add(`req_id,category,requirement_text
`)
	f.Add(`req_id,category,requirement_text
REQ-001,CAT,"Text with ""quotes"" and commas, inside"
`)
	f.Add(`req_id,category,requirement_text
REQ-001,CAT,` + strings.Repeat("A", 10000) + `
`)
	f.Add(`req_id,category,requirement_text
REQ-001,,Empty category
`)
	f.Add(`req_id,category,requirement_text
,MISSING,No ID
`)

	// Malformed inputs
	f.Add(`invalid header only`)
	f.Add(``)
	f.Add("\x00\x00\x00")
	f.Add(`req_id,category,requirement_text
"unclosed quote`)
	f.Add(`req_id,category,requirement_text
REQ-001,CAT,text
,,extra,fields,that,dont,match,header
`)

	f.Fuzz(func(t *testing.T, data string) {
		// The function should not panic on any input
		db, err := ReadCSV(strings.NewReader(data))

		// If parsing succeeded, the database should be usable
		if err == nil && db != nil {
			// Exercise the database methods without panicking
			_ = db.Len()
			_ = db.All()
			_ = db.StatusCounts()
			_ = db.CompletionPercentage()

			// Try to write back to CSV
			var buf bytes.Buffer
			_ = db.WriteCSV(&buf)
		}
	})
}

// FuzzParseStatus tests status string parsing with arbitrary input.
func FuzzParseStatus(f *testing.F) {
	// Seed corpus with valid statuses
	f.Add("COMPLETE")
	f.Add("complete")
	f.Add("Complete")
	f.Add("PARTIAL")
	f.Add("partial")
	f.Add("MISSING")
	f.Add("missing")
	f.Add("NOT_STARTED")
	f.Add("not_started")
	f.Add("")

	// Edge cases
	f.Add("  COMPLETE  ")
	f.Add("\tPARTIAL\n")
	f.Add("COMPLETE\x00")
	f.Add(strings.Repeat("X", 10000))
	f.Add("COM\x00PLETE")
	f.Add("\r\nMISSING\r\n")

	// Unicode edge cases
	f.Add("\u200BCOMPLETE") // Zero-width space
	f.Add("COMPLETE\u200B")
	f.Add("\uFEFFCOMPLETE") // BOM

	f.Fuzz(func(t *testing.T, input string) {
		// The function should never panic
		status, err := ParseStatus(input)

		// If no error, status should be valid
		if err == nil {
			// Should be one of the known statuses
			valid := false
			for _, s := range AllStatuses() {
				if status == s {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("ParseStatus(%q) returned invalid status %v", input, status)
			}
		}

		// Status methods should not panic
		_ = status.String()
		_ = status.IsComplete()
		_ = status.IsIncomplete()
		_ = status.Weight()
		_ = status.CompletionPercent()
	})
}

// FuzzParsePriority tests priority string parsing with arbitrary input.
func FuzzParsePriority(f *testing.F) {
	// Seed corpus with valid priorities
	f.Add("P0")
	f.Add("p0")
	f.Add("HIGH")
	f.Add("high")
	f.Add("MEDIUM")
	f.Add("medium")
	f.Add("LOW")
	f.Add("low")
	f.Add("")

	// Edge cases
	f.Add("  HIGH  ")
	f.Add("\tMEDIUM\n")
	f.Add("P0\x00")
	f.Add(strings.Repeat("P", 10000))

	f.Fuzz(func(t *testing.T, input string) {
		// The function should never panic
		priority, err := ParsePriority(input)

		// If no error, priority should be valid
		if err == nil {
			valid := false
			for _, p := range AllPriorities() {
				if priority == p {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("ParsePriority(%q) returned invalid priority %v", input, priority)
			}
		}

		// Priority methods should not panic
		_ = priority.String()
		_ = priority.Weight()
		_ = priority.IsHighPriority()
	})
}

// FuzzParseStringSet tests pipe-separated string parsing.
func FuzzParseStringSet(f *testing.F) {
	// Seed corpus with valid pipe-separated strings
	f.Add("REQ-001")
	f.Add("REQ-001|REQ-002")
	f.Add("REQ-001|REQ-002|REQ-003")
	f.Add("")
	f.Add("|||")
	f.Add("  REQ-001  |  REQ-002  ")

	// Edge cases
	f.Add("|")
	f.Add("||||||")
	f.Add("a|b|c|d|e|f|g|h|i|j|k|l|m|n|o|p")
	f.Add(strings.Repeat("REQ-001|", 100) + "REQ-001")
	f.Add("item\x00with\x00null|normal")
	f.Add("  |  |  |  ")
	f.Add("a||b||c") // Empty elements between pipes
	f.Add("\n\t|item|\r\n")

	// Very long items
	f.Add(strings.Repeat("A", 10000) + "|B")
	f.Add("A|" + strings.Repeat("B", 10000))

	f.Fuzz(func(t *testing.T, input string) {
		// The function should never panic
		set := ParseStringSet(input)

		// Set operations should not panic
		_ = set.Len()
		_ = set.String()
		_ = set.Slice()

		// Adding and checking should work
		set.Add("TEST")
		_ = set.Contains("TEST")
		set.Remove("TEST")
	})
}

// FuzzNormalizeColumnName tests column name normalization.
func FuzzNormalizeColumnName(f *testing.F) {
	// Seed corpus with various column name formats
	f.Add("req_id")
	f.Add("ReqId")
	f.Add("RequirementText")
	f.Add("REQUIREMENT_TEXT")
	f.Add("testModule")
	f.Add("")
	f.Add("  req_id  ")
	f.Add("REQ_ID")

	// Edge cases
	f.Add("A")
	f.Add("a")
	f.Add("ABC")
	f.Add("abc")
	f.Add("aBC")
	f.Add("ABc")
	f.Add("___")
	f.Add("a_b_c")
	f.Add("A_B_C")
	f.Add(strings.Repeat("A", 1000))
	f.Add(strings.Repeat("a", 1000))
	f.Add("\t\ncolumn\t\n")

	// Unicode edge cases
	f.Add("column\u200B") // Zero-width space
	f.Add("\uFEFFcolumn") // BOM

	f.Fuzz(func(t *testing.T, input string) {
		// The function should never panic
		result := normalizeColumnName(input)

		// Result should be lowercase
		if result != strings.ToLower(result) {
			t.Errorf("normalizeColumnName(%q) = %q, should be lowercase", input, result)
		}
	})
}

// FuzzCSVRoundTrip tests that valid data survives a round trip through CSV.
func FuzzCSVRoundTrip(f *testing.F) {
	// Seed with requirement-like data
	f.Add("REQ-001", "CLI", "Basic requirement", "COMPLETE", "HIGH", 1)
	f.Add("REQ-002", "DATA", "With special chars: comma, \"quotes\"", "PARTIAL", "MEDIUM", 2)
	f.Add("REQ-003", "", "", "MISSING", "", 0)
	f.Add("REQ-TEST-WITH-LONG-ID-001", "CATEGORY-WITH-DASHES", "Multiline\nrequirement\ntext", "NOT_STARTED", "LOW", 10)

	f.Fuzz(func(t *testing.T, reqID, category, text, statusStr, priorityStr string, phase int) {
		// Skip if req_id is empty (required field)
		reqID = strings.TrimSpace(reqID)
		if reqID == "" {
			return
		}

		// Create a requirement with the fuzzed data
		req := NewRequirement(reqID)
		req.Category = category
		req.RequirementText = text
		req.Phase = phase

		// Parse status and priority (they have defaults for invalid input)
		status, _ := ParseStatus(statusStr)
		req.Status = status

		priority, _ := ParsePriority(priorityStr)
		req.Priority = priority

		// Add to database
		db := NewDatabase()
		if err := db.Add(req); err != nil {
			// Duplicate ID is expected in fuzz testing
			return
		}

		// Write to CSV
		var buf bytes.Buffer
		if err := db.WriteCSV(&buf); err != nil {
			t.Fatalf("WriteCSV failed: %v", err)
		}

		// Read back
		db2, err := ReadCSV(&buf)
		if err != nil {
			t.Fatalf("ReadCSV failed: %v", err)
		}

		// Verify round-trip
		if db2.Len() != 1 {
			t.Errorf("Round-trip: got %d requirements, want 1", db2.Len())
			return
		}

		req2 := db2.Get(reqID)
		if req2 == nil {
			t.Errorf("Round-trip: requirement %q not found", reqID)
			return
		}

		// Verify fields survived
		if req2.Category != req.Category {
			t.Errorf("Category: got %q, want %q", req2.Category, req.Category)
		}
		if req2.Status != req.Status {
			t.Errorf("Status: got %v, want %v", req2.Status, req.Status)
		}
		if req2.Priority != req.Priority {
			t.Errorf("Priority: got %v, want %v", req2.Priority, req.Priority)
		}
		if req2.Phase != req.Phase {
			t.Errorf("Phase: got %d, want %d", req2.Phase, req.Phase)
		}
	})
}
