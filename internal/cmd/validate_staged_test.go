package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateStagedValidCSV(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid CSV file
	validCSV := filepath.Join(tmpDir, "valid.csv")
	validContent := `req_id,category,subcategory,requirement_text,status,priority,phase
REQ-TEST-001,TEST,Unit,Test requirement 1,COMPLETE,HIGH,1
REQ-TEST-002,TEST,Unit,Test requirement 2,MISSING,MEDIUM,2
`
	if err := os.WriteFile(validCSV, []byte(validContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test validation
	errors := validateCSVFile(validCSV)
	if len(errors) > 0 {
		t.Errorf("Expected no errors for valid CSV, got: %v", errors)
	}
}

func TestValidateStagedMissingColumns(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CSV missing required columns
	invalidCSV := filepath.Join(tmpDir, "missing_cols.csv")
	invalidContent := `req_id,category
REQ-TEST-001,TEST
`
	if err := os.WriteFile(invalidCSV, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	errors := validateCSVFile(invalidCSV)
	if len(errors) == 0 {
		t.Error("Expected errors for missing columns")
	}

	// Should report missing status and requirement_text
	foundStatus := false
	foundReqText := false
	for _, err := range errors {
		if containsSubstr(err, "status") {
			foundStatus = true
		}
		if containsSubstr(err, "requirement_text") {
			foundReqText = true
		}
	}
	if !foundStatus {
		t.Error("Expected error about missing status column")
	}
	if !foundReqText {
		t.Error("Expected error about missing requirement_text column")
	}
}

func TestValidateStagedDuplicateIDs(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CSV with duplicate IDs
	duplicateCSV := filepath.Join(tmpDir, "duplicates.csv")
	duplicateContent := `req_id,category,subcategory,requirement_text,status,priority,phase
REQ-TEST-001,TEST,Unit,Test requirement 1,COMPLETE,HIGH,1
REQ-TEST-001,TEST,Unit,Duplicate ID,MISSING,MEDIUM,2
`
	if err := os.WriteFile(duplicateCSV, []byte(duplicateContent), 0644); err != nil {
		t.Fatal(err)
	}

	errors := validateCSVFile(duplicateCSV)
	if len(errors) == 0 {
		t.Error("Expected duplicate ID error")
	}

	found := false
	for _, err := range errors {
		if containsSubstr(err, "Duplicate") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected duplicate ID error, got: %v", errors)
	}
}

func TestValidateStagedInvalidStatus(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CSV with invalid status
	invalidCSV := filepath.Join(tmpDir, "invalid_status.csv")
	invalidContent := `req_id,category,subcategory,requirement_text,status,priority,phase
REQ-TEST-001,TEST,Unit,Test requirement 1,INVALID_STATUS,HIGH,1
`
	if err := os.WriteFile(invalidCSV, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	errors := validateCSVFile(invalidCSV)
	if len(errors) == 0 {
		t.Error("Expected invalid status error")
	}

	found := false
	for _, err := range errors {
		if containsSubstr(err, "Invalid status") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected invalid status error, got: %v", errors)
	}
}

func TestValidateStagedInvalidPriority(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CSV with invalid priority
	invalidCSV := filepath.Join(tmpDir, "invalid_priority.csv")
	invalidContent := `req_id,category,subcategory,requirement_text,status,priority,phase
REQ-TEST-001,TEST,Unit,Test requirement 1,COMPLETE,INVALID_PRIORITY,1
`
	if err := os.WriteFile(invalidCSV, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	errors := validateCSVFile(invalidCSV)
	if len(errors) == 0 {
		t.Error("Expected invalid priority error")
	}

	found := false
	for _, err := range errors {
		if containsSubstr(err, "Invalid priority") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected invalid priority error, got: %v", errors)
	}
}

func TestValidateStagedCycleDetection(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-validate-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create CSV with dependency cycle
	cycleCSV := filepath.Join(tmpDir, "cycle.csv")
	cycleContent := `req_id,category,subcategory,requirement_text,status,priority,phase,dependencies
REQ-A,TEST,Unit,Requirement A,MISSING,HIGH,1,REQ-B
REQ-B,TEST,Unit,Requirement B,MISSING,HIGH,1,REQ-C
REQ-C,TEST,Unit,Requirement C,MISSING,HIGH,1,REQ-A
`
	if err := os.WriteFile(cycleCSV, []byte(cycleContent), 0644); err != nil {
		t.Fatal(err)
	}

	errors := validateCSVFile(cycleCSV)
	if len(errors) == 0 {
		t.Error("Expected cycle detection error")
	}

	found := false
	for _, err := range errors {
		if containsSubstr(err, "circular") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected circular dependency error, got: %v", errors)
	}
}

func TestValidateStagedFileNotFound(t *testing.T) {
	errors := validateCSVFile("/nonexistent/file.csv")
	if len(errors) == 0 {
		t.Error("Expected file not found error")
	}

	found := false
	for _, err := range errors {
		if containsSubstr(err, "not found") || containsSubstr(err, "no such file") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected file not found error, got: %v", errors)
	}
}

func TestValidateStagedNoFiles(t *testing.T) {
	// Test with empty args - should succeed silently
	cmd := validateStagedCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("validate-staged with no files should not error: %v", err)
	}
}

// Helper function
func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			// Case insensitive comparison
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
