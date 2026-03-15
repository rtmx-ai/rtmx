# REQ-VERIFY-001: Cross-Language Results File Support

## Metadata
- **Category**: VERIFY
- **Subcategory**: Cross-Language
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007
- **Blocks**: REQ-LANG-004, REQ-LANG-005, REQ-LANG-006, REQ-LANG-008

## Requirement

`rtmx verify` shall accept a `--results` flag to consume the language-agnostic RTMX results JSON format, enabling cross-language closed-loop verification.

## Rationale

Currently `rtmx verify` only parses Go test JSON output (`go test -json`). The `from-go` command reads the RTMX results JSON format but operates independently from the verify workflow. To enable cross-language verification (Python, Rust, JavaScript, Java), verify must support the common results format defined by REQ-LANG-007.

This is the critical path blocker for all language extensions. Without this, we cannot dogfood RTMX as we build the Python extension.

## Gap Analysis

| Command | Input Format | Status Update | Streaming |
|---------|-------------|---------------|-----------|
| `verify` | Go test JSON only | Yes | Yes |
| `from-go` | RTMX results JSON | Yes | No (file) |
| **`verify --results`** | RTMX results JSON | Yes | No (file) |

## Design

### CLI Interface

```bash
# Current: Go-only streaming
rtmx verify                              # go test -json ./...
rtmx verify --command "go test -json"    # custom command

# New: Language-agnostic results file
rtmx verify --results rtmx-results.json  # read results file
rtmx verify --results -                  # read from stdin

# Combined workflows
pytest --rtmx-output=results.json && rtmx verify --results results.json --update
cargo test && rtmx verify --results rtmx-results.json --update
```

### Results File Format (from REQ-LANG-007)

```json
[
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
]
```

### Implementation

Extend `verify.go` to:
1. Accept `--results` flag (path to results file or `-` for stdin)
2. Parse results file using existing `GoTestResult` struct from `from_go.go`
3. Feed parsed results into existing `mapTestsToRequirements` pipeline
4. Reuse all existing status determination and update logic

### Relationship to `from-go`

`from-go` should be refactored to call the same underlying logic as `verify --results`. Eventually `from-go` becomes an alias or is deprecated in favor of `verify --results`.

## Acceptance Criteria

1. `rtmx verify --results rtmx-results.json` reads and processes results
2. `rtmx verify --results - < results.json` reads from stdin
3. Status updates follow same rules as streaming verify (pass→COMPLETE, fail→PARTIAL)
4. `--update`, `--dry-run`, `--verbose` flags work with `--results`
5. Invalid/malformed results files produce clear error messages
6. `--results` and default `go test` mode are mutually exclusive

## Files to Modify

- `internal/cmd/verify.go` - Add `--results` flag, results file parsing
- `internal/cmd/verify_test.go` - Tests for results file mode
- `internal/cmd/from_go.go` - Extract shared logic (optional refactor)

## Test Strategy

- Unit tests: Parse valid/invalid results files
- Integration tests: Full verify pipeline with results file
- BDD: `features/verify/cross-language.feature`
