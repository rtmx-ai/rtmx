package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExportJSONLDocumentAndRemigrate(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-003")
	reqs, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err := ExportDir(out, reqs); err != nil {
		t.Fatal(err)
	}
	jsonl := filepath.Join(out, "requirements.jsonl")
	doc := filepath.Join(out, "requirements.document.json")
	for _, p := range []string{jsonl, doc, filepath.Join(out, "companions", "REQ-VERIFY-013.md")} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing %s: %v", p, err)
		}
	}
	loaded, err := LoadJSONL(jsonl)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != len(reqs) {
		t.Fatalf("loaded %d want %d", len(loaded), len(reqs))
	}
	fresh, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	stable := Remigrate(loaded, fresh)
	byID := map[string][]AC{}
	for _, r := range loaded {
		byID[r.ReqID] = r.ACs
	}
	for _, r := range stable {
		prev := byID[r.ReqID]
		if len(prev) == 0 {
			continue
		}
		for i := range r.ACs {
			if i < len(prev) && normalizeStatement(r.ACs[i].Statement) == normalizeStatement(prev[i].Statement) && r.ACs[i].ID != prev[i].ID {
				t.Fatalf("%s AC id churn: %s -> %s", r.ReqID, prev[i].ID, r.ACs[i].ID)
			}
		}
	}
}
