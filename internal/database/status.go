// Package database provides the core data models and CSV parsing for RTMX.
package database

import (
	"fmt"
	"strings"
)

// Status represents the completion status of a requirement.
type Status string

const (
	StatusComplete   Status = "COMPLETE"
	StatusPartial    Status = "PARTIAL"
	StatusMissing    Status = "MISSING"
	StatusNotStarted Status = "NOT_STARTED"
)

// AllStatuses returns all valid status values.
func AllStatuses() []Status {
	return []Status{StatusComplete, StatusPartial, StatusMissing, StatusNotStarted}
}

// ParseStatus parses a string into a Status, case-insensitive.
func ParseStatus(s string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "COMPLETE":
		return StatusComplete, nil
	case "PARTIAL":
		return StatusPartial, nil
	case "MISSING", "":
		return StatusMissing, nil
	case "NOT_STARTED":
		return StatusNotStarted, nil
	default:
		return StatusMissing, fmt.Errorf("invalid status: %q", s)
	}
}

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// IsComplete returns true if the status is COMPLETE.
func (s Status) IsComplete() bool {
	return s == StatusComplete
}

// IsIncomplete returns true if the status is not COMPLETE.
func (s Status) IsIncomplete() bool {
	return s != StatusComplete
}

// Weight returns a numeric weight for sorting (0=COMPLETE, 1=PARTIAL, 2=MISSING, 3=NOT_STARTED).
func (s Status) Weight() int {
	switch s {
	case StatusComplete:
		return 0
	case StatusPartial:
		return 1
	case StatusMissing:
		return 2
	case StatusNotStarted:
		return 3
	default:
		return 4
	}
}

// CompletionPercent returns the completion percentage (100, 50, 0, 0).
func (s Status) CompletionPercent() float64 {
	switch s {
	case StatusComplete:
		return 100.0
	case StatusPartial:
		return 50.0
	default:
		return 0.0
	}
}
