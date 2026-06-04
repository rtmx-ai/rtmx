package results

import (
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func validMarker(scope, technique, env string) Result {
	return Result{
		Passed: true,
		Marker: Marker{
			ReqID:     "REQ-X-001",
			TestName:  "t",
			TestFile:  "t.go",
			Scope:     scope,
			Technique: technique,
			Env:       env,
		},
	}
}

func TestValidateWithVocabulary(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-010",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)
	// A phoenix-style env value is rejected by the built-in vocabulary...
	res := []Result{validMarker("unit", "nominal", "static_field")}
	if errs := Validate(res); len(errs) == 0 {
		t.Fatal("expected built-in validation to reject env=static_field")
	}

	// ...but accepted once the project configures it.
	vocab := Vocabulary{Envs: []string{"static_field", "dynamic_field"}}
	if errs := ValidateWithVocabulary(res, vocab); len(errs) != 0 {
		t.Errorf("expected static_field to validate with custom vocabulary, got %v", errs)
	}

	// Built-in values still validate when a custom vocabulary is set.
	builtin := []Result{validMarker("unit", "nominal", "simulation")}
	if errs := ValidateWithVocabulary(builtin, vocab); len(errs) != 0 {
		t.Errorf("expected built-in env to remain valid, got %v", errs)
	}

	// Custom scope, technique, and env values are honored too.
	custom := []Result{validMarker("field_test", "soak", "dynamic_field")}
	v2 := Vocabulary{Scopes: []string{"field_test"}, Techniques: []string{"soak"}, Envs: []string{"dynamic_field"}}
	if errs := ValidateWithVocabulary(custom, v2); len(errs) != 0 {
		t.Errorf("expected custom scope/technique/env to validate, got %v", errs)
	}

	// An empty vocabulary reproduces the default behavior exactly.
	if errs := ValidateWithVocabulary(builtin, Vocabulary{}); len(errs) != 0 {
		t.Errorf("empty vocabulary should match default Validate, got %v", errs)
	}
}
