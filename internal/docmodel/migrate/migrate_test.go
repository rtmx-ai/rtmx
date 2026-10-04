package migrate

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func fixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "fixture")
}

func TestAC1ParseCheckboxAndNumberedLists(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001d")
	reqs, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]Requirement{}
	for _, r := range reqs {
		byID[r.ReqID] = r
	}
	v := byID["REQ-VERIFY-013"]
	if len(v.ACs) != 5 {
		t.Fatalf("VERIFY-013 acs = %d %#v", len(v.ACs), v.ACs)
	}
	joined := ""
	for _, ac := range v.ACs {
		joined += ac.Statement + "\n"
	}
	if !strings.Contains(joined, "stays COMPLETE") {
		t.Fatalf("missing wrapped checkbox AC: %#v", v.ACs)
	}
	s := byID["REQ-SYNC-002"]
	if len(s.ACs) != 6 {
		t.Fatalf("SYNC-002 acs = %d %#v", len(s.ACs), s.ACs)
	}
	if s.ACs[0].ID != "AC-1" || s.ACs[5].ID != "AC-6" {
		t.Fatalf("ordinal ids: %#v", s.ACs)
	}
}


func TestAC2StableACIDsWhenTextUnchanged(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001d")
	first, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	secondFresh, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	second := Remigrate(first, secondFresh)
	for i := range first {
		if first[i].ReqID != second[i].ReqID {
			t.Fatalf("order drift")
		}
		if len(first[i].ACs) != len(second[i].ACs) {
			t.Fatalf("ac count drift for %s", first[i].ReqID)
		}
		for j := range first[i].ACs {
			if first[i].ACs[j].ID != second[i].ACs[j].ID {
				t.Fatalf("%s AC %d id %s -> %s", first[i].ReqID, j, first[i].ACs[j].ID, second[i].ACs[j].ID)
			}
		}
	}
}

func TestAC3CSVExportDropsNestingKeepsScalars(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001d")
	reqs, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out.csv")
	if err := ProjectCSV(out, reqs); err != nil {
		t.Fatal(err)
	}
	got, err := LoadCSVRequirements(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("rows = %d", len(got))
	}
	byID := map[string]Requirement{}
	for _, r := range got {
		byID[r.ReqID] = r
		if len(r.ACs) != 0 {
			t.Fatalf("CSV projection must drop AC nesting")
		}
	}
	v := byID["REQ-VERIFY-013"]
	if v.Status != "MISSING" || v.Priority != "HIGH" {
		t.Fatalf("%+v", v)
	}
	if len(v.Dependencies) != 1 || v.Dependencies[0] != "REQ-VERIFY-012" {
		t.Fatalf("deps = %#v", v.Dependencies)
	}
	s := byID["REQ-SYNC-002"]
	if s.Status != "COMPLETE" || s.RequirementText == "" {
		t.Fatalf("%+v", s)
	}
}

func TestAC4SecondMigrateIdempotentACIDs(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001d")
	a, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	b := Remigrate(a, mustMigrate(t))
	c := Remigrate(b, mustMigrate(t))
	for i := range a {
		for j := range a[i].ACs {
			if a[i].ACs[j].ID != c[i].ACs[j].ID {
				t.Fatalf("id churn %s: %s -> %s", a[i].ReqID, a[i].ACs[j].ID, c[i].ACs[j].ID)
			}
		}
	}
}

func mustMigrate(t *testing.T) []Requirement {
	t.Helper()
	reqs, err := MigrateFixtureDir(fixtureDir(t))
	if err != nil {
		t.Fatal(err)
	}
	return reqs
}
