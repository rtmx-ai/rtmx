# REQ-VERIFY-006: Verify Shall Warn on Unmatched Test References by Default

## Metadata
- **Category**: VERIFY
- **Subcategory**: Audit
- **Priority**: MEDIUM
- **Phase**: 24
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-005

## Requirement

`rtmx verify` (without `--audit`) shall print a summary warning line
when unmatched test references are detected, prompting the user to run
`--audit` for details. This ensures false negatives are surfaced during
normal verification without requiring the user to remember the audit flag.

## Rationale

The `--audit` flag (REQ-VERIFY-005) provides detailed diagnostics, but
the silent-skip problem is dangerous precisely because users do not know
to look for it. A single summary line in normal verify output -- e.g.,
"Warning: 4 requirements have test references that did not match any
test result. Run with --audit for details." -- makes the gap visible
without cluttering the standard output.

## Acceptance Criteria

1. `rtmx verify` prints a warning line when unmatched references exist
2. Warning includes count of unmatched references
3. Warning suggests `--audit` for details
4. Warning does not appear when all references match
5. Warning does not change exit code

## Files to Create/Modify

- `internal/cmd/verify.go` -- Add summary warning after main verify pass
- `internal/cmd/verify_test.go` -- Tests for warning presence/absence

## Effort Estimate

0.25 weeks
