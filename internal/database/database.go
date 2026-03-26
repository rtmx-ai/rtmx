package database

import (
	"fmt"
	"sort"
)

// Database is the in-memory RTM database.
type Database struct {
	// requirements stores all requirements by ID.
	requirements map[string]*Requirement

	// order preserves insertion order for consistent output.
	order []string

	// path is the file path this database was loaded from.
	path string

	// dirty tracks if the database has been modified.
	dirty bool
}

// NewDatabase creates a new empty database.
func NewDatabase() *Database {
	return &Database{
		requirements: make(map[string]*Requirement),
		order:        make([]string, 0),
	}
}

// Path returns the file path this database was loaded from.
func (db *Database) Path() string {
	return db.path
}

// SetPath sets the file path for saving.
func (db *Database) SetPath(path string) {
	db.path = path
}

// IsDirty returns true if the database has unsaved changes.
func (db *Database) IsDirty() bool {
	return db.dirty
}

// MarkClean marks the database as saved.
func (db *Database) MarkClean() {
	db.dirty = false
}

// Len returns the number of requirements.
func (db *Database) Len() int {
	return len(db.requirements)
}

// Get retrieves a requirement by ID.
func (db *Database) Get(reqID string) *Requirement {
	return db.requirements[reqID]
}

// Exists checks if a requirement exists.
func (db *Database) Exists(reqID string) bool {
	_, ok := db.requirements[reqID]
	return ok
}

// Add adds a new requirement to the database.
func (db *Database) Add(req *Requirement) error {
	if req.ReqID == "" {
		return fmt.Errorf("requirement ID cannot be empty")
	}
	if db.Exists(req.ReqID) {
		return fmt.Errorf("requirement %q already exists", req.ReqID)
	}
	db.requirements[req.ReqID] = req
	db.order = append(db.order, req.ReqID)
	db.dirty = true
	return nil
}

// Update updates fields on an existing requirement.
func (db *Database) Update(reqID string, updates map[string]interface{}) error {
	req := db.Get(reqID)
	if req == nil {
		return fmt.Errorf("requirement %q not found", reqID)
	}

	for key, value := range updates {
		switch key {
		case "status":
			if s, ok := value.(Status); ok {
				req.Status = s
			} else if s, ok := value.(string); ok {
				status, err := ParseStatus(s)
				if err != nil {
					return err
				}
				req.Status = status
			}
		case "priority":
			if p, ok := value.(Priority); ok {
				req.Priority = p
			} else if s, ok := value.(string); ok {
				priority, err := ParsePriority(s)
				if err != nil {
					return err
				}
				req.Priority = priority
			}
		case "phase":
			if p, ok := value.(int); ok {
				req.Phase = p
			}
		case "assignee":
			if s, ok := value.(string); ok {
				req.Assignee = s
			}
		case "sprint":
			if s, ok := value.(string); ok {
				req.Sprint = s
			}
		case "test_module":
			if s, ok := value.(string); ok {
				req.TestModule = s
			}
		case "test_function":
			if s, ok := value.(string); ok {
				req.TestFunction = s
			}
		case "started_date":
			if s, ok := value.(string); ok {
				req.StartedDate = s
			}
		case "completed_date":
			if s, ok := value.(string); ok {
				req.CompletedDate = s
			}
		// Add more fields as needed
		default:
			// Store unknown fields in Extra
			if s, ok := value.(string); ok {
				if req.Extra == nil {
					req.Extra = make(map[string]string)
				}
				req.Extra[key] = s
			}
		}
	}

	db.dirty = true
	return nil
}

// Remove removes a requirement from the database.
func (db *Database) Remove(reqID string) error {
	if !db.Exists(reqID) {
		return fmt.Errorf("requirement %q not found", reqID)
	}

	delete(db.requirements, reqID)

	// Remove from order
	newOrder := make([]string, 0, len(db.order)-1)
	for _, id := range db.order {
		if id != reqID {
			newOrder = append(newOrder, id)
		}
	}
	db.order = newOrder
	db.dirty = true
	return nil
}

// All returns all requirements in insertion order.
func (db *Database) All() []*Requirement {
	reqs := make([]*Requirement, 0, len(db.order))
	for _, id := range db.order {
		if req := db.requirements[id]; req != nil {
			reqs = append(reqs, req)
		}
	}
	return reqs
}

// IDs returns all requirement IDs in insertion order.
func (db *Database) IDs() []string {
	return append([]string{}, db.order...)
}

// Filter returns requirements matching the given criteria.
func (db *Database) Filter(opts FilterOptions) []*Requirement {
	var results []*Requirement

	for _, req := range db.All() {
		if opts.Status != nil && req.Status != *opts.Status {
			continue
		}
		if opts.Priority != nil && req.Priority != *opts.Priority {
			continue
		}
		if opts.Category != "" && req.Category != opts.Category {
			continue
		}
		if opts.Phase != nil && req.Phase != *opts.Phase {
			continue
		}
		if opts.HasTest != nil {
			if *opts.HasTest && !req.HasTest() {
				continue
			}
			if !*opts.HasTest && req.HasTest() {
				continue
			}
		}
		if opts.IsComplete != nil {
			if *opts.IsComplete && !req.IsComplete() {
				continue
			}
			if !*opts.IsComplete && req.IsComplete() {
				continue
			}
		}
		if opts.IsBlocked != nil {
			blocked := req.IsBlocked(db)
			if *opts.IsBlocked && !blocked {
				continue
			}
			if !*opts.IsBlocked && blocked {
				continue
			}
		}
		if opts.Assignee != "" && req.Assignee != opts.Assignee {
			continue
		}

		results = append(results, req)
	}

	return results
}

// FilterOptions specifies criteria for filtering requirements.
type FilterOptions struct {
	Status     *Status
	Priority   *Priority
	Category   string
	Phase      *int
	HasTest    *bool
	IsComplete *bool
	IsBlocked  *bool
	Assignee   string
}

// StatusCounts returns a map of status to count.
func (db *Database) StatusCounts() map[Status]int {
	counts := make(map[Status]int)
	for _, req := range db.All() {
		counts[req.Status]++
	}
	return counts
}

// PriorityCounts returns a map of priority to count.
func (db *Database) PriorityCounts() map[Priority]int {
	counts := make(map[Priority]int)
	for _, req := range db.All() {
		counts[req.Priority]++
	}
	return counts
}

// Categories returns all unique categories.
func (db *Database) Categories() []string {
	seen := make(map[string]bool)
	var cats []string
	for _, req := range db.All() {
		if !seen[req.Category] {
			seen[req.Category] = true
			cats = append(cats, req.Category)
		}
	}
	sort.Strings(cats)
	return cats
}

// Phases returns all unique phases.
func (db *Database) Phases() []int {
	seen := make(map[int]bool)
	var phases []int
	for _, req := range db.All() {
		if req.Phase > 0 && !seen[req.Phase] {
			seen[req.Phase] = true
			phases = append(phases, req.Phase)
		}
	}
	sort.Ints(phases)
	return phases
}

// CompletionPercentage returns the overall completion percentage.
// COMPLETE = 100%, PARTIAL = 50%, others = 0%
func (db *Database) CompletionPercentage() float64 {
	if db.Len() == 0 {
		return 0
	}

	var total float64
	for _, req := range db.All() {
		total += req.Status.CompletionPercent()
	}

	return total / float64(db.Len())
}

// ByCategory returns requirements grouped by category.
func (db *Database) ByCategory() map[string][]*Requirement {
	result := make(map[string][]*Requirement)
	for _, req := range db.All() {
		result[req.Category] = append(result[req.Category], req)
	}
	return result
}

// ByPhase returns requirements grouped by phase.
func (db *Database) ByPhase() map[int][]*Requirement {
	result := make(map[int][]*Requirement)
	for _, req := range db.All() {
		result[req.Phase] = append(result[req.Phase], req)
	}
	return result
}

// Incomplete returns all incomplete requirements.
func (db *Database) Incomplete() []*Requirement {
	complete := false
	return db.Filter(FilterOptions{IsComplete: &complete})
}

// Complete returns all complete requirements.
func (db *Database) Complete() []*Requirement {
	complete := true
	return db.Filter(FilterOptions{IsComplete: &complete})
}

// Backlog returns incomplete requirements sorted by priority.
func (db *Database) Backlog() []*Requirement {
	incomplete := db.Incomplete()
	sort.Slice(incomplete, func(i, j int) bool {
		// Sort by priority weight (lower = higher priority)
		pi := incomplete[i].Priority.Weight()
		pj := incomplete[j].Priority.Weight()
		if pi != pj {
			return pi < pj
		}
		// Then by phase
		if incomplete[i].Phase != incomplete[j].Phase {
			return incomplete[i].Phase < incomplete[j].Phase
		}
		// Then by ID
		return incomplete[i].ReqID < incomplete[j].ReqID
	})
	return incomplete
}
