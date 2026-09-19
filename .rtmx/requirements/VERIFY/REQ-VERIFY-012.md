# REQ-VERIFY-012: Skipped results are non-evidence on the --results path

## Metadata
- **Category**: VERIFY
- **Subcategory**: COMPLETENESS
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-004
- **Blocks**: REQ-VERIFY-013
- **External ID**:

## Requirement

`rtmx verify --results` shall treat a skipped test result as non-evidence: it
shall neither promote nor downgrade a requirement's status, and shall not count
toward failing tests for the command's exit status. A result decodes as skipped
when it carries `"status": "skip"`/`"skipped"`, an explicit `"skipped": true`,
or supplies neither `passed` nor `status` at all (an absent outcome is treated
as non-evidence, not a silent failure). This mirrors the native go-test path
(`determineNewStatus`) and the `from-pytest` scanner, which already omit skips.

## Rationale

The cross-language results decoder had no representation for a skipped test: the
`Result` type exposed only `Passed bool`, and the decoder folded `status:"skip"`
— and any record that supplied neither `passed` nor `status` — into
`Passed=false`. The `--results` status engine then counted every non-passing
record as a failure, so under `require_all_pass` (the default) a single skipped
test **downgraded a COMPLETE requirement to PARTIAL**, and the shared count loop
put the skip in `TestsFailed`, making `rtmx verify --results` exit non-zero even
without `--update`.

This is asymmetric with the native go-test path, which keeps current status on a
skip. Sharded or matrix CI (e.g. a config-variant that skips a test in one leg)
routinely produces such results, so the bug silently corrupted status on every
run. A skipped test conveys no evidence and must move nothing.

## Acceptance Criteria

- [ ] A COMPLETE requirement with `>= min_combinations` passing combos plus one
      skipped result stays COMPLETE (combinations policy).
- [ ] A COMPLETE requirement with a passing test plus a skipped result stays
      COMPLETE (simple policy).
- [ ] A skipped result does not increment `TestsFailed` and does not cause a
      non-zero verify exit.
- [ ] A results record with `status:"skip"`/`"skipped"`, `"skipped":true`, or no
      outcome field at all decodes as skipped (non-evidence), not as a failure.
- [ ] A genuine failing result still downgrades COMPLETE→PARTIAL (regression guard).
- [ ] Verified by `internal/cmd/verify_completeness_test.go::TestDetermineStatusWithPolicy_SkipNonEvidence`
      and `internal/results/schema_test.go::TestParseSkippedResults`.
