# REQ-BENCH-010: Diagnostic-on-exit

## Metadata
- **Category**: BENCH
- **Subcategory**: Resilience
- **Priority**: P0
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: (none)

## Requirement
run-benchmark.sh and report.sh shall install an ERR trap that prints script name, line number, failed command, and exit code before exiting non-zero.

## Rationale
During the 2026-04-11..16 outage, all 23 language benchmarks failed silently with zero diagnostic output because set -euo pipefail killed the script mid-assignment without printing anything. Five days of nightly failures produced only "make: *** [Makefile:11: run] Error 1". The ERR trap ensures any future failure is immediately diagnosable from CI logs.

## Acceptance Criteria
1. Force a failure at any point in the script; CI log contains the trap line with file, line number, and failed command.
2. Trap message is printed to stderr.
3. Both run-benchmark.sh and report.sh have the trap installed.

## Files to Create/Modify
- benchmarks/scripts/run-benchmark.sh
- benchmarks/scripts/report.sh

## Effort Estimate
0.25 weeks

## Test Strategy
- Inject a parse error, verify trap output appears in stderr.
