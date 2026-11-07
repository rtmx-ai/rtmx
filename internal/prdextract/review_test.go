package prdextract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/prdextract"
)

func TestNoCandidateWrittenUntilHumanAccepts(t *testing.T) {
	dir := t.TempDir()
	rtm := filepath.Join(dir, "rtm.csv")
	before := "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"
	if err := os.WriteFile(rtm, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	c := prdextract.Candidate{Statement: "The system shall gate writes on accept.", Start: 0, End: 38}
	// No decision yet: RTM unchanged.
	raw, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != before {
		t.Fatal("RTM changed without a decision")
	}
	id, err := prdextract.ApplyDecision(rtm, c, prdextract.DecisionAccept)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected requirement id")
	}
	after, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), id) {
		t.Fatalf("accepted candidate not written: %s", after)
	}
}

func TestRejectingCandidateWritesNothing(t *testing.T) {
	dir := t.TempDir()
	rtm := filepath.Join(dir, "rtm.csv")
	before := "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"
	if err := os.WriteFile(rtm, []byte(before), 0o644); err != nil {
		t.Fatal(err)
	}
	c := prdextract.Candidate{Statement: "The system shall ignore rejected candidates.", Start: 0, End: 44}
	id, err := prdextract.ApplyDecision(rtm, c, prdextract.DecisionReject)
	if err != nil {
		t.Fatal(err)
	}
	if id != "" {
		t.Fatalf("reject returned id %q", id)
	}
	after, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Fatalf("reject mutated RTM:\nbefore=%q\nafter=%q", before, after)
	}
}


func TestAcceptingSameCandidateYieldsSameRequirementID(t *testing.T) {
	dir := t.TempDir()
	rtm := filepath.Join(dir, "rtm.csv")
	c := prdextract.Candidate{Statement: "The system shall keep stable IDs.", Start: 0, End: 33}
	id1, err := prdextract.ApplyDecision(rtm, c, prdextract.DecisionAccept)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := prdextract.ApplyDecision(rtm, c, prdextract.DecisionAccept)
	if err != nil {
		t.Fatal(err)
	}
	if id1 == "" || id1 != id2 {
		t.Fatalf("ids %q vs %q", id1, id2)
	}
	raw, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(raw), id1) != 1 {
		t.Fatalf("expected one row for %s: %s", id1, raw)
	}
}

func TestAcceptedRowHasAtLeastOneAcceptanceCriterion(t *testing.T) {
	dir := t.TempDir()
	rtm := filepath.Join(dir, "rtm.csv")
	c := prdextract.Candidate{Statement: "The system shall include acceptance criteria.", Start: 0, End: 46}
	id, err := prdextract.ApplyDecision(rtm, c, prdextract.DecisionAccept)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, id) {
		t.Fatalf("missing id: %s", body)
	}
	if !strings.Contains(body, "Acceptance Criteria:") || !strings.Contains(body, "1. [ ]") {
		t.Fatalf("missing acceptance criterion: %s", body)
	}
}

func TestExtractionNeverSetsStatusComplete(t *testing.T) {
	dir := t.TempDir()
	rtm := filepath.Join(dir, "rtm.csv")
	c := prdextract.Candidate{Statement: "The system shall stay incomplete until verified.", Start: 0, End: 48}
	id, err := prdextract.ApplyDecision(rtm, c, prdextract.DecisionAccept)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(rtm)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, id) {
		t.Fatalf("missing row: %s", body)
	}
	if strings.Contains(body, "COMPLETE") {
		t.Fatalf("extraction set COMPLETE: %s", body)
	}
	if !strings.Contains(body, "NOT_STARTED") {
		t.Fatalf("expected NOT_STARTED: %s", body)
	}
}
