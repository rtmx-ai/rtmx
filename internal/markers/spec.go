// Package markers defines the language-agnostic marker annotation
// specification for RTMX requirement traceability.
package markers

import (
	"fmt"
	"regexp"
)

// Marker represents a requirement annotation in a test function.
// This is the canonical definition of the marker format shared
// across all language integrations (Go, Python, Rust, JS, Java, C#).
type Marker struct {
	ReqID     string `json:"req_id"`
	TestName  string `json:"test_name"`
	TestFile  string `json:"test_file"`
	Scope     string `json:"scope,omitempty"`
	Technique string `json:"technique,omitempty"`
	Env       string `json:"env,omitempty"`
	Line      int    `json:"line,omitempty"`
}

// reqIDPattern validates the requirement ID format. The category prefix may be
// one or more uppercase-alphanumeric segments (e.g. REQ-SW-009, REQ-E2E-010,
// REQ-INFRA-DT-002, REQ-MODE-S-006); the final segment is the numeric index.
var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[0-9]+$`)

// ValidScopes lists all valid scope values.
var ValidScopes = []string{"unit", "integration", "system", "acceptance"}

// ValidTechniques lists all valid technique values.
var ValidTechniques = []string{"nominal", "parametric", "monte_carlo", "stress", "boundary"}

// ValidEnvs lists all valid environment values.
var ValidEnvs = []string{"simulation", "hil", "anechoic", "field"}

// Validate checks that a Marker conforms to the specification.
// Returns a list of validation errors (empty if valid).
func (m *Marker) Validate() []error {
	var errs []error

	if m.ReqID == "" {
		errs = append(errs, fmt.Errorf("req_id is required"))
	} else if !reqIDPattern.MatchString(m.ReqID) {
		errs = append(errs, fmt.Errorf("invalid req_id %q: must match REQ-<CATEGORY>-<NUMBER>, e.g. REQ-SW-009 or REQ-INFRA-DT-002", m.ReqID))
	}

	if m.TestName == "" {
		errs = append(errs, fmt.Errorf("test_name is required"))
	}

	if m.TestFile == "" {
		errs = append(errs, fmt.Errorf("test_file is required"))
	}

	if m.Scope != "" && !contains(ValidScopes, m.Scope) {
		errs = append(errs, fmt.Errorf("invalid scope %q: must be one of %v", m.Scope, ValidScopes))
	}

	if m.Technique != "" && !contains(ValidTechniques, m.Technique) {
		errs = append(errs, fmt.Errorf("invalid technique %q: must be one of %v", m.Technique, ValidTechniques))
	}

	if m.Env != "" && !contains(ValidEnvs, m.Env) {
		errs = append(errs, fmt.Errorf("invalid env %q: must be one of %v", m.Env, ValidEnvs))
	}

	return errs
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
