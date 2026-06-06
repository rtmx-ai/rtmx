# REQ-ADAPT-008: GitLab Bidirectional Sync

## Metadata
- **Category**: ADAPT
- **Subcategory**: GitLab
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-007
- **Blocks**: REQ-ADAPT-009

## Requirement

The GitLab adapter shall support bidirectional synchronization between
GitLab issues and RTMX requirements, mapping issue state and labels to
RTMX status, assignee to RTMX assignee, and milestones to RTMX sprint/
target version.

## Rationale

Bidirectional sync is essential for teams using GitLab Issues as their
primary task tracker. The GitLab issue lifecycle (opened/closed with
label refinement) maps naturally to RTMX statuses. Milestone mapping
enables release planning alignment.

## Design

### Status Mapping

```yaml
rtmx:
  adapters:
    gitlab:
      status_mapping:
        labels:
          "doing": "PARTIAL"
          "blocked": "MISSING"
          "review": "PARTIAL"
        # Default: opened -> MISSING, closed -> COMPLETE
```

### Milestone-to-Version Mapping

GitLab milestones map to RTMX sprint/version:

```yaml
rtmx:
  adapters:
    gitlab:
      version_mapping:
        milestones:
          "v1.2.0": "v1.2.0"  # direct mapping
          "Sprint 1": "v1.2.0"  # alias mapping
```

### Requirement ID Linking

Requirement IDs are stored as `[REQ-XXX-NNN]` markers in the issue
description (same pattern as GitHub adapter). The adapter parses these
during fetch and writes them during create/update.

### Sync Command

```bash
rtmx sync --adapter gitlab              # sync all linked issues
rtmx sync --adapter gitlab --dry-run    # preview changes
rtmx sync --adapter gitlab --pull-only  # one-way pull from GitLab
```

## Acceptance Criteria

1. `rtmx sync --adapter gitlab` syncs status bidirectionally.
2. Issue state + labels map to RTMX status per configuration.
3. GitLab milestones map to RTMX sprint/version.
4. Requirement ID in issue description preserved across syncs.
5. `--dry-run` shows planned changes without mutations.
6. Assignee synced bidirectionally (GitLab username to RTMX assignee).
7. New requirements pushed as new GitLab issues.
8. Sync report shows counts: updated, created, skipped, conflicted.

## Files to Create/Modify

- `internal/adapters/gitlab.go` -- Sync logic
- `internal/adapters/gitlab_test.go` -- Sync scenario tests

## Effort Estimate

1 week

## Test Strategy

- Sync scenarios: GitLab newer, RTMX newer, in-sync
- Label-based status mapping: each label combination tested
- Milestone mapping: verify version assignment from milestone
- Dry run: verify no API mutations
- Req ID linking: verify marker parsing and generation
