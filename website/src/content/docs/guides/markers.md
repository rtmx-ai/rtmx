---
title: Test Markers
description: Link pytest tests to requirements with RTMX markers
---

RTMX provides pytest markers to link tests directly to requirements, enabling automatic traceability tracking.

## Basic Usage

```python
import pytest

@pytest.mark.req("REQ-AUTH-001")
@pytest.mark.scope_unit
def test_oauth_login():
    """Validates REQ-AUTH-001: OAuth 2.0 authentication."""
    assert authenticate_user(token) == expected_user
```

## Available Markers

### Requirement Link

| Marker | Purpose |
|--------|---------|
| `@pytest.mark.req("REQ-XXX-NNN")` | Links test to requirement |

The requirement ID must match a row in your RTM database.

### Scope Markers

| Marker | Purpose |
|--------|---------|
| `@pytest.mark.scope_unit` | Unit test scope |
| `@pytest.mark.scope_integration` | Integration test scope |
| `@pytest.mark.scope_system` | System/E2E test scope |

### Technique Markers

| Marker | Purpose |
|--------|---------|
| `@pytest.mark.technique_nominal` | Nominal/happy path testing |
| `@pytest.mark.technique_parametric` | Parametric testing |
| `@pytest.mark.technique_monte_carlo` | Monte Carlo testing |
| `@pytest.mark.technique_stress` | Stress testing |

### Environment Markers

| Marker | Purpose |
|--------|---------|
| `@pytest.mark.env_simulation` | Simulation environment |
| `@pytest.mark.env_hil` | Hardware-in-the-loop |
| `@pytest.mark.env_anechoic` | Anechoic chamber |
| `@pytest.mark.env_field` | Field testing |

## Configuration

Enable markers in `rtmx.yaml`:

```yaml
rtmx:
  pytest:
    marker_prefix: "req"
    register_markers: true
```

Or in `pyproject.toml`:

```toml
[tool.pytest.ini_options]
markers = [
    "req(id): Link test to requirement",
    "scope_unit: Unit test scope",
    "scope_integration: Integration test scope",
    "scope_system: System test scope",
]
```

## Scanning Tests

Update your RTM database from test markers:

```bash
rtmx from-tests              # Scan and display results
rtmx from-tests --update     # Update RTM database
rtmx from-tests --dry-run    # Preview changes
```

## Best Practices

### One Test Per Requirement

Each requirement should have at least one test:

```python
@pytest.mark.req("REQ-AUTH-001")
def test_login_success():
    """Happy path test for REQ-AUTH-001."""
    pass

@pytest.mark.req("REQ-AUTH-001")
def test_login_invalid_token():
    """Error handling for REQ-AUTH-001."""
    pass
```

### Descriptive Docstrings

Include the requirement ID in docstrings for clarity:

```python
@pytest.mark.req("REQ-DATA-003")
def test_export_csv():
    """REQ-DATA-003: Export data as CSV.

    Validates that users can export their data in CSV format
    with proper encoding and column headers.
    """
    pass
```

### Multiple Requirements

A test can validate multiple requirements:

```python
@pytest.mark.req("REQ-AUTH-001")
@pytest.mark.req("REQ-AUDIT-002")
def test_login_audit():
    """Validates login and audit logging."""
    pass
```

## CI Integration

Check for unmarked tests in CI:

```bash
# Fail if tests don't have required markers
pytest --strict-markers

# Report marker coverage
rtmx from-tests --report
```
