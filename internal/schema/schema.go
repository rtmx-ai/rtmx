// Package schema defines the column schema framework for RTMX databases.
// Schemas declare which columns a database should contain, their types,
// and validation rules. The core schema matches the 21 standard columns;
// extension schemas (e.g., Phoenix) add domain-specific columns.
package schema

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ColumnType represents the data type of a column.
type ColumnType int

const (
	TypeString ColumnType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeDate   // YYYY-MM-DD
	TypeEnum   // one of EnumValues
	TypeSet    // pipe-separated set of values
)

// String returns the human-readable name of a ColumnType.
func (t ColumnType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeBool:
		return "bool"
	case TypeDate:
		return "date"
	case TypeEnum:
		return "enum"
	case TypeSet:
		return "set"
	default:
		return "unknown"
	}
}

// Column defines a single column in a schema.
type Column struct {
	Name        string
	Type        ColumnType
	Required    bool
	Description string
	EnumValues  []string // for TypeEnum and TypeSet
}

// Schema defines the column layout and validation rules for a database.
type Schema struct {
	Name        string
	Description string
	Columns     []Column
	colIndex    map[string]int // lazy index
}

// New creates a schema with the given name and columns.
func New(name string, columns []Column) *Schema {
	s := &Schema{
		Name:    name,
		Columns: columns,
	}
	s.buildIndex()
	return s
}

func (s *Schema) buildIndex() {
	s.colIndex = make(map[string]int, len(s.Columns))
	for i, c := range s.Columns {
		s.colIndex[c.Name] = i
	}
}

// ColumnNames returns the ordered list of column names.
func (s *Schema) ColumnNames() []string {
	names := make([]string, len(s.Columns))
	for i, c := range s.Columns {
		names[i] = c.Name
	}
	return names
}

// HasColumn returns true if the schema defines the named column.
func (s *Schema) HasColumn(name string) bool {
	_, ok := s.colIndex[name]
	return ok
}

// Extend returns a new schema with additional columns appended.
func (s *Schema) Extend(name string, extra []Column) *Schema {
	cols := make([]Column, len(s.Columns), len(s.Columns)+len(extra))
	copy(cols, s.Columns)
	cols = append(cols, extra...)
	extended := New(name, cols)
	extended.Description = s.Description + " (extended)"
	return extended
}

// Validate checks a CSV row (as a map of column name -> value) against
// the schema. Returns a list of validation errors.
func (s *Schema) Validate(row map[string]string) []error {
	var errs []error

	for _, col := range s.Columns {
		val, present := row[col.Name]

		if col.Required && (!present || val == "") {
			errs = append(errs, fmt.Errorf("missing required column %q", col.Name))
			continue
		}

		if !present || val == "" {
			continue
		}

		if err := validateValue(col, val); err != nil {
			errs = append(errs, fmt.Errorf("column %q: %w", col.Name, err))
		}
	}

	return errs
}

// ValidateHeader checks that a CSV header contains all required columns
// and reports any extra or missing columns.
func (s *Schema) ValidateHeader(header []string) (missing, extra []string) {
	headerSet := make(map[string]bool, len(header))
	for _, h := range header {
		headerSet[h] = true
	}

	for _, col := range s.Columns {
		if !headerSet[col.Name] {
			missing = append(missing, col.Name)
		}
	}

	for _, h := range header {
		if !s.HasColumn(h) {
			extra = append(extra, h)
		}
	}

	return missing, extra
}

func validateValue(col Column, val string) error {
	switch col.Type {
	case TypeInt:
		if _, err := strconv.Atoi(val); err != nil {
			return fmt.Errorf("expected int, got %q", val)
		}
	case TypeFloat:
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return fmt.Errorf("expected float, got %q", val)
		}
	case TypeBool:
		lower := strings.ToLower(val)
		if lower != "true" && lower != "false" && lower != "1" && lower != "0" && lower != "" {
			return fmt.Errorf("expected bool, got %q", val)
		}
	case TypeDate:
		if _, err := time.Parse("2006-01-02", val); err != nil {
			return fmt.Errorf("expected date (YYYY-MM-DD), got %q", val)
		}
	case TypeEnum:
		if len(col.EnumValues) > 0 {
			found := false
			for _, ev := range col.EnumValues {
				if strings.EqualFold(val, ev) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("expected one of %v, got %q", col.EnumValues, val)
			}
		}
	case TypeSet:
		parts := strings.Split(val, "|")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if len(col.EnumValues) > 0 {
				found := false
				for _, ev := range col.EnumValues {
					if strings.EqualFold(part, ev) {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("set element %q not in %v", part, col.EnumValues)
				}
			}
		}
	}
	return nil
}
