# REQ-LANG-004: Python pytest Integration

## Metadata
- **Category**: LANG
- **Subcategory**: Python
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007

## Requirement

RTMX shall provide Python testing integration via a pytest plugin that enables requirement markers on test functions, maintaining full compatibility with the original rtmx Python package.

## Rationale

Python/pytest is a primary target for RTMX due to:
1. Parity with original rtmx package (pytest plugin)
2. Large Python testing ecosystem
3. Data science and ML workflows often use Python

## Design

### Installation

```bash
# From PyPI (wrapper that ensures Go CLI is installed)
pip install rtmx

# The package provides:
# 1. pytest plugin for markers
# 2. Automatic installation of rtmx Go binary
```

### Marker Syntax

```python
import pytest

@pytest.mark.req("REQ-AUTH-001")
def test_login_success():
    """Test successful login."""
    pass

@pytest.mark.req("REQ-AUTH-002", scope="integration", technique="boundary")
def test_login_invalid_password():
    """Test login with invalid password."""
    pass

# Multiple requirements
@pytest.mark.req("REQ-AUTH-001")
@pytest.mark.req("REQ-AUDIT-001")
def test_login_audited():
    """Test that login is audited."""
    pass
```

### Marker Registration

```python
# conftest.py (auto-generated or user-provided)
def pytest_configure(config):
    config.addinivalue_line(
        "markers", "req(id, scope=None, technique=None, env=None): Link test to requirement"
    )
```

### Output Integration

```bash
# Run tests with RTMX output
pytest --rtmx-output=rtmx-results.json

# Or use rtmx verify directly
rtmx verify --command "pytest -v"
```

### Parity with Original rtmx

The Python package must support all features from original rtmx:

| Feature | Original rtmx | Go CLI + Python Plugin |
|---------|---------------|------------------------|
| `@pytest.mark.req()` | Yes | Yes |
| `@pytest.mark.scope_*` | Yes | Yes (via scope param) |
| `@pytest.mark.technique_*` | Yes | Yes (via technique param) |
| `@pytest.mark.env_*` | Yes | Yes (via env param) |
| `rtmx from-tests` | Yes | Yes (Go CLI) |
| `rtmx verify` | No | Yes (Go CLI) |
| Marker registration | Yes | Yes |
| JSON output | Yes | Yes |

## Acceptance Criteria

1. `pip install rtmx` installs pytest plugin
2. `@pytest.mark.req("REQ-XXX")` works on test functions
3. `pytest --rtmx-output=results.json` produces compatible JSON
4. `rtmx verify --command "pytest"` correctly updates status
5. All original rtmx pytest features are supported
6. Plugin auto-registers markers (no manual conftest.py needed)

## Test Strategy

- Unit tests for marker parsing
- Integration tests with pytest
- Parity tests comparing original rtmx vs new plugin output

## Package Structure

```
rtmx/                    # PyPI package
├── __init__.py          # Version, ensure CLI installed
├── pytest/
│   ├── __init__.py
│   └── plugin.py        # pytest plugin implementation
├── markers/
│   └── __init__.py      # Marker utilities
└── cli/
    └── __init__.py      # CLI wrapper (calls Go binary)
```

## References

- Original rtmx pytest plugin: rtmx-ai/rtmx
- pytest marker documentation
- REQ-LANG-007 marker specification
