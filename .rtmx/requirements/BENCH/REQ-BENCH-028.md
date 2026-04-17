# REQ-BENCH-028: Monorepo Dashboard Integration

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: LOW
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-026
- **Blocks**: (none)

## Requirement
Benchmark workflow health shall be aggregated into make workspace-status at the monorepo root so nightly CI health is visible alongside RTM completion.

## Rationale
Benchmark status is currently visible only by navigating to the rtmx repo Actions tab. Operators using workspace-status have no visibility into benchmark health.

## Acceptance Criteria
1. make workspace-status output includes benchmark health section with latest run status per language.

## Files to Create/Modify
- system/scripts/workspace-status.sh

## Effort Estimate
0.5 weeks

## Test Strategy
- Run workspace-status after a benchmark run; verify benchmark section appears.
