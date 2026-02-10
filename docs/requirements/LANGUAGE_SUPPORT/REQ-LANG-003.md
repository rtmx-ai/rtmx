# REQ-LANG-003: Go Testing Integration

## Status: MISSING
## Priority: MEDIUM
## Phase: 14

## Description
System shall provide Go testing integration that enables requirement traceability through **helper functions and struct tags as the primary mechanisms**, with comment markers as a deprecated fallback for existing codebases.

## Marker Strategy: Native Go Idioms First

Go lacks annotation/attribute syntax like Java or Rust. The idiomatic Go approach uses:

**Primary (Recommended)**:
1. Helper functions: `rtmx.Req(t, "REQ-XXX-NNN")` - explicit, type-safe, runtime validated
2. Struct tags: `rtmx:"REQ-XXX-NNN"` - for table-driven tests

**Secondary (Deprecated Fallback)**: Comment markers for legacy codebases.

## Acceptance Criteria

### Native Go Integration (PRIMARY)
- [ ] Helper function API for runtime marker registration:
  ```go
  import "github.com/rtmx-ai/rtmx-go/rtmx"

  func TestLoginSuccess(t *testing.T) {
      rtmx.Req(t, "REQ-AUTH-001",
          rtmx.Scope("integration"),
          rtmx.Technique("nominal"),
          rtmx.Env("simulation"),
      )

      // Test implementation
  }
  ```
- [ ] Table-driven test struct tags:
  ```go
  func TestCalculations(t *testing.T) {
      tests := []struct {
          name     string
          rtmx     string `rtmx:"REQ-MATH-001"`
          input    int
          expected int
      }{
          {"positive", "", 5, 25},
          {"zero", "", 0, 0},
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              rtmx.FromTag(t, tt)
              // Test implementation
          })
      }
  }
  ```
- [ ] Subtest support with `t.Run()` and requirement inheritance
- [ ] Compile-time validation via `go vet` analyzer (optional)
- [ ] Runtime validation ensures requirement ID format matches `^REQ-[A-Z]+-[0-9]+$`

### Companion Module Distribution
- [ ] Go module `github.com/rtmx-ai/rtmx-go` published and importable
- [ ] Sub-packages: `rtmx`, `rtmx/reporter`, `rtmx/analyzer`
- [ ] Zero CGO dependencies for maximum portability
- [ ] Go 1.21+ module compatibility

### Test Output Integration
- [ ] `rtmx-go` CLI tool wraps `go test -json` and enriches output with requirement data
- [ ] JSON output compatible with RTMX marker specification (REQ-LANG-007)
- [ ] JSON Lines format for streaming large test suites
- [ ] `rtmx from-go <results.json>` imports Go test results into RTM database

### Testify Integration (Optional)
- [ ] Support for `testify/suite.Suite` based tests
- [ ] `suite.SetupTest()` integration for requirement registration
- [ ] Inherited requirements from suite-level `Req()` calls

### Legacy Fallback (DEPRECATED)
- [ ] Comment marker format: `// rtmx:req=REQ-XXX-NNN` parsed before test functions
- [ ] Deprecation warnings emitted when comment markers detected
- [ ] Migration tool: `rtmx migrate-markers --from=comments --to=helper` rewrites tests

## Technical Notes

### Why Helper Functions Over Comments

| Aspect | Helper Function | Comments |
|--------|-----------------|----------|
| Type safety | Compile-time validation | None |
| IDE support | go-to-definition, autocomplete | None |
| Runtime validation | Validates at test execution | Requires separate parsing |
| Refactoring | Works with `gorename` | Breaks on refactors |
| Discoverability | Explicit in test body | Hidden in comments |
| Test context | Access to `*testing.T` | No test context |

### Helper Function Implementation

```go
// github.com/rtmx-ai/rtmx-go/rtmx/req.go
package rtmx

import (
    "regexp"
    "testing"
)

var reqIDPattern = regexp.MustCompile(`^REQ-[A-Z]+-[0-9]+$`)

type Option func(*marker)

func Scope(s string) Option   { return func(m *marker) { m.scope = s } }
func Technique(t string) Option { return func(m *marker) { m.technique = t } }
func Env(e string) Option     { return func(m *marker) { m.env = e } }

// Req registers a requirement marker for the current test
func Req(t testing.TB, reqID string, opts ...Option) {
    t.Helper()
    if !reqIDPattern.MatchString(reqID) {
        t.Fatalf("invalid requirement ID format: %s", reqID)
    }
    m := &marker{reqID: reqID}
    for _, opt := range opts {
        opt(m)
    }
    register(t, m)
}
```

### Struct Tag Processing

```go
// FromTag extracts requirement from struct tag and registers it
func FromTag(t testing.TB, testCase interface{}) {
    t.Helper()
    v := reflect.ValueOf(testCase)
    if v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    typ := v.Type()
    for i := 0; i < typ.NumField(); i++ {
        field := typ.Field(i)
        if tag := field.Tag.Get("rtmx"); tag != "" {
            Req(t, tag)
            return
        }
    }
}
```

### go vet Analyzer (Optional)

```go
// analyzer/analyzer.go - validates rtmx.Req calls at compile time
package analyzer

import (
    "go/ast"
    "golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
    Name: "rtmxcheck",
    Doc:  "checks rtmx.Req calls have valid requirement IDs",
    Run:  run,
}
```

### Migration from Comments

```bash
# Analyze current comment-based markers
rtmx markers discover --format json > current-markers.json

# Generate migration script
rtmx migrate-markers --lang=go --from=comments --to=helper --dry-run

# Apply migration
rtmx migrate-markers --lang=go --from=comments --to=helper
```

## Test Cases
1. `tests/test_lang_go.py::test_helper_function_extraction` - Parse rtmx.Req() calls
2. `tests/test_lang_go.py::test_struct_tag_extraction` - Parse struct tag markers
3. `tests/test_lang_go.py::test_extended_marker_attributes` - Parse scope/technique/env options
4. `tests/test_lang_go.py::test_subtest_marker_inheritance` - t.Run() inheritance
5. `tests/test_lang_go.py::test_table_driven_test_support` - Table test extraction
6. `tests/test_lang_go.py::test_go_test_json_enrichment` - JSON output enrichment
7. `tests/test_lang_go.py::test_runtime_validation` - Invalid req ID fails test
8. `tests/test_lang_go.py::test_testify_suite_support` - Testify compatibility
9. `tests/test_lang_go.py::test_from_go_import` - CLI import command
10. `tests/test_lang_go.py::test_comment_deprecation_warning` - Deprecation for comments

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec
- REQ-DIST-002: Standalone binary CLI (Go CLI shares codebase)

## Blocks
- None

## Effort
2.5 weeks
