# REQ-ORCH-018: MCP Tool parallel-plan

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-010, REQ-ORCH-011, REQ-ORCH-012, REQ-MCP-005
- **Blocks**: (none)

## Summary

RTMX shall expose a new MCP tool `parallel-plan` that returns the parallel execution plan as structured JSON. AI agents call this tool to understand which web to claim, in what order to work, and when to merge. The tool returns the complete plan including parallel groups, web details, merge order, cross-web dependencies, file overlaps, and active worktree assignments. This integrates with the existing MCP server and follows the established tool registration pattern.

## Acceptance Criteria

1. The MCP tool `parallel-plan` is registered in the MCP server tool list.
2. The tool accepts optional parameters: `agents` (integer, default: auto-detect from group size).
3. The tool returns JSON with: phases (array of groups with web details), merge_order (array of web IDs), cross_web_dependencies (map of web ID to dependency web IDs), overlaps (array of overlap pairs), active_worktrees (array of assignments), critical_path_weeks, total_effort_weeks.
4. Each web in the response includes: id, requirements, unblocked, blocked, total_effort, file_surface, parallel_group.
5. The tool returns an error response if the database cannot be loaded.
6. The tool is read-only (no mutations, no authorization required beyond MCP connection).
7. The tool description is clear and sufficient for an AI agent to understand how to use it.

## Dependencies
- REQ-ORCH-010 (cross-web dependency graph)
- REQ-ORCH-011 (merge order computation)
- REQ-ORCH-012 (parallel group assignment)
- REQ-MCP-005 (MCP server tool registration pattern)

## Blocks
- (none)

## Files to Modify
- internal/adapters/mcp/tools_query.go (or appropriate file for read-only MCP tools)
- internal/adapters/mcp/tools_query_test.go (or appropriate test file)
- internal/adapters/mcp/server.go (register the new tool)

## Test Strategy
- Unit test verifying tool registration and discovery.
- Unit test verifying JSON response structure with a multi-web fixture database.
- Unit test verifying error response when database is unavailable.
- Unit test verifying optional `agents` parameter is passed through to plan computation.

## Effort
- 1.0 weeks

## Priority
- HIGH
