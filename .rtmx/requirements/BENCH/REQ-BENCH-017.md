# REQ-BENCH-017: Distinguish Infra Failure from Benchmark Regression

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: REQ-BENCH-018, REQ-BENCH-019, REQ-BENCH-027

## Requirement
The benchmark workflow issue-creation step shall differentiate (a) report.sh exit 1 (true regression) from (b) any step failing before report.sh runs (infra/setup failure). Infra failures create "benchmark-infra: <lang>" issues; regressions create "benchmark-regression: <lang>" issues.

## Rationale
During the outage, if: failure() fired on every job failure regardless of cause. 30 issues labeled "regression" were created, all caused by the same awk parse bug. This desensitized monitoring and polluted the tracker.

## Acceptance Criteria
1. Inject a parse error; exactly one infra issue is created, no regression issue.
2. Inject a marker count regression; exactly one regression issue is created, no infra issue.

## Files to Create/Modify
- .github/workflows/benchmark.yml

## Effort Estimate
0.5 weeks

## Test Strategy
- Simulate both failure modes; verify issue labels.
