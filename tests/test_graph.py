"""Tests for rtmx.graph module."""

from rtmx.graph import DependencyGraph


class TestDependencyGraph:
    """Tests for DependencyGraph class."""

    def test_empty_graph(self):
        """Test empty graph initialization."""
        graph = DependencyGraph()
        assert graph.node_count == 0
        assert graph.edge_count == 0
        assert graph.find_cycles() == []

    def test_add_edge(self):
        """Test adding edges to graph."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")

        assert graph.node_count == 2
        assert graph.edge_count == 1
        assert "B" in graph.dependencies("A")
        assert "A" in graph.dependents("B")

    def test_add_multiple_edges(self):
        """Test adding multiple edges."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")
        graph.add_edge("A", "C")
        graph.add_edge("B", "C")

        assert graph.node_count == 3
        assert graph.edge_count == 3
        assert graph.dependencies("A") == {"B", "C"}
        assert graph.dependencies("B") == {"C"}
        assert graph.dependencies("C") == set()

    def test_remove_edge(self):
        """Test removing edges from graph."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")
        graph.add_edge("A", "C")

        graph.remove_edge("A", "B")

        assert graph.dependencies("A") == {"C"}
        assert "A" not in graph.dependents("B")

    def test_no_cycles(self):
        """Test graph with no cycles."""
        graph = DependencyGraph()
        # A -> B -> C (linear chain)
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")

        cycles = graph.find_cycles()
        assert len(cycles) == 0

    def test_simple_cycle(self):
        """Test detecting a simple cycle."""
        graph = DependencyGraph()
        # A -> B -> C -> A (cycle)
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")
        graph.add_edge("C", "A")

        cycles = graph.find_cycles()
        assert len(cycles) == 1
        assert set(cycles[0]) == {"A", "B", "C"}

    def test_two_node_cycle(self):
        """Test detecting a two-node cycle."""
        graph = DependencyGraph()
        # A <-> B
        graph.add_edge("A", "B")
        graph.add_edge("B", "A")

        cycles = graph.find_cycles()
        assert len(cycles) == 1
        assert set(cycles[0]) == {"A", "B"}

    def test_multiple_independent_cycles(self):
        """Test detecting multiple independent cycles."""
        graph = DependencyGraph()
        # Cycle 1: A <-> B
        graph.add_edge("A", "B")
        graph.add_edge("B", "A")
        # Cycle 2: C <-> D
        graph.add_edge("C", "D")
        graph.add_edge("D", "C")
        # Link between cycles (no additional cycle)
        graph.add_edge("A", "C")

        cycles = graph.find_cycles()
        assert len(cycles) == 2

    def test_transitive_dependencies(self):
        """Test transitive dependency calculation."""
        graph = DependencyGraph()
        # A -> B -> C -> D
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")
        graph.add_edge("C", "D")

        deps = graph.transitive_dependencies("A")
        assert "B" in deps
        assert "C" in deps
        assert "D" in deps
        assert "A" not in deps

    def test_transitive_dependencies_empty(self):
        """Test transitive dependencies for node with no dependencies."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")

        deps = graph.transitive_dependencies("B")
        assert len(deps) == 0

    def test_transitive_blocks(self):
        """Test transitive blocking calculation."""
        graph = DependencyGraph()
        # A -> B means A depends on B, so B blocks A
        # A -> B -> C
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")

        # C transitively blocks A (via B) and directly blocks B
        blocked = graph.transitive_blocks("C")
        assert "B" in blocked
        assert "A" in blocked

    def test_transitive_blocks_empty(self):
        """Test transitive blocks for node with no dependents."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")

        blocked = graph.transitive_blocks("A")
        assert len(blocked) == 0

    def test_critical_path(self):
        """Test critical path analysis."""
        graph = DependencyGraph()
        # A -> B -> D
        # A -> C -> D
        # D has the most dependents transitively
        graph.add_edge("A", "B")
        graph.add_edge("A", "C")
        graph.add_edge("B", "D")
        graph.add_edge("C", "D")

        critical = graph.critical_path()
        assert len(critical) > 0
        # D should be most critical (A, B, C all depend on it directly or transitively)
        assert critical[0] == "D"

    def test_topological_sort_linear(self):
        """Test topological sort on linear graph."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")
        graph.add_edge("C", "D")

        order = graph.topological_sort()
        assert order is not None
        # A depends on B, B on C, C on D
        # Topological sort processes nodes with no incoming edges first
        # In this graph, A has no dependents, so A comes first
        # Valid topological order has A before B, B before C, C before D
        assert order.index("A") < order.index("B")
        assert order.index("B") < order.index("C")
        assert order.index("C") < order.index("D")

    def test_topological_sort_with_cycle(self):
        """Test topological sort returns None with cycle."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")
        graph.add_edge("C", "A")

        order = graph.topological_sort()
        assert order is None

    def test_statistics(self):
        """Test graph statistics."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")
        graph.add_edge("A", "C")
        graph.add_edge("B", "C")

        stats = graph.statistics()
        assert stats["nodes"] == 3
        assert stats["edges"] == 3
        assert stats["avg_dependencies"] == 1.0
        assert stats["cycles"] == 0


class TestTarjanAlgorithm:
    """Tests specifically for Tarjan's SCC algorithm edge cases."""

    def test_diamond_pattern_no_cycle(self):
        """Test diamond pattern (no cycle)."""
        graph = DependencyGraph()
        # A -> B, A -> C, B -> D, C -> D
        graph.add_edge("A", "B")
        graph.add_edge("A", "C")
        graph.add_edge("B", "D")
        graph.add_edge("C", "D")

        cycles = graph.find_cycles()
        assert len(cycles) == 0

    def test_y_pattern_no_double_counting(self):
        """Test Y-pattern doesn't double-count shared edges.

        Pattern: A -> B -> D and C -> B -> D
        The B -> D edge should only be counted once when calculating
        transitive blocks. This is a regression test for a bug found
        in related projects (cyclone, phoenix).
        """
        graph = DependencyGraph()
        graph.add_edge("A", "B")  # A depends on B
        graph.add_edge("C", "B")  # C depends on B
        graph.add_edge("B", "D")  # B depends on D

        # Edge count should be exactly 3, not 4
        assert graph.edge_count == 3

        # D transitively blocks B, A, and C (3 nodes, not 4)
        blocked_by_d = graph.transitive_blocks("D")
        assert blocked_by_d == {"A", "B", "C"}
        assert len(blocked_by_d) == 3

        # B transitively blocks A and C only
        blocked_by_b = graph.transitive_blocks("B")
        assert blocked_by_b == {"A", "C"}
        assert len(blocked_by_b) == 2

        # Critical path: D should be first (blocks 3), B second (blocks 2)
        critical = graph.critical_path()
        assert critical[0] == "D"
        assert critical[1] == "B"

    def test_nested_cycles(self):
        """Test nested/overlapping cycles."""
        graph = DependencyGraph()
        # A -> B -> C -> A (outer cycle)
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")
        graph.add_edge("C", "A")
        # B -> D -> B (inner cycle from B)
        graph.add_edge("B", "D")
        graph.add_edge("D", "B")

        cycles = graph.find_cycles()
        # All connected in one large SCC
        assert len(cycles) >= 1

    def test_disconnected_components(self):
        """Test graph with disconnected components."""
        graph = DependencyGraph()
        # Component 1: A -> B
        graph.add_edge("A", "B")
        # Component 2: C -> D (disconnected)
        graph.add_edge("C", "D")

        cycles = graph.find_cycles()
        assert len(cycles) == 0

    def test_large_cycle(self):
        """Test detecting a large cycle."""
        graph = DependencyGraph()
        nodes = ["N" + str(i) for i in range(10)]
        # Create cycle: N0 -> N1 -> N2 -> ... -> N9 -> N0
        for i in range(len(nodes)):
            graph.add_edge(nodes[i], nodes[(i + 1) % len(nodes)])

        cycles = graph.find_cycles()
        assert len(cycles) == 1
        assert len(cycles[0]) == 10

    def test_find_cycle_path(self):
        """Test finding actual cycle path."""
        graph = DependencyGraph()
        graph.add_edge("A", "B")
        graph.add_edge("B", "C")
        graph.add_edge("C", "A")

        cycles = graph.find_cycles()
        assert len(cycles) == 1

        path = graph.find_cycle_path(set(cycles[0]))
        # Path should start and end with same node
        assert path[0] == path[-1]
        # All cycle members should be in path
        assert set(path[:-1]) == set(cycles[0])


class TestGraphFromDatabase:
    """Tests for creating graphs from RTMDatabase."""

    def test_from_database(self, core_rtm_path):
        """Test creating graph from database fixture."""
        from rtmx import RTMDatabase

        db = RTMDatabase.load(core_rtm_path)
        graph = DependencyGraph.from_database(db)

        assert graph.node_count > 0
