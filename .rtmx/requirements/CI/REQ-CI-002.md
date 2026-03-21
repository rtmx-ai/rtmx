# REQ-CI-002: PR RTM Diff Comments

## Metadata
- **Category**: CI
- **Subcategory**: PR
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-022, REQ-GO-012
- **Blocks**: REQ-GO-047

## Requirement

CI pipeline shall post RTM diff and health status as PR comments when PRs modify the RTM database, requirements, or tests.

## Rationale

The Python CI (rtmx-pr.yml) triggers on PRs that modify `docs/rtm_database.csv`, tests, or requirements. It runs `rtmx diff` against the base branch and posts a formatted comment showing what changed. This gives reviewers visibility into requirement impact before merging.

## Design

### New Workflow: `.github/workflows/rtmx-pr.yml`

Triggers on pull_request when these paths change:
- `.rtmx/database.csv`
- `.rtmx/requirements/**`
- `internal/cmd/*_test.go`
- `test/**`
- `pkg/rtmx/**`

Steps:
1. Checkout PR with full history
2. Build rtmx binary
3. Get base branch RTM database
4. Run `rtmx diff base-rtm.csv .rtmx/database.csv`
5. Run `rtmx health --json`
6. Post/update PR comment with diff and health status

### PR Comment Format

```markdown
## 📋 RTM Changes

**Health:** ✅ HEALTHY

### Changes
| Requirement | Field | Before | After |
|-------------|-------|--------|-------|
| REQ-PAR-001 | status | MISSING | COMPLETE |
| REQ-PAR-002 | status | MISSING | COMPLETE |

*2 requirements changed*
```

## Acceptance Criteria

1. PR comment appears on PRs modifying RTM-related files
2. Comment shows RTM diff in table format
3. Comment shows health status with emoji
4. Existing comment is updated (not duplicated) on push
5. PR fails if health check returns UNHEALTHY

## Files to Create

- `.github/workflows/rtmx-pr.yml` - New workflow
