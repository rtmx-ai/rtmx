package results

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// REQ-VERIFY-002: RTMX Results JSON Schema Validation

func TestResultsSchemaValidation(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	// End-to-end: parse, validate, group
	input := `[
		{
			"marker": {"req_id": "REQ-AUTH-001", "scope": "unit", "test_name": "test_login", "test_file": "test_auth.py", "line": 10},
			"passed": true, "duration_ms": 5.0, "timestamp": "2026-02-20T18:45:00Z"
		},
		{
			"marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login_edge", "test_file": "test_auth.py"},
			"passed": false, "error": "AssertionError"
		},
		{
			"marker": {"req_id": "REQ-DATA-001", "test_name": "test_parse", "test_file": "test_data.py"},
			"passed": true
		}
	]`

	results, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Parse() got %d results, want 3", len(results))
	}

	errs := Validate(results)
	if len(errs) != 0 {
		t.Errorf("Validate() got errors: %v", errs)
	}

	grouped := GroupByRequirement(results)
	if len(grouped) != 2 {
		t.Errorf("GroupByRequirement() got %d groups, want 2", len(grouped))
	}
}

func TestParseValidResultsMinimal(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	input := `[
		{
			"marker": {
				"req_id": "REQ-AUTH-001",
				"test_name": "test_login",
				"test_file": "test_auth.py"
			},
			"passed": true
		}
	]`

	results, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Parse() got %d results, want 1", len(results))
	}
	if results[0].Marker.ReqID != "REQ-AUTH-001" {
		t.Errorf("ReqID = %q, want %q", results[0].Marker.ReqID, "REQ-AUTH-001")
	}
	if results[0].Marker.TestName != "test_login" {
		t.Errorf("TestName = %q, want %q", results[0].Marker.TestName, "test_login")
	}
	if !results[0].Passed {
		t.Error("Passed = false, want true")
	}
}

func TestParseValidResultsAllFields(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	input := `[
		{
			"marker": {
				"req_id": "REQ-AUTH-001",
				"scope": "unit",
				"technique": "nominal",
				"env": "simulation",
				"test_name": "test_login_success",
				"test_file": "test_auth.py",
				"line": 42
			},
			"passed": true,
			"duration_ms": 15.5,
			"error": "",
			"timestamp": "2026-02-20T18:45:00Z"
		}
	]`

	results, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Parse() got %d results, want 1", len(results))
	}

	r := results[0]
	if r.Marker.Scope != "unit" {
		t.Errorf("Scope = %q, want %q", r.Marker.Scope, "unit")
	}
	if r.Marker.Technique != "nominal" {
		t.Errorf("Technique = %q, want %q", r.Marker.Technique, "nominal")
	}
	if r.Marker.Env != "simulation" {
		t.Errorf("Env = %q, want %q", r.Marker.Env, "simulation")
	}
	if r.Marker.Line != 42 {
		t.Errorf("Line = %d, want %d", r.Marker.Line, 42)
	}
	if r.Duration != 15.5 {
		t.Errorf("Duration = %f, want %f", r.Duration, 15.5)
	}
}

func TestParseMultipleResults(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	input := `[
		{
			"marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
			"passed": true
		},
		{
			"marker": {"req_id": "REQ-AUTH-002", "test_name": "test_logout", "test_file": "test_auth.py"},
			"passed": false,
			"error": "AssertionError"
		}
	]`

	results, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Parse() got %d results, want 2", len(results))
	}
	if results[0].Passed != true {
		t.Error("results[0].Passed = false, want true")
	}
	if results[1].Passed != false {
		t.Error("results[1].Passed = true, want false")
	}
	if results[1].Error != "AssertionError" {
		t.Errorf("results[1].Error = %q, want %q", results[1].Error, "AssertionError")
	}
}

func TestParseEmptyArray(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results, err := Parse(strings.NewReader("[]"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Parse() got %d results, want 0", len(results))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	_, err := Parse(strings.NewReader("not json"))
	if err == nil {
		t.Fatal("Parse() expected error for invalid JSON")
	}
}

func TestValidateValidResults(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:    "REQ-AUTH-001",
				TestName: "test_login",
				TestFile: "test_auth.py",
				Scope:    "unit",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) != 0 {
		t.Errorf("Validate() got %d errors, want 0: %v", len(errs), errs)
	}
}

func TestValidateInvalidReqID(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:    "INVALID-ID",
				TestName: "test_foo",
				TestFile: "test.py",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for invalid req_id")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "req_id") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() errors should mention req_id: %v", errs)
	}
}

func TestValidateMissingTestName(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:    "REQ-AUTH-001",
				TestFile: "test.py",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for missing test_name")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "test_name") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() errors should mention test_name: %v", errs)
	}
}

func TestValidateMissingTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:    "REQ-AUTH-001",
				TestName: "test_foo",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for missing test_file")
	}
}

func TestValidateInvalidScope(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:    "REQ-AUTH-001",
				TestName: "test_foo",
				TestFile: "test.py",
				Scope:    "not_a_valid_scope",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for invalid scope")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "scope") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() errors should mention scope: %v", errs)
	}
}

func TestValidateInvalidTechnique(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:     "REQ-AUTH-001",
				TestName:  "test_foo",
				TestFile:  "test.py",
				Technique: "invalid_technique",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for invalid technique")
	}
}

func TestValidateInvalidEnv(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{
			Marker: Marker{
				ReqID:    "REQ-AUTH-001",
				TestName: "test_foo",
				TestFile: "test.py",
				Env:      "invalid_env",
			},
			Passed: true,
		},
	}

	errs := Validate(results)
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for invalid env")
	}
}

func TestValidateOptionalFieldsOmitted(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	// All optional fields empty - should be valid
	results := []Result{
		{
			Marker: Marker{
				ReqID:    "REQ-AUTH-001",
				TestName: "test_foo",
				TestFile: "test.py",
			},
			Passed: false,
		},
	}

	errs := Validate(results)
	if len(errs) != 0 {
		t.Errorf("Validate() got %d errors for valid minimal result: %v", len(errs), errs)
	}
}

func TestParseAndValidateRoundTrip(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	input := `[
		{
			"marker": {
				"req_id": "REQ-AUTH-001",
				"scope": "integration",
				"technique": "nominal",
				"env": "simulation",
				"test_name": "test_login",
				"test_file": "test_auth.py",
				"line": 10
			},
			"passed": true,
			"duration_ms": 42.0,
			"timestamp": "2026-02-20T18:45:00Z"
		},
		{
			"marker": {
				"req_id": "REQ-AUTH-002",
				"test_name": "test_logout",
				"test_file": "test_auth.py"
			},
			"passed": false,
			"error": "timeout"
		}
	]`

	results, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	errs := Validate(results)
	if len(errs) != 0 {
		t.Errorf("Validate() got %d errors: %v", len(errs), errs)
	}
}

func TestGroupByRequirement(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	results := []Result{
		{Marker: Marker{ReqID: "REQ-AUTH-001", TestName: "test_a", TestFile: "t.py"}, Passed: true},
		{Marker: Marker{ReqID: "REQ-AUTH-001", TestName: "test_b", TestFile: "t.py"}, Passed: true},
		{Marker: Marker{ReqID: "REQ-AUTH-002", TestName: "test_c", TestFile: "t.py"}, Passed: false},
	}

	grouped := GroupByRequirement(results)
	if len(grouped) != 2 {
		t.Fatalf("GroupByRequirement() got %d groups, want 2", len(grouped))
	}
	if len(grouped["REQ-AUTH-001"]) != 2 {
		t.Errorf("REQ-AUTH-001 got %d results, want 2", len(grouped["REQ-AUTH-001"]))
	}
	if len(grouped["REQ-AUTH-002"]) != 1 {
		t.Errorf("REQ-AUTH-002 got %d results, want 1", len(grouped["REQ-AUTH-002"]))
	}
}

// REQ-VERIFY-004: lenient and strict results JSON parsing.
//
// Covers the v0.2.4 bug where flat-form payloads silently produced
// zero-valued markers, plus the new strictness on unknown fields.
func TestResultUnmarshalForms(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-004",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		check       func(t *testing.T, r Result)
	}{
		{
			name:  "canonical nested form",
			input: `{"marker":{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go"},"passed":true}`,
			check: func(t *testing.T, r Result) {
				if r.Marker.ReqID != "REQ-X-1" || r.Marker.TestName != "t" || r.Marker.TestFile != "t.go" || !r.Passed {
					t.Errorf("nested decoded wrong: %+v", r)
				}
			},
		},
		{
			name:  "flat form (bug repro from v0.2.4)",
			input: `{"req_id":"REQ-INGEST-030","test_name":"test_foo","test_file":"tests/unit/test_foo.py","status":"pass"}`,
			check: func(t *testing.T, r Result) {
				if r.Marker.ReqID != "REQ-INGEST-030" {
					t.Errorf("flat req_id not promoted: %+v", r)
				}
				if r.Marker.TestName != "test_foo" || r.Marker.TestFile != "tests/unit/test_foo.py" {
					t.Errorf("flat test_name/test_file not promoted: %+v", r)
				}
				if !r.Passed {
					t.Errorf("status pass should set Passed=true: %+v", r)
				}
			},
		},
		{
			name:  "status fail maps to Passed=false",
			input: `{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go","status":"fail"}`,
			check: func(t *testing.T, r Result) {
				if r.Passed {
					t.Errorf("status fail should set Passed=false")
				}
			},
		},
		{
			name:        "unknown top-level field rejected",
			input:       `{"reqid":"REQ-X-1","test_name":"t","test_file":"t.go","passed":true}`,
			wantErr:     true,
			errContains: "reqid",
		},
		{
			name:        "unknown status string rejected",
			input:       `{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go","status":"weird"}`,
			wantErr:     true,
			errContains: "weird",
		},
		{
			name:  "mixed nested and flat: nested wins",
			input: `{"marker":{"req_id":"REQ-NESTED-1","test_name":"n","test_file":"n.go"},"req_id":"REQ-FLAT-1","passed":true}`,
			check: func(t *testing.T, r Result) {
				if r.Marker.ReqID != "REQ-NESTED-1" {
					t.Errorf("nested should win, got %q", r.Marker.ReqID)
				}
			},
		},
		{
			name:  "mixed nested and flat: flat fills blanks",
			input: `{"marker":{"req_id":"REQ-NESTED-1"},"test_name":"flat_t","test_file":"flat.go","passed":true}`,
			check: func(t *testing.T, r Result) {
				if r.Marker.TestName != "flat_t" || r.Marker.TestFile != "flat.go" {
					t.Errorf("flat should fill blanks, got %+v", r.Marker)
				}
			},
		},
		{
			name:  "boolean passed wins over status",
			input: `{"req_id":"REQ-X-1","test_name":"t","test_file":"t.go","passed":false,"status":"pass"}`,
			check: func(t *testing.T, r Result) {
				if r.Passed {
					t.Errorf("explicit passed:false should win, got %+v", r)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Result
			err := r.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil; result=%+v", r)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, r)
			}
		})
	}
}

// TestParseRejectsUnknownFieldArray ensures Parse propagates the strict
// per-element decode error for the v0.2.4 bug repro shape with a typo.
func TestParseRejectsUnknownFieldArray(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-004",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	input := `[{"reqid":"REQ-X-1","test_name":"t","test_file":"t.go","passed":true}]`
	if _, err := Parse(strings.NewReader(input)); err == nil {
		t.Fatal("expected Parse to reject unknown field, got nil")
	}
}
