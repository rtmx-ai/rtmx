# REQ-PAR-005: Parse Python-format conftest.py Markers

## Metadata
- **Category**: PARITY
- **Subcategory**: Config
- **Priority**: HIGH
- **Phase**: 14
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-008

## Requirement

Go CLI shall detect and parse Python conftest.py files that register rtmx markers (e.g., `config.addinivalue_line("markers", "req(id): ...")`) and extract `@pytest.mark.req()` decorators from fixture functions defined in conftest.py.

## Problem

The `from-tests` command only scanned `test_*.py` files, missing conftest.py files that:
1. Register rtmx markers via `pytest_configure` / `addinivalue_line`
2. Contain `@pytest.mark.req()` decorators on fixture functions

## Design

### Changes to `scanTestDirectory`
- Also scan `conftest.py` files during directory walks
- Extract `@pytest.mark.req()` markers from fixture functions (not just `test_*` functions)

### New Functions
- `extractConftestRegistrations()` - Parse `addinivalue_line("markers", ...)` patterns including multiline calls
- `scanConftestFiles()` - Walk directory tree finding and parsing all conftest.py files

### New Types
- `ConftestMarkerRegistration` - Represents a marker registration (name, args, help text, file, line)

### Behavioral Details
- In conftest.py, any function with `@pytest.mark.req()` is captured (not just `test_*`)
- In non-conftest test files, only `test_*` functions are captured; markers on helper functions are discarded
- Multiline `addinivalue_line` calls are accumulated up to 5 lines

## Acceptance Criteria

1. `scanTestDirectory` picks up conftest.py files
2. `extractConftestRegistrations` parses single-line marker registrations
3. `extractConftestRegistrations` parses multiline marker registrations
4. Both single-quoted and double-quoted strings are handled
5. Markers with args, without args, with help text, without help text all parse correctly
6. `@pytest.mark.req()` on fixture functions in conftest.py is detected
7. Non-conftest files still only match `test_*` function names

## Files Modified

- `internal/cmd/from_tests.go` - Added conftest.py parsing logic
- `internal/cmd/from_tests_test.go` - Added comprehensive test coverage
