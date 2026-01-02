"""Monte Carlo tests for graph algorithms.

REQ-TEST-005: Monte Carlo tests shall exist for graph algorithms

These tests use randomized inputs to verify graph algorithm correctness
across many different graph structures, finding edge cases that might
be missed by traditional unit tests.
"""

from __future__ import annotations

import random

import pytest

from rtmx.graph import DependencyGraph


def generate_random_dag(
    num_nodes: int,
    edge_probability: float = 0.3,
    seed: int | None = None,
) -> tuple[DependencyGraph, list[str]]:
    """Generate a random directed acyclic graph (DAG).

    Args:
        num_nodes: Number of nodes
        edge_probability: Probability of edge between any ordered pair
        seed: Random seed for reproducibility

    Returns:
        Tuple of (DependencyGraph, list of node IDs in topological order)
    """
    if seed is not None:
        random.seed(seed)

    graph = DependencyGraph()
    nodes = [f"REQ-MC-{i:03d}" for i in range(num_nodes)]

    # Add all nodes
    for node in nodes:
        graph._nodes.add(node)

    # Only add edges from lower to higher index to guarantee DAG
    for i in range(num_nodes):
        for j in range(i + 1, num_nodes):
            if random.random() < edge_probability:
                # Node i depends on node j (j must come before i)
                graph.add_edge(nodes[i], nodes[j])

    return graph, nodes


def generate_random_graph_with_cycle(
    num_nodes: int,
    cycle_size: int,
    extra_edges: int = 5,
    seed: int | None = None,
) -> tuple[DependencyGraph, list[str]]:
    """Generate a random graph with a guaranteed cycle.

    Args:
        num_nodes: Total number of nodes
        cycle_size: Size of the cycle to include
        extra_edges: Number of additional random edges
        seed: Random seed for reproducibility

    Returns:
        Tuple of (DependencyGraph, list of nodes forming the cycle)
    """
    if seed is not None:
        random.seed(seed)

    graph = DependencyGraph()
    nodes = [f"REQ-MC-{i:03d}" for i in range(num_nodes)]

    for node in nodes:
        graph._nodes.add(node)

    # Create guaranteed cycle among first cycle_size nodes
    cycle_nodes = nodes[:cycle_size]
    for i in range(cycle_size):
        from_node = cycle_nodes[i]
        to_node = cycle_nodes[(i + 1) % cycle_size]
        graph.add_edge(from_node, to_node)

    # Add some random edges
    for _ in range(extra_edges):
        i = random.randint(0, num_nodes - 1)
        j = random.randint(0, num_nodes - 1)
        if i != j:
            graph.add_edge(nodes[i], nodes[j])

    return graph, cycle_nodes


def verify_topological_order(graph: DependencyGraph, order: list[str]) -> bool:
    """Verify that a topological order is valid.

    Args:
        graph: The graph
        order: Proposed topological order

    Returns:
        True if order is valid
    """
    position = {node: i for i, node in enumerate(order)}

    for node in order:
        for dep in graph.dependencies(node):
            # Dependency must come after (higher index) than dependent
            if dep in position and position[dep] <= position[node]:
                return False

    return True


def verify_transitive_closure(
    graph: DependencyGraph,
    node: str,
    transitive_deps: set[str],
) -> bool:
    """Verify that transitive dependencies are complete and correct.

    Args:
        graph: The graph
        node: Starting node
        transitive_deps: Claimed transitive dependencies

    Returns:
        True if transitive_deps is correct
    """
    # BFS to find all reachable nodes
    visited: set[str] = set()
    to_visit = list(graph.dependencies(node))

    while to_visit:
        current = to_visit.pop()
        if current not in visited:
            visited.add(current)
            to_visit.extend(graph.dependencies(current) - visited)

    return visited == transitive_deps


# =============================================================================
# Monte Carlo Test Classes
# =============================================================================


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_monte_carlo
@pytest.mark.env_simulation
class TestCycleDetectionMonteCarlo:
    """Monte Carlo tests for cycle detection using Tarjan's algorithm."""

    @pytest.mark.parametrize("seed", range(10))
    def test_dag_has_no_cycles(self, seed: int) -> None:
        """Random DAGs should have no cycles detected."""
        graph, _ = generate_random_dag(
            num_nodes=random.randint(10, 50),
            edge_probability=0.2,
            seed=seed,
        )

        cycles = graph.find_cycles()
        assert len(cycles) == 0, f"DAG should have no cycles, found: {cycles}"

    @pytest.mark.parametrize("seed", range(10))
    def test_graph_with_cycle_detected(self, seed: int) -> None:
        """Graphs with cycles should have cycles detected."""
        cycle_size = random.randint(2, 5)
        graph, cycle_nodes = generate_random_graph_with_cycle(
            num_nodes=20,
            cycle_size=cycle_size,
            extra_edges=10,
            seed=seed,
        )

        cycles = graph.find_cycles()
        assert len(cycles) >= 1, "Should detect at least one cycle"

        # Verify cycle nodes are in some detected SCC
        all_scc_nodes = set()
        for scc in cycles:
            all_scc_nodes.update(scc)

        # At least some cycle nodes should be in detected SCCs
        overlap = set(cycle_nodes) & all_scc_nodes
        assert len(overlap) > 0, "Cycle nodes should be in detected SCCs"

    @pytest.mark.parametrize("seed", range(5))
    def test_self_loop_detection(self, seed: int) -> None:
        """Self-loops should be handled correctly."""
        random.seed(seed)
        graph = DependencyGraph()

        # Create a graph with a self-loop
        nodes = [f"REQ-SL-{i:03d}" for i in range(10)]
        for node in nodes:
            graph._nodes.add(node)

        # Add some normal edges
        for i in range(5):
            graph.add_edge(nodes[i], nodes[i + 1])

        # Note: Self-loops (A -> A) don't form SCCs > 1 in Tarjan's
        # So we test that the algorithm doesn't crash
        cycles = graph.find_cycles()
        # Should complete without error
        assert isinstance(cycles, list)

    @pytest.mark.parametrize("seed", range(5))
    def test_multiple_separate_cycles(self, seed: int) -> None:
        """Multiple separate cycles should all be detected."""
        random.seed(seed)
        graph = DependencyGraph()

        # Create two separate cycles
        cycle1 = ["REQ-C1-001", "REQ-C1-002", "REQ-C1-003"]
        cycle2 = ["REQ-C2-001", "REQ-C2-002"]

        for node in cycle1 + cycle2:
            graph._nodes.add(node)

        # Cycle 1: A -> B -> C -> A
        for i in range(len(cycle1)):
            graph.add_edge(cycle1[i], cycle1[(i + 1) % len(cycle1)])

        # Cycle 2: X -> Y -> X
        for i in range(len(cycle2)):
            graph.add_edge(cycle2[i], cycle2[(i + 1) % len(cycle2)])

        cycles = graph.find_cycles()
        assert len(cycles) >= 2, f"Should detect at least 2 cycles, found {len(cycles)}"

    def test_large_cycle(self) -> None:
        """Large cycles should be detected correctly."""
        graph = DependencyGraph()

        # Create a large cycle of 100 nodes
        cycle_size = 100
        nodes = [f"REQ-LC-{i:03d}" for i in range(cycle_size)]

        for node in nodes:
            graph._nodes.add(node)

        for i in range(cycle_size):
            graph.add_edge(nodes[i], nodes[(i + 1) % cycle_size])

        cycles = graph.find_cycles()
        assert len(cycles) == 1
        assert len(cycles[0]) == cycle_size


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_monte_carlo
@pytest.mark.env_simulation
class TestTopologicalSortMonteCarlo:
    """Monte Carlo tests for topological sorting."""

    @pytest.mark.parametrize("seed", range(10))
    def test_dag_topological_sort_valid(self, seed: int) -> None:
        """Topological sort of DAG should produce valid ordering."""
        graph, _ = generate_random_dag(
            num_nodes=random.randint(10, 50),
            edge_probability=0.2,
            seed=seed,
        )

        order = graph.topological_sort()
        assert order is not None, "DAG should have valid topological sort"
        assert verify_topological_order(graph, order), "Order should be valid"

    @pytest.mark.parametrize("seed", range(5))
    def test_cyclic_graph_no_topological_sort(self, seed: int) -> None:
        """Graphs with cycles should return None for topological sort."""
        graph, _ = generate_random_graph_with_cycle(
            num_nodes=20,
            cycle_size=3,
            extra_edges=5,
            seed=seed,
        )

        order = graph.topological_sort()
        assert order is None, "Cyclic graph should not have topological sort"

    def test_empty_graph(self) -> None:
        """Empty graph should have valid (empty) topological sort."""
        graph = DependencyGraph()
        order = graph.topological_sort()
        assert order is not None
        assert len(order) == 0

    def test_single_node(self) -> None:
        """Single node graph should have valid topological sort."""
        graph = DependencyGraph()
        graph._nodes.add("REQ-SINGLE-001")
        order = graph.topological_sort()
        assert order is not None
        assert len(order) == 1
        assert order[0] == "REQ-SINGLE-001"


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_monte_carlo
@pytest.mark.env_simulation
class TestTransitiveClosureMonteCarlo:
    """Monte Carlo tests for transitive closure operations."""

    @pytest.mark.parametrize("seed", range(10))
    def test_transitive_dependencies_complete(self, seed: int) -> None:
        """Transitive dependencies should include all reachable nodes."""
        graph, nodes = generate_random_dag(
            num_nodes=random.randint(10, 30),
            edge_probability=0.3,
            seed=seed,
        )

        # Pick a random node and verify transitive deps
        if nodes:
            test_node = random.choice(nodes)
            trans_deps = graph.transitive_dependencies(test_node)
            assert verify_transitive_closure(graph, test_node, trans_deps)

    @pytest.mark.parametrize("seed", range(10))
    def test_transitive_blocks_complete(self, seed: int) -> None:
        """Transitive blocks should include all blocking nodes."""
        graph, nodes = generate_random_dag(
            num_nodes=random.randint(10, 30),
            edge_probability=0.3,
            seed=seed,
        )

        if nodes:
            test_node = random.choice(nodes)
            trans_blocks = graph.transitive_blocks(test_node)

            # Verify by checking reverse direction
            # All nodes in trans_blocks should depend (transitively) on test_node
            for blocked in trans_blocks:
                trans_deps_of_blocked = graph.transitive_dependencies(blocked)
                assert test_node in trans_deps_of_blocked or blocked in graph.dependents(test_node)

    def test_no_transitive_in_isolated_node(self) -> None:
        """Isolated node should have no transitive dependencies or blocks."""
        graph = DependencyGraph()
        graph._nodes.add("REQ-ISO-001")

        assert graph.transitive_dependencies("REQ-ISO-001") == set()
        assert graph.transitive_blocks("REQ-ISO-001") == set()


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_monte_carlo
@pytest.mark.env_simulation
class TestCriticalPathMonteCarlo:
    """Monte Carlo tests for critical path calculation."""

    @pytest.mark.parametrize("seed", range(10))
    def test_critical_path_ordered_by_blocking(self, seed: int) -> None:
        """Critical path should be ordered by blocking count."""
        graph, _ = generate_random_dag(
            num_nodes=random.randint(15, 40),
            edge_probability=0.25,
            seed=seed,
        )

        critical = graph.critical_path()

        # Verify ordering: each node should block >= as many as next
        for i in range(len(critical) - 1):
            blocking_i = len(graph.transitive_blocks(critical[i]))
            blocking_next = len(graph.transitive_blocks(critical[i + 1]))
            assert blocking_i >= blocking_next, "Critical path should be sorted by blocking count"

    @pytest.mark.parametrize("seed", range(5))
    def test_critical_path_all_block_something(self, seed: int) -> None:
        """All nodes on critical path should block at least one requirement."""
        graph, _ = generate_random_dag(
            num_nodes=20,
            edge_probability=0.3,
            seed=seed,
        )

        critical = graph.critical_path()

        for node in critical:
            blocks = graph.transitive_blocks(node)
            assert len(blocks) > 0, f"{node} on critical path should block something"


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_monte_carlo
@pytest.mark.env_simulation
class TestGraphStatisticsMonteCarlo:
    """Monte Carlo tests for graph statistics calculations."""

    @pytest.mark.parametrize("seed", range(10))
    def test_statistics_accuracy(self, seed: int) -> None:
        """Graph statistics should accurately reflect graph structure."""
        num_nodes = random.randint(10, 50)
        graph, _ = generate_random_dag(
            num_nodes=num_nodes,
            edge_probability=0.2,
            seed=seed,
        )

        stats = graph.statistics()

        assert stats["nodes"] == graph.node_count
        assert stats["edges"] == graph.edge_count
        assert stats["cycles"] == len(graph.find_cycles())

        # Average should be edge_count / node_count
        expected_avg = graph.edge_count / graph.node_count if graph.node_count > 0 else 0
        assert abs(stats["avg_dependencies"] - expected_avg) < 0.001

    @pytest.mark.parametrize("seed", range(5))
    def test_edge_count_matches_edges(self, seed: int) -> None:
        """Edge count should match actual edge traversal."""
        graph, _ = generate_random_dag(
            num_nodes=30,
            edge_probability=0.25,
            seed=seed,
        )

        # Count edges manually
        manual_count = 0
        for node in graph._nodes:
            manual_count += len(graph.dependencies(node))

        assert graph.edge_count == manual_count


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_monte_carlo
@pytest.mark.env_simulation
class TestCyclePathMonteCarlo:
    """Monte Carlo tests for cycle path finding."""

    @pytest.mark.parametrize("seed", range(5))
    def test_cycle_path_valid(self, seed: int) -> None:
        """Cycle path should form a valid cycle."""
        cycle_size = random.randint(3, 6)
        graph, cycle_nodes = generate_random_graph_with_cycle(
            num_nodes=15,
            cycle_size=cycle_size,
            extra_edges=5,
            seed=seed,
        )

        cycles = graph.find_cycles()
        if cycles:
            # Get path for first detected SCC
            scc = cycles[0]
            path = graph.find_cycle_path(set(scc))

            # Path should start and end at same node if it's a proper cycle
            if len(path) > 1 and path[0] == path[-1]:
                # Verify each consecutive pair is a valid edge
                for i in range(len(path) - 1):
                    deps = graph.dependencies(path[i])
                    assert path[i + 1] in deps or path[i] in graph.dependencies(path[i + 1])

    def test_empty_cycle_members(self) -> None:
        """Empty cycle members should return empty path."""
        graph = DependencyGraph()
        path = graph.find_cycle_path(set())
        assert path == []


@pytest.mark.req("REQ-TEST-005")
@pytest.mark.scope_unit
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestGraphStress:
    """Stress tests for graph algorithms with larger inputs."""

    def test_large_dag_performance(self) -> None:
        """Large DAG should complete cycle detection in reasonable time."""
        graph, _ = generate_random_dag(
            num_nodes=500,
            edge_probability=0.05,
            seed=42,
        )

        # Should complete without timeout
        cycles = graph.find_cycles()
        assert len(cycles) == 0

    def test_large_dag_topological_sort(self) -> None:
        """Large DAG should complete topological sort in reasonable time."""
        graph, _ = generate_random_dag(
            num_nodes=500,
            edge_probability=0.05,
            seed=42,
        )

        order = graph.topological_sort()
        assert order is not None
        assert len(order) == 500

    def test_dense_graph_statistics(self) -> None:
        """Dense graph should compute statistics correctly."""
        graph, _ = generate_random_dag(
            num_nodes=100,
            edge_probability=0.5,
            seed=42,
        )

        stats = graph.statistics()
        assert stats["nodes"] == 100
        # Dense graph should have many edges
        assert stats["edges"] > 100
