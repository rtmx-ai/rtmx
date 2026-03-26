# REQ-LANG-003: Go Testing Integration

## Metadata
- **Category**: LANG
- **Subcategory**: Go
- **Priority**: P0
- **Phase**: 14
- **Status**: COMPLETE
- **Completed**: 2026-02-17
- **Dependencies**: REQ-GO-018, REQ-LANG-007

## Requirement

RTMX shall provide Go testing integration with helper functions and struct tags as primary marker mechanisms.

## Rationale

Go is the implementation language for the RTMX CLI. Native Go integration demonstrates the marker pattern and serves as the reference implementation for other language bindings.

## Design

### Installation

```go
import "github.com/rtmx-ai/rtmx/pkg/rtmx"
```

### Helper Function

```go
func TestLogin(t *testing.T) {
    rtmx.Req(t, "REQ-AUTH-001")
    // test implementation
}

// With options
func TestLogin(t *testing.T) {
    rtmx.Req(t, "REQ-AUTH-001",
        rtmx.Scope("integration"),
        rtmx.Technique("nominal"),
        rtmx.Env("simulation"),
    )
    // test implementation
}
```

### Struct Tags (Table-Driven Tests)

```go
func TestCalculations(t *testing.T) {
    tests := []struct {
        name     string
        rtmx     string `rtmx:"REQ-MATH-001"`
        input    int
        expected int
    }{
        {"positive", "", 5, 25},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rtmx.FromTag(t, tt)
            // test implementation
        })
    }
}
```

### TestMain Integration

```go
func TestMain(m *testing.M) {
    code := m.Run()
    rtmx.WriteResultsJSON("rtmx-results.json")
    os.Exit(code)
}
```

## Implementation

Located in `pkg/rtmx/rtmx.go`:
- `Req(t, reqID, opts...)` - Register marker
- `FromTag(t, testCase)` - Extract from struct tag
- `Scope(s)`, `Technique(t)`, `Env(e)` - Options
- `Results()` - Get all results
- `WriteResultsJSON(path)` - Write to file

## Acceptance Criteria

1. `rtmx.Req()` registers markers for tests
2. `rtmx.FromTag()` extracts markers from struct tags
3. Test results include pass/fail status
4. JSON output compatible with `rtmx verify`
5. Works with `go test -json`

## Test Strategy

- Unit tests in `pkg/rtmx/rtmx_test.go`
- Integration with `rtmx verify`

## References

- Implementation: `pkg/rtmx/rtmx.go`
- Tests: `pkg/rtmx/rtmx_test.go`
- REQ-LANG-007 marker specification
