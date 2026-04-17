# REQ-BENCH-012: Real YAML Parser for Benchmark Configs

## Metadata
- **Category**: BENCH
- **Subcategory**: Robustness
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: (none)

## Requirement
Benchmark config parsing shall use a real YAML parser (yq, Python yaml module, or Go helper) rather than grep/awk/sed line parsing.

## Rationale
The root cause of the 2026-04-11..16 outage was get_nested_field() using an awk range that terminates on the same line it begins. This class of bug is inherent to line-based YAML "parsing" -- any nested structure is one edge case away from silent misparse.

## Acceptance Criteria
1. Nested keys three levels deep parse correctly.
2. All 23 benchmark configs parse identically to Go yaml.v3.

## Files to Create/Modify
- benchmarks/scripts/run-benchmark.sh (or new Go helper)

## Effort Estimate
1 week

## Test Strategy
- Unit test parsing every config against a reference parser.
