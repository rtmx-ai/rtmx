# REQ-BENCH-026: Job Summary on Every Run

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: REQ-BENCH-028

## Requirement
Each benchmark matrix job shall write a markdown table to $GITHUB_STEP_SUMMARY regardless of outcome (pass, fail, error), including language, markers found/expected, verify status, regression verdict, and exemplar SHA.

## Rationale
The current workflow produces no summary view. Operators must click into individual job logs to find results. A job summary provides single-glance dashboard.

## Acceptance Criteria
1. A failed run still shows a summary table with failure reason visible without expanding logs.

## Files to Create/Modify
- .github/workflows/benchmark.yml
- benchmarks/scripts/run-benchmark.sh

## Effort Estimate
0.25 weeks

## Test Strategy
- Run workflow; verify GITHUB_STEP_SUMMARY contains well-formed markdown table.
