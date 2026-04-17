# REQ-BENCH-022: Bounded Retry on Transient Network Failure

## Metadata
- **Category**: BENCH
- **Subcategory**: Resilience
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-010
- **Blocks**: (none)

## Requirement
The clone and setup steps in run-benchmark.sh shall retry up to 2 times with exponential backoff on transient failures. The result JSON shall record "network-failure" distinct from "benchmark-failure".

## Rationale
A single transient DNS failure or GitHub rate-limit currently kills the benchmark with no retry. This is an infrastructure failure, not a benchmark regression.

## Acceptance Criteria
1. Simulated first-try DNS failure still produces a valid result JSON after retry.
2. After 2 retries fail, result JSON shows "network-failure".

## Files to Create/Modify
- benchmarks/scripts/run-benchmark.sh

## Effort Estimate
0.5 weeks

## Test Strategy
- Mock a transient clone failure; verify retry behavior and result JSON status field.
