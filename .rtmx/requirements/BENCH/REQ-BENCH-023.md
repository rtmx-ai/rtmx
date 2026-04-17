# REQ-BENCH-023: Baseline Provenance

## Metadata
- **Category**: BENCH
- **Subcategory**: Integrity
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: REQ-BENCH-024

## Requirement
Every baseline JSON shall carry provenance fields: source_run_id, rtmx_version, exemplar_sha, generated_at. report.sh shall reject baselines without provenance.

## Rationale
Current baselines were hand-written with "timestamp: 2026-04-12T00:00:00Z" and exact expected values, never generated from an actual successful run. There is no way to determine whether a baseline is meaningful or fabricated.

## Acceptance Criteria
1. Current hand-written baselines fail report.sh validation until regenerated.
2. A properly-generated baseline passes validation.

## Files to Create/Modify
- benchmarks/scripts/report.sh
- benchmarks/results/baselines/*.json

## Effort Estimate
0.5 weeks

## Test Strategy
- Run report.sh against a baseline missing provenance; verify rejection.
