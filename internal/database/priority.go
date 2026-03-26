package database

import (
	"fmt"
	"strings"
)

// Priority represents the priority level of a requirement.
type Priority string

const (
	PriorityP0     Priority = "P0"
	PriorityHigh   Priority = "HIGH"
	PriorityMedium Priority = "MEDIUM"
	PriorityLow    Priority = "LOW"
)

// AllPriorities returns all valid priority values in order.
func AllPriorities() []Priority {
	return []Priority{PriorityP0, PriorityHigh, PriorityMedium, PriorityLow}
}

// ParsePriority parses a string into a Priority, case-insensitive.
func ParsePriority(s string) (Priority, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "P0":
		return PriorityP0, nil
	case "HIGH":
		return PriorityHigh, nil
	case "MEDIUM", "":
		return PriorityMedium, nil
	case "LOW":
		return PriorityLow, nil
	default:
		return PriorityMedium, fmt.Errorf("invalid priority: %q", s)
	}
}

// String returns the string representation of the priority.
func (p Priority) String() string {
	return string(p)
}

// Weight returns a numeric weight for sorting (lower = higher priority).
func (p Priority) Weight() int {
	switch p {
	case PriorityP0:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 3
	default:
		return 4
	}
}

// IsHighPriority returns true if priority is P0 or HIGH.
func (p Priority) IsHighPriority() bool {
	return p == PriorityP0 || p == PriorityHigh
}
