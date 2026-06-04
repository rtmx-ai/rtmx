package config

import (
	"os"
	"testing"
)

func TestCompletenessDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.RTMX.Completeness.Policy != "simple" {
		t.Errorf("default completeness policy = %q, want simple", cfg.RTMX.Completeness.Policy)
	}
	if cfg.RTMX.Completeness.IsCombinations() {
		t.Error("default policy should not be combinations")
	}
}

func TestCompletenessHelpers(t *testing.T) {
	c := CompletenessConfig{Policy: "combinations"}
	if !c.IsCombinations() {
		t.Error("IsCombinations() should be true")
	}
	// ShouldRequireAllPass defaults to true when unset.
	if !c.ShouldRequireAllPass() {
		t.Error("ShouldRequireAllPass() should default to true")
	}
	f := false
	c.RequireAllPass = &f
	if c.ShouldRequireAllPass() {
		t.Error("ShouldRequireAllPass() should honor explicit false")
	}
	// EffectiveMinCombinations defaults to 1.
	if got := (CompletenessConfig{}).EffectiveMinCombinations(); got != 1 {
		t.Errorf("EffectiveMinCombinations() default = %d, want 1", got)
	}
	if got := (CompletenessConfig{MinCombinations: 3}).EffectiveMinCombinations(); got != 3 {
		t.Errorf("EffectiveMinCombinations() = %d, want 3", got)
	}
}

func TestCompletenessParsesFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	yaml := `rtmx:
  database: .rtmx/database.csv
  completeness:
    policy: combinations
    dimensions: [scope, technique]
    min_combinations: 3
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	c := cfg.RTMX.Completeness
	if !c.IsCombinations() {
		t.Error("expected combinations policy")
	}
	if len(c.Dimensions) != 2 || c.Dimensions[0] != "scope" || c.Dimensions[1] != "technique" {
		t.Errorf("dimensions = %v", c.Dimensions)
	}
	if c.EffectiveMinCombinations() != 3 {
		t.Errorf("min_combinations = %d, want 3", c.MinCombinations)
	}
}
