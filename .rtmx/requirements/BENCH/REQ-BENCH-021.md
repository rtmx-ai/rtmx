# REQ-BENCH-021: Exemplar Clone Cache

## Metadata
- **Category**: BENCH
- **Subcategory**: Resilience
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-020
- **Blocks**: (none)

## Requirement
The benchmark workflow shall use actions/cache keyed by (repo, SHA) to cache the shallow clone between runs.

## Rationale
Each nightly run currently does 7 fresh git clones from github.com. Rate limits, network hiccups, or upstream repo deletion break all benchmarks simultaneously.

## Acceptance Criteria
1. Second nightly run for unchanged pins performs no git clone (cache hit).
2. Cache key includes repo and SHA.

## Files to Create/Modify
- .github/workflows/benchmark.yml

## Effort Estimate
0.5 weeks

## Test Strategy
- Run workflow twice with same pins; verify cache hit message.
