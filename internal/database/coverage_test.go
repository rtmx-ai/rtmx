package database

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDatabasePathAndDirty tests Path, SetPath, IsDirty, MarkClean methods.
func TestDatabasePathAndDirty(t *testing.T) {
	db := NewDatabase()

	// Path should be empty initially
	if db.Path() != "" {
		t.Errorf("Path() = %q, want empty", db.Path())
	}

	// SetPath should set path
	db.SetPath("/tmp/test.csv")
	if db.Path() != "/tmp/test.csv" {
		t.Errorf("Path() = %q, want /tmp/test.csv", db.Path())
	}

	// IsDirty should be false initially
	if db.IsDirty() {
		t.Error("IsDirty() should be false initially")
	}

	// After Add, dirty should be true
	req := NewRequirement("REQ-001")
	req.Category = "TEST"
	_ = db.Add(req)
	if !db.IsDirty() {
		t.Error("IsDirty() should be true after Add")
	}

	// MarkClean should clear dirty flag
	db.MarkClean()
	if db.IsDirty() {
		t.Error("IsDirty() should be false after MarkClean")
	}
}

// TestDatabaseIDs tests the IDs method.
func TestDatabaseIDs(t *testing.T) {
	db := NewDatabase()

	req1 := NewRequirement("REQ-001")
	req1.Category = "A"
	_ = db.Add(req1)

	req2 := NewRequirement("REQ-002")
	req2.Category = "B"
	_ = db.Add(req2)

	ids := db.IDs()
	if len(ids) != 2 {
		t.Fatalf("IDs() len = %d, want 2", len(ids))
	}
	if ids[0] != "REQ-001" || ids[1] != "REQ-002" {
		t.Errorf("IDs() = %v, want [REQ-001 REQ-002]", ids)
	}

	// Modifying returned slice should not affect database
	ids[0] = "MODIFIED"
	ids2 := db.IDs()
	if ids2[0] != "REQ-001" {
		t.Error("Modifying IDs() return value should not affect database")
	}
}

// TestDatabaseUpdate tests the Update method.
func TestDatabaseUpdate(t *testing.T) {
	db := NewDatabase()

	req := NewRequirement("REQ-001")
	req.Category = "TEST"
	_ = db.Add(req)

	// Update status with string
	err := db.Update("REQ-001", map[string]interface{}{
		"status": "COMPLETE",
	})
	if err != nil {
		t.Fatalf("Update status string failed: %v", err)
	}
	if db.Get("REQ-001").Status != StatusComplete {
		t.Errorf("Status = %v, want COMPLETE", db.Get("REQ-001").Status)
	}

	// Update status with Status type
	err = db.Update("REQ-001", map[string]interface{}{
		"status": StatusPartial,
	})
	if err != nil {
		t.Fatalf("Update status type failed: %v", err)
	}
	if db.Get("REQ-001").Status != StatusPartial {
		t.Errorf("Status = %v, want PARTIAL", db.Get("REQ-001").Status)
	}

	// Update priority with string
	err = db.Update("REQ-001", map[string]interface{}{
		"priority": "HIGH",
	})
	if err != nil {
		t.Fatalf("Update priority string failed: %v", err)
	}
	if db.Get("REQ-001").Priority != PriorityHigh {
		t.Errorf("Priority = %v, want HIGH", db.Get("REQ-001").Priority)
	}

	// Update priority with Priority type
	err = db.Update("REQ-001", map[string]interface{}{
		"priority": PriorityLow,
	})
	if err != nil {
		t.Fatalf("Update priority type failed: %v", err)
	}
	if db.Get("REQ-001").Priority != PriorityLow {
		t.Errorf("Priority = %v, want LOW", db.Get("REQ-001").Priority)
	}

	// Update phase
	err = db.Update("REQ-001", map[string]interface{}{
		"phase": 3,
	})
	if err != nil {
		t.Fatalf("Update phase failed: %v", err)
	}
	if db.Get("REQ-001").Phase != 3 {
		t.Errorf("Phase = %d, want 3", db.Get("REQ-001").Phase)
	}

	// Update string fields
	err = db.Update("REQ-001", map[string]interface{}{
		"assignee":       "dev1",
		"sprint":         "s1",
		"test_module":    "test.go",
		"test_function":  "TestX",
		"started_date":   "2024-01-01",
		"completed_date": "2024-02-01",
	})
	if err != nil {
		t.Fatalf("Update string fields failed: %v", err)
	}
	got := db.Get("REQ-001")
	if got.Assignee != "dev1" {
		t.Errorf("Assignee = %q, want dev1", got.Assignee)
	}
	if got.Sprint != "s1" {
		t.Errorf("Sprint = %q, want s1", got.Sprint)
	}
	if got.TestModule != "test.go" {
		t.Errorf("TestModule = %q, want test.go", got.TestModule)
	}
	if got.TestFunction != "TestX" {
		t.Errorf("TestFunction = %q, want TestX", got.TestFunction)
	}

	// Update unknown field (goes to Extra)
	err = db.Update("REQ-001", map[string]interface{}{
		"custom_field": "custom_value",
	})
	if err != nil {
		t.Fatalf("Update custom field failed: %v", err)
	}
	if db.Get("REQ-001").Extra["custom_field"] != "custom_value" {
		t.Errorf("Extra[custom_field] = %q, want custom_value", db.Get("REQ-001").Extra["custom_field"])
	}

	// Update non-existing requirement
	err = db.Update("REQ-999", map[string]interface{}{"status": "COMPLETE"})
	if err == nil {
		t.Error("Update non-existing requirement should fail")
	}

	// Update with invalid status string
	err = db.Update("REQ-001", map[string]interface{}{"status": "INVALID_STATUS"})
	if err == nil {
		t.Error("Update with invalid status should fail")
	}

	// Update with invalid priority string
	err = db.Update("REQ-001", map[string]interface{}{"priority": "INVALID_PRIORITY"})
	if err == nil {
		t.Error("Update with invalid priority should fail")
	}
}

// TestPriorityCounts tests the PriorityCounts method.
func TestPriorityCounts(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "A", Status: StatusComplete, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "A", Status: StatusMissing, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "B", Status: StatusComplete, Priority: PriorityLow,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	counts := db.PriorityCounts()
	if counts[PriorityHigh] != 2 {
		t.Errorf("PriorityCounts[HIGH] = %d, want 2", counts[PriorityHigh])
	}
	if counts[PriorityLow] != 1 {
		t.Errorf("PriorityCounts[LOW] = %d, want 1", counts[PriorityLow])
	}
}

// TestCategories tests the Categories method.
func TestCategories(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "CLI", Status: StatusComplete, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "DATA", Status: StatusMissing, Priority: PriorityMedium,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "CLI", Status: StatusPartial, Priority: PriorityLow,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	cats := db.Categories()
	if len(cats) != 2 {
		t.Fatalf("Categories() len = %d, want 2", len(cats))
	}
	// Sorted alphabetically
	if cats[0] != "CLI" || cats[1] != "DATA" {
		t.Errorf("Categories() = %v, want [CLI DATA]", cats)
	}
}

// TestPhases tests the Phases method.
func TestPhases(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "A", Phase: 2, Status: StatusComplete, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "A", Phase: 1, Status: StatusMissing, Priority: PriorityMedium,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "B", Phase: 2, Status: StatusPartial, Priority: PriorityLow,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-004", Category: "B", Phase: 0, Status: StatusMissing, Priority: PriorityMedium,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	phases := db.Phases()
	if len(phases) != 2 {
		t.Fatalf("Phases() len = %d, want 2", len(phases))
	}
	// Sorted numerically, phase 0 excluded
	if phases[0] != 1 || phases[1] != 2 {
		t.Errorf("Phases() = %v, want [1 2]", phases)
	}
}

// TestByCategory tests the ByCategory method.
func TestByCategory(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "CLI", Status: StatusComplete, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "DATA", Status: StatusMissing, Priority: PriorityMedium,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "CLI", Status: StatusPartial, Priority: PriorityLow,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	byCat := db.ByCategory()
	if len(byCat["CLI"]) != 2 {
		t.Errorf("ByCategory[CLI] len = %d, want 2", len(byCat["CLI"]))
	}
	if len(byCat["DATA"]) != 1 {
		t.Errorf("ByCategory[DATA] len = %d, want 1", len(byCat["DATA"]))
	}
}

// TestByPhase tests the ByPhase method.
func TestByPhase(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "A", Phase: 1, Status: StatusComplete, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "A", Phase: 2, Status: StatusMissing, Priority: PriorityMedium,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "B", Phase: 1, Status: StatusPartial, Priority: PriorityLow,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	byPhase := db.ByPhase()
	if len(byPhase[1]) != 2 {
		t.Errorf("ByPhase[1] len = %d, want 2", len(byPhase[1]))
	}
	if len(byPhase[2]) != 1 {
		t.Errorf("ByPhase[2] len = %d, want 1", len(byPhase[2]))
	}
}

// TestIncompleteAndComplete tests the Incomplete and Complete methods.
func TestIncompleteAndComplete(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "A", Status: StatusComplete, Priority: PriorityHigh,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "A", Status: StatusMissing, Priority: PriorityMedium,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "B", Status: StatusPartial, Priority: PriorityLow,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	incomplete := db.Incomplete()
	if len(incomplete) != 2 {
		t.Errorf("Incomplete() len = %d, want 2", len(incomplete))
	}

	complete := db.Complete()
	if len(complete) != 1 {
		t.Errorf("Complete() len = %d, want 1", len(complete))
	}
}

// TestBacklog tests the Backlog method.
func TestBacklog(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "A", Status: StatusComplete, Priority: PriorityHigh, Phase: 1,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "A", Status: StatusMissing, Priority: PriorityLow, Phase: 2,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "B", Status: StatusPartial, Priority: PriorityP0, Phase: 1,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-004", Category: "B", Status: StatusMissing, Priority: PriorityP0, Phase: 2,
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	backlog := db.Backlog()
	// Should not include complete items
	if len(backlog) != 3 {
		t.Fatalf("Backlog() len = %d, want 3", len(backlog))
	}
	// P0 items first, then sorted by phase, then by ID
	if backlog[0].ReqID != "REQ-003" {
		t.Errorf("Backlog[0] = %s, want REQ-003 (P0, phase 1)", backlog[0].ReqID)
	}
	if backlog[1].ReqID != "REQ-004" {
		t.Errorf("Backlog[1] = %s, want REQ-004 (P0, phase 2)", backlog[1].ReqID)
	}
}

// TestFilterAdvanced tests advanced filter options.
func TestFilterAdvanced(t *testing.T) {
	db := NewDatabase()

	reqs := []*Requirement{
		{ReqID: "REQ-001", Category: "CLI", Status: StatusComplete, Priority: PriorityHigh, Phase: 1,
			TestModule: "test.go", TestFunction: "TestA", Assignee: "dev1",
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-002", Category: "DATA", Status: StatusMissing, Priority: PriorityMedium, Phase: 2,
			Assignee: "dev2",
			Dependencies: make(StringSet), Blocks: make(StringSet), Extra: make(map[string]string)},
		{ReqID: "REQ-003", Category: "CLI", Status: StatusMissing, Priority: PriorityLow, Phase: 1,
			Dependencies: database_newStringSetWithDep("REQ-002"), Blocks: make(StringSet), Extra: make(map[string]string)},
	}
	// REQ-002 is incomplete so REQ-003 is blocked
	for _, req := range reqs {
		_ = db.Add(req)
	}

	// Filter by HasTest
	hasTest := true
	filtered := db.Filter(FilterOptions{HasTest: &hasTest})
	if len(filtered) != 1 {
		t.Errorf("Filter HasTest=true: got %d, want 1", len(filtered))
	}

	noTest := false
	filtered = db.Filter(FilterOptions{HasTest: &noTest})
	if len(filtered) != 2 {
		t.Errorf("Filter HasTest=false: got %d, want 2", len(filtered))
	}

	// Filter by IsComplete
	isComplete := true
	filtered = db.Filter(FilterOptions{IsComplete: &isComplete})
	if len(filtered) != 1 {
		t.Errorf("Filter IsComplete=true: got %d, want 1", len(filtered))
	}

	notComplete := false
	filtered = db.Filter(FilterOptions{IsComplete: &notComplete})
	if len(filtered) != 2 {
		t.Errorf("Filter IsComplete=false: got %d, want 2", len(filtered))
	}

	// Filter by IsBlocked
	isBlocked := true
	filtered = db.Filter(FilterOptions{IsBlocked: &isBlocked})
	if len(filtered) != 1 {
		t.Errorf("Filter IsBlocked=true: got %d, want 1", len(filtered))
	}

	notBlocked := false
	filtered = db.Filter(FilterOptions{IsBlocked: &notBlocked})
	if len(filtered) != 2 {
		t.Errorf("Filter IsBlocked=false: got %d, want 2", len(filtered))
	}

	// Filter by Assignee
	filtered = db.Filter(FilterOptions{Assignee: "dev1"})
	if len(filtered) != 1 {
		t.Errorf("Filter Assignee=dev1: got %d, want 1", len(filtered))
	}

	// Filter by Priority
	high := PriorityHigh
	filtered = db.Filter(FilterOptions{Priority: &high})
	if len(filtered) != 1 {
		t.Errorf("Filter Priority=HIGH: got %d, want 1", len(filtered))
	}
}

// helper to create a StringSet with a dependency
func database_newStringSetWithDep(dep string) StringSet {
	s := make(StringSet)
	s.Add(dep)
	return s
}

// TestAddEmptyID tests that Add rejects empty IDs.
func TestAddEmptyID(t *testing.T) {
	db := NewDatabase()
	req := NewRequirement("")
	err := db.Add(req)
	if err == nil {
		t.Error("Add with empty ID should fail")
	}
}

// TestCompletionPercentageEmpty tests CompletionPercentage on empty database.
func TestCompletionPercentageEmpty(t *testing.T) {
	db := NewDatabase()
	pct := db.CompletionPercentage()
	if pct != 0 {
		t.Errorf("CompletionPercentage on empty db = %f, want 0", pct)
	}
}

// TestSaveAndLoad tests Save and Load round-trip.
func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.csv")

	db := NewDatabase()
	req := NewRequirement("REQ-001")
	req.Category = "TEST"
	req.Status = StatusComplete
	_ = db.Add(req)

	// Save
	err := db.Save(dbPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify not dirty after save
	if db.IsDirty() {
		t.Error("Should not be dirty after save")
	}

	// Load
	loaded, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Len() != 1 {
		t.Errorf("Loaded db has %d reqs, want 1", loaded.Len())
	}
	if loaded.Get("REQ-001").Status != StatusComplete {
		t.Errorf("Loaded status = %v, want COMPLETE", loaded.Get("REQ-001").Status)
	}
}

// TestSaveNoPath tests Save with no path.
func TestSaveNoPath(t *testing.T) {
	db := NewDatabase()
	err := db.Save("")
	if err == nil {
		t.Error("Save with no path should fail")
	}
}

// TestFindDatabase tests the FindDatabase function.
func TestFindDatabase(t *testing.T) {
	tmpDir := t.TempDir()

	// No database present
	_, err := FindDatabase(tmpDir)
	if err == nil {
		t.Error("FindDatabase should fail when no database exists")
	}

	// Create a database file
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	dbPath := filepath.Join(rtmxDir, "database.csv")
	_ = os.WriteFile(dbPath, []byte("req_id,category,requirement_text\nREQ-001,TEST,Test\n"), 0644)

	found, err := FindDatabase(tmpDir)
	if err != nil {
		t.Fatalf("FindDatabase failed: %v", err)
	}
	if found != tmpDir+"/.rtmx/database.csv" {
		t.Errorf("FindDatabase = %q, want %q", found, tmpDir+"/.rtmx/database.csv")
	}
}

// TestCSVMissingRequiredColumn tests ReadCSV with missing required columns.
func TestCSVMissingRequiredColumn(t *testing.T) {
	csvData := "category,requirement_text\nTEST,Test\n"
	_, err := ReadCSV(strings.NewReader(csvData))
	if err == nil {
		t.Error("ReadCSV should fail when req_id column is missing")
	}
}

// TestWriteCSVExtraColumns tests WriteCSV with extra columns.
func TestWriteCSVExtraColumns(t *testing.T) {
	db := NewDatabase()

	req1 := NewRequirement("REQ-001")
	req1.Category = "TEST"
	req1.Extra["zfield"] = "zval"
	req1.Extra["afield"] = "aval"
	_ = db.Add(req1)

	var buf bytes.Buffer
	err := db.WriteCSV(&buf)
	if err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	output := buf.String()
	// Extra columns should be in the header sorted alphabetically
	if !strings.Contains(output, "afield") || !strings.Contains(output, "zfield") {
		t.Errorf("WriteCSV should include extra columns, got:\n%s", output)
	}
}

// TestRequirementBlockingDeps tests BlockingDeps method.
func TestRequirementBlockingDeps(t *testing.T) {
	db := NewDatabase()

	dep := NewRequirement("DEP-001")
	dep.Category = "TEST"
	dep.Status = StatusMissing
	_ = db.Add(dep)

	completeDep := NewRequirement("DEP-002")
	completeDep.Category = "TEST"
	completeDep.Status = StatusComplete
	_ = db.Add(completeDep)

	req := NewRequirement("REQ-001")
	req.Category = "TEST"
	req.Dependencies.Add("DEP-001")
	req.Dependencies.Add("DEP-002")
	req.Dependencies.Add("external:DEP-003") // cross-repo dep
	_ = db.Add(req)

	blocking := req.BlockingDeps(db)
	if len(blocking) != 1 {
		t.Errorf("BlockingDeps len = %d, want 1", len(blocking))
	}
	if len(blocking) > 0 && blocking[0] != "DEP-001" {
		t.Errorf("BlockingDeps[0] = %q, want DEP-001", blocking[0])
	}
}

// TestRequirementDates tests SetStartedDate and SetCompletedDate.
func TestRequirementDates(t *testing.T) {
	req := NewRequirement("REQ-001")

	// SetStartedDate should set date if empty
	req.SetStartedDate()
	if req.StartedDate == "" {
		t.Error("SetStartedDate should set date")
	}

	// SetStartedDate should not overwrite existing date
	req.StartedDate = "2024-01-01"
	req.SetStartedDate()
	if req.StartedDate != "2024-01-01" {
		t.Errorf("SetStartedDate should not overwrite, got %q", req.StartedDate)
	}

	// SetCompletedDate should always set
	req.SetCompletedDate()
	if req.CompletedDate == "" {
		t.Error("SetCompletedDate should set date")
	}
}

// TestRequirementIsHighPriority tests the IsHighPriority method.
func TestRequirementIsHighPriority(t *testing.T) {
	tests := []struct {
		priority Priority
		expected bool
	}{
		{PriorityP0, true},
		{PriorityHigh, true},
		{PriorityMedium, false},
		{PriorityLow, false},
	}

	for _, tt := range tests {
		req := NewRequirement("REQ-001")
		req.Priority = tt.priority
		if got := req.IsHighPriority(); got != tt.expected {
			t.Errorf("IsHighPriority() for %s = %v, want %v", tt.priority, got, tt.expected)
		}
	}
}

// TestPriorityWeight tests the Weight method for all priority values.
func TestPriorityWeight(t *testing.T) {
	tests := []struct {
		priority Priority
		weight   int
	}{
		{PriorityP0, 0},
		{PriorityHigh, 1},
		{PriorityMedium, 2},
		{PriorityLow, 3},
		{Priority("UNKNOWN"), 4},
	}

	for _, tt := range tests {
		if got := tt.priority.Weight(); got != tt.weight {
			t.Errorf("Weight() for %s = %d, want %d", tt.priority, got, tt.weight)
		}
	}
}

// TestStatusWeight tests the Weight method for all status values.
func TestStatusWeight(t *testing.T) {
	tests := []struct {
		status Status
		weight int
	}{
		{StatusComplete, 0},
		{StatusPartial, 1},
		{StatusMissing, 2},
		{StatusNotStarted, 3},
		{Status("UNKNOWN"), 4},
	}

	for _, tt := range tests {
		if got := tt.status.Weight(); got != tt.weight {
			t.Errorf("Weight() for %s = %d, want %d", tt.status, got, tt.weight)
		}
	}
}

// TestStatusCompletionPercent tests CompletionPercent for all status values.
func TestStatusCompletionPercent(t *testing.T) {
	tests := []struct {
		status  Status
		percent float64
	}{
		{StatusComplete, 100.0},
		{StatusPartial, 50.0},
		{StatusMissing, 0.0},
		{StatusNotStarted, 0.0},
	}

	for _, tt := range tests {
		if got := tt.status.CompletionPercent(); got != tt.percent {
			t.Errorf("CompletionPercent() for %s = %f, want %f", tt.status, got, tt.percent)
		}
	}
}

// TestAllStatusesAndPriorities tests AllStatuses and AllPriorities.
func TestAllStatusesAndPriorities(t *testing.T) {
	statuses := AllStatuses()
	if len(statuses) != 4 {
		t.Errorf("AllStatuses() len = %d, want 4", len(statuses))
	}

	priorities := AllPriorities()
	if len(priorities) != 4 {
		t.Errorf("AllPriorities() len = %d, want 4", len(priorities))
	}
}

// TestStringSetEdgeCases tests StringSet edge cases.
func TestStringSetEdgeCases(t *testing.T) {
	// Empty string set
	s := NewStringSet()
	if s.Len() != 0 {
		t.Errorf("Empty StringSet len = %d, want 0", s.Len())
	}
	if s.String() != "" {
		t.Errorf("Empty StringSet String() = %q, want empty", s.String())
	}

	// Add empty/whitespace strings
	s.Add("")
	s.Add("  ")
	if s.Len() != 0 {
		t.Errorf("StringSet with empty adds len = %d, want 0", s.Len())
	}

	// Parse empty string
	s2 := ParseStringSet("")
	if s2.Len() != 0 {
		t.Errorf("ParseStringSet('') len = %d, want 0", s2.Len())
	}

	// NewStringSet with whitespace items
	s3 := NewStringSet("  a  ", "", "  b  ")
	if s3.Len() != 2 {
		t.Errorf("NewStringSet with whitespace len = %d, want 2", s3.Len())
	}
	if !s3.Contains("a") || !s3.Contains("b") {
		t.Error("NewStringSet should trim whitespace")
	}
}

// TestLoadNonExistent tests Load with non-existent file.
func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/database.csv")
	if err == nil {
		t.Error("Load should fail for non-existent file")
	}
}
