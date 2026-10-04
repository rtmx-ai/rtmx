// Package store reads and writes the JSONL + Markdown companion document
// store (REQ-DATA-004). It does not change the default CSV database path.
package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AC is one acceptance criterion in the structured store.
type AC struct {
	ID        string `json:"ac_id"`
	Statement string `json:"statement"`
	Required  *bool  `json:"required,omitempty"`
}

// Binding links a test to an AC.
type Binding struct {
	ID           string `json:"binding_id"`
	ACID         string `json:"ac_id,omitempty"`
	TestFunction string `json:"test_function,omitempty"`
}

// Record is one JSONL line (structured fields only).
type Record struct {
	ReqID           string    `json:"req_id"`
	RequirementText string    `json:"requirement_text"`
	Status          string    `json:"status"`
	Priority        string    `json:"priority"`
	RequirementFile string    `json:"requirement_file,omitempty"`
	Narrative       string    `json:"narrative,omitempty"`
	ACs             []AC      `json:"acs"`
	TestBindings    []Binding `json:"test_bindings,omitempty"`
}

// Projection is the agent/verify surface: metadata + ACs + bindings, no rationale.
type Projection struct {
	ReqID           string    `json:"req_id"`
	RequirementText string    `json:"requirement_text"`
	Status          string    `json:"status"`
	Priority        string    `json:"priority"`
	ACs             []AC      `json:"acs"`
	TestBindings    []Binding `json:"test_bindings,omitempty"`
}

// Store is a directory containing requirements.jsonl and optional companions/.
type Store struct {
	Dir string
}

func (s Store) jsonlPath() string {
	return filepath.Join(s.Dir, "requirements.jsonl")
}

// Load reads all records from requirements.jsonl.
func (s Store) Load() ([]Record, error) {
	f, err := os.Open(s.jsonlPath())
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var out []Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("jsonl: %w", err)
		}
		out = append(out, rec)
	}
	return out, sc.Err()
}

// Save writes records as JSONL (replaces the file). Companion Markdown is not rewritten.
func (s Store) Save(recs []Record) error {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(s.jsonlPath())
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, rec := range recs {
		if rec.ACs == nil {
			rec.ACs = []AC{}
		}
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}

// Put upserts one record by req_id.
func (s Store) Put(rec Record) error {
	recs, err := s.Load()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	replaced := false
	for i := range recs {
		if recs[i].ReqID == rec.ReqID {
			recs[i] = rec
			replaced = true
			break
		}
	}
	if !replaced {
		recs = append(recs, rec)
	}
	return s.Save(recs)
}

// ReadCompanion loads the Markdown body for a record's requirement_file, if set.
func (s Store) ReadCompanion(rec Record) (string, error) {
	if rec.RequirementFile == "" {
		return "", nil
	}
	path := rec.RequirementFile
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.Dir, path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

// Project drops narrative and companion body for agent/verify payloads.
func Project(rec Record) Projection {
	acs := rec.ACs
	if acs == nil {
		acs = []AC{}
	}
	return Projection{
		ReqID:           rec.ReqID,
		RequirementText: rec.RequirementText,
		Status:          rec.Status,
		Priority:        rec.Priority,
		ACs:             acs,
		TestBindings:    rec.TestBindings,
	}
}
