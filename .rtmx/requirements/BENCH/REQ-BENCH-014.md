# REQ-BENCH-014: Config Parse Test on Every PR

## Metadata
- **Category**: BENCH
- **Subcategory**: Shift-Left
- **Priority**: P0
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: (none)

## Requirement
A Go test shall load every file in benchmarks/configs/*.yaml through the Go YAML parser and verify that all required fields are present and non-empty. This test shall run in normal PR CI, not just nightly.

## Rationale
The benchmark script was merged via subtree with a latent parsing bug. No PR-level test exercised the config parsing. Five days of nightly failures went unnoticed because the only verification was the nightly schedule.

## Acceptance Criteria
1. Breaking any config file fails a PR check in under 30 seconds.
2. Test covers all 23 configs.
3. Test validates language, exemplar.repo, exemplar.ref, expected_markers, scan_command.

## Files to Create/Modify
- internal/benchmark/config_test.go

## Effort Estimate
0.5 weeks

## Test Strategy
- The test itself is the verification. Also: intentionally break a config and confirm CI catches it.
