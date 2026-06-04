# REQ-VERIFY-009: Configurable Multi-Dimensional Completeness Policy

## Metadata
- **Category**: VERIFY
- **Subcategory**: COMPLETENESS
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-001
- **Blocks**:
- **External ID**:

## Requirement

`rtmx verify` shall support a configurable completeness policy, set via
`rtmx.completeness` in config, that determines when a requirement is COMPLETE
from its test evidence:

- `policy: simple` (default) — a requirement is COMPLETE when it has at least
  one passing test and no failing tests (the historical rule, unchanged).
- `policy: combinations` — a requirement is COMPLETE only when its passing
  tests cover at least `min_combinations` distinct tuples of the configured
  `dimensions` (a subset of `scope`, `technique`, `env`); passing-but-
  insufficient evidence is PARTIAL.

`require_all_pass` (default true) shall downgrade a COMPLETE requirement to
PARTIAL on any failing test under either policy. The policy applies to the
cross-language results path (`verify --results`), where dimension markers are
present; the native go-test path retains the simple rule.

## Rationale

The `phoenix`, `do178c`, and `iso26262` built-in schemas already define
multi-dimensional verification columns (scope/technique/env markers, DAL/ASIL
levels), but completeness determination was one-dimensional: a single passing
test marked a requirement COMPLETE. Systems-engineering and safety-critical
projects require evidence across multiple test dimensions before declaring a
requirement done (e.g. Phoenix's rule of >= 3 distinct scope x technique
combinations). This requirement closes that gap while remaining fully backward
compatible: absent configuration, behavior is identical to prior releases.

## Acceptance Criteria

- [ ] `policy: simple` (default) reproduces the prior single-test rule exactly.
- [ ] `policy: combinations` marks COMPLETE only at >= `min_combinations`
      distinct dimension tuples; otherwise PARTIAL.
- [ ] Duplicate tuples are counted once.
- [ ] `require_all_pass` downgrade behavior is preserved.
- [ ] Misconfiguration (combinations with no dimensions) falls back to simple
      with a warning.
- [ ] Verified by `internal/cmd/verify_completeness_test.go::TestDetermineStatusWithPolicy`.
