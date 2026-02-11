package database

import (
	"bytes"
	"strings"
	"testing"
)

func TestStatusParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Status
		wantErr  bool
	}{
		{"COMPLETE", StatusComplete, false},
		{"complete", StatusComplete, false},
		{"Complete", StatusComplete, false},
		{"PARTIAL", StatusPartial, false},
		{"MISSING", StatusMissing, false},
		{"", StatusMissing, false},
		{"NOT_STARTED", StatusNotStarted, false},
		{"INVALID", StatusMissing, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatus(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseStatus(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPriorityParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
		wantErr  bool
	}{
		{"P0", PriorityP0, false},
		{"HIGH", PriorityHigh, false},
		{"high", PriorityHigh, false},
		{"MEDIUM", PriorityMedium, false},
		{"", PriorityMedium, false},
		{"LOW", PriorityLow, false},
		{"INVALID", PriorityMedium, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePriority(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePriority(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParsePriority(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStringSet(t *testing.T) {
	// Test creation
	s := NewStringSet("a", "b", "c")
	if s.Len() != 3 {
		t.Errorf("NewStringSet: expected 3 items, got %d", s.Len())
	}

	// Test Contains
	if !s.Contains("a") {
		t.Error("StringSet should contain 'a'")
	}
	if s.Contains("d") {
		t.Error("StringSet should not contain 'd'")
	}

	// Test Add
	s.Add("d")
	if !s.Contains("d") {
		t.Error("StringSet should contain 'd' after Add")
	}

	// Test Remove
	s.Remove("a")
	if s.Contains("a") {
		t.Error("StringSet should not contain 'a' after Remove")
	}

	// Test String (pipe-separated)
	s2 := NewStringSet("REQ-A", "REQ-B")
	str := s2.String()
	if str != "REQ-A|REQ-B" && str != "REQ-B|REQ-A" {
		t.Errorf("StringSet.String() = %q, want pipe-separated", str)
	}

	// Test ParseStringSet
	s3 := ParseStringSet("REQ-A|REQ-B|REQ-C")
	if s3.Len() != 3 {
		t.Errorf("ParseStringSet: expected 3 items, got %d", s3.Len())
	}
}

func TestRequirement(t *testing.T) {
	req := NewRequirement("REQ-TEST-001")

	// Test defaults
	if req.Status != StatusMissing {
		t.Errorf("Default status should be MISSING, got %v", req.Status)
	}
	if req.Priority != PriorityMedium {
		t.Errorf("Default priority should be MEDIUM, got %v", req.Priority)
	}

	// Test HasTest
	if req.HasTest() {
		t.Error("New requirement should not have test")
	}
	req.TestModule = "tests/test.py"
	req.TestFunction = "test_func"
	if !req.HasTest() {
		t.Error("Requirement with test info should have test")
	}

	// Test IsComplete
	if req.IsComplete() {
		t.Error("New requirement should not be complete")
	}
	req.Status = StatusComplete
	if !req.IsComplete() {
		t.Error("Complete requirement should be complete")
	}

	// Test Clone
	clone := req.Clone()
	if clone.ReqID != req.ReqID {
		t.Error("Clone should have same ReqID")
	}
	clone.ReqID = "MODIFIED"
	if req.ReqID == "MODIFIED" {
		t.Error("Clone should not affect original")
	}
}

func TestDatabase(t *testing.T) {
	db := NewDatabase()

	// Test Add
	req1 := NewRequirement("REQ-001")
	req1.Category = "TEST"
	if err := db.Add(req1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Test Get
	got := db.Get("REQ-001")
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.Category != "TEST" {
		t.Errorf("Get returned wrong category: %s", got.Category)
	}

	// Test Exists
	if !db.Exists("REQ-001") {
		t.Error("Exists returned false for existing requirement")
	}
	if db.Exists("REQ-999") {
		t.Error("Exists returned true for non-existing requirement")
	}

	// Test duplicate Add
	if err := db.Add(req1); err == nil {
		t.Error("Add should fail for duplicate ID")
	}

	// Test Len
	if db.Len() != 1 {
		t.Errorf("Len = %d, want 1", db.Len())
	}

	// Test Remove
	if err := db.Remove("REQ-001"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if db.Exists("REQ-001") {
		t.Error("Requirement should not exist after Remove")
	}

	// Test Remove non-existing
	if err := db.Remove("REQ-999"); err == nil {
		t.Error("Remove should fail for non-existing requirement")
	}
}

func TestDatabaseFilter(t *testing.T) {
	db := NewDatabase()

	// Add test requirements
	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "CLI", Status: StatusComplete, Priority: PriorityHigh, Phase: 1},
		{ReqID: "REQ-002", Category: "CLI", Status: StatusMissing, Priority: PriorityMedium, Phase: 1},
		{ReqID: "REQ-003", Category: "DATA", Status: StatusComplete, Priority: PriorityLow, Phase: 2},
		{ReqID: "REQ-004", Category: "DATA", Status: StatusPartial, Priority: PriorityP0, Phase: 2},
	}

	for _, req := range reqs {
		req.Dependencies = make(StringSet)
		req.Blocks = make(StringSet)
		req.Extra = make(map[string]string)
		db.Add(req)
	}

	// Test filter by status
	complete := StatusComplete
	filtered := db.Filter(FilterOptions{Status: &complete})
	if len(filtered) != 2 {
		t.Errorf("Filter by COMPLETE: got %d, want 2", len(filtered))
	}

	// Test filter by category
	filtered = db.Filter(FilterOptions{Category: "CLI"})
	if len(filtered) != 2 {
		t.Errorf("Filter by CLI category: got %d, want 2", len(filtered))
	}

	// Test filter by phase
	phase := 2
	filtered = db.Filter(FilterOptions{Phase: &phase})
	if len(filtered) != 2 {
		t.Errorf("Filter by phase 2: got %d, want 2", len(filtered))
	}

	// Test StatusCounts
	counts := db.StatusCounts()
	if counts[StatusComplete] != 2 {
		t.Errorf("StatusCounts[COMPLETE] = %d, want 2", counts[StatusComplete])
	}

	// Test CompletionPercentage
	pct := db.CompletionPercentage()
	// 2 complete (200%) + 1 partial (50%) + 1 missing (0%) = 250% / 4 = 62.5%
	if pct != 62.5 {
		t.Errorf("CompletionPercentage = %f, want 62.5", pct)
	}
}

func TestCSVRoundTrip(t *testing.T) {
	// Create a database
	db := NewDatabase()

	req := NewRequirement("REQ-TEST-001")
	req.Category = "CLI"
	req.Subcategory = "Foundation"
	req.RequirementText = "Test requirement with special chars: comma, \"quotes\""
	req.Status = StatusComplete
	req.Priority = PriorityHigh
	req.Phase = 1
	req.EffortWeeks = 2.5
	req.Dependencies.Add("REQ-A")
	req.Dependencies.Add("REQ-B")
	req.Blocks.Add("REQ-C")
	req.Extra["custom_field"] = "custom_value"

	if err := db.Add(req); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := db.WriteCSV(&buf); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	// Read back
	db2, err := ReadCSV(&buf)
	if err != nil {
		t.Fatalf("ReadCSV failed: %v", err)
	}

	// Verify
	if db2.Len() != 1 {
		t.Errorf("Round-trip: got %d requirements, want 1", db2.Len())
	}

	req2 := db2.Get("REQ-TEST-001")
	if req2 == nil {
		t.Fatal("Round-trip: requirement not found")
	}

	if req2.Category != req.Category {
		t.Errorf("Category: got %q, want %q", req2.Category, req.Category)
	}
	if req2.RequirementText != req.RequirementText {
		t.Errorf("RequirementText: got %q, want %q", req2.RequirementText, req.RequirementText)
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
	if req2.EffortWeeks != req.EffortWeeks {
		t.Errorf("EffortWeeks: got %f, want %f", req2.EffortWeeks, req.EffortWeeks)
	}
	if !req2.Dependencies.Contains("REQ-A") || !req2.Dependencies.Contains("REQ-B") {
		t.Errorf("Dependencies not preserved: %v", req2.Dependencies)
	}
	if !req2.Blocks.Contains("REQ-C") {
		t.Errorf("Blocks not preserved: %v", req2.Blocks)
	}
}

func TestLoadRealDatabase(t *testing.T) {
	// Try multiple paths to find the database
	paths := []string{
		".rtmx/database.csv",
		"../../.rtmx/database.csv", // From internal/database/
	}

	var db *Database
	var err error
	for _, path := range paths {
		db, err = Load(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Skipf("Skipping real database test: %v", err)
	}

	// Verify we loaded some requirements
	if db.Len() == 0 {
		t.Error("Loaded database is empty")
	}

	// Check for expected requirement
	req := db.Get("REQ-GO-001")
	if req == nil {
		t.Error("Expected REQ-GO-001 in database")
	} else {
		if req.Category != "CLI" {
			t.Errorf("REQ-GO-001 category = %q, want CLI", req.Category)
		}
		if req.Status != StatusComplete {
			t.Errorf("REQ-GO-001 status = %v, want COMPLETE", req.Status)
		}
	}

	// Test statistics
	counts := db.StatusCounts()
	t.Logf("Database stats: %d total, %d complete, %d partial, %d missing",
		db.Len(),
		counts[StatusComplete],
		counts[StatusPartial],
		counts[StatusMissing])

	pct := db.CompletionPercentage()
	t.Logf("Completion: %.1f%%", pct)
}

func TestNormalizeColumnName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"req_id", "req_id"},
		{"ReqId", "req_id"},
		{"RequirementText", "requirement_text"},
		{"REQUIREMENT_TEXT", "requirement_text"},
		{"testModule", "test_module"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeColumnName(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeColumnName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseCSVWithHeaders(t *testing.T) {
	csvData := `req_id,category,requirement_text,status,priority,phase,dependencies
REQ-001,CLI,First requirement,COMPLETE,HIGH,1,
REQ-002,DATA,Second requirement,MISSING,MEDIUM,2,REQ-001
REQ-003,TEST,Third requirement,PARTIAL,LOW,3,REQ-001|REQ-002
`

	db, err := ReadCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ReadCSV failed: %v", err)
	}

	if db.Len() != 3 {
		t.Errorf("Expected 3 requirements, got %d", db.Len())
	}

	// Check dependencies
	req3 := db.Get("REQ-003")
	if req3 == nil {
		t.Fatal("REQ-003 not found")
	}
	if !req3.Dependencies.Contains("REQ-001") || !req3.Dependencies.Contains("REQ-002") {
		t.Errorf("REQ-003 dependencies = %v, want REQ-001 and REQ-002", req3.Dependencies)
	}

	// Check blocking analysis
	req1 := db.Get("REQ-001")
	if req1 == nil {
		t.Fatal("REQ-001 not found")
	}
	if req1.IsBlocked(db) {
		t.Error("REQ-001 should not be blocked")
	}

	req2 := db.Get("REQ-002")
	if req2 == nil {
		t.Fatal("REQ-002 not found")
	}
	// REQ-002 depends on REQ-001 which is COMPLETE, so not blocked
	if req2.IsBlocked(db) {
		t.Error("REQ-002 should not be blocked (REQ-001 is complete)")
	}
}
