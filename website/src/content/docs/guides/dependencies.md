---
title: Dependency Analysis
description: Detect cycles, find critical paths, and analyze blockers
---

RTMX provides powerful dependency analysis tools to understand how your requirements relate to each other.

## Commands Overview

| Command | Description |
|---------|-------------|
| `rtmx deps` | Show dependency graph sorted by blocking impact |
| `rtmx cycles` | Detect circular dependencies |
| `rtmx reconcile` | Check and fix dependency reciprocity |

## Dependency Graph

The `rtmx deps` command shows all requirements sorted by how many other requirements they block:

```bash
rtmx deps
```

Output shows:
- **ID**: Requirement identifier
- **Deps**: Number of dependencies this requirement has
- **Blocks**: Number of requirements blocked by this one
- **Description**: Truncated requirement text

### Filtering

```bash
rtmx deps --phase 1        # Filter by phase
rtmx deps --category CORE  # Filter by category
rtmx deps --req REQ-001    # Show specific requirement
```

## Cycle Detection

Circular dependencies create impossible situations where A depends on B, and B depends on A. The `rtmx cycles` command uses Tarjan's algorithm to detect these:

```bash
rtmx cycles
```

If cycles exist, you'll see:

```
CIRCULAR DEPENDENCIES DETECTED

Cycle 1:
  REQ-AUTH-001 → REQ-AUTH-003 → REQ-AUTH-001

Recommendation: Review these requirements and break the cycle
by removing one dependency.
```

A healthy project shows:

```
✓ NO CIRCULAR DEPENDENCIES FOUND

The dependency graph is acyclic (DAG). This is ideal for
requirements management.
```

## Reciprocity Check

Dependencies should be reciprocal: if A depends on B, then B should list A in its `blocks` field. The `rtmx reconcile` command checks this:

```bash
rtmx reconcile           # Check for violations
rtmx reconcile --fix     # Auto-fix violations
```

## Critical Path Analysis

The critical path identifies which requirements, if delayed, would delay the entire project. Requirements with high "Blocks" counts are on the critical path.

```bash
rtmx deps | head -10     # Top 10 blocking requirements
```

Focus effort on completing these first to unblock the most downstream work.

## Best Practices

### Keep Dependencies Minimal

Each dependency is a potential blocker. Only add dependencies that are truly necessary.

### Avoid Deep Chains

Long dependency chains (A → B → C → D → E) create fragile projects. If any link breaks, everything downstream is blocked.

### Phase Dependencies

Use phases to create natural dependency boundaries:

```csv
REQ-PHASE2-001,...,phase=2,dependencies=REQ-PHASE1-FINAL
```

### Regular Health Checks

Run dependency analysis as part of CI:

```yaml
- run: rtmx cycles
- run: rtmx reconcile
```

## Integration with Backlog

The `rtmx backlog` command uses dependency analysis to prioritize work. Requirements that block many others appear first, ensuring you tackle high-impact items early.

```bash
rtmx backlog    # Prioritized by blocking impact
```
