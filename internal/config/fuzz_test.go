package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// FuzzConfigParse tests YAML config parsing with arbitrary input.
// It ensures that malformed YAML doesn't cause panics.
// REQ-GO-070: Go CLI shall include fuzz tests for CSV and YAML parsing
func FuzzConfigParse(f *testing.F) {
	// Seed corpus with valid YAML configs
	f.Add(`rtmx:
  database: .rtmx/database.csv
  schema: core
`)
	f.Add(`rtmx:
  database: custom/database.csv
  schema: phoenix
  phases:
    1: Foundation
    2: Integration
    3: Testing
`)
	f.Add(`rtmx:
  adapters:
    github:
      enabled: true
      repo: rtmx-ai/rtmx-go
    jira:
      enabled: false
`)
	f.Add(`rtmx:
  pytest:
    marker_prefix: req
    register_markers: true
  agents:
    claude:
      enabled: true
      config_path: CLAUDE.md
`)
	f.Add(`rtmx:
  sync:
    conflict_resolution: manual
    remotes:
      upstream:
        repo: rtmx-ai/rtmx
        database: docs/rtm_database.csv
`)
	f.Add(`rtmx:
  mcp:
    enabled: true
    port: 3000
    host: localhost
`)
	f.Add(`rtmx:
  auth:
    provider: zitadel
    scopes:
      - openid
      - profile
      - email
    callback_port: 8765
`)

	// Edge cases
	f.Add(``) // Empty
	f.Add(`rtmx:`) // Empty rtmx section
	f.Add(`rtmx:
  database: ` + strings.Repeat("a", 10000) + `
`) // Very long string
	f.Add(`# Just a comment`) // Only comment
	f.Add(`---
rtmx:
  database: test.csv
`) // Document start marker
	f.Add(`rtmx:
  phases:
    1: One
    2: Two
    99999999: Very High
`) // Large phase number

	// Malformed YAML
	f.Add(`{invalid json-like}`)
	f.Add(`rtmx:
    database: bad indent
  schema: wrong
`)
	f.Add(`rtmx:
  - this
  - is
  - a list
`)
	f.Add("\x00\x00\x00") // Null bytes
	f.Add(`rtmx:
  database: "unclosed string
`)
	f.Add(`rtmx: [1, 2, 3]`) // Array instead of object
	f.Add(`rtmx:
  phases:
    not_a_number: Foundation
`)

	// Unicode edge cases
	f.Add(`rtmx:
  database: test.csv
  notes: ` + "\u200B" + `zero-width space
`)
	f.Add("\uFEFF" + `rtmx:
  database: test.csv
`) // BOM at start

	f.Fuzz(func(t *testing.T, data string) {
		// The function should not panic on any input
		config := DefaultConfig()
		err := yaml.Unmarshal([]byte(data), config)

		// If parsing succeeded, the config should be usable
		if err == nil {
			// Exercise config methods without panicking
			_ = config.DatabasePath("/test")
			_ = config.RequirementsPath("/test")
			_ = config.PhaseDescription(1)
			_ = config.PhaseDescription(99)

			// Check that RTMX section is accessible
			_ = config.RTMX.Database
			_ = config.RTMX.Schema
			_ = config.RTMX.Phases
			_ = config.RTMX.Adapters.GitHub.Enabled
			_ = config.RTMX.Adapters.Jira.Enabled
			_ = config.RTMX.MCP.Port
			_ = config.RTMX.Auth.Provider
		}
	})
}

// FuzzConfigRoundTrip tests that valid configs survive serialization round trip.
func FuzzConfigRoundTrip(f *testing.F) {
	// Seed with config-like data
	f.Add(".rtmx/database.csv", "core", 1, "Foundation")
	f.Add("custom/db.csv", "phoenix", 2, "Integration")
	f.Add("", "", 0, "")
	f.Add("path/with spaces/db.csv", "custom-schema", 99, "Final Phase")

	f.Fuzz(func(t *testing.T, database, schema string, phaseNum int, phaseName string) {
		// Create config with fuzzed data
		config := DefaultConfig()
		config.RTMX.Database = database
		config.RTMX.Schema = schema
		if phaseNum > 0 && phaseName != "" {
			config.RTMX.Phases[phaseNum] = phaseName
		}

		// Serialize to YAML
		data, err := yaml.Marshal(config)
		if err != nil {
			// Some inputs may not be serializable
			return
		}

		// Parse back
		config2 := DefaultConfig()
		err = yaml.Unmarshal(data, config2)
		if err != nil {
			t.Fatalf("Failed to parse serialized config: %v", err)
		}

		// Verify round-trip
		if config2.RTMX.Database != config.RTMX.Database {
			t.Errorf("Database: got %q, want %q", config2.RTMX.Database, config.RTMX.Database)
		}
		if config2.RTMX.Schema != config.RTMX.Schema {
			t.Errorf("Schema: got %q, want %q", config2.RTMX.Schema, config.RTMX.Schema)
		}
		if phaseNum > 0 && phaseName != "" {
			if config2.RTMX.Phases[phaseNum] != phaseName {
				t.Errorf("Phase %d: got %q, want %q", phaseNum, config2.RTMX.Phases[phaseNum], phaseName)
			}
		}
	})
}

// FuzzDatabasePath tests database path resolution.
func FuzzDatabasePath(f *testing.F) {
	f.Add(".rtmx/database.csv", "/project")
	f.Add("/absolute/path/db.csv", "/project")
	f.Add("relative/path.csv", "/home/user/project")
	f.Add("", "/project")
	f.Add("path with spaces/db.csv", "/project/with spaces")
	f.Add(strings.Repeat("a/", 100)+"db.csv", "/project")

	f.Fuzz(func(t *testing.T, dbPath, baseDir string) {
		config := DefaultConfig()
		config.RTMX.Database = dbPath

		// Should not panic
		result := config.DatabasePath(baseDir)

		// If absolute path, should return as-is
		if len(dbPath) > 0 && dbPath[0] == '/' {
			if result != dbPath {
				t.Errorf("Absolute path not preserved: got %q, want %q", result, dbPath)
			}
		}
	})
}

// FuzzPhaseDescription tests phase description retrieval.
func FuzzPhaseDescription(f *testing.F) {
	f.Add(1)
	f.Add(2)
	f.Add(3)
	f.Add(0)
	f.Add(-1)
	f.Add(100)
	f.Add(999999)
	f.Add(-999999)

	f.Fuzz(func(t *testing.T, phase int) {
		config := DefaultConfig()

		// Should not panic
		desc := config.PhaseDescription(phase)

		// Should return something
		if desc == "" {
			t.Errorf("PhaseDescription(%d) returned empty string", phase)
		}

		// If phase is in map, should return that description
		if expected, ok := config.RTMX.Phases[phase]; ok {
			if desc != expected {
				t.Errorf("PhaseDescription(%d) = %q, want %q", phase, desc, expected)
			}
		}
	})
}
