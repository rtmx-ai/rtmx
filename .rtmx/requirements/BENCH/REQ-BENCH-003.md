# REQ-BENCH-003: Python Language Benchmark (psf/requests)

## Metadata
- **Category**: BENCH
- **Subcategory**: Python
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-004
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the Python scanner against the `requests` library, confirming pytest marker extraction and test output parsing on a production Python project.

## Rationale

Python was the original RTMX implementation language. The `requests` library uses pytest extensively with fixtures, parametrize, and class-based tests -- patterns the Python scanner must handle correctly in the wild.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [psf/requests](https://github.com/psf/requests) |
| Pinned ref | v2.32.0 (or latest stable at implementation time) |
| License | Apache-2.0 |
| Test count | ~500 |
| Test framework | pytest |
| Build time | ~1 min |

## Design

### Marker Patch

Add `@pytest.mark.req("REQ-BENCH-PY-NNN")` decorators to a representative sample:
- `tests/test_requests.py` -- core HTTP tests (~10 tests)
- `tests/test_utils.py` -- utility functions (~5 tests)
- `tests/test_structures.py` -- data structures (~5 tests)
- `tests/test_hooks.py` -- hook system (~5 tests)

Minimum 20 markers across at least 3 test modules.

### Marker Patch also adds conftest.py registration

```python
# conftest.py addition
import pytest
def pytest_configure(config):
    config.addinivalue_line("markers", "req(id): RTMX requirement marker")
```

### Benchmark Config

```yaml
language: python
exemplar:
  repo: psf/requests
  ref: v2.32.0
  license: Apache-2.0
clone_depth: 1
setup_commands:
  - python -m pip install -e ".[dev]"
marker_patch: patches/python/requests.patch
expected_markers: 20
scan_command: rtmx from-tests --format json .
verify_command: pytest --tb=short -q
timeout_minutes: 5
```

### Validation Checks

1. `rtmx from-tests` extracts >= 20 markers from patched source
2. Markers span >= 3 test modules
3. `pytest` runs successfully on patched source
4. `rtmx verify --command "pytest --tb=short"` parses output correctly
5. No marker insertion side effects (all patched tests pass)

## Acceptance Criteria

1. `benchmarks/configs/python.yaml` exists with valid config
2. `benchmarks/patches/python/requests.patch` applies cleanly to pinned ref
3. `make -C benchmarks run LANG=python` completes successfully
4. Extracted marker count matches or exceeds baseline
5. Verify output maps all markers to correct status

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=python` in CI
- Baseline stored in `benchmarks/results/baselines/python.json`
