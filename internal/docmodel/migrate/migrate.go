// Package migrate prototypes CSV+Markdown → document model → CSV (REQ-DATA-001d).
package migrate

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reNumbered = regexp.MustCompile(`^\s*(\d+)[.)]\s+(.+)$`)
	reCheckbox = regexp.MustCompile(`^\s*-\s*\[([ xX])\]\s+(.+)$`)
	reBullet   = regexp.MustCompile(`^\s*-\s+(.+)$`)
)

// AC is a parsed acceptance criterion with stable id.
type AC struct {
	ID        string `json:"ac_id"`
	Statement string `json:"statement"`
}

// Requirement is the structured spike document (subset of 001a).
type Requirement struct {
	ReqID           string   `json:"req_id"`
	RequirementText string   `json:"requirement_text"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	Dependencies    []string `json:"dependencies"`
	Blocks          []string `json:"blocks"`
	ACs             []AC     `json:"acs"`
}

// ParseAcceptanceCriteria extracts ACs from Markdown body after "## Acceptance Criteria".
// Supports numbered lists (1. / 1)), checkbox lists (- [ ] / - [x]), and plain bullets.
func ParseAcceptanceCriteria(md string) []AC {
	lines := strings.Split(md, "\n")
	inSection := false
	var raw []string
	var cont strings.Builder

	flushCont := func() {
		if cont.Len() == 0 {
			return
		}
		raw[len(raw)-1] = strings.TrimSpace(raw[len(raw)-1] + " " + strings.TrimSpace(cont.String()))
		cont.Reset()
	}

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "## ") {
			if inSection {
				break
			}
			if strings.EqualFold(trim, "## Acceptance Criteria") {
				inSection = true
			}
			continue
		}
		if !inSection {
			continue
		}
		if trim == "" {
			flushCont()
			continue
		}
		if m := reNumbered.FindStringSubmatch(line); m != nil {
			flushCont()
			raw = append(raw, strings.TrimSpace(m[2]))
			continue
		}
		if m := reCheckbox.FindStringSubmatch(line); m != nil {
			flushCont()
			raw = append(raw, strings.TrimSpace(m[2]))
			continue
		}
		if m := reBullet.FindStringSubmatch(line); m != nil && !strings.HasPrefix(strings.TrimSpace(m[1]), "[") {
			flushCont()
			raw = append(raw, strings.TrimSpace(m[1]))
			continue
		}
		// Continuation of previous AC (wrapped lines).
		if len(raw) > 0 {
			cont.WriteString(" ")
			cont.WriteString(trim)
		}
	}
	flushCont()

	out := make([]AC, 0, len(raw))
	for i, stmt := range raw {
		stmt = normalizeStatement(stmt)
		out = append(out, AC{ID: ordinalACID(i + 1), Statement: stmt})
	}
	return out
}

func normalizeStatement(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func ordinalACID(n int) string {
	return fmt.Sprintf("AC-%d", n)
}

// AssignStableACIDs reapplies ordinal ids, reusing prior ids when statement text matches.
func AssignStableACIDs(previous, current []AC) []AC {
	byStmt := map[string]string{}
	for _, ac := range previous {
		byStmt[normalizeStatement(ac.Statement)] = ac.ID
	}
	used := map[string]bool{}
	out := make([]AC, len(current))
	nextOrdinal := 1
	for i, ac := range current {
		stmt := normalizeStatement(ac.Statement)
		if id, ok := byStmt[stmt]; ok && !used[id] {
			out[i] = AC{ID: id, Statement: stmt}
			used[id] = true
			continue
		}
		for used[ordinalACID(nextOrdinal)] {
			nextOrdinal++
		}
		id := ordinalACID(nextOrdinal)
		out[i] = AC{ID: id, Statement: stmt}
		used[id] = true
		nextOrdinal++
	}
	return out
}

// ContentHashACID is documented as deferred; kept for tests of stability alternatives.
func ContentHashACID(statement string) string {
	sum := sha256.Sum256([]byte(normalizeStatement(statement)))
	return "AC-" + hex.EncodeToString(sum[:4])
}

// ProjectCSV writes req-level rows without AC nesting.
func ProjectCSV(path string, reqs []Requirement) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"req_id", "requirement_text", "status", "priority", "dependencies", "blocks"})
	for _, r := range reqs {
		_ = w.Write([]string{
			r.ReqID,
			r.RequirementText,
			r.Status,
			r.Priority,
			strings.Join(r.Dependencies, "|"),
			strings.Join(r.Blocks, "|"),
		})
	}
	w.Flush()
	return w.Error()
}

// LoadCSVRequirements reads the spike CSV projection.
func LoadCSVRequirements(path string) ([]Requirement, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}
	var out []Requirement
	for _, row := range rows[1:] {
		if len(row) < 6 {
			continue
		}
		var deps, blocks []string
		if row[4] != "" {
			deps = strings.Split(row[4], "|")
		}
		if row[5] != "" {
			blocks = strings.Split(row[5], "|")
		}
		out = append(out, Requirement{
			ReqID:           row[0],
			RequirementText: row[1],
			Status:          row[2],
			Priority:        row[3],
			Dependencies:    deps,
			Blocks:          blocks,
		})
	}
	return out, nil
}

// ReadMarkdownFile loads a requirement markdown file.
func ReadMarkdownFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MigrateFixtureDir reads CSV+MD under dir and returns structured requirements with ACs.
func MigrateFixtureDir(dir string) ([]Requirement, error) {
	csvPath := filepath.Join(dir, "database.csv")
	reqs, err := LoadCSVRequirements(csvPath)
	if err != nil {
		return nil, err
	}
	for i := range reqs {
		mdPath := filepath.Join(dir, "requirements", reqs[i].ReqID+".md")
		md, err := ReadMarkdownFile(mdPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		reqs[i].ACs = ParseAcceptanceCriteria(md)
	}
	return reqs, nil
}

// Remigrate applies a second pass, stabilizing ac_ids from previous.
func Remigrate(previous, fresh []Requirement) []Requirement {
	prevByID := map[string]Requirement{}
	for _, r := range previous {
		prevByID[r.ReqID] = r
	}
	out := make([]Requirement, len(fresh))
	for i, r := range fresh {
		if p, ok := prevByID[r.ReqID]; ok {
			r.ACs = AssignStableACIDs(p.ACs, r.ACs)
		}
		out[i] = r
	}
	return out
}

// ScanFile is a small helper for tests reading fixtures line-by-line.
func ScanFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines, s.Err()
}
