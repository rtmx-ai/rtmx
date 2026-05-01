# REQ-PLAN-011: Historical Velocity Calculation

## Metadata
- **Category**: PLAN
- **Subcategory**: Forecast
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-010, REQ-PLAN-002
- **Blocks**: REQ-PLAN-012

## Requirement

`rtmx velocity` shall compute team velocity from completed requirements
that have both `effort_weeks` and `completed_date` populated. Velocity is
defined as total effort-weeks completed divided by calendar-weeks elapsed.
`--window <weeks>` limits calculation to recent history.

## Acceptance Criteria

1. `rtmx velocity` displays velocity in effort-weeks/calendar-week
2. `--window 4` limits to last 4 calendar weeks
3. Gracefully handles zero data ("not enough data for velocity calculation")
4. `--json` outputs machine-readable velocity data
5. Only counts requirements with both effort_weeks > 0 and completed_date set

## Files to Create

- `internal/cmd/velocity.go`
- `internal/cmd/velocity_test.go`
