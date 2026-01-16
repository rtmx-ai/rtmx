# REQ-GIT-003: Post-merge hook for semantic conflict detection

## Status: NOT STARTED
## Priority: HIGH
## Phase: 15
## Effort: 1.5 weeks

## Description

Implement a post-merge hook that detects semantic conflicts in RTM files after Git merge operations. The hook shall identify status regressions, orphaned dependencies, and other semantic issues that may have been introduced by the merge, and optionally auto-resolve conflicts using CRDT semantics.

## Acceptance Criteria

- [ ] Hook runs automatically after `git merge` and `git pull`
- [ ] Detects status regressions (COMPLETE -> PARTIAL, PARTIAL -> MISSING)
- [ ] Detects orphaned dependencies (merged req references deleted req)
- [ ] Detects duplicate requirements introduced by merge
- [ ] Reports conflicts with specific req_ids and field details
- [ ] `--auto-resolve` flag attempts CRDT-based conflict resolution
- [ ] Auto-resolve creates fixup commit with resolution details
- [ ] Non-auto mode exits with warnings, does not modify files
- [ ] `rtmx install --hooks --post-merge` installs the hook
- [ ] Hook integrates with merge driver conflict markers

## Test Cases

- `tests/test_hooks.py::test_postmerge_status_regression` - Detects COMPLETE->PARTIAL
- `tests/test_hooks.py::test_postmerge_orphaned_deps` - Detects deleted dependency refs
- `tests/test_hooks.py::test_postmerge_duplicate_reqs` - Detects duplicate req_ids
- `tests/test_hooks.py::test_postmerge_auto_resolve_regression` - Auto-fixes status
- `tests/test_hooks.py::test_postmerge_auto_resolve_orphan` - Auto-removes orphan refs
- `tests/test_hooks.py::test_postmerge_conflict_markers` - Processes merge driver markers

## Technical Notes

### Post-merge Hook Script

```bash
#!/bin/sh
# RTMX post-merge hook
# Installed by: rtmx install --hooks --post-merge

# Check if this was a squash merge (no second parent)
if git rev-parse HEAD^2 > /dev/null 2>&1; then
    MERGE_TYPE="merge"
else
    MERGE_TYPE="fast-forward"
fi

echo "Running RTMX post-merge analysis..."
rtmx post-merge-check --type "$MERGE_TYPE"

if [ $? -ne 0 ]; then
    echo ""
    echo "Semantic conflicts detected. Run 'rtmx resolve' to fix."
fi
```

### Semantic Conflict Types

1. **Status Regression**
   - COMPLETE -> PARTIAL or MISSING
   - PARTIAL -> MISSING
   - Indicates work was lost or reverted

2. **Orphaned Dependencies**
   - `depends_on` or `blocks` references non-existent req_id
   - Usually caused by one branch deleting, other branch adding refs

3. **Duplicate Requirements**
   - Same req_id appears multiple times
   - Usually merge driver failure or manual conflict resolution error

4. **Conflict Markers**
   - `<<<<<<<`, `=======`, `>>>>>>>` markers in CSV
   - Indicates merge driver could not resolve

### Auto-Resolution Rules

1. **Status Regression**: Keep higher status (COMPLETE > PARTIAL > MISSING)
2. **Orphaned Dependencies**: Remove reference, add note field
3. **Duplicates**: Keep version with higher status, merge notes
4. **Conflict Markers**: Parse both versions, apply CRDT merge

### Integration with Merge Driver

Post-merge hook reads `.rtmx/merge-conflicts.json` if present, which contains detailed conflict information from the merge driver for guided resolution.

## Files to Create/Modify

- `src/rtmx/cli/post_merge.py` - New post-merge-check command
- `src/rtmx/cli/resolve.py` - New resolve command for auto-resolution
- `src/rtmx/cli/install.py` - Add `--post-merge` option
- `src/rtmx/conflicts.py` - Conflict detection and resolution logic
- `tests/test_hooks.py` - Post-merge hook tests

## Dependencies

- REQ-GIT-001: Custom merge driver (provides conflict context)
- REQ-GIT-002: Pre-commit hook (shared validation infrastructure)

## Blocks

- None
