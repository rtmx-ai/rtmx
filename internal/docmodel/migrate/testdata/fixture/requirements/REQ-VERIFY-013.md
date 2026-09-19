# REQ-VERIFY-013 (fixture excerpt)

## Acceptance Criteria

- [ ] With `--no-demote`, a COMPLETE requirement whose result set yields PARTIAL
      stays COMPLETE.
- [ ] With `--no-demote`, a MISSING/PARTIAL requirement whose result set yields
      COMPLETE is still promoted (raises are not blocked).
- [ ] An equal computed status is unchanged.
- [ ] Without the flag, the default demotion behavior is preserved.
- [ ] Verified by `internal/cmd/verify_completeness_test.go::TestClampNoDemote`.

## Implementation History

Free-form notes that must NOT round-trip into CSV scalars.
