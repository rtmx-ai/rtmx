package docmodel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func authoringUXDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, ".rtmx", "requirements", "DATA", "spikes", "authoring-ux")
}

func TestAuthoringUXSpikeThreeEncodings(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001c")
	dir := authoringUXDir(t)
	paths := []string{
		filepath.Join(dir, "jsonl-pure", "REQ-AUTH-001.jsonl"),
		filepath.Join(dir, "jsonl-plus-md", "REQ-AUTH-001.jsonl"),
		filepath.Join(dir, "jsonl-plus-md", "REQ-AUTH-001.md"),
		filepath.Join(dir, "md-frontmatter", "REQ-AUTH-001.md"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing encoding artifact %s: %v", p, err)
		}
	}
	scorecard, err := os.ReadFile(filepath.Join(dir, "SCORECARD.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(scorecard)
	for _, need := range []string{
		"Pure JSONL",
		"JSONL + MD",
		"frontmatter",
		"Winner",
		"Recommendation",
		"losers fail",
		"ATDD",
	} {
		if !strings.Contains(text, need) {
			t.Errorf("SCORECARD.md missing %q", need)
		}
	}
}

func TestAuthoringUXSpikeAgentProjection(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001c")
	raw, err := os.ReadFile(filepath.Join(authoringUXDir(t), "projection-agent-atdd-v0.json"))
	if err != nil {
		t.Fatal(err)
	}
	var proj struct {
		Requirements []struct {
			ReqID         string `json:"req_id"`
			Rationale     string `json:"rationale"`
			ACs           []any  `json:"acs"`
			TestBindings  []any  `json:"test_bindings"`
			RequirementTx string `json:"requirement_text"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal(raw, &proj); err != nil {
		t.Fatal(err)
	}
	if len(proj.Requirements) != 1 || proj.Requirements[0].ReqID != "REQ-AUTH-001" {
		t.Fatalf("expected one REQ-AUTH-001 projection, got %+v", proj.Requirements)
	}
	r := proj.Requirements[0]
	if r.Rationale != "" {
		t.Fatal("agent projection must omit bulk rationale")
	}
	if len(r.ACs) < 2 || len(r.TestBindings) < 2 {
		t.Fatalf("projection must include ACs and bindings; acs=%d bindings=%d", len(r.ACs), len(r.TestBindings))
	}
	if r.RequirementTx == "" {
		t.Fatal("projection must keep requirement_text metadata")
	}
	if strings.Contains(string(raw), "must not observe room CRDT") {
		t.Fatal("projection leaked companion rationale prose")
	}
}
