# REQ-LANG-007: Language-Agnostic Marker Specification

## Metadata
- **Category**: LANG
- **Subcategory**: Specification
- **Priority**: HIGH
- **Phase**: 14
- **Status**: COMPLETE
- **Completed**: 2026-02-10

## Requirement

RTMX shall define a language-agnostic marker annotation specification that enables consistent requirement traceability across all supported programming languages.

## Rationale

Like Cucumber's Gherkin specification provides a common language for BDD across implementations (Cucumber-JVM, Cucumber.js, Behave, Godog), RTMX needs a common specification for requirement markers that can be implemented idiomatically in each language while maintaining semantic compatibility.

## Specification

### Marker Semantics

Every language implementation must support these core marker attributes:

| Attribute | Required | Description | Example |
|-----------|----------|-------------|---------|
| `req_id` | Yes | Requirement identifier | `REQ-AUTH-001` |
| `scope` | No | Test scope | `unit`, `integration`, `system`, `e2e` |
| `technique` | No | Test technique | `nominal`, `boundary`, `error`, `stress` |
| `env` | No | Test environment | `simulation`, `hil`, `field` |

### Requirement ID Format

```regex
^REQ-[A-Z]+-[0-9]+$
```

Examples: `REQ-AUTH-001`, `REQ-GO-042`, `REQ-INT-003`

### Output Format (JSON)

All language implementations must produce compatible JSON output:

```json
{
  "results": [
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
      "timestamp": "2026-02-20T18:45:00Z"
    }
  ]
}
```

### Language Idioms

Each implementation should use language-native patterns:

| Language | Marker Style | Example |
|----------|--------------|---------|
| Go | Helper function | `rtmx.Req(t, "REQ-001")` |
| Python | pytest marker | `@pytest.mark.req("REQ-001")` |
| Rust | Attribute macro | `#[rtmx::req("REQ-001")]` |
| JavaScript | Decorator/helper | `rtmx.req("REQ-001")` |
| Java | Annotation | `@Req("REQ-001")` |

## Acceptance Criteria

1. JSON Schema defined for marker output format
2. Specification document published
3. At least two language implementations demonstrate compatibility
4. `rtmx verify` consumes output from all implementations identically

## References

- Cucumber Gherkin specification
- JUnit 5 annotation model
- pytest marker system
