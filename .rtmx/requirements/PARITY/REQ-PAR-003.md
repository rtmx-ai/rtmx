# REQ-PAR-003: Transitive Blocking Analysis in Backlog

## Metadata
- **Category**: PARITY
- **Subcategory**: Graph
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-013, REQ-GO-011
- **Blocks**: REQ-GO-047

## Requirement

`rtmx backlog` shall use transitive dependency analysis for blocking counts and shall exclude blocked requirements from quick-wins.

## Rationale

The Go CLI currently only counts direct blocking dependencies, undercounting the true project impact. Python uses `transitive_blocks()` for accurate analysis. Additionally, Go's quick-wins section includes blocked requirements that can't be acted upon, which is misleading.

## Design

### Transitive Blocking

If A blocks B, and B blocks C, then A transitively blocks 2 requirements (B and C), not just 1.

```
Blocks column: "5 (2)" = 5 transitive, 2 direct
```

### Quick-Wins Filter

```go
// CURRENT (broken): includes blocked items
if priority.IsHighOrP0() && effort <= 1.0 {
    quickWins = append(quickWins, req)
}

// FIXED: exclude items blocked by incomplete deps
if priority.IsHighOrP0() && effort <= 1.0 && !hasIncompleteDependencies(req, db) {
    quickWins = append(quickWins, req)
}
```

## Acceptance Criteria

1. Blocking counts use transitive closure (not just direct deps)
2. Blocks column shows "X (Y)" format (X=transitive, Y=direct)
3. Quick-wins excludes requirements with incomplete dependencies
4. Critical path sorted by transitive blocking impact
5. Results match Python CLI on identical database

## Files to Modify

- `internal/cmd/backlog.go` - Use transitive analysis, fix quick-wins filter
- `internal/graph/graph.go` - Add/use TransitiveDependents()
- `internal/cmd/backlog_test.go` - Transitive blocking tests
