// Package results defines the language-agnostic RTMX test results format
// and provides parsing and validation for cross-language verification.
package results

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// Result represents a single test result in the common RTMX format.
//
// The canonical JSON shape nests marker fields under a "marker" object:
//
//	{"marker":{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go"},"passed":true}
//
// For convenience, UnmarshalJSON also accepts a "flat" form where marker
// fields appear at the top level, and a "status" string ("pass"/"fail")
// in place of the boolean "passed". Unknown fields are rejected so that
// typos surface immediately rather than producing silent zero values.
// See REQ-VERIFY-004.
type Result struct {
	Marker    Marker  `json:"marker"`
	Passed    bool    `json:"passed"`
	Duration  float64 `json:"duration_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
	Timestamp string  `json:"timestamp,omitempty"`
}

// rawResult is the wire representation accepted by Result.UnmarshalJSON.
// It exposes both the canonical nested marker and the convenience flat
// fields so either form decodes successfully.
type rawResult struct {
	Marker    *Marker `json:"marker"`
	Passed    *bool   `json:"passed"`
	Status    *string `json:"status"`
	Duration  float64 `json:"duration_ms"`
	Error     string  `json:"error"`
	Timestamp string  `json:"timestamp"`

	// Flat-form fallbacks promoted into Marker if marker is absent.
	ReqID     *string `json:"req_id"`
	Scope     *string `json:"scope"`
	Technique *string `json:"technique"`
	Env       *string `json:"env"`
	TestName  *string `json:"test_name"`
	TestFile  *string `json:"test_file"`
	Line      *int    `json:"line"`
}

// UnmarshalJSON implements strict decoding with a flat-form compatibility
// shim. Unknown fields cause a decode error.
func (r *Result) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var raw rawResult
	if err := dec.Decode(&raw); err != nil {
		return fmt.Errorf("decode result: %w", err)
	}

	if raw.Marker != nil {
		r.Marker = *raw.Marker
	}

	// Promote flat fields. Nested values win; flat fills blanks.
	promote := func(dst *string, src *string) {
		if src != nil && *dst == "" {
			*dst = *src
		}
	}
	promote(&r.Marker.ReqID, raw.ReqID)
	promote(&r.Marker.Scope, raw.Scope)
	promote(&r.Marker.Technique, raw.Technique)
	promote(&r.Marker.Env, raw.Env)
	promote(&r.Marker.TestName, raw.TestName)
	promote(&r.Marker.TestFile, raw.TestFile)
	if raw.Line != nil && r.Marker.Line == 0 {
		r.Marker.Line = *raw.Line
	}

	switch {
	case raw.Passed != nil:
		r.Passed = *raw.Passed
	case raw.Status != nil:
		switch strings.ToLower(strings.TrimSpace(*raw.Status)) {
		case "pass", "passed", "ok", "success":
			r.Passed = true
		case "fail", "failed", "error", "errored", "skip", "skipped":
			r.Passed = false
		default:
			return fmt.Errorf("decode result: unrecognized status %q (expected pass/fail)", *raw.Status)
		}
	}

	r.Duration = raw.Duration
	r.Error = raw.Error
	r.Timestamp = raw.Timestamp
	return nil
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

// Vocabulary defines project-configured accepted values for the marker
// dimensions. Values in each field are accepted in addition to the built-in
// vocabulary for that dimension; an empty field uses only the built-ins. An
// entirely empty Vocabulary reproduces the default validation behavior. See
// REQ-VERIFY-010.
type Vocabulary struct {
	Scopes     []string
	Techniques []string
	Envs       []string
}

// Validate checks results against the RTMX results schema using the built-in
// dimension vocabularies.
func Validate(results []Result) []error {
	return ValidateWithVocabulary(results, Vocabulary{})
}

// ValidateWithVocabulary checks results against the schema, accepting the given
// per-dimension vocabularies in addition to the built-in values.
func ValidateWithVocabulary(results []Result, vocab Vocabulary) []error {
	scopes := mergeVocabulary(validScopes, vocab.Scopes)
	techniques := mergeVocabulary(validTechniques, vocab.Techniques)
	envs := mergeVocabulary(validEnvs, vocab.Envs)

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
		if r.Marker.Scope != "" && !scopes[r.Marker.Scope] {
			errs = append(errs, fmt.Errorf("%s: invalid scope %q (expected one of: %s)", prefix, r.Marker.Scope, vocabularyList(scopes)))
		}
		if r.Marker.Technique != "" && !techniques[r.Marker.Technique] {
			errs = append(errs, fmt.Errorf("%s: invalid technique %q (expected one of: %s)", prefix, r.Marker.Technique, vocabularyList(techniques)))
		}
		if r.Marker.Env != "" && !envs[r.Marker.Env] {
			errs = append(errs, fmt.Errorf("%s: invalid env %q (expected one of: %s)", prefix, r.Marker.Env, vocabularyList(envs)))
		}
	}
	return errs
}

// mergeVocabulary returns a set containing the built-in values plus any
// non-empty extras.
func mergeVocabulary(builtin map[string]bool, extra []string) map[string]bool {
	out := make(map[string]bool, len(builtin)+len(extra))
	for k := range builtin {
		out[k] = true
	}
	for _, e := range extra {
		if e != "" {
			out[e] = true
		}
	}
	return out
}

// vocabularyList renders a set as a sorted, comma-separated string for messages.
func vocabularyList(set map[string]bool) string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

// GroupByRequirement groups results by requirement ID.
func GroupByRequirement(results []Result) map[string][]Result {
	grouped := make(map[string][]Result)
	for _, r := range results {
		grouped[r.Marker.ReqID] = append(grouped[r.Marker.ReqID], r)
	}
	return grouped
}
