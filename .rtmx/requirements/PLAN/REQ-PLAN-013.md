# REQ-PLAN-013: Git Attribution for Requirement Completion

## Metadata
- **Category**: PLAN
- **Subcategory**: Attribution
- **Priority**: LOW
- **Phase**: 24
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-010, REQ-ORCH-005

## Requirement

When `rtmx verify --update` marks a requirement COMPLETE, it shall extract
the git author of the most recent commit that modified the test file (from
`git log -1 --format='%an' -- <test_module>`). If `assignee` is not already
set, it shall be auto-populated with the git author. This creates an
automatic attribution trail linking code authors to requirement completion.

## Rationale

Deferred until REQ-ORCH-005 (claim protocol with agent identity) is
implemented. The agent identity model and git attribution model must be
compatible -- agent claims should take precedence over git author when
both are available.

## Design

Uses existing git integration patterns from `internal/cmd/verify_meta.go`
(getGitHEAD, getCommitDistance).

## Acceptance Criteria

1. Git author extracted from test file commit history
2. Assignee auto-set only when empty (existing assignee never overwritten)
3. Agent identity (from claims.json) takes precedence over git author
4. Gracefully handles missing git history (non-git repos, shallow clones)
5. Works with both human authors and bot/agent commit authors

## Files to Modify

- `internal/cmd/verify.go`
- `internal/cmd/verify_meta.go`
