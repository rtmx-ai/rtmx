# REQ-TEST-006: Fix Known Test Bugs

## Status: NOT_STARTED
## Priority: HIGH
## Phase: 4
## Effort: 0.5 weeks

## Description

Fix 6 `UnboundLocalError` bugs that cause tests to be skipped when files don't exist.

## Acceptance Criteria

- [ ] `test_diff_baseline_not_found` passes (both variants in test_cli_additional.py)
- [ ] `test_from_tests_csv_not_found` passes
- [ ] `test_sync_csv_not_found` passes
- [ ] `test_integrate_csv_not_found` passes
- [ ] `test_health_csv_not_found` passes
- [ ] All 6 previously skipped tests are unskipped and passing
- [ ] Error messages are user-friendly (not stack traces)

## Root Cause

Variable used before assignment in conditional error handling paths:

```python
# Before (buggy):
def run_command():
    if condition:
        result = do_something()
    # result used here but may be undefined if condition was False
    return result
```

## Implementation

```python
# After (fixed):
def run_command():
    result = None  # Initialize before conditional
    if condition:
        result = do_something()
    if result is None:
        raise click.ClickException("File not found: ...")
    return result
```

## Files to Modify

- `src/rtmx/cli/diff.py`
- `src/rtmx/cli/sync.py`
- `src/rtmx/cli/integrate.py`
- `src/rtmx/cli/health.py`
- `src/rtmx/cli/from_tests.py`

## Test Cases

- `tests/test_cli_commands.py::test_from_tests_csv_not_found`
- `tests/test_cli_commands.py::test_sync_csv_not_found`
- `tests/test_cli_commands.py::test_integrate_csv_not_found`
- `tests/test_cli_commands.py::test_health_csv_not_found`
- `tests/test_cli_additional.py::test_diff_baseline_not_found` (2 variants)

## Notes

These are quick wins that improve test reliability and unblock proper E2E testing of error conditions.
