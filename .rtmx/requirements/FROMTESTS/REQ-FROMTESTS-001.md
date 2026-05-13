# REQ-FROMTESTS-001: from-tests --verify Flag for Combined Linking and Status Transition

## Metadata
- **Category**: FROMTESTS
- **Subcategory**: Workflow
- **Priority**: HIGH
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-GO-017|REQ-VERIFY-001
- **Blocks**: 
- **External ID**: aegis-cli

## Requirement

`rtmx from-tests --update --verify` shall combine test marker scanning,
database linking, and status verification into a single command, eliminating
the need to run `from-tests --update` followed by `verify --update` as
separate steps.

## Rationale

The current workflow for updating test linkage and verifying requirement
statuses requires two commands:

```bash
rtmx from-tests --update   # scan markers, link tests to requirements
rtmx verify --update        # run tests, match results, transition statuses
```

Since `from-tests` already knows which tests link to which requirements
via marker scanning, adding a `--verify` flag that also runs the test suite
and transitions statuses would cut the workflow in half. This is particularly
valuable in CI pipelines and for projects with large test suites where
running the full cycle is a common operation.

The aegis-cli team (rtmx-ai/aegis-cli) reported this as a workflow friction
point: every development cycle requires remembering and running both commands
in the correct order.

## Design

### New Flag

`--verify` (boolean, default false) on the `from-tests` command. When set
in combination with `--update`:

1. Scan source files for `// rtmx:req` markers (existing behavior).
2. Update `test_module` and `test_function` fields in the database (existing
   `--update` behavior).
3. Run the test suite using the configured test command (same as
   `verify --command` would use).
4. Match test results against requirements (same matching logic as `verify`).
5. Transition requirement statuses based on test results (same as
   `verify --update`).
6. Print combined output showing both linkage updates and status transitions.

### Configuration

The test command is determined by the same logic as `verify --command`:
- Explicit `--command` flag if provided.
- `test_command` from `rtmx.yaml` configuration.
- Auto-detected from project type (cargo test, go test, pytest, etc.).

### Error Handling

- If marker scanning succeeds but test execution fails, the linkage updates
  are still written but status transitions are skipped. An error message
  indicates which phase failed.
- `--verify` without `--update` is an error (verify needs the linkage to
  be current).

## Acceptance Criteria

1. `rtmx from-tests --update --verify` scans markers, updates linkage, runs
   tests, and transitions statuses in a single invocation.
2. Output includes both linkage changes and status transitions.
3. `--verify` without `--update` produces an error message.
4. If test execution fails, linkage updates are preserved but status
   transitions are skipped.
5. Exit code reflects the overall result (non-zero if tests fail or
   verification finds issues).
6. `--verify` respects `--command` for custom test commands.
7. JSON output (`--json`) includes both linkage and verification sections.

## Files to Create/Modify

- `internal/cmd/from_tests.go` -- Add `--verify` flag and orchestration
- `internal/cmd/from_tests_test.go` -- Tests for combined workflow
- `internal/cmd/verify.go` -- Extract reusable verification logic if needed
- `.rtmx/database.csv` -- Add this requirement

## Effort Estimate

1.5 weeks
