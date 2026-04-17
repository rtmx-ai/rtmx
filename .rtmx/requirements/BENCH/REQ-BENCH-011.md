# REQ-BENCH-011: No Silent Make Rules

## Metadata
- **Category**: BENCH
- **Subcategory**: Resilience
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: (none)

## Requirement
The benchmarks Makefile run and compare rules shall not suppress script invocation with the @ prefix, ensuring CI logs show the full command line on failure.

## Rationale
The @ prefix in Makefile:11 hid the run-benchmark.sh invocation, so when the script exited 1 without producing output, CI showed only the make error line with no preceding context.

## Acceptance Criteria
1. Induced failure produces the command echo in CI logs.
2. The @ prefix is removed from run and compare rules.

## Files to Create/Modify
- benchmarks/Makefile

## Effort Estimate
0.25 weeks

## Test Strategy
- Induce a script failure; verify the make output includes the full command line.
