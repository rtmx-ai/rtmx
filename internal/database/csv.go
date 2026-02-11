package database

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CSVColumn represents a column in the CSV file.
type CSVColumn struct {
	Name     string
	Index    int
	Required bool
}

// Standard column names (snake_case).
var standardColumns = []string{
	"req_id",
	"category",
	"subcategory",
	"requirement_text",
	"target_value",
	"test_module",
	"test_function",
	"validation_method",
	"status",
	"priority",
	"phase",
	"notes",
	"effort_weeks",
	"dependencies",
	"blocks",
	"assignee",
	"sprint",
	"started_date",
	"completed_date",
	"requirement_file",
	"external_id",
}

// Load loads a database from a CSV file.
func Load(path string) (*Database, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer file.Close()

	db, err := ReadCSV(file)
	if err != nil {
		return nil, err
	}

	db.path = path
	return db, nil
}

// Save saves the database to a CSV file.
func (db *Database) Save(path string) error {
	if path == "" {
		path = db.path
	}
	if path == "" {
		return fmt.Errorf("no path specified for saving database")
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer file.Close()

	if err := db.WriteCSV(file); err != nil {
		return err
	}

	db.path = path
	db.dirty = false
	return nil
}

// ReadCSV reads requirements from a CSV reader.
func ReadCSV(r io.Reader) (*Database, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1 // Allow variable fields

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Build column index
	colIndex := make(map[string]int)
	extraCols := make([]string, 0)
	for i, col := range header {
		normalized := normalizeColumnName(col)
		colIndex[normalized] = i

		// Track extra columns not in standard set
		isStandard := false
		for _, std := range standardColumns {
			if normalized == std {
				isStandard = true
				break
			}
		}
		if !isStandard {
			extraCols = append(extraCols, col)
		}
	}

	// Verify required columns
	requiredCols := []string{"req_id", "category", "requirement_text"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	db := NewDatabase()

	// Read rows
	lineNum := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", lineNum+1, err)
		}
		lineNum++

		req, err := parseRow(record, colIndex, extraCols)
		if err != nil {
			return nil, fmt.Errorf("failed to parse row %d: %w", lineNum, err)
		}

		if err := db.Add(req); err != nil {
			return nil, fmt.Errorf("row %d: %w", lineNum, err)
		}
	}

	return db, nil
}

// WriteCSV writes the database to a CSV writer.
func (db *Database) WriteCSV(w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Collect all extra columns used
	extraCols := make(map[string]bool)
	for _, req := range db.All() {
		for k := range req.Extra {
			extraCols[k] = true
		}
	}

	// Build header
	header := make([]string, len(standardColumns))
	copy(header, standardColumns)
	extraColsList := make([]string, 0, len(extraCols))
	for col := range extraCols {
		extraColsList = append(extraColsList, col)
	}
	// Sort extra columns for deterministic output
	for i := 0; i < len(extraColsList)-1; i++ {
		for j := i + 1; j < len(extraColsList); j++ {
			if extraColsList[i] > extraColsList[j] {
				extraColsList[i], extraColsList[j] = extraColsList[j], extraColsList[i]
			}
		}
	}
	header = append(header, extraColsList...)

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, req := range db.All() {
		row := formatRow(req, header)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row for %s: %w", req.ReqID, err)
		}
	}

	return nil
}

// normalizeColumnName converts column names to snake_case.
func normalizeColumnName(name string) string {
	name = strings.TrimSpace(name)

	// If already snake_case or all lowercase, just lowercase it
	if strings.Contains(name, "_") || strings.ToLower(name) == name {
		return strings.ToLower(name)
	}

	// Handle PascalCase/camelCase -> snake_case
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Don't add underscore if previous char was also uppercase
			// (handles acronyms like "REQId" -> "req_id")
			prev := rune(name[i-1])
			if prev >= 'a' && prev <= 'z' {
				result.WriteByte('_')
			}
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// parseRow parses a CSV row into a Requirement.
func parseRow(record []string, colIndex map[string]int, extraCols []string) (*Requirement, error) {
	getValue := func(col string) string {
		if idx, ok := colIndex[col]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	req := NewRequirement(getValue("req_id"))
	if req.ReqID == "" {
		return nil, fmt.Errorf("req_id is required")
	}

	req.Category = getValue("category")
	req.Subcategory = getValue("subcategory")
	req.RequirementText = getValue("requirement_text")
	req.TargetValue = getValue("target_value")
	req.TestModule = getValue("test_module")
	req.TestFunction = getValue("test_function")
	req.ValidationMethod = getValue("validation_method")
	req.Notes = getValue("notes")
	req.Assignee = getValue("assignee")
	req.Sprint = getValue("sprint")
	req.StartedDate = getValue("started_date")
	req.CompletedDate = getValue("completed_date")
	req.RequirementFile = getValue("requirement_file")
	req.ExternalID = getValue("external_id")

	// Parse status
	statusStr := getValue("status")
	status, err := ParseStatus(statusStr)
	if err != nil {
		// Log warning but don't fail
		status = StatusMissing
	}
	req.Status = status

	// Parse priority
	priorityStr := getValue("priority")
	priority, err := ParsePriority(priorityStr)
	if err != nil {
		priority = PriorityMedium
	}
	req.Priority = priority

	// Parse phase
	phaseStr := getValue("phase")
	if phaseStr != "" {
		if phase, err := strconv.Atoi(phaseStr); err == nil {
			req.Phase = phase
		}
	}

	// Parse effort_weeks
	effortStr := getValue("effort_weeks")
	if effortStr != "" {
		if effort, err := strconv.ParseFloat(effortStr, 64); err == nil {
			req.EffortWeeks = effort
		}
	}

	// Parse dependencies
	req.Dependencies = ParseStringSet(getValue("dependencies"))

	// Parse blocks
	req.Blocks = ParseStringSet(getValue("blocks"))

	// Parse extra columns
	for _, col := range extraCols {
		normalized := normalizeColumnName(col)
		if idx, ok := colIndex[normalized]; ok && idx < len(record) {
			value := strings.TrimSpace(record[idx])
			if value != "" {
				req.Extra[col] = value
			}
		}
	}

	return req, nil
}

// formatRow formats a Requirement as a CSV row.
func formatRow(req *Requirement, header []string) []string {
	row := make([]string, len(header))

	for i, col := range header {
		switch col {
		case "req_id":
			row[i] = req.ReqID
		case "category":
			row[i] = req.Category
		case "subcategory":
			row[i] = req.Subcategory
		case "requirement_text":
			row[i] = req.RequirementText
		case "target_value":
			row[i] = req.TargetValue
		case "test_module":
			row[i] = req.TestModule
		case "test_function":
			row[i] = req.TestFunction
		case "validation_method":
			row[i] = req.ValidationMethod
		case "status":
			row[i] = req.Status.String()
		case "priority":
			row[i] = req.Priority.String()
		case "phase":
			if req.Phase > 0 {
				row[i] = strconv.Itoa(req.Phase)
			}
		case "notes":
			row[i] = req.Notes
		case "effort_weeks":
			if req.EffortWeeks > 0 {
				row[i] = strconv.FormatFloat(req.EffortWeeks, 'f', -1, 64)
			}
		case "dependencies":
			row[i] = req.Dependencies.String()
		case "blocks":
			row[i] = req.Blocks.String()
		case "assignee":
			row[i] = req.Assignee
		case "sprint":
			row[i] = req.Sprint
		case "started_date":
			row[i] = req.StartedDate
		case "completed_date":
			row[i] = req.CompletedDate
		case "requirement_file":
			row[i] = req.RequirementFile
		case "external_id":
			row[i] = req.ExternalID
		default:
			// Extra column
			if val, ok := req.Extra[col]; ok {
				row[i] = val
			}
		}
	}

	return row
}

// FindDatabase searches for a database file starting from the given path.
func FindDatabase(startPath string) (string, error) {
	// Check common locations
	candidates := []string{
		".rtmx/database.csv",
		"docs/rtm_database.csv",
		"rtm_database.csv",
	}

	for _, candidate := range candidates {
		path := startPath + "/" + candidate
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// TODO: Search parent directories

	return "", fmt.Errorf("no RTM database found")
}
