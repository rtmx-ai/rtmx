// Package results defines the language-agnostic RTMX test results format
// and provides parsing and validation for cross-language verification.
package results

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

// Result represents a single test result in the common RTMX format.
type Result struct {
	Marker    Marker  `json:"marker"`
	Passed    bool    `json:"passed"`
	Duration  float64 `json:"duration_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
	Timestamp string  `json:"timestamp,omitempty"`
}

// Marker represents requirement marker metadata.
type Marker struct {
	ReqID     string `json:"req_id"`
	Scope     string `json:"scope,omitempty"`
	Technique string `json:"technique,omitempty"`
	Env       string `json:"env,omitempty"`
	TestName  string `json:"test_name"`
	TestFile  string `json:"test_file"`
	Line      int    `json:"line,omitempty"`
}

var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z]+-[0-9]+$`)

var validScopes = map[string]bool{
	"unit": true, "integration": true, "system": true, "acceptance": true,
}

var validTechniques = map[string]bool{
	"nominal": true, "parametric": true, "monte_carlo": true, "stress": true, "boundary": true,
}

var validEnvs = map[string]bool{
	"simulation": true, "hil": true, "anechoic": true, "field": true,
}

// Parse reads and decodes an RTMX results JSON file.
func Parse(r io.Reader) ([]Result, error) {
	var results []Result
	if err := json.NewDecoder(r).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse results JSON: %w", err)
	}
	return results, nil
}

// Validate checks results against the RTMX results schema.
func Validate(results []Result) []error {
	var errs []error
	for i, r := range results {
		prefix := fmt.Sprintf("result[%d]", i)

		if !reqIDPattern.MatchString(r.Marker.ReqID) {
			errs = append(errs, fmt.Errorf("%s: invalid req_id %q (expected REQ-[A-Z]+-[0-9]+)", prefix, r.Marker.ReqID))
		}
		if r.Marker.TestName == "" {
			errs = append(errs, fmt.Errorf("%s: missing required field test_name", prefix))
		}
		if r.Marker.TestFile == "" {
			errs = append(errs, fmt.Errorf("%s: missing required field test_file", prefix))
		}
		if r.Marker.Scope != "" && !validScopes[r.Marker.Scope] {
			errs = append(errs, fmt.Errorf("%s: invalid scope %q (expected one of: unit, integration, system, acceptance)", prefix, r.Marker.Scope))
		}
		if r.Marker.Technique != "" && !validTechniques[r.Marker.Technique] {
			errs = append(errs, fmt.Errorf("%s: invalid technique %q (expected one of: nominal, parametric, monte_carlo, stress, boundary)", prefix, r.Marker.Technique))
		}
		if r.Marker.Env != "" && !validEnvs[r.Marker.Env] {
			errs = append(errs, fmt.Errorf("%s: invalid env %q (expected one of: simulation, hil, anechoic, field)", prefix, r.Marker.Env))
		}
	}
	return errs
}

// GroupByRequirement groups results by requirement ID.
func GroupByRequirement(results []Result) map[string][]Result {
	grouped := make(map[string][]Result)
	for _, r := range results {
		grouped[r.Marker.ReqID] = append(grouped[r.Marker.ReqID], r)
	}
	return grouped
}
