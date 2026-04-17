# REQ-BENCH-015: Dry-run Mode for Benchmark Scripts

## Metadata
- **Category**: BENCH
- **Subcategory**: Shift-Left
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-010, REQ-BENCH-013
- **Blocks**: (none)

## Requirement
run-benchmark.sh shall support a --dry-run flag that parses the config, validates all required fields, resolves the exemplar SHA, verifies URL reachability, and exits 0 without cloning or running the workload.

## Rationale
A lightweight pre-flight check would have caught the parse bug instantly during development without requiring a full benchmark run.

## Acceptance Criteria
1. make -C benchmarks dry-run-all completes in under 60 seconds.
2. No workdir/ directories are created.
3. All configs report OK.

## Files to Create/Modify
- benchmarks/scripts/run-benchmark.sh
- benchmarks/Makefile

## Effort Estimate
0.5 weeks

## Test Strategy
- Run dry-run-all; verify no workdir/ directories are created and all configs report OK.
