"""Dependency graph operations for RTMX.

This module provides graph algorithms for analyzing requirement dependencies:
- Cycle detection using Tarjan's strongly connected components algorithm
- Transitive closure for blocking analysis
- Critical path identification
"""

from __future__ import annotations

from collections import defaultdict
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.models import RTMDatabase


class DependencyGraph:
    """Directed graph representing requirement dependencies.

    The graph supports both forward edges (dependencies) and reverse edges (blocks).
    Provides algorithms for cycle detection and transitive analysis.
    """

    def __init__(self) -> None:
        """Initialize empty dependency graph."""
        # Forward edges: req_id -> set of requirements it depends on
        self._forward: dict[str, set[str]] = defaultdict(set)
        # Reverse edges: req_id -> set of requirements that depend on it
        self._reverse: dict[str, set[str]] = defaultdict(set)
        # All known nodes
        self._nodes: set[str] = set()

    @classmethod
    def from_database(cls, db: RTMDatabase) -> DependencyGraph:
        """Create dependency graph from RTM database.

        Args:
            db: RTM database

        Returns:
            DependencyGraph instance
        """
        graph = cls()

        for req in db:
            graph._nodes.add(req.req_id)

            # Add forward edges (dependencies)
            for dep_id in req.dependencies:
                graph._forward[req.req_id].add(dep_id)
                graph._reverse[dep_id].add(req.req_id)
                graph._nodes.add(dep_id)

            # Note: we don't add blocks as edges since they should be
            # reciprocal to dependencies. If A blocks B, then B depends on A.

        return graph

    def add_edge(self, from_id: str, to_id: str) -> None:
        """Add a dependency edge.

        Args:
            from_id: Requirement that depends
            to_id: Requirement that is depended upon
        """
        self._nodes.add(from_id)
        self._nodes.add(to_id)
        self._forward[from_id].add(to_id)
        self._reverse[to_id].add(from_id)

    def remove_edge(self, from_id: str, to_id: str) -> None:
        """Remove a dependency edge.

        Args:
            from_id: Requirement that depends
            to_id: Requirement that is depended upon
        """
        if to_id in self._forward[from_id]:
            self._forward[from_id].remove(to_id)
        if from_id in self._reverse[to_id]:
            self._reverse[to_id].remove(from_id)

    def dependencies(self, req_id: str) -> set[str]:
        """Get direct dependencies of a requirement.

        Args:
            req_id: Requirement identifier

        Returns:
            Set of requirement IDs this depends on
        """
        return self._forward.get(req_id, set()).copy()

    def dependents(self, req_id: str) -> set[str]:
        """Get requirements that directly depend on this one.

        Args:
            req_id: Requirement identifier

        Returns:
            Set of requirement IDs that depend on this
        """
        return self._reverse.get(req_id, set()).copy()

    def find_cycles(self) -> list[list[str]]:
        """Find all circular dependency cycles using Tarjan's SCC algorithm.

        Returns:
            List of cycles, where each cycle is a list of requirement IDs
            forming a strongly connected component with more than one node.
        """
        # Tarjan's algorithm for finding strongly connected components
        index_counter = [0]
        stack: list[str] = []
        lowlink: dict[str, int] = {}
        index: dict[str, int] = {}
        on_stack: dict[str, bool] = defaultdict(bool)
        sccs: list[list[str]] = []

        def strongconnect(req_id: str) -> None:
            index[req_id] = index_counter[0]
            lowlink[req_id] = index_counter[0]
            index_counter[0] += 1
            stack.append(req_id)
            on_stack[req_id] = True

            # Consider successors (dependencies)
            for dep in self._forward.get(req_id, set()):
                if dep not in index:
                    # Successor not yet visited; recurse
                    strongconnect(dep)
                    lowlink[req_id] = min(lowlink[req_id], lowlink[dep])
                elif on_stack[dep]:
                    # Successor is on stack and hence in current SCC
                    lowlink[req_id] = min(lowlink[req_id], index[dep])

            # If req_id is a root node, pop the stack to get SCC
            if lowlink[req_id] == index[req_id]:
                scc: list[str] = []
                while True:
                    w = stack.pop()
                    on_stack[w] = False
                    scc.append(w)
                    if w == req_id:
                        break
                # Only include SCCs with more than one node (actual cycles)
                if len(scc) > 1:
                    sccs.append(scc)

        # Find all SCCs
        for req_id in self._nodes:
            if req_id not in index:
                strongconnect(req_id)

        return sccs

    def find_cycle_path(self, cycle_members: set[str]) -> list[str]:
        """Find an actual cycle path through the given cycle members.

        Args:
            cycle_members: Set of requirement IDs forming a cycle

        Returns:
            List forming a cycle path (first element repeated at end)
        """
        if not cycle_members:
            return []

        # Start from arbitrary member
        start = next(iter(cycle_members))
        path = [start]
        current = start
        visited = {start}

        # Follow dependencies until we return to start
        while True:
            # Find next node in cycle
            next_nodes = self._forward.get(current, set()) & cycle_members
            if not next_nodes:
                break

            # Prefer unvisited nodes
            unvisited = next_nodes - visited
            if unvisited:
                next_node = next(iter(unvisited))
            else:
                # Allow revisiting to complete cycle
                next_node = next(iter(next_nodes))
                if next_node == start:
                    path.append(start)
                    return path
                break

            path.append(next_node)
            visited.add(next_node)
            current = next_node

            if current == start:
                path.append(start)
                return path

        # Fallback: return members as-is
        return list(cycle_members)

    def transitive_blocks(self, req_id: str) -> set[str]:
        """Get all requirements transitively blocked by a requirement.

        If A blocks B, and B blocks C, then A transitively blocks both B and C.

        Args:
            req_id: Requirement identifier

        Returns:
            Set of all transitively blocked requirement IDs
        """
        blocked: set[str] = set()
        to_visit = list(self._reverse.get(req_id, set()))

        while to_visit:
            current = to_visit.pop()
            if current not in blocked:
                blocked.add(current)
                # Add dependents of current
                to_visit.extend(self._reverse.get(current, set()) - blocked)

        return blocked

    def transitive_dependencies(self, req_id: str) -> set[str]:
        """Get all requirements this transitively depends on.

        Args:
            req_id: Requirement identifier

        Returns:
            Set of all transitive dependency requirement IDs
        """
        deps: set[str] = set()
        to_visit = list(self._forward.get(req_id, set()))

        while to_visit:
            current = to_visit.pop()
            if current not in deps:
                deps.add(current)
                to_visit.extend(self._forward.get(current, set()) - deps)

        return deps

    def critical_path(self) -> list[str]:
        """Identify requirements on the critical path.

        The critical path consists of requirements that:
        1. Are incomplete (not COMPLETE status)
        2. Block the most other incomplete requirements

        Returns:
            List of requirement IDs on critical path, sorted by blocking count
        """
        # Calculate blocking counts for each requirement
        blocking_counts: dict[str, int] = {}

        for req_id in self._nodes:
            blocked = self.transitive_blocks(req_id)
            blocking_counts[req_id] = len(blocked)

        # Sort by blocking count (descending)
        sorted_reqs = sorted(
            blocking_counts.keys(),
            key=lambda r: blocking_counts[r],
            reverse=True,
        )

        # Return requirements that block at least one other
        return [r for r in sorted_reqs if blocking_counts[r] > 0]

    def topological_sort(self) -> list[str] | None:
        """Perform topological sort of requirements.

        Returns:
            List of requirement IDs in topological order, or None if cycles exist
        """
        # Kahn's algorithm
        in_degree: dict[str, int] = {node: 0 for node in self._nodes}

        for node in self._nodes:
            for dep in self._forward.get(node, set()):
                if dep in in_degree:
                    in_degree[dep] += 1

        # Start with nodes that have no incoming edges
        queue = [node for node in self._nodes if in_degree[node] == 0]
        result: list[str] = []

        while queue:
            node = queue.pop(0)
            result.append(node)

            for dep in self._forward.get(node, set()):
                if dep in in_degree:
                    in_degree[dep] -= 1
                    if in_degree[dep] == 0:
                        queue.append(dep)

        # If we didn't process all nodes, there's a cycle
        if len(result) != len(self._nodes):
            return None

        return result

    @property
    def node_count(self) -> int:
        """Get number of nodes in graph."""
        return len(self._nodes)

    @property
    def edge_count(self) -> int:
        """Get number of edges in graph."""
        return sum(len(deps) for deps in self._forward.values())

    def statistics(self) -> dict[str, int | float]:
        """Get graph statistics.

        Returns:
            Dictionary with graph statistics
        """
        return {
            "nodes": self.node_count,
            "edges": self.edge_count,
            "avg_dependencies": self.edge_count / self.node_count if self.node_count else 0,
            "cycles": len(self.find_cycles()),
        }
