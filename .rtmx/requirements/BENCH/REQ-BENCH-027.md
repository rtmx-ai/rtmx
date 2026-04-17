# REQ-BENCH-027: Consecutive-failure Escalation

## Metadata
- **Category**: BENCH
- **Subcategory**: Observability
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-017, REQ-BENCH-018
- **Blocks**: (none)

## Requirement
After N consecutive nightly failures for the same language (N=2), the workflow shall open a P1 issue with "blocker" label and ping the configured team.

## Rationale
The 2026-04-11..16 outage ran for 5+ consecutive nights with no escalation beyond issue creation. Regular regression issues were ignored. A P1 escalation ensures human attention within 48 hours.

## Acceptance Criteria
1. Simulate two consecutive failures; P1 issue is created with "blocker" label.
2. Team mention is included.

## Files to Create/Modify
- .github/workflows/benchmark.yml

## Effort Estimate
0.5 weeks

## Test Strategy
- Track consecutive failure count and verify P1 is created at threshold.
