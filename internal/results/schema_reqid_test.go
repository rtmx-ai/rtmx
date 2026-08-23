package results

import (
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func reqIDResult(id string) Result {
	return Result{
		Passed: true,
		Marker: Marker{ReqID: id, TestName: "t", TestFile: "t.go"},
	}
}

// TestDefaultReqIDPattern verifies the built-in requirement-ID grammar accepts
// single- and multi-segment category prefixes while rejecting malformed IDs.
//
// REQ-VERIFY-011: Configurable requirement-ID pattern (default behavior).
func TestDefaultReqIDPattern(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-011",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	accepted := []string{
		"REQ-SW-009",         // legacy single-segment
		"REQ-VERIFY-011",     // single-segment, multi-char
		"REQ-E2E-010",        // alphanumeric single-segment
		"REQ-V2-001",         // leading-letter alphanumeric
		"REQ-INFRA-DT-002",   // two-segment category
		"REQ-MODE-S-006",     // two-segment, short tail segment
		"REQ-SW-DSP-015",     // two-segment category
		"REQ-A-B-C-001",      // three-segment category
		"REQ-HW-RF-001b",     // decomposition child: optional lowercase suffix
		"REQ-HW-STRUCT-002c", // decomposition child: optional lowercase suffix
	}
	for _, id := range accepted {
		if errs := Validate([]Result{reqIDResult(id)}); len(errs) != 0 {
			t.Errorf("expected req_id %q to validate, got %v", id, errs)
		}
	}

	rejected := []string{
		"REQ-sw-009",   // lowercase category
		"REQ-SW-",      // missing number
		"REQ-SW",       // no number segment
		"REQ-123-001",  // category starts with a digit
		"SW-009",       // missing REQ- prefix
		"REQ--009",     // empty category segment
		"REQ-SW-001ab", // suffix is a SINGLE optional letter, not two
	}
	for _, id := range rejected {
		if errs := Validate([]Result{reqIDResult(id)}); len(errs) == 0 {
			t.Errorf("expected req_id %q to be rejected", id)
		}
	}
}

// TestConfigurableReqIDPattern verifies that a project can override the
// requirement-ID pattern via the Vocabulary, that an empty override reproduces
// the default, and that an invalid override is reported once and falls back to
// the default.
//
// REQ-VERIFY-011: Configurable requirement-ID pattern (override behavior).
func TestConfigurableReqIDPattern(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-011",
		rtmx.Scope("unit"),
		rtmx.Technique("stress"),
		rtmx.Env("simulation"),
	)

	// A custom convention (lowercase, dot-separated) that the default rejects.
	custom := Vocabulary{ReqIDPattern: `^req\.[a-z]+\.[0-9]+$`}
	if errs := ValidateWithVocabulary([]Result{reqIDResult("req.cli.001")}, custom); len(errs) != 0 {
		t.Errorf("expected custom-pattern id to validate, got %v", errs)
	}
	// The default-form ID is rejected once the custom pattern is in force.
	if errs := ValidateWithVocabulary([]Result{reqIDResult("REQ-SW-009")}, custom); len(errs) == 0 {
		t.Error("expected REQ-SW-009 to be rejected under the custom pattern")
	}

	// An empty override reproduces the default behavior exactly.
	if errs := ValidateWithVocabulary([]Result{reqIDResult("REQ-INFRA-DT-002")}, Vocabulary{}); len(errs) != 0 {
		t.Errorf("empty override should accept the default grammar, got %v", errs)
	}

	// An invalid pattern is reported (and validation still runs on the default).
	bad := Vocabulary{ReqIDPattern: `^REQ-(unclosed`}
	errs := ValidateWithVocabulary([]Result{reqIDResult("REQ-SW-009")}, bad)
	if len(errs) == 0 {
		t.Fatal("expected an error for an invalid req_id_pattern")
	}
	// REQ-SW-009 matches the fallback default, so the only error is the bad
	// pattern itself, not a rejected id.
	if len(errs) != 1 {
		t.Errorf("expected exactly one error (the invalid pattern), got %d: %v", len(errs), errs)
	}
}
