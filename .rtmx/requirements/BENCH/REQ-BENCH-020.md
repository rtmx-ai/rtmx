# REQ-BENCH-020: SHA-pinned Exemplars

## Metadata
- **Category**: BENCH
- **Subcategory**: Resilience
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: REQ-BENCH-021

## Requirement
exemplar.ref in every benchmark config shall be a 40-character commit SHA, not a tag or branch. A lint step shall reject non-SHA refs.

## Rationale
Tags are mutable on GitHub (force-pushed). A ref like "v2.32.3" can be moved, silently changing what the benchmark validates. SHA-pinning ensures reproducibility.

## Acceptance Criteria
1. Changing ref to "v2.32.3" fails a PR check.
2. All existing configs updated to SHAs.

## Files to Create/Modify
- benchmarks/configs/*.yaml
- lint script or Go test

## Effort Estimate
0.5 weeks

## Test Strategy
- Verify all configs have 40-char hex refs; verify lint rejects a non-SHA ref.
