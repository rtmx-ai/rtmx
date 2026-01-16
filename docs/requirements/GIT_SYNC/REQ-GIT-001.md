# REQ-GIT-001: Custom merge driver for RTM CSV 3-way merge

## Status: NOT STARTED
## Priority: HIGH
## Phase: 15
## Effort: 2.0 weeks

## Description

Implement a custom Git merge driver that performs semantic 3-way merges of RTM CSV files. The driver shall merge by requirement ID rather than line-by-line, using CRDT conflict resolution semantics when the same requirement is modified in both branches.

## Acceptance Criteria

- [ ] `rtmx merge-driver %O %A %B %L %P` command implements Git merge driver protocol
- [ ] Merge driver parses base (%O), ours (%A), and theirs (%B) CSV files
- [ ] Requirements are matched by `req_id` field for semantic merge
- [ ] New requirements in either branch are included in merged output
- [ ] Deleted requirements in one branch are removed unless modified in other
- [ ] Concurrent modifications to same requirement use CRDT LWW resolution
- [ ] Merge driver writes merged result to %A (ours) file
- [ ] Exit code 0 on successful merge, non-zero on unresolvable conflict
- [ ] Conflict markers are inserted for unresolvable semantic conflicts
- [ ] Merge preserves CSV header and column order from ours (%A)

## Test Cases

- `tests/test_merge_driver.py::test_merge_disjoint_additions` - Different reqs added in each branch
- `tests/test_merge_driver.py::test_merge_same_req_different_fields` - Same req, different fields modified
- `tests/test_merge_driver.py::test_merge_same_field_lww` - Same field modified, LWW resolution
- `tests/test_merge_driver.py::test_merge_deletion_vs_modification` - Delete vs modify conflict
- `tests/test_merge_driver.py::test_merge_preserves_order` - Output preserves logical order
- `tests/test_merge_driver.py::test_merge_conflict_markers` - Unresolvable conflicts marked

## Technical Notes

### Git Merge Driver Protocol

Git invokes custom merge drivers with arguments:
- `%O` - Ancestor (base) version path
- `%A` - Current branch (ours) version path - also the output file
- `%B` - Other branch (theirs) version path
- `%L` - Conflict marker size
- `%P` - Pathname in the repository

### Merge Algorithm

1. Parse all three CSV files into requirement dictionaries keyed by `req_id`
2. For each requirement in base:
   - If deleted in ours but not theirs: mark for conflict or delete
   - If deleted in theirs but not ours: mark for conflict or delete
   - If modified in both: apply field-level CRDT merge
3. Add new requirements from ours (not in base)
4. Add new requirements from theirs (not in base, not conflicting with ours)
5. Write merged result to ours path

### CRDT Field Resolution

For same-field conflicts, use Last-Writer-Wins based on:
1. If both branches have timestamps, use most recent
2. If no timestamps, prefer ours (current branch)
3. For status field, use status progression rules (MISSING < PARTIAL < COMPLETE)

## Files to Create/Modify

- `src/rtmx/cli/merge_driver.py` - New CLI command for merge driver
- `src/rtmx/merge.py` - Core merge logic
- `tests/test_merge_driver.py` - Comprehensive merge tests

## Dependencies

- REQ-CRDT-001: CRDT layer for conflict resolution semantics
- REQ-CRDT-006: Concurrent edit merge without conflicts

## Blocks

- REQ-GIT-003: Post-merge hook for semantic conflict detection
- REQ-GIT-004: Git attributes configuration
