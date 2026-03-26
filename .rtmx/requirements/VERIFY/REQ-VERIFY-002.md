# REQ-VERIFY-002: RTMX Results JSON Schema Validation

## Metadata
- **Category**: VERIFY
- **Subcategory**: Schema
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007
- **Blocks**: REQ-VERIFY-001

## Requirement

RTMX shall define and enforce a JSON Schema for the common results format, ensuring all language integrations produce compatible output.

## Rationale

The results JSON format is the contract between language-specific test integrations and the Go CLI. A formal schema ensures:
1. Language authors know exactly what to produce
2. The CLI can validate input before processing
3. Schema violations produce actionable error messages

## Design

### Schema Definition

The schema is embedded in the Go CLI binary and also published at `https://rtmx.ai/schemas/results/v1`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://rtmx.ai/schemas/results/v1",
  "title": "RTMX Test Results",
  "description": "Language-agnostic test results format for closed-loop verification",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["marker", "passed"],
    "properties": {
      "marker": {
        "type": "object",
        "required": ["req_id", "test_name", "test_file"],
        "properties": {
          "req_id": {
            "type": "string",
            "pattern": "^REQ-[A-Z]+-[0-9]+$"
          },
          "scope": {
            "type": "string",
            "enum": ["unit", "integration", "system", "acceptance"]
          },
          "technique": {
            "type": "string",
            "enum": ["nominal", "parametric", "monte_carlo", "stress", "boundary"]
          },
          "env": {
            "type": "string",
            "enum": ["simulation", "hil", "anechoic", "field"]
          },
          "test_name": { "type": "string" },
          "test_file": { "type": "string" },
          "line": { "type": "integer", "minimum": 0 }
        }
      },
      "passed": { "type": "boolean" },
      "duration_ms": { "type": "number", "minimum": 0 },
      "error": { "type": "string" },
      "timestamp": { "type": "string", "format": "date-time" }
    }
  }
}
```

### Alignment with OG rtmx

The schema aligns with:
- `pkg/rtmx/rtmx.go` `testResult` struct (Go integration)
- `src/rtmx/markers/schema.py` `MARKER_SCHEMA` (Python OG)
- `internal/cmd/from_go.go` `GoTestResult` struct (current consumer)

### Implementation

```go
// internal/results/schema.go
package results

// Result represents a single test result in the common format.
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

// Parse reads and validates an RTMX results file.
func Parse(r io.Reader) ([]Result, error) { ... }

// Validate checks results against the JSON Schema.
func Validate(results []Result) []error { ... }
```

## Acceptance Criteria

1. JSON Schema defined and embedded in Go binary
2. `Parse()` reads valid results files without error
3. `Parse()` returns clear errors for invalid files (missing req_id, bad format)
4. `Validate()` checks enum values (scope, technique, env)
5. Go `pkg/rtmx` `WriteResultsJSON()` output validates against schema
6. Schema published alongside Go CLI release

## Files to Create/Modify

- `internal/results/schema.go` - New package for results parsing
- `internal/results/schema_test.go` - Schema validation tests
- `internal/cmd/from_go.go` - Migrate to use `results.Parse()`
- `internal/cmd/verify.go` - Use `results.Parse()` for `--results` mode
- `pkg/rtmx/rtmx.go` - Verify output matches schema

## Test Strategy

- Unit tests: Valid/invalid JSON parsing, schema validation
- Fuzz tests: Malformed JSON inputs
- Compatibility tests: Go output validates, Python output validates
- BDD: `features/verify/results-format.feature`
