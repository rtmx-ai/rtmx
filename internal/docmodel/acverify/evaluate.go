package acverify

import (
	"fmt"
	"sort"
	"strings"
)

// Evidence is one test result that may bind to an AC or requirement.
type Evidence struct {
	BindingID string
	Passed    bool
}

// Binding links evidence to an AC and/or requirement (REQ-DATA-001b).
type Binding struct {
	ID     string
	ACID   string
	ReqID  string
	Passed *bool // nil = unbound / no evidence yet
}

// AC is a required or optional acceptance criterion.
type AC struct {
	ID       string
	Required bool
}

// Requirement is the fixture unit under ATDD policy.
type Requirement struct {
	ID       string
	ACs      []AC
	Bindings []Binding
}

// ACCell is one row of the AC evidence matrix.
type ACCell struct {
	ACID   string
	Status string // pass | fail | gap
}

// Result is per-AC matrix plus aggregated requirement status.
type Result struct {
	ReqID      string
	Matrix     []ACCell
	Status     string // COMPLETE | PARTIAL | MISSING
	PolicyName string
}

// PolicyATDDRequiredAll is COMPLETE iff every required AC passes;
// PARTIAL if any required AC passes or fails (but not all pass);
// MISSING if no required AC has pass evidence.
const PolicyATDDRequiredAll = "atdd-required-all-v0"

// Evaluate applies PolicyATDDRequiredAll to a requirement fixture.
func Evaluate(req Requirement) Result {
	byAC := map[string]string{} // ac_id -> pass|fail|gap
	for _, ac := range req.ACs {
		byAC[ac.ID] = "gap"
	}

	reqLevelPass := false
	reqLevelFail := false
	for _, b := range req.Bindings {
		if b.Passed == nil {
			continue
		}
		if b.ACID != "" {
			if *b.Passed {
				byAC[b.ACID] = "pass"
			} else if byAC[b.ACID] != "pass" {
				byAC[b.ACID] = "fail"
			}
			continue
		}
		if b.ReqID == req.ID || (b.ReqID == "" && b.ACID == "") {
			if *b.Passed {
				reqLevelPass = true
			} else {
				reqLevelFail = true
			}
		}
	}

	// Compat: req-level-only bindings with no ACs (or as fallback when ACs empty).
	if len(req.ACs) == 0 {
		status := "MISSING"
		if reqLevelPass && !reqLevelFail {
			status = "COMPLETE"
		} else if reqLevelPass || reqLevelFail {
			status = "PARTIAL"
		}
		return Result{ReqID: req.ID, Status: status, PolicyName: PolicyATDDRequiredAll}
	}

	matrix := make([]ACCell, 0, len(req.ACs))
	ids := make([]string, 0, len(req.ACs))
	required := map[string]bool{}
	for _, ac := range req.ACs {
		ids = append(ids, ac.ID)
		required[ac.ID] = ac.Required
	}
	sort.Strings(ids)
	for _, id := range ids {
		matrix = append(matrix, ACCell{ACID: id, Status: byAC[id]})
	}

	passN, failN, gapN := 0, 0, 0
	for _, ac := range req.ACs {
		if !ac.Required {
			continue
		}
		switch byAC[ac.ID] {
		case "pass":
			passN++
		case "fail":
			failN++
		default:
			gapN++
		}
	}
	totalReq := passN + failN + gapN
	status := "MISSING"
	if totalReq > 0 && passN == totalReq {
		status = "COMPLETE"
	} else if passN > 0 || failN > 0 {
		// Unbound (gap) or fail on a required AC ⇒ never silent COMPLETE.
		status = "PARTIAL"
	}

	return Result{
		ReqID:      req.ID,
		Matrix:     matrix,
		Status:     status,
		PolicyName: PolicyATDDRequiredAll,
	}
}

// FormatMatrix renders the AC pass/fail/gap table for spike output.
func FormatMatrix(r Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Requirement %s  status=%s  policy=%s\n", r.ReqID, r.Status, r.PolicyName)
	fmt.Fprintf(&b, "AC matrix:\n")
	for _, cell := range r.Matrix {
		fmt.Fprintf(&b, "  %-8s  %s\n", cell.ACID, cell.Status)
	}
	return b.String()
}
