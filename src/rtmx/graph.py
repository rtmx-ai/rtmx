"""Dependency graph operations for RTMX.

This module provides graph algorithms for analyzing requirement dependencies:
- Cycle detection using Tarjan's strongly connected components algorithm
- Transitive closure for blocking analysis
- Critical path identification
- Cross-repository edge tracking for federated requirements
"""

from __future__ import annotations

from collections import defaultdict
from dataclasses import dataclass
from enum import Enum
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.models import RTMDatabase


class EdgeType(str, Enum):
    """Type of dependency edge."""

    LOCAL = "local"  # Both endpoints in same repository
    CROSS_REPO = "cross_repo"  # Endpoints in different repositories
    SHADOW = "shadow"  # Destination is a shadow requirement


@dataclass
class CrossRepoEdge:
    """Represents a dependency edge that spans repository boundaries.

    Cross-repo edges track dependencies between requirements in different
    repositories, enabling federated requirements traceability across
    trust boundaries.

    Attributes:
        from_id: Source requirement ID (full format: repo:req_id or just req_id)
        to_id: Destination requirement ID (full format: repo:req_id or just req_id)
        from_repo: Source repository path (empty string for local)
        to_repo: Destination repository path (empty string for local)
        edge_type: Classification of the edge
        verified: Whether the destination has been verified accessible
        shadow_hash: Hash of shadow requirement if edge_type is SHADOW
    """

    from_id: str
    to_id: str
    from_repo: str = ""
    to_repo: str = ""
    edge_type: EdgeType = EdgeType.LOCAL
    verified: bool = False
    shadow_hash: str = ""

    @property
    def is_cross_repo(self) -> bool:
        """Check if edge crosses repository boundaries."""
        return self.edge_type in (EdgeType.CROSS_REPO, EdgeType.SHADOW)

    @property
    def from_full_id(self) -> str:
        """Get fully qualified source ID."""
        if self.from_repo:
            return f"{self.from_repo}:{self.from_id}"
        return self.from_id

    @property
    def to_full_id(self) -> str:
        """Get fully qualified destination ID."""
        if self.to_repo:
            return f"{self.to_repo}:{self.to_id}"
        return self.to_id

    def __hash__(self) -> int:
        """Enable use in sets and as dict keys."""
        return hash((self.from_full_id, self.to_full_id))

    def __eq__(self, other: object) -> bool:
        """Check equality based on full IDs."""
        if not isinstance(other, CrossRepoEdge):
            return NotImplemented
        return self.from_full_id == other.from_full_id and self.to_full_id == other.to_full_id


class DependencyGraph:
    """Directed graph representing requirement dependencies.

    The graph supports both forward edges (dependencies) and reverse edges (blocks).
    Provides algorithms for cycle detection and transitive analysis.
    Tracks cross-repository edges for federated requirements.
    """

    def __init__(self) -> None:
        """Initialize empty dependency graph."""
        # Forward edges: req_id -> set of requirements it depends on
        self._forward: dict[str, set[str]] = defaultdict(set)
        # Reverse edges: req_id -> set of requirements that depend on it
        self._reverse: dict[str, set[str]] = defaultdict(set)
        # All known nodes
        self._nodes: set[str] = set()
        # Cross-repo edges for federated tracking
        self._cross_repo_edges: set[CrossRepoEdge] = set()
        # Repository for this graph (empty string for local-only)
        self._repo: str = ""

    @classmethod
    def from_database(cls, db: RTMDatabase, repo: str = "") -> DependencyGraph:
        """Create dependency graph from RTM database.

        Args:
            db: RTM database
            repo: Repository identifier for this database (empty for local)

        Returns:
            DependencyGraph instance
        """
        from rtmx.parser import parse_requirement_ref

        graph = cls()
        graph._repo = repo

        for req in db:
            graph._nodes.add(req.req_id)

            # Add forward edges (dependencies)
            for dep_str in req.dependencies:
                ref = parse_requirement_ref(dep_str)

                if ref.is_local:
                    # Local dependency
                    graph._forward[req.req_id].add(ref.req_id)
                    graph._reverse[ref.req_id].add(req.req_id)
                    graph._nodes.add(ref.req_id)
                else:
                    # Cross-repo dependency
                    to_repo = ref.full_repo or ""
                    edge = CrossRepoEdge(
                        from_id=req.req_id,
                        to_id=ref.req_id,
                        from_repo=repo,
                        to_repo=to_repo,
                        edge_type=EdgeType.CROSS_REPO,
                    )
                    graph._cross_repo_edges.add(edge)
                    # Also add to forward/reverse for graph algorithms
                    full_to_id = dep_str  # Keep original reference format
                    graph._forward[req.req_id].add(full_to_id)
                    graph._reverse[full_to_id].add(req.req_id)
                    graph._nodes.add(full_to_id)

            # Note: we don't add blocks as edges since they should be
            # reciprocal to dependencies. If A blocks B, then B depends on A.

        return graph

    def add_cross_repo_edge(self, edge: CrossRepoEdge) -> None:
        """Add a cross-repository dependency edge.

        Args:
            edge: CrossRepoEdge to add
        """
        self._cross_repo_edges.add(edge)
        self._nodes.add(edge.from_full_id)
        self._nodes.add(edge.to_full_id)
        self._forward[edge.from_full_id].add(edge.to_full_id)
        self._reverse[edge.to_full_id].add(edge.from_full_id)

    def get_cross_repo_edges(self) -> set[CrossRepoEdge]:
        """Get all cross-repository edges.

        Returns:
            Set of CrossRepoEdge instances
        """
        return self._cross_repo_edges.copy()

    def cross_repo_dependencies(self, req_id: str) -> list[CrossRepoEdge]:
        """Get cross-repo dependencies for a requirement.

        Args:
            req_id: Requirement identifier

        Returns:
            List of CrossRepoEdge instances for this requirement's cross-repo deps
        """
        return [e for e in self._cross_repo_edges if e.from_id == req_id]

    def cross_repo_dependents(self, req_id: str) -> list[CrossRepoEdge]:
        """Get requirements from other repos that depend on this one.

        Args:
            req_id: Requirement identifier

        Returns:
            List of CrossRepoEdge instances where this requirement is the target
        """
        return [e for e in self._cross_repo_edges if e.to_id == req_id]

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
        in_degree: dict[str, int] = dict.fromkeys(self._nodes, 0)

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
        cross_repo_count = len(self._cross_repo_edges)
        return {
            "nodes": self.node_count,
            "edges": self.edge_count,
            "cross_repo_edges": cross_repo_count,
            "avg_dependencies": self.edge_count / self.node_count if self.node_count else 0,
            "cycles": len(self.find_cycles()),
        }

    @property
    def cross_repo_edge_count(self) -> int:
        """Get number of cross-repository edges."""
        return len(self._cross_repo_edges)
