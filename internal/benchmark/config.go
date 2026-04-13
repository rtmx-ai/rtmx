// Package benchmark provides configuration parsing, validation, and regression
// detection for the RTMX language scanner benchmark framework.
package benchmark

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ExemplarConfig describes the open source project used as a benchmark target.
type ExemplarConfig struct {
	Repo    string `yaml:"repo"`
	Ref     string `yaml:"ref"`
	License string `yaml:"license"`
}

// BenchmarkConfig describes a single language benchmark.
type BenchmarkConfig struct {
	Language        string         `yaml:"language"`
	Exemplar        ExemplarConfig `yaml:"exemplar"`
	CloneDepth      int            `yaml:"clone_depth"`
	SetupCommands   []string       `yaml:"setup_commands"`
	MarkerPatch     string         `yaml:"marker_patch"`
	ExpectedMarkers int            `yaml:"expected_markers"`
	ScanCommand     string         `yaml:"scan_command"`
	VerifyCommand     string         `yaml:"verify_command"`
	TimeoutMinutes    int            `yaml:"timeout_minutes"`
	ExpectedBuildFile string         `yaml:"expected_build_file,omitempty"`
}

// ParseConfig parses YAML data into a BenchmarkConfig.
func ParseConfig(data []byte) (*BenchmarkConfig, error) {
	var cfg BenchmarkConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing benchmark config: %w", err)
	}
	return &cfg, nil
}

// LoadConfig reads a YAML file and returns a validated BenchmarkConfig.
func LoadConfig(path string) (*BenchmarkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading benchmark config %s: %w", path, err)
	}
	cfg, err := ParseConfig(data)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating benchmark config %s: %w", path, err)
	}
	return cfg, nil
}

// Validate checks that required fields are present and applies defaults.
func (c *BenchmarkConfig) Validate() error {
	if c.Language == "" {
		return fmt.Errorf("language is required")
	}
	if c.Exemplar.Repo == "" {
		return fmt.Errorf("exemplar.repo is required")
	}
	if c.Exemplar.Ref == "" {
		return fmt.Errorf("exemplar.ref is required")
	}
	if c.ExpectedMarkers <= 0 {
		return fmt.Errorf("expected_markers must be positive")
	}
	if c.ScanCommand == "" {
		return fmt.Errorf("scan_command is required")
	}
	if c.CloneDepth <= 0 {
		c.CloneDepth = 1
	}
	if c.TimeoutMinutes <= 0 {
		c.TimeoutMinutes = 10
	}
	return nil
}
