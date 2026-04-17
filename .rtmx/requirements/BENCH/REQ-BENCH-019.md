# REQ-BENCH-019: Auto-close Benchmark Issues on Recovery

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-017, REQ-BENCH-018
- **Blocks**: (none)

## Requirement
On a green nightly benchmark run, any open benchmark-regression or benchmark-infra issue for the recovering language shall be closed with a comment linking the successful run.

## Rationale
Without auto-close, resolved issues accumulate as stale noise. The 30 open issues from the outage must be manually closed.

## Acceptance Criteria
1. Fix root cause; next green nightly run closes existing tickets automatically.
2. Closing comment links the successful run.

## Files to Create/Modify
- .github/workflows/benchmark.yml

## Effort Estimate
0.25 weeks

## Test Strategy
- Open a test issue, simulate green run, verify issue is closed with comment.
