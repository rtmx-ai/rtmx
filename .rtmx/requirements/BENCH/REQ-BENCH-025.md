# REQ-BENCH-025: Regression Tolerance Bands

## Metadata
- **Category**: BENCH
- **Subcategory**: Integrity
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-023
- **Blocks**: (none)

## Requirement
Regression thresholds in report.sh shall be configurable with explicit tolerance rather than strict inequality.

## Rationale
Current report.sh uses strict -lt for marker count. A one-marker fluctuation due to an upstream test rename fires a "regression" issue when no actual regression occurred.

## Acceptance Criteria
1. A one-marker reduction with tolerance 1 produces OK.
2. With tolerance 0, it produces REGRESSION.

## Files to Create/Modify
- benchmarks/scripts/report.sh
- benchmarks/configs/*.yaml

## Effort Estimate
0.25 weeks

## Test Strategy
- Run report.sh with various marker counts at and beyond tolerance thresholds.
