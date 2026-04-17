# REQ-BENCH-013: Config Schema Validation

## Metadata
- **Category**: BENCH
- **Subcategory**: Robustness
- **Priority**: P0
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001
- **Blocks**: (none)

## Requirement
Every benchmark config shall be validated at script entry against required fields (language, exemplar.repo, exemplar.ref, scan_command, expected_markers). Missing or empty required fields shall cause exit 2 with message "<config> missing required field <field>".

## Rationale
In the 2026-04-11..16 outage, exemplar.repo parsed as empty string. The script continued into git clone with an empty URL instead of catching the empty field upfront. Validation at the boundary prevents cascading failures.

## Acceptance Criteria
1. Remove exemplar.repo from python.yaml copy; script exits 2 with "configs/python.yaml missing required field exemplar.repo".
2. Each of the 5 required fields is validated.
3. Validation runs before any network call.

## Files to Create/Modify
- benchmarks/scripts/run-benchmark.sh

## Effort Estimate
0.25 weeks

## Test Strategy
- Create a config with each required field missing in turn; verify each produces the correct error and exit code.
