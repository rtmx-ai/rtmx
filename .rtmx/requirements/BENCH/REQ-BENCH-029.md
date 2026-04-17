# REQ-BENCH-029: Benchmark Coverage per Slice

## Metadata
- **Category**: BENCH
- **Subcategory**: Traceability
- **Priority**: LOW
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: (none)

## Requirement
When a slice in system/features/ adds new scan or verify behavior, the slice's SLICE.md shall list affected benchmark configs or declare "no benchmark impact". A lint step shall flag slices touching from-tests or database code without a benchmark impact statement.

## Rationale
Changes to scanner behavior can silently break benchmarks. Requiring an impact statement forces developers to consider benchmark implications during slice planning.

## Acceptance Criteria
1. A slice touching from-tests without a benchmark impact line fails a SLICE lint check.

## Files to Create/Modify
- system/ lint script
- SLICE.md template

## Effort Estimate
0.25 weeks

## Test Strategy
- Create a SLICE.md touching from-tests without benchmark-impact section; verify lint fails.
