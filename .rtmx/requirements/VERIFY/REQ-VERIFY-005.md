# REQ-VERIFY-005: Verify Audit Diagnostics for Unmatched Test References

## Metadata
- **Category**: VERIFY
- **Subcategory**: Audit
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-001
- **Blocks**: REQ-VERIFY-006

## Requirement

`rtmx verify --audit` shall report diagnostic warnings for requirements
whose `test_function` or `test_module` fields reference tests that do not
exist, enabling detection of false negatives (implemented features stuck
at MISSING due to stale references) and false positives (COMPLETE
requirements whose tests have been renamed or deleted).

## Rationale

The current verify command silently skips requirements when
`test_function` does not match any test result (`verify.go:522-524`).
The `test_module` field is never validated against the filesystem.
This silent-skip behavior caused REQ-PLAN-014 to remain MISSING despite
having a full implementation, because the database referenced a
non-existent test function name in a non-existent file. An audit of all
214 requirements found 6 with stale references pointing to wrong files
or function names, and 17 with empty test_function fields that cannot
participate in verification at all.

Without diagnostics, the only way to discover these gaps is manual
cross-referencing of every requirement against the codebase -- exactly
the work that an RTM tool should automate.

## Design

### New Flag

`rtmx verify --audit` runs the standard verification pass, then performs
three additional checks and appends an audit section to the output:

1. **Unmatched references** -- Requirements where `test_function` is set
   but no test result matched. These are potential false negatives:
   either the test was renamed, or the database reference is aspirational.

2. **Stale test_module paths** -- Requirements where `test_module` is set
   but the file does not exist on disk. The test_module field is not used
   for matching, but stale paths indicate the database has drifted from
   the codebase.

3. **Unverified COMPLETE** -- Requirements with status COMPLETE that had
   no test match in this verification run. These are potential false
   positives: the requirement may have been manually marked COMPLETE, or
   its test was removed.

4. **Empty test references** -- Requirements with both `test_module` and
   `test_function` empty. These cannot participate in closed-loop
   verification.

### Output Format

```
=================== Audit Diagnostics ===================

  Unmatched test references (4):
    REQ-PLAN-014  test_function=TestReleaseGateVersionPolicy  (file missing: internal/cmd/release_test.go)
    REQ-PLAN-006  test_function=TestReleaseScope               (file missing: internal/cmd/release_test.go)
    REQ-PLAN-012  test_function=TestReleaseForecast             (file missing: internal/cmd/release_test.go)
    REQ-BENCH-008 test_function=TestSyncConfig                  (file exists, function not found)

  Stale test_module paths (3):
    REQ-PLAN-014  internal/cmd/release_test.go
    REQ-PLAN-006  internal/cmd/release_test.go
    REQ-PLAN-012  internal/cmd/release_test.go

  Unverified COMPLETE requirements (12):
    REQ-GO-003    test_function=TestVersionCommand              (no test result)
    ...

  Empty test references (17):
    REQ-BENCH-011 (no test_function set)
    ...

  Summary: 4 unmatched, 3 stale paths, 12 unverified, 17 empty
```

### JSON Output

`rtmx verify --audit --json` includes an `audit` object:

```json
{
  "audit": {
    "unmatched_references": [...],
    "stale_test_modules": [...],
    "unverified_complete": [...],
    "empty_references": [...],
    "summary": {
      "unmatched": 4,
      "stale_paths": 3,
      "unverified_complete": 12,
      "empty": 17
    }
  }
}
```

### Exit Code

`--audit` does not change the exit code by default. The existing
`--fail-under` flag controls exit code based on completion percentage.
A future requirement may add `--fail-on-audit` to fail when audit
findings exceed a threshold.

### Implementation Notes

- File existence checks use the injected `FileSystem` interface, not
  direct `os.Stat`, to maintain testability.
- Function existence checks use `go/parser` to parse the test file AST
  and look for the function name, or fall back to a simple grep if the
  file is not Go source.
- The audit runs after the main verification pass so it has access to
  both the test results and the database state.

## Acceptance Criteria

1. `rtmx verify --audit` reports requirements with unmatched test_function
2. `rtmx verify --audit` reports requirements with non-existent test_module files
3. `rtmx verify --audit` reports COMPLETE requirements with no test match
4. `rtmx verify --audit` reports requirements with empty test references
5. `rtmx verify --audit --json` includes audit object in JSON output
6. `--audit` does not change exit code behavior
7. `--audit` combined with `--update` runs both audit and status updates
8. Audit uses FileSystem interface for file existence checks (testable)
9. Output groups findings by category with counts and summary line

## Files to Create/Modify

- `internal/cmd/verify.go` -- Add audit logic after main verify pass
- `internal/cmd/verify_test.go` -- Table-driven tests for audit diagnostics
- `.rtmx/database.csv` -- Add this requirement

## Effort Estimate

1.0 week
