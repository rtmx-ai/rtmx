package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestFallingEdgeCursor(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019b")

	cursor := LoopCursor{
		MergedPRNumbers: []int{10},
		CompleteIDs:     []string{"REQ-EX-001"},
	}
	merged := []PullRequest{
		{Number: 10, Title: "feat(REQ-EX-001): old", Merged: true, SHA: "aaa"},
		{Number: 11, Title: "feat(REQ-EX-002): new", Body: "Closes REQ-EX-002", Merged: true, SHA: "bbb"},
		{Number: 12, Title: "feat(REQ-EX-003): still open", Merged: false},
	}
	complete := []string{"REQ-EX-001", "REQ-EX-004"}

	edges, next := DetectFallingEdges(cursor, merged, complete)

	var byID = map[string]LoopEdge{}
	for _, e := range edges {
		byID[e.ReqID+"/"+e.Source] = e
	}
	if _, ok := byID["REQ-EX-001/pr"]; ok {
		t.Fatal("old merged PR must not replay")
	}
	if e, ok := byID["REQ-EX-002/pr"]; !ok || e.PRNumber != 11 {
		t.Fatalf("expected new PR edge for REQ-EX-002, got %+v", edges)
	}
	if _, ok := byID["REQ-EX-003/pr"]; ok {
		t.Fatal("open PR must not be an edge")
	}
	if _, ok := byID["REQ-EX-004/complete"]; !ok {
		t.Fatalf("COMPLETE transition must be an edge: %+v", edges)
	}

	// Timer-only wake: same merged list and same COMPLETE set → zero edges.
	edges2, _ := DetectFallingEdges(next, merged, complete)
	if len(edges2) != 0 {
		t.Fatalf("timer wake with no change must yield 0 edges, got %+v", edges2)
	}
}

func TestLoopCursorRoundTripAndGitignore(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019b")
	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureLoopGitignore(rtmxDir); err != nil {
		t.Fatal(err)
	}
	gi, err := os.ReadFile(filepath.Join(rtmxDir, ".gitignore"))
	if err != nil || !containsLoopIgnore(string(gi)) {
		t.Fatalf("gitignore missing loop/: %q err=%v", gi, err)
	}
	path := LoopCursorPath(dir)
	c := LoopCursor{MergedPRNumbers: []int{1}, CompleteIDs: []string{"REQ-A-1"}}
	if err := SaveLoopCursor(path, c); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadLoopCursor(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.MergedPRNumbers) != 1 || loaded.MergedPRNumbers[0] != 1 {
		t.Fatalf("cursor = %+v", loaded)
	}
}

func containsLoopIgnore(s string) bool {
	return len(s) > 0 && (filepath.Base(s) == "loop/" || // never
		true) && (stringContains(s, "loop/"))
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func TestExtractReqIDs(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019b")
	ids := ExtractReqIDs("feat(REQ-EX-002): also REQ-EX-002 and REQ-DATA-001a")
	if len(ids) != 2 || ids[0] != "REQ-DATA-001a" || ids[1] != "REQ-EX-002" {
		t.Fatalf("ids = %v", ids)
	}
}
