package acverify

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func boolPtr(v bool) *bool { return &v }

// Fixture: one requirement, three ACs, one binding per AC (REQ-DATA-001b AC1).
func threeACFixture(ac1, ac2, ac3 *bool) Requirement {
	return Requirement{
		ID: "REQ-DEMO-001",
		ACs: []AC{
			{ID: "AC-1", Required: true},
			{ID: "AC-2", Required: true},
			{ID: "AC-3", Required: true},
		},
		Bindings: []Binding{
			{ID: "TB-1", ACID: "AC-1", Passed: ac1},
			{ID: "TB-2", ACID: "AC-2", Passed: ac2},
			{ID: "TB-3", ACID: "AC-3", Passed: ac3},
		},
	}
}

func TestAC1MatrixAllPass(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001b")
	r := Evaluate(threeACFixture(boolPtr(true), boolPtr(true), boolPtr(true)))
	out := FormatMatrix(r)
	if r.Status != "COMPLETE" {
		t.Fatalf("status = %s", r.Status)
	}
	for _, id := range []string{"AC-1", "AC-2", "AC-3"} {
		if !strings.Contains(out, id) || !strings.Contains(out, "pass") {
			t.Fatalf("matrix missing pass row:\n%s", out)
		}
	}
	if !strings.Contains(out, "AC matrix:") {
		t.Fatalf("expected matrix header:\n%s", out)
	}
}

func TestAC2FailOrGapNeverSilentComplete(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001b")

	t.Run("one_fail", func(t *testing.T) {
		r := Evaluate(threeACFixture(boolPtr(true), boolPtr(false), boolPtr(true)))
		if r.Status != "PARTIAL" {
			t.Fatalf("status = %s, want PARTIAL", r.Status)
		}
		out := FormatMatrix(r)
		if !strings.Contains(out, "AC-2") || !strings.Contains(out, "fail") {
			t.Fatalf("expected fail cell:\n%s", out)
		}
	})

	t.Run("one_unbound_gap", func(t *testing.T) {
		r := Evaluate(threeACFixture(boolPtr(true), boolPtr(true), nil))
		if r.Status != "PARTIAL" {
			t.Fatalf("status = %s, want PARTIAL", r.Status)
		}
		out := FormatMatrix(r)
		if !strings.Contains(out, "AC-3") || !strings.Contains(out, "gap") {
			t.Fatalf("expected gap cell:\n%s", out)
		}
	})
}

func TestAC4ReqLevelOnlyBindings(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-001b")
	r := Evaluate(Requirement{
		ID:  "REQ-LEGACY-001",
		ACs: nil,
		Bindings: []Binding{
			{ID: "TB-req", ReqID: "REQ-LEGACY-001", Passed: boolPtr(true)},
		},
	})
	if r.Status != "COMPLETE" {
		t.Fatalf("status = %s", r.Status)
	}

	r = Evaluate(Requirement{
		ID:  "REQ-LEGACY-002",
		ACs: nil,
		Bindings: []Binding{
			{ID: "TB-a", ReqID: "REQ-LEGACY-002", Passed: boolPtr(true)},
			{ID: "TB-b", ReqID: "REQ-LEGACY-002", Passed: boolPtr(false)},
		},
	})
	if r.Status != "PARTIAL" {
		t.Fatalf("status = %s, want PARTIAL", r.Status)
	}
}
