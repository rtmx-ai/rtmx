# REQ-BENCH-016: PR-level Smoke Benchmark

## Metadata
- **Category**: BENCH
- **Subcategory**: Shift-Left
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-014
- **Blocks**: (none)

## Requirement
One fast language benchmark (Go against a small pinned exemplar) shall run on every PR that touches benchmarks/**, internal/cmd/from_tests*, internal/database/**, or cmd/rtmx/**.

## Rationale
Schedule-only triggers mean breakages surface at 04:00 UTC the next day. A PR-level smoke catches regressions before merge.

## Acceptance Criteria
1. PR touching scan command behavior produces a benchmark result in CI within 5 minutes.
2. Only triggers on relevant path changes.

## Files to Create/Modify
- .github/workflows/ci.yml or new .github/workflows/benchmark-smoke.yml

## Effort Estimate
0.5 weeks

## Test Strategy
- Push a PR touching benchmarks/configs/go.yaml; verify smoke job runs.
