# REQ-PLAN-010: Auto-set Dates on Status Transitions

## Metadata
- **Category**: PLAN
- **Subcategory**: Automation
- **Priority**: MEDIUM
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-PLAN-011

## Requirement

When `rtmx verify --update` transitions a requirement from MISSING to
PARTIAL or COMPLETE, it shall call `req.SetStartedDate()` to record when
work began. When transitioning to COMPLETE, it shall additionally call
`req.SetCompletedDate()`. Both methods already exist and are tested at
`internal/database/requirement.go:181-193`.

## Rationale

These methods exist but are never called from any command. Adding ~4 lines
to the verify update loop populates temporal data that enables velocity
calculation and forecasting (REQ-PLAN-011, REQ-PLAN-012).

## Acceptance Criteria

1. MISSING -> PARTIAL sets started_date to today
2. MISSING -> COMPLETE sets started_date and completed_date to today
3. PARTIAL -> COMPLETE sets completed_date to today (started_date preserved)
4. Existing started_date is never overwritten (SetStartedDate() already handles this)
5. Date format is YYYY-MM-DD consistent with database convention

## Files to Modify

- `internal/cmd/verify.go` -- add SetStartedDate/SetCompletedDate calls in update loop
