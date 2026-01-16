# REQ-GIT-006: Branch-aware requirement status

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 15
## Effort: 2.0 weeks

## Description

RTMX shall provide branch-aware requirement tracking, showing per-branch status, differences between branches, and merge previews. This enables feature branches to track independent requirement progress while maintaining visibility into the overall project state.

## Acceptance Criteria

- [ ] `rtmx status --branch <name>` shows status for specific branch
- [ ] `rtmx status` shows current branch status by default
- [ ] `rtmx diff --branch <name>` shows requirement differences vs branch
- [ ] `rtmx diff --branch main` compares current branch to main
- [ ] `rtmx merge-preview <branch>` shows what merge would produce
- [ ] Status output indicates requirements changed in current branch
- [ ] New/modified/deleted requirements highlighted in diff
- [ ] Branch comparison works entirely offline via Git
- [ ] Supports comparing any two branches with `--from` and `--to`

## Test Cases

- `tests/test_branch.py::test_status_current_branch` - Default branch status
- `tests/test_branch.py::test_status_specific_branch` - Named branch status
- `tests/test_branch.py::test_diff_vs_main` - Diff against main branch
- `tests/test_branch.py::test_diff_between_branches` - Diff two branches
- `tests/test_branch.py::test_merge_preview` - Preview merge result
- `tests/test_branch.py::test_branch_diff_offline` - Works offline
- `tests/test_branch.py::test_changed_reqs_highlight` - Visual indicators

## Technical Notes

### Command Interface

```bash
# Status for current branch (default)
rtmx status

# Status for specific branch
rtmx status --branch feature/new-api

# Diff current branch vs main
rtmx diff --branch main

# Diff between any two branches
rtmx diff --from feature/api --to main

# Preview merge result
rtmx merge-preview main

# Show requirements changed in this branch
rtmx status --changes-only
```

### Git Integration

Branch operations use Git directly without network:

```python
def get_branch_rtm(branch: str, rtm_path: str = "docs/rtm_database.csv") -> str:
    """Get RTM content from a specific branch."""
    result = subprocess.run(
        ["git", "show", f"{branch}:{rtm_path}"],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise BranchNotFoundError(f"Cannot read RTM from branch: {branch}")
    return result.stdout


def get_branches() -> list[str]:
    """Get list of all branches."""
    result = subprocess.run(
        ["git", "branch", "-a", "--format=%(refname:short)"],
        capture_output=True,
        text=True,
    )
    return result.stdout.strip().split("\n")
```

### Diff Output Format

```
RTMX Branch Diff: feature/new-api -> main
============================================

New Requirements (in feature/new-api, not in main):
  + REQ-API-005: New authentication endpoint
  + REQ-API-006: Rate limiting implementation

Modified Requirements:
  ~ REQ-API-001: Status changed PARTIAL -> COMPLETE
  ~ REQ-API-003: Priority changed MEDIUM -> HIGH

Deleted Requirements (in main, not in feature/new-api):
  - REQ-LEGACY-001: Deprecated auth method

Summary: +2 new, ~2 modified, -1 deleted
```

### Merge Preview

```python
def merge_preview(ours: str, theirs: str, base: str | None = None) -> MergeResult:
    """Preview what a merge would produce without modifying files."""
    if base is None:
        # Find merge base
        result = subprocess.run(
            ["git", "merge-base", ours, theirs],
            capture_output=True,
            text=True,
        )
        base = result.stdout.strip()

    # Get RTM from each branch
    ours_rtm = RTMDatabase.from_string(get_branch_rtm(ours))
    theirs_rtm = RTMDatabase.from_string(get_branch_rtm(theirs))
    base_rtm = RTMDatabase.from_string(get_branch_rtm(base))

    # Simulate merge
    merged, conflicts = merge_rtm(base_rtm, ours_rtm, theirs_rtm)

    return MergeResult(
        merged=merged,
        conflicts=conflicts,
        added=find_additions(base_rtm, merged),
        removed=find_removals(base_rtm, merged),
        modified=find_modifications(base_rtm, merged),
    )
```

### Status with Branch Context

```
RTMX Status (branch: feature/new-api)
=====================================

Requirements: 45 total
  COMPLETE: 28 (62%)
  PARTIAL:  12 (27%)
  MISSING:   5 (11%)

Branch Changes (vs main):
  +3 new requirements
  ~5 modified requirements
  -1 deleted requirement

Run 'rtmx diff --branch main' for details.
```

### Data Model

```python
@dataclass
class BranchDiff:
    """Difference between two branches."""
    from_branch: str
    to_branch: str
    added: list[Requirement]      # In from, not in to
    removed: list[Requirement]    # In to, not in from
    modified: list[RequirementDiff]  # Changed between branches


@dataclass
class RequirementDiff:
    """Difference in a single requirement between branches."""
    req_id: str
    field_changes: dict[str, tuple[Any, Any]]  # field -> (old, new)
```

## Files to Create/Modify

- `src/rtmx/cli/status.py` - Add `--branch` option
- `src/rtmx/cli/diff.py` - New diff command
- `src/rtmx/cli/merge_preview.py` - New merge-preview command
- `src/rtmx/branch.py` - Branch comparison logic
- `tests/test_branch.py` - Branch operation tests

## Dependencies

- REQ-GIT-005: Offline-first operation (all branch ops must work offline)

## Blocks

- None
