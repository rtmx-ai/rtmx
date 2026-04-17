# REQ-BENCH-024: Baseline Regeneration Workflow

## Metadata
- **Category**: BENCH
- **Subcategory**: Integrity
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-023
- **Blocks**: (none)

## Requirement
A dispatchable workflow "benchmarks-bless" shall regenerate baselines from the latest green run and open a PR with the updated baseline JSONs including full provenance.

## Rationale
Without a blessed regeneration path, baselines either drift stale or are hand-edited without accountability.

## Acceptance Criteria
1. Manual dispatch produces a PR with updated baselines.
2. Baselines contain valid provenance fields matching the run.

## Files to Create/Modify
- .github/workflows/benchmark-bless.yml

## Effort Estimate
0.5 weeks

## Test Strategy
- Dispatch the workflow; verify PR is created with updated baselines.
