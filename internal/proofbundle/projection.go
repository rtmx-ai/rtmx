package proofbundle

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Counts is the non-technical verification projection (REQ-MONO-065c).
type Counts struct {
	Verified   int `json:"verified"`
	Unverified int `json:"unverified"`
	Stale      int `json:"stale"`
}

// ProjectCounts reads a requirement matrix and evidence log and returns
// verified / unverified / stale counts relative to revisionSHA.
//
// Verified: latest evidence commit_sha equals revisionSHA.
// Stale: latest evidence exists but its commit_sha differs from revisionSHA.
// Unverified: requirement has no evidence record.
func ProjectCounts(matrixPath, evidencePath, revisionSHA string) (Counts, error) {
	reqs, err := readRequirementIDs(matrixPath)
	if err != nil {
		return Counts{}, err
	}
	latest, err := latestEvidenceSHA(evidencePath)
	if err != nil {
		return Counts{}, err
	}
	rev := strings.ToLower(strings.TrimSpace(revisionSHA))
	var c Counts
	for _, id := range reqs {
		sha, ok := latest[id]
		if !ok {
			c.Unverified++
			continue
		}
		if sha == rev {
			c.Verified++
			continue
		}
		c.Stale++
	}
	return c, nil
}

func readRequirementIDs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open matrix: %w", err)
	}
	defer func() { _ = f.Close() }()
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read matrix: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	idx := -1
	for i, h := range rows[0] {
		if strings.EqualFold(strings.TrimSpace(h), "req_id") {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("matrix missing req_id column")
	}
	var ids []string
	for _, row := range rows[1:] {
		if idx >= len(row) {
			continue
		}
		id := strings.TrimSpace(row[idx])
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func latestEvidenceSHA(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("open evidence: %w", err)
	}
	defer func() { _ = f.Close() }()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec evidenceRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("parse evidence: %w", err)
		}
		if rec.RequirementID == "" || rec.CommitSHA == "" {
			continue
		}
		out[rec.RequirementID] = strings.ToLower(rec.CommitSHA)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan evidence: %w", err)
	}
	return out, nil
}
