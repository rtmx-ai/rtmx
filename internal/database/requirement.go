package database

import (
	"strings"
	"time"
)

// Requirement represents a single requirement in the RTM database.
type Requirement struct {
	// Core identification
	ReqID       string `csv:"req_id" json:"req_id"`
	Category    string `csv:"category" json:"category"`
	Subcategory string `csv:"subcategory" json:"subcategory"`

	// Description
	RequirementText string `csv:"requirement_text" json:"requirement_text"`
	TargetValue     string `csv:"target_value" json:"target_value"`
	Notes           string `csv:"notes" json:"notes"`

	// Testing
	TestModule       string `csv:"test_module" json:"test_module"`
	TestFunction     string `csv:"test_function" json:"test_function"`
	ValidationMethod string `csv:"validation_method" json:"validation_method"`

	// Status and priority
	Status   Status   `csv:"status" json:"status"`
	Priority Priority `csv:"priority" json:"priority"`
	Phase    int      `csv:"phase" json:"phase"`

	// Planning
	EffortWeeks float64 `csv:"effort_weeks" json:"effort_weeks"`
	Assignee    string  `csv:"assignee" json:"assignee"`
	Sprint      string  `csv:"sprint" json:"sprint"`

	// Dependencies (stored as pipe-separated strings in CSV)
	Dependencies StringSet `csv:"dependencies" json:"dependencies"`
	Blocks       StringSet `csv:"blocks" json:"blocks"`

	// Dates
	StartedDate   string `csv:"started_date" json:"started_date"`
	CompletedDate string `csv:"completed_date" json:"completed_date"`

	// External references
	RequirementFile string `csv:"requirement_file" json:"requirement_file"`
	ExternalID      string `csv:"external_id" json:"external_id"`

	// Extensible fields
	Extra map[string]string `csv:"-" json:"extra,omitempty"`
}

// StringSet is a set of strings, stored as pipe-separated in CSV.
type StringSet map[string]struct{}

// NewStringSet creates a new StringSet from strings.
func NewStringSet(items ...string) StringSet {
	s := make(StringSet)
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			s[item] = struct{}{}
		}
	}
	return s
}

// ParseStringSet parses a pipe-separated string into a StringSet.
func ParseStringSet(s string) StringSet {
	set := make(StringSet)
	for _, item := range strings.Split(s, "|") {
		if item = strings.TrimSpace(item); item != "" {
			set[item] = struct{}{}
		}
	}
	return set
}

// Add adds an item to the set.
func (s StringSet) Add(item string) {
	if item = strings.TrimSpace(item); item != "" {
		s[item] = struct{}{}
	}
}

// Remove removes an item from the set.
func (s StringSet) Remove(item string) {
	delete(s, item)
}

// Contains checks if an item is in the set.
func (s StringSet) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

// Slice returns the items as a sorted slice.
func (s StringSet) Slice() []string {
	items := make([]string, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	// Sort for deterministic output
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i] > items[j] {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

// String returns a pipe-separated string.
func (s StringSet) String() string {
	return strings.Join(s.Slice(), "|")
}

// Len returns the number of items.
func (s StringSet) Len() int {
	return len(s)
}

// NewRequirement creates a new Requirement with defaults.
func NewRequirement(reqID string) *Requirement {
	return &Requirement{
		ReqID:        reqID,
		Status:       StatusMissing,
		Priority:     PriorityMedium,
		Dependencies: make(StringSet),
		Blocks:       make(StringSet),
		Extra:        make(map[string]string),
	}
}

// HasTest returns true if the requirement has a test assigned.
func (r *Requirement) HasTest() bool {
	return r.TestModule != "" && r.TestFunction != ""
}

// IsComplete returns true if the requirement is complete.
func (r *Requirement) IsComplete() bool {
	return r.Status.IsComplete()
}

// IsIncomplete returns true if the requirement is not complete.
func (r *Requirement) IsIncomplete() bool {
	return r.Status.IsIncomplete()
}

// IsHighPriority returns true if the requirement is P0 or HIGH priority.
func (r *Requirement) IsHighPriority() bool {
	return r.Priority.IsHighPriority()
}

// IsBlocked returns true if any dependency is incomplete.
func (r *Requirement) IsBlocked(db *Database) bool {
	for dep := range r.Dependencies {
		// Skip cross-repo dependencies for now
		if strings.Contains(dep, ":") {
			continue
		}
		if depReq := db.Get(dep); depReq != nil && depReq.IsIncomplete() {
			return true
		}
	}
	return false
}

// BlockingDeps returns the list of incomplete dependencies.
func (r *Requirement) BlockingDeps(db *Database) []string {
	var blocking []string
	for dep := range r.Dependencies {
		if strings.Contains(dep, ":") {
			continue
		}
		if depReq := db.Get(dep); depReq != nil && depReq.IsIncomplete() {
			blocking = append(blocking, dep)
		}
	}
	return blocking
}

// SetStartedDate sets the started date to today if not already set.
func (r *Requirement) SetStartedDate() {
	if r.StartedDate == "" {
		r.StartedDate = time.Now().Format("2006-01-02")
	}
}

// SetCompletedDate sets the completed date to today.
func (r *Requirement) SetCompletedDate() {
	r.CompletedDate = time.Now().Format("2006-01-02")
}

// Clone creates a deep copy of the requirement.
func (r *Requirement) Clone() *Requirement {
	clone := *r
	clone.Dependencies = make(StringSet)
	for k := range r.Dependencies {
		clone.Dependencies[k] = struct{}{}
	}
	clone.Blocks = make(StringSet)
	for k := range r.Blocks {
		clone.Blocks[k] = struct{}{}
	}
	clone.Extra = make(map[string]string)
	for k, v := range r.Extra {
		clone.Extra[k] = v
	}
	return &clone
}
