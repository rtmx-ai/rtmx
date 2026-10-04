# REQ-BENCH-028: Monorepo Dashboard Integration

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: LOW
- **Phase**: 23
- **Status**: COMPLETE
- **Dependencies**: REQ-BENCH-026
- **Blocks**: (none)

## Requirement
Benchmark workflow health shall be aggregated into make workspace-status at the monorepo root so nightly CI health is visible alongside RTM completion.

## Rationale
Benchmark status is currently visible only by navigating to the rtmx repo Actions tab. Operators using workspace-status have no visibility into benchmark health.

## Acceptance Criteria
1. make workspace-status output includes a Benchmark Resilience section with REQ-BENCH-* completion and config counts.

## Files to Create/Modify
- system/scripts/workspace-status.sh
- system/tests/test_benchmark_resilience.py

## Effort Estimate
0.5 weeks

## Test Strategy
- `make test-integration SLICE=benchmark-resilience`
