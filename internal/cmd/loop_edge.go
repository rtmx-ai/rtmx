package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// PullRequest is a PR snapshot used by the falling-edge watcher (REQ-ORCH-019b).
type PullRequest struct {
	Number int
	Title  string
	Body   string
	Merged bool
	SHA    string
}

// PRSource discovers pull requests. Production may shell out to gh; tests inject fakes.
type PRSource interface {
	ListMerged() ([]PullRequest, error)
}

// LoopCursor persists what the delivery loop has already seen.
type LoopCursor struct {
	MergedPRNumbers []int    `json:"merged_pr_numbers"`
	MergedSHAs      []string `json:"merged_shas"`
	CompleteIDs     []string `json:"complete_ids"`
}

// LoopEdge is one falling edge: a requirement that just merged or became COMPLETE.
type LoopEdge struct {
	ReqID    string `json:"req_id"`
	Source   string `json:"source"` // "pr" | "complete"
	PRNumber int    `json:"pr_number,omitempty"`
}

var reqIDInText = regexp.MustCompile(`REQ-[A-Z0-9]+-\d+[a-z]?`)

// ExtractReqIDs returns unique requirement IDs mentioned in text.
func ExtractReqIDs(text string) []string {
	found := reqIDInText.FindAllString(text, -1)
	seen := map[string]bool{}
	var out []string
	for _, id := range found {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// LoopCursorPath is .rtmx/loop/cursor.json under projectDir.
func LoopCursorPath(projectDir string) string {
	return filepath.Join(projectDir, ".rtmx", "loop", "cursor.json")
}

// LoadLoopCursor reads the cursor or returns an empty one.
func LoadLoopCursor(path string) (LoopCursor, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LoopCursor{}, nil
		}
		return LoopCursor{}, err
	}
	var c LoopCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return LoopCursor{}, fmt.Errorf("loop cursor: %w", err)
	}
	return c, nil
}

// SaveLoopCursor writes the cursor atomically enough for a single agent.
func SaveLoopCursor(path string, c LoopCursor) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

// EnsureLoopGitignore adds loop/ to .rtmx/.gitignore when missing.
func EnsureLoopGitignore(rtmxDir string) error {
	path := filepath.Join(rtmxDir, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	text := string(existing)
	if strings.Contains(text, "loop/") {
		return nil
	}
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	text += "loop/\n"
	return os.WriteFile(path, []byte(text), 0o644)
}

// DetectFallingEdges compares the cursor to newly merged PRs and the current
// COMPLETE set. A timer wake with no new data yields zero edges.
func DetectFallingEdges(cursor LoopCursor, merged []PullRequest, completeNow []string) (edges []LoopEdge, next LoopCursor) {
	seenPR := map[int]bool{}
	for _, n := range cursor.MergedPRNumbers {
		seenPR[n] = true
	}
	seenSHA := map[string]bool{}
	for _, s := range cursor.MergedSHAs {
		if s != "" {
			seenSHA[s] = true
		}
	}
	seenComplete := map[string]bool{}
	for _, id := range cursor.CompleteIDs {
		seenComplete[id] = true
	}

	next = LoopCursor{
		MergedPRNumbers: append([]int{}, cursor.MergedPRNumbers...),
		MergedSHAs:      append([]string{}, cursor.MergedSHAs...),
		CompleteIDs:     append([]string{}, completeNow...),
	}
	edgeSeen := map[string]bool{}

	for _, pr := range merged {
		if !pr.Merged {
			continue // open PRs are never edges
		}
		already := seenPR[pr.Number]
		if pr.SHA != "" && seenSHA[pr.SHA] {
			already = true
		}
		if already {
			continue
		}
		ids := ExtractReqIDs(pr.Title + "\n" + pr.Body)
		for _, id := range ids {
			key := "pr:" + id
			if edgeSeen[key] {
				continue
			}
			edgeSeen[key] = true
			edges = append(edges, LoopEdge{ReqID: id, Source: "pr", PRNumber: pr.Number})
		}
		next.MergedPRNumbers = append(next.MergedPRNumbers, pr.Number)
		if pr.SHA != "" && !seenSHA[pr.SHA] {
			next.MergedSHAs = append(next.MergedSHAs, pr.SHA)
			seenSHA[pr.SHA] = true
		}
		seenPR[pr.Number] = true
	}

	for _, id := range completeNow {
		if seenComplete[id] {
			continue
		}
		key := "complete:" + id
		if edgeSeen[key] {
			continue
		}
		edgeSeen[key] = true
		edges = append(edges, LoopEdge{ReqID: id, Source: "complete"})
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].ReqID != edges[j].ReqID {
			return edges[i].ReqID < edges[j].ReqID
		}
		return edges[i].Source < edges[j].Source
	})
	return edges, next
}

// CompleteIDsFromStatuses builds the COMPLETE set from a map of req_id → status.
func CompleteIDsFromStatuses(statuses map[string]string) []string {
	var out []string
	for id, st := range statuses {
		if strings.EqualFold(st, "COMPLETE") {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}
