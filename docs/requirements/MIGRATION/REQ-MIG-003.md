# REQ-MIG-003: Legacy Python Package Deprecation

## Status: MISSING
## Priority: HIGH
## Phase: 21

## Description

The Python `rtmx` PyPI package shall be formally deprecated and reduced to a minimal pytest plugin. The full Python CLI shall be removed from the package. Users installing `pip install rtmx` shall receive the marker-only pytest plugin and a clear deprecation notice directing them to the Go binary for CLI functionality.

## Rationale

After the main branch migrates to Go, the Python package must not remain as a stale, unmaintained full CLI. Reducing it to the minimal pytest plugin prevents user confusion, avoids security liability from unmaintained code, and ensures the PyPI package serves its ongoing purpose (test marker registration and result capture).

## Acceptance Criteria

- [ ] PyPI `rtmx` package contains only: `rtmx.pytest.plugin`, `rtmx.pytest.reporter`, `rtmx.markers`
- [ ] Package metadata includes deprecation notice and link to Go CLI install instructions
- [ ] `pip install rtmx` emits post-install message directing users to Go binary
- [ ] Version bumped to indicate breaking change (e.g., v2.0.0)
- [ ] Old CLI entry points removed from pyproject.toml
- [ ] README updated with migration instructions
- [ ] PyPI project description updated with deprecation banner
- [ ] Final Python CLI version archived on PyPI (not yanked, for reproducibility)

## Test Cases

1. `tests/test_deprecation_pkg.py::test_minimal_package_contents` - Only pytest plugin modules present
2. `tests/test_deprecation_pkg.py::test_no_cli_entrypoint` - No `rtmx` CLI command installed
3. `tests/test_deprecation_pkg.py::test_pytest_markers_functional` - Markers still register correctly
4. `tests/test_deprecation_pkg.py::test_json_reporter_functional` - JSON result output works
5. `tests/test_deprecation_pkg.py::test_deprecation_metadata` - Package metadata includes notice

## Dependencies

- REQ-MIG-002: Main branch migration (Go CLI must be the primary before deprecating Python)
- REQ-PYTEST-001: Minimal pytest plugin implementation

## Blocks

None

## Effort

1.5 weeks
