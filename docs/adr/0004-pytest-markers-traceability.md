# ADR-0004: Pytest Markers for Requirements Traceability

## Status

Accepted

## Context

Requirements traceability demands a bidirectional link between requirements and tests. We needed a mechanism to:

1. Link test functions to specific requirements
2. Query which tests cover a requirement
3. Report test results per requirement
4. Categorize tests by scope and technique

## Decision

We use **pytest markers** to annotate test functions with requirement metadata.

## Rationale

### Native Pytest Integration
- Uses pytest's built-in marker system
- No custom test runners needed
- Works with all pytest plugins (coverage, parallel, etc.)

### Queryable
- `pytest -m "req" --collect-only` lists all marked tests
- Filter tests by requirement: `pytest -m "REQ-SW-001"`
- Filter by scope: `pytest -m scope_unit`

### Self-Documenting
- Markers visible in test code
- IDE support for marker navigation
- Generates traceability reports automatically

### Multi-Dimensional Classification
```python
@pytest.mark.req("REQ-SW-001")           # Links to requirement
@pytest.mark.scope_unit                   # Unit/Integration/System
@pytest.mark.technique_nominal            # Nominal/Boundary/Stress
@pytest.mark.env_simulation               # Test environment
def test_feature():
    pass
```

## Implementation

### Marker Registration (pyproject.toml)
```toml
[tool.pytest.ini_options]
markers = [
    "req(id): Links test to requirement ID",
    "scope_unit: Unit test scope",
    "scope_integration: Integration test scope",
    "scope_system: System/E2E test scope",
    "technique_nominal: Nominal value testing",
    "technique_monte_carlo: Randomized testing",
]
```

### Pytest Plugin
```python
# rtmx/pytest/plugin.py
def pytest_report_teststatus(report, config):
    """Track test results by requirement."""
    for marker in report.item.iter_markers("req"):
        req_id = marker.args[0]
        # Record result for requirement
```

## Consequences

### Positive
- Zero friction for developers (just add decorators)
- Standard pytest commands work
- Enables CI/CD traceability gates
- Coverage visible in pytest output

### Negative
- Markers must be maintained manually
- Requires discipline to add markers
- No automatic requirement-to-test linking

### Mitigations
- CI enforces marker presence (80% minimum)
- `rtmx from-tests` discovers unmarked tests
- Pre-commit hooks check marker format

## References

- [Pytest Markers Documentation](https://docs.pytest.org/en/latest/how-to/mark.html)
- [DO-178C Traceability Requirements](https://en.wikipedia.org/wiki/DO-178C)
