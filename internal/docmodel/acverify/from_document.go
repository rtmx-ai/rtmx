package acverify

import (
	"encoding/json"
	"fmt"
	"os"
)

// DocumentFile is the on-disk requirement-document/v0 shape (subset for verify).
type DocumentFile struct {
	SchemaVersion string                `json:"schema_version"`
	Requirements  []DocumentRequirement `json:"requirements"`
}

// DocumentRequirement holds ACs and bindings from the document model.
type DocumentRequirement struct {
	ReqID        string            `json:"req_id"`
	ACs          []DocumentAC      `json:"acs"`
	TestBindings []DocumentBinding `json:"test_bindings"`
}

// DocumentAC is one acceptance criterion record.
type DocumentAC struct {
	ID        string `json:"ac_id"`
	Statement string `json:"statement,omitempty"`
	Required  *bool  `json:"required"`
}

// DocumentBinding links a test identity to an AC or req.
type DocumentBinding struct {
	ID           string `json:"binding_id"`
	ACID         string `json:"ac_id"`
	ReqID        string `json:"req_id"`
	TestFunction string `json:"test_function"`
	Marker       string `json:"marker"`
}

// EvidenceHit is one test result used to satisfy a binding.
type EvidenceHit struct {
	TestName string
	ReqID    string
	Passed   bool
}

// LoadDocumentFile reads a requirement-document/v0 JSON file.
func LoadDocumentFile(path string) (*DocumentFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc DocumentFile
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse ac-document: %w", err)
	}
	if doc.SchemaVersion != "" && doc.SchemaVersion != "requirement-document/v0" {
		return nil, fmt.Errorf("unsupported schema_version %q", doc.SchemaVersion)
	}
	if len(doc.Requirements) == 0 {
		return nil, fmt.Errorf("ac-document has no requirements")
	}
	return &doc, nil
}

// EvaluateDocument applies atdd-required-all-v0 using optional evidence hits.
// Bindings match evidence when TestFunction equals EvidenceHit.TestName
// (and ReqID matches when both are set on the hit).
func EvaluateDocument(doc *DocumentFile, evidence []EvidenceHit) []Result {
	byName := map[string][]EvidenceHit{}
	for _, e := range evidence {
		byName[e.TestName] = append(byName[e.TestName], e)
	}

	out := make([]Result, 0, len(doc.Requirements))
	for _, dr := range doc.Requirements {
		req := Requirement{ID: dr.ReqID}
		for _, ac := range dr.ACs {
			required := true
			if ac.Required != nil {
				required = *ac.Required
			}
			req.ACs = append(req.ACs, AC{ID: ac.ID, Required: required})
		}
		for _, tb := range dr.TestBindings {
			b := Binding{ID: tb.ID, ACID: tb.ACID, ReqID: tb.ReqID}
			if tb.TestFunction != "" {
				if hits := byName[tb.TestFunction]; len(hits) > 0 {
					passed := false
					any := false
					for _, h := range hits {
						if h.ReqID != "" && h.ReqID != dr.ReqID {
							continue
						}
						any = true
						if h.Passed {
							passed = true
						}
					}
					if any {
						b.Passed = &passed
					}
				}
			}
			req.Bindings = append(req.Bindings, b)
		}
		out = append(out, Evaluate(req))
	}
	return out
}
