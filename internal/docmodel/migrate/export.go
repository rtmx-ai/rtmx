package migrate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DocumentEnvelope is requirement-document/v0 for --ac-document consumers.
type DocumentEnvelope struct {
	SchemaVersion string              `json:"schema_version"`
	Requirements  []DocumentRequirement `json:"requirements"`
}

// DocumentRequirement is the JSON shape written for 2.0 structured store export.
type DocumentRequirement struct {
	ReqID           string   `json:"req_id"`
	RequirementText string   `json:"requirement_text"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	Dependencies    []string `json:"dependencies,omitempty"`
	Blocks          []string `json:"blocks,omitempty"`
	RequirementFile string   `json:"requirement_file,omitempty"`
	ACs             []AC     `json:"acs"`
	TestBindings    []any    `json:"test_bindings,omitempty"`
}

// ToDocument converts spike requirements into a v0 envelope.
func ToDocument(reqs []Requirement) DocumentEnvelope {
	out := DocumentEnvelope{
		SchemaVersion: "requirement-document/v0",
		Requirements:  make([]DocumentRequirement, 0, len(reqs)),
	}
	for _, r := range reqs {
		acs := r.ACs
		if acs == nil {
			acs = []AC{}
		}
		out.Requirements = append(out.Requirements, DocumentRequirement{
			ReqID:           r.ReqID,
			RequirementText: r.RequirementText,
			Status:          r.Status,
			Priority:        r.Priority,
			Dependencies:    r.Dependencies,
			Blocks:          r.Blocks,
			RequirementFile: fmt.Sprintf("companions/%s.md", r.ReqID),
			ACs:             acs,
		})
	}
	return out
}

// WriteJSONL writes one DocumentRequirement JSON object per line.
func WriteJSONL(path string, reqs []Requirement) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	doc := ToDocument(reqs)
	for _, r := range doc.Requirements {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

// WriteDocumentJSON writes a single requirement-document/v0 file.
func WriteDocumentJSON(path string, reqs []Requirement) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(ToDocument(reqs), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

// LoadJSONL reads DocumentRequirement lines back into spike Requirements (for Remigrate).
func LoadJSONL(path string) ([]Requirement, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var out []Requirement
	s := bufio.NewScanner(f)
	// Allow long lines
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		var dr DocumentRequirement
		if err := json.Unmarshal([]byte(line), &dr); err != nil {
			return nil, fmt.Errorf("jsonl: %w", err)
		}
		out = append(out, Requirement{
			ReqID:           dr.ReqID,
			RequirementText: dr.RequirementText,
			Status:          dr.Status,
			Priority:        dr.Priority,
			Dependencies:    dr.Dependencies,
			Blocks:          dr.Blocks,
			ACs:             dr.ACs,
		})
	}
	return out, s.Err()
}

// WriteCompanionStubs writes short Markdown companions pointing at structured ACs.
func WriteCompanionStubs(dir string, reqs []Requirement) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, r := range reqs {
		path := filepath.Join(dir, r.ReqID+".md")
		body := fmt.Sprintf("# %s\n\n## Rationale\n\nMigrated companion stub. Acceptance criteria and bindings live in the structured JSONL/document store; edit `ac_id` values there.\n", r.ReqID)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ExportDir writes JSONL, document JSON, optional companions, and CSV projection.
func ExportDir(outDir string, reqs []Requirement) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := WriteJSONL(filepath.Join(outDir, "requirements.jsonl"), reqs); err != nil {
		return err
	}
	if err := WriteDocumentJSON(filepath.Join(outDir, "requirements.document.json"), reqs); err != nil {
		return err
	}
	if err := WriteCompanionStubs(filepath.Join(outDir, "companions"), reqs); err != nil {
		return err
	}
	return ProjectCSV(filepath.Join(outDir, "database.projection.csv"), reqs)
}
