package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/graph"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var validateStagedVerbose bool

var validateStagedCmd = &cobra.Command{
	Use:   "validate-staged [FILES...]",
	Short: "Validate staged RTM CSV files (pre-commit hook)",
	Long: `Validate staged RTM CSV files for use with pre-commit hooks.

Validates only the specified CSV files. Designed to be called from a
pre-commit hook to validate staged RTM database files.

Examples:
    rtmx validate-staged docs/rtm_database.csv
    rtmx validate-staged *.csv`,
	RunE: runValidateStaged,
}

func init() {
	validateStagedCmd.Flags().BoolVarP(&validateStagedVerbose, "verbose", "v", false, "show detailed output")

	rootCmd.AddCommand(validateStagedCmd)
}

// Valid status values for strict validation
var validStatusValues = map[string]bool{
	"COMPLETE":    true,
	"PARTIAL":     true,
	"MISSING":     true,
	"NOT_STARTED": true,
}

// Valid priority values for strict validation
var validPriorityValues = map[string]bool{
	"P0":     true,
	"HIGH":   true,
	"MEDIUM": true,
	"LOW":    true,
}

func runValidateStaged(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if len(args) == 0 {
		// No files to validate - success
		return nil
	}

	if validateStagedVerbose {
		cmd.Printf("Validating %d file(s)...\n", len(args))
	}

	var allErrors []string

	for _, filePath := range args {
		// Skip non-CSV files
		if !strings.HasSuffix(filePath, ".csv") {
			continue
		}

		errors := validateCSVFile(filePath)
		allErrors = append(allErrors, errors...)
	}

	if len(allErrors) > 0 {
		cmd.Printf("%s\n", output.Color("Validation failed:", output.Red))
		for _, err := range allErrors {
			cmd.Printf("  %s %s\n", output.Color("✗", output.Red), err)
		}
		return NewExitError(1, "validation failed")
	}

	if validateStagedVerbose {
		cmd.Printf("%s\n", output.Color("✓ All files passed validation", output.Green))
	}

	return nil
}

func validateCSVFile(filePath string) []string {
	var errors []string

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("%s: File not found", filePath))
		return errors
	}

	// Open and validate raw CSV content
	file, err := os.Open(filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("%s: Failed to open: %v", filePath, err))
		return errors
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		errors = append(errors, fmt.Sprintf("%s: Failed to read header: %v", filePath, err))
		return errors
	}

	// Build column index
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	// Check required columns
	requiredCols := []string{"req_id", "category", "requirement_text", "status"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			errors = append(errors, fmt.Sprintf("%s: Missing required column: %s", filePath, col))
		}
	}

	if len(errors) > 0 {
		return errors // Can't continue without required columns
	}

	// Validate rows
	seenIDs := make(map[string]bool)
	rowNum := 1 // Header is row 1

	for {
		rowNum++
		row, err := reader.Read()
		if err != nil {
			break // EOF or error
		}

		// Get req_id
		reqID := ""
		if idx, ok := colIndex["req_id"]; ok && idx < len(row) {
			reqID = strings.TrimSpace(row[idx])
		}

		// Check for duplicate IDs
		if reqID != "" {
			if seenIDs[reqID] {
				errors = append(errors, fmt.Sprintf("%s: Row %d: Duplicate requirement ID '%s'", filePath, rowNum, reqID))
			}
			seenIDs[reqID] = true
		}

		// Validate status value (strict)
		if idx, ok := colIndex["status"]; ok && idx < len(row) {
			statusVal := strings.ToUpper(strings.TrimSpace(row[idx]))
			if statusVal != "" && !validStatusValues[statusVal] {
				errors = append(errors, fmt.Sprintf("%s: Row %d (%s): Invalid status '%s' (valid: COMPLETE, PARTIAL, MISSING, NOT_STARTED)",
					filePath, rowNum, reqID, statusVal))
			}
		}

		// Validate priority value (strict)
		if idx, ok := colIndex["priority"]; ok && idx < len(row) {
			priorityVal := strings.ToUpper(strings.TrimSpace(row[idx]))
			if priorityVal != "" && !validPriorityValues[priorityVal] {
				errors = append(errors, fmt.Sprintf("%s: Row %d (%s): Invalid priority '%s' (valid: P0, HIGH, MEDIUM, LOW)",
					filePath, rowNum, reqID, priorityVal))
			}
		}
	}

	// If raw validation passed, try model-level validation
	if len(errors) == 0 {
		db, err := database.Load(filePath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: Failed to parse CSV: %v", filePath, err))
			return errors
		}

		// Check for cycles (blocking error)
		g := graph.NewGraph(db)
		cycles := g.FindCycles()
		if len(cycles) > 0 {
			errors = append(errors, fmt.Sprintf("%s: Found %d circular dependency(s)", filePath, len(cycles)))
		}

		// Validate dependencies exist
		for _, req := range db.All() {
			for dep := range req.Dependencies {
				if db.Get(dep) == nil {
					errors = append(errors, fmt.Sprintf("%s: %s references unknown dependency: %s", filePath, req.ReqID, dep))
				}
			}
			for blocked := range req.Blocks {
				if db.Get(blocked) == nil {
					errors = append(errors, fmt.Sprintf("%s: %s references unknown blocked requirement: %s", filePath, req.ReqID, blocked))
				}
			}
		}
	}

	return errors
}
