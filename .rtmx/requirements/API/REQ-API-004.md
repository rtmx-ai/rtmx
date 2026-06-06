# REQ-API-004: Dependency Graph Endpoint

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-TUI-004, REQ-DASH-004

## Requirement

The RTMX serve command shall expose a `GET /api/graph` endpoint that returns
the full dependency graph as a JSON structure of nodes and edges, suitable
for rendering by both ASCII (TUI) and interactive (GUI) graph visualizers.

## Rationale

Dependency visualization is central to project management -- it answers
"what is the critical path?", "where are the bottlenecks?", and "what can
run in parallel?". Both the TUI and GUI need the same graph data in a
format that supports layout algorithms. Exposing the graph as a structured
API avoids duplicating graph traversal logic in frontend code.

## Design

### Response Schema

```json
{
  "nodes": [
    {
      "id": "REQ-MCP-007",
      "category": "MCP",
      "status": "MISSING",
      "priority": "P0",
      "effort_weeks": 0.5,
      "label": "Response Size Logging",
      "blocked": false,
      "depth": 3
    }
  ],
  "edges": [
    {
      "from": "REQ-MCP-003",
      "to": "REQ-MCP-007",
      "type": "blocks"
    }
  ],
  "metadata": {
    "total_nodes": 241,
    "total_edges": 412,
    "critical_path": ["REQ-GO-001", "REQ-GO-009", "..."],
    "critical_path_length": 14,
    "max_depth": 8,
    "independent_webs": 5
  }
}
```

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| category | string | (all) | Filter graph to single category |
| root | string | (all) | Return subgraph rooted at this req_id (upstream + downstream) |
| depth | int | (all) | Limit traversal depth from root |
| status | string | (all) | Filter to specific status |

### Implementation

Reuses the existing `graph.Graph` and `graph.Web` packages. The endpoint
serializes the internal graph representation to the node/edge JSON format.
Critical path comes from `graph.CriticalPath()`. Independent webs come
from `graph.DetectWebs()`.

## Acceptance Criteria

1. `GET /api/graph` returns full graph with all nodes and edges.
2. Node `blocked` field is true when any upstream dependency is incomplete.
3. `depth` field reflects topological depth from root nodes.
4. `critical_path` is the longest path through the dependency DAG.
5. `?root=REQ-MCP-007` returns only the connected subgraph.
6. `?depth=2` limits traversal to 2 hops from root.
7. `?category=MCP` filters to MCP-category nodes and their cross-category edges.
8. Response time < 100ms for 500-node graph.

## Files to Create/Modify

- `internal/cmd/serve_api.go` -- Add /api/graph handler
- `internal/cmd/serve_api_test.go` -- Graph endpoint tests
- `internal/graph/export.go` -- GraphToJSON export function

## Effort Estimate

0.5 weeks

## Test Strategy

- Full graph returns all nodes and edges with correct counts
- Subgraph filtering returns connected component only
- Depth limiting works correctly
- Critical path matches `rtmx deps --critical` output
- Empty database returns empty graph (not error)
- Cyclic graph returns error with cycle description
