# REQ-BENCH-018: Deduplicate Benchmark Issues by Title

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-017
- **Blocks**: (none)

## Requirement
Before creating a benchmark issue, the workflow shall search open issues with the same title and label set. If found, it shall post a comment instead of creating a new issue.

## Rationale
Five consecutive nightly failures created 30 duplicate issues (5 nights x 6 matrix languages), each titled "Benchmark regression: <lang>" with identical content.

## Acceptance Criteria
1. Five consecutive failures produce one issue with five comments, not five issues.
2. New issues are only created when no matching open issue exists.

## Files to Create/Modify
- .github/workflows/benchmark.yml

## Effort Estimate
0.25 weeks

## Test Strategy
- Run the workflow twice with the same failure; verify issue count does not increase.
