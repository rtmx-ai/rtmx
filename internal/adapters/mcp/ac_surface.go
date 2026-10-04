package mcp

import "github.com/rtmx-ai/rtmx/internal/docmodel/acverify"

// acView is one AC on an MCP payload. Rationale is never included.
type acView struct {
	ACID      string `json:"ac_id"`
	Statement string `json:"statement,omitempty"`
	Required  bool   `json:"required"`
	Evidence  string `json:"evidence,omitempty"`
}

type bindingView struct {
	BindingID    string `json:"binding_id"`
	ACID         string `json:"ac_id,omitempty"`
	TestFunction string `json:"test_function,omitempty"`
}

type acGap struct {
	ACID   string `json:"ac_id"`
	Reason string `json:"reason"`
}

// attachACSurface fills MCP backlog/verify fields from the optional document.
// Narrative and companion prose are omitted.
func (s *Server) attachACSurface(acs *[]acView, bindings *[]bindingView, gaps *[]acGap, gapCount *int, reqID string, evidence []acverify.EvidenceHit) {
	if s == nil || s.acDocument == nil || reqID == "" {
		return
	}
	var matched *acverify.DocumentRequirement
	for i := range s.acDocument.Requirements {
		if s.acDocument.Requirements[i].ReqID == reqID {
			matched = &s.acDocument.Requirements[i]
			break
		}
	}
	if matched == nil {
		return
	}

	results := acverify.EvaluateDocument(&acverify.DocumentFile{
		SchemaVersion: s.acDocument.SchemaVersion,
		Requirements:  []acverify.DocumentRequirement{*matched},
	}, evidence)
	evidenceByAC := map[string]string{}
	if len(results) == 1 {
		for _, cell := range results[0].Matrix {
			evidenceByAC[cell.ACID] = cell.Status
		}
	}

	outACs := make([]acView, 0, len(matched.ACs))
	outGaps := make([]acGap, 0)
	for _, ac := range matched.ACs {
		required := true
		if ac.Required != nil {
			required = *ac.Required
		}
		ev := evidenceByAC[ac.ID]
		outACs = append(outACs, acView{
			ACID:      ac.ID,
			Statement: ac.Statement,
			Required:  required,
			Evidence:  ev,
		})
		if required && ev != "pass" {
			reason := "no_passing_binding"
			if ev == "fail" {
				reason = "fail"
			}
			outGaps = append(outGaps, acGap{ACID: ac.ID, Reason: reason})
		}
	}
	outBindings := make([]bindingView, 0, len(matched.TestBindings))
	for _, b := range matched.TestBindings {
		outBindings = append(outBindings, bindingView{
			BindingID:    b.ID,
			ACID:         b.ACID,
			TestFunction: b.TestFunction,
		})
	}
	*acs = outACs
	*bindings = outBindings
	*gaps = outGaps
	*gapCount = len(outGaps)
}
