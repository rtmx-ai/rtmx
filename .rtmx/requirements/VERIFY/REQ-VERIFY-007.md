# REQ-VERIFY-007: Verify Test Name Matching for Rust Module Paths

## Metadata
- **Category**: VERIFY
- **Subcategory**: Matching
- **Priority**: P0
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-001|REQ-LANG-005
- **Blocks**: 
- **External ID**: aegis-cli

## Requirement

`rtmx verify --update` shall correctly match test results from `cargo test`
output that include Rust module path prefixes against `test_function` values
stored in the CSV database, even when the database omits the module prefix.

## Rationale

The aegis-cli project (rtmx-ai/aegis-cli) has 479 requirements and all tests
pass, but `verify --update` only matches 142 of them. The root cause is a
mismatch between how `cargo test` reports test names and how the CSV database
stores them:

- `cargo test` output: `embedding::tests::test_file_chunker_overlap`
- CSV `test_function`: `tests::test_file_chunker_overlap`

The module prefix (`embedding::`) is present in the test output but absent
from the database. The current verify matching logic requires an exact string
match between the test result function name and the database `test_function`
field, so 337 requirements show "test references that did not match any test
result" despite having passing tests.

The project reports 94.6% verified but is likely 97%+ in reality -- the gap
is entirely due to test name matching failures.

## Design

### Option A: Suffix Matching (Recommended)

When matching test results against database entries, treat a match as valid
if the test result function name ends with the database `test_function` value
(after splitting on `::` or `.` path separators). This handles Rust module
paths, Python package paths, and similar hierarchical naming.

Specifically:
1. If exact match: accept.
2. If test result name ends with `::` + database value: accept.
3. If test result name ends with `.` + database value: accept.

Guard against false positives by requiring the match to occur at a path
separator boundary (not a substring of an identifier).

### Option B: Marker-Authoritative Matching

Use the `// rtmx:req` marker linkage established by `from-tests` as the
authoritative source for test-to-requirement mapping. The markers already
establish which test covers which requirement. Verify should trust that
linkage rather than re-deriving it from test output parsing.

This would mean: if a requirement has a `test_function` value that was
populated by `from-tests` (via marker scanning), and the test name appears
in the test output (possibly with a module prefix), the match is valid.

### Recommended Approach

Implement Option A (suffix matching) as it is simpler, backward-compatible,
and handles the immediate problem without requiring changes to the from-tests
workflow. Option B can be pursued as a future enhancement.

### Edge Cases

- Multiple database entries could suffix-match the same test result. This is
  not a problem: each requirement independently tracks its own test reference.
- A short `test_function` like `test_add` could match `math::test_add` and
  `string::test_add`. The boundary-separator guard mitigates this, and in
  practice database entries include enough context (e.g., `tests::test_add`)
  to be unique.

## Acceptance Criteria

1. `cargo test` output with module-prefixed names matches database entries
   that store only the suffix portion of the test name.
2. Exact matches continue to work as before (no regression).
3. Suffix matching respects path separator boundaries (`::` and `.`).
4. No false positives from substring matches within identifiers.
5. `verify --update` with Rust test output updates status for all matching
   requirements.
6. Matching logic works for both `--command` and `--results` modes.

## Files to Create/Modify

- `internal/cmd/verify.go` -- Update test name matching logic
- `internal/cmd/verify_test.go` -- Table-driven tests for suffix matching
- `.rtmx/database.csv` -- Add this requirement

## Effort Estimate

1.0 week
