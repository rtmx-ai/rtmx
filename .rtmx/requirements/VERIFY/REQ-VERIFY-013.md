# REQ-VERIFY-013: Raise-only --no-demote mode for partial/sharded result sets

## Metadata
- **Category**: VERIFY
- **Subcategory**: COMPLETENESS
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-012
- **Blocks**:
- **External ID**:

## Requirement

`rtmx verify --results --no-demote` shall never lower a requirement's status. For
each requirement it computes the new status under the completeness policy, then
keeps whichever of {current, computed} is more complete (by `Status.Weight`),
promoting freely while never demoting. Without the flag, existing behavior is
unchanged.

## Rationale

Even with skipped results correctly ignored (REQ-VERIFY-012), a partial or
sharded result set legitimately carries fewer passing `(scope, technique)`
combinations for a requirement than its full validation does — because the other
combinations ran in a different CI leg, or in an environment filtered out of this
run. Under the combinations policy the status engine correctly returns PARTIAL
for such a partial view, so `--update` would demote a requirement that is
genuinely COMPLETE when considered across all legs.

Projects that shard verification across many result files therefore need a mode
that treats each result set as *additive evidence*: it may raise a requirement's
status as new combinations are observed, but must never lower it on the basis of
a single partial view. `--no-demote` provides exactly that raise-only semantics,
letting sharded CI safely run `--update` and commit status forward without
spurious downgrades.

## Acceptance Criteria

- [ ] With `--no-demote`, a COMPLETE requirement whose result set yields PARTIAL
      stays COMPLETE.
- [ ] With `--no-demote`, a MISSING/PARTIAL requirement whose result set yields
      COMPLETE is still promoted (raises are not blocked).
- [ ] An equal computed status is unchanged.
- [ ] Without the flag, the default demotion behavior is preserved.
- [ ] Verified by `internal/cmd/verify_completeness_test.go::TestClampNoDemote`.
