# REQ-INT-004: Correct Stale Test References in RTM Database

## Metadata
- **Category**: INTEGRITY
- **Subcategory**: DataQuality
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-VERIFY-005

## Requirement

All `test_module` and `test_function` fields in the RTM database shall
reference files and functions that exist in the codebase. Stale
references that point to non-existent files or function names shall be
corrected to point to the actual test locations, or cleared if no
corresponding test exists.

## Rationale

An audit of all 214 requirements found 6 with stale test references
that prevent `rtmx verify` from discovering passing tests:

| Req ID | Problem | Correct Reference |
|--------|---------|-------------------|
| REQ-PLAN-014 | test_module `internal/cmd/release_test.go` does not exist; test_function `TestReleaseGateVersionPolicy` does not exist | Config tests exist at `internal/config/config_test.go` (`TestVersionPolicyIncrementLevel` et al.) but no gate integration test exists -- test_function should be cleared until gate test is written |
| REQ-PLAN-006 | test_module `internal/cmd/release_test.go` does not exist; `TestReleaseScope` does not exist | Genuinely MISSING -- clear aspirational references |
| REQ-PLAN-012 | test_module `internal/cmd/release_test.go` does not exist; `TestReleaseForecast` does not exist | Genuinely MISSING -- clear aspirational references |
| REQ-BENCH-008 | test_function `TestSyncConfig` does not exist in `internal/benchmark/config_test.go` | Likely meant `TestParseBenchmarkConfig` or similar -- verify and correct |
| REQ-BENCH-010 | Feature implemented (ERR trap in scripts) but empty test_function prevents verify match | Script test -- add validation method note |
| REQ-BENCH-016 | test_function contains a file path instead of a function name (column alignment error) | Fix column alignment |

Aspirational references (pointing to tests that should be written but
do not yet exist) cause `rtmx verify` to silently skip the requirement,
making the database claim a status that cannot be verified. The correct
practice is: if the test does not exist yet, leave test_function empty
and set status to MISSING. When the test is written, populate
test_function so verify can close the loop.

## Acceptance Criteria

1. Every `test_module` in the database references a file that exists, or is empty
2. Every `test_function` in the database references a function that exists in the
   corresponding test_module file, or is empty
3. REQ-PLAN-014 test references corrected or cleared
4. REQ-PLAN-006, REQ-PLAN-012 aspirational references cleared
5. REQ-BENCH-008 test_function corrected to actual function name
6. REQ-BENCH-010 validation method updated
7. REQ-BENCH-016 column alignment fixed
8. `rtmx verify --update` run after corrections with no false status changes

## Files to Create/Modify

- `.rtmx/database.csv` -- Correct stale fields

## Effort Estimate

0.25 weeks
