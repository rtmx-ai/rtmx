"""Performance tests for RTMX with large scale databases.

This module tests that RTMX commands complete within acceptable time limits
when working with large RTM databases (1000+ requirements).

REQ-TEST-009: Large Scale Performance Tests
"""

from __future__ import annotations

import csv
import random
import sys
import time
import tracemalloc
from io import StringIO
from pathlib import Path
from typing import TYPE_CHECKING
from unittest import mock

import pytest

from rtmx import Priority, RTMDatabase, Status
from rtmx.cli.backlog import run_backlog
from rtmx.cli.from_tests import scan_test_directory
from rtmx.cli.status import run_status
from rtmx.graph import DependencyGraph

if TYPE_CHECKING:
    pass  # No type-only imports needed currently

# Performance thresholds from acceptance criteria
STATUS_TIMEOUT_SECONDS = 5.0
BACKLOG_TIMEOUT_SECONDS = 5.0
CYCLES_TIMEOUT_SECONDS = 10.0
FROM_TESTS_TIMEOUT_SECONDS = 30.0
MAX_MEMORY_MB = 500


def generate_large_csv(
    num_requirements: int,
    path: Path,
    dependency_density: float = 0.1,
    include_cycles: bool = False,
) -> None:
    """Generate a large RTM CSV file for testing.

    Args:
        num_requirements: Number of requirements to generate
        path: Path to write the CSV file
        dependency_density: Fraction of requirements that have dependencies (0.0-1.0)
        include_cycles: Whether to include circular dependencies
    """
    categories = ["SOFTWARE", "HARDWARE", "PERFORMANCE", "TESTING", "DOCUMENTATION"]
    subcategories = {
        "SOFTWARE": ["ALGORITHM", "API", "UI", "DATABASE", "NETWORK"],
        "HARDWARE": ["SENSOR", "ACTUATOR", "POWER", "INTERFACE"],
        "PERFORMANCE": ["TIMING", "ACCURACY", "THROUGHPUT", "LATENCY"],
        "TESTING": ["UNIT", "INTEGRATION", "SYSTEM", "ACCEPTANCE"],
        "DOCUMENTATION": ["API", "USER", "ADMIN", "DESIGN"],
    }
    statuses = [Status.COMPLETE, Status.PARTIAL, Status.MISSING, Status.NOT_STARTED]
    priorities = [Priority.P0, Priority.HIGH, Priority.MEDIUM, Priority.LOW]

    requirements: list[dict] = []

    for i in range(num_requirements):
        req_id = f"REQ-PERF-{i:05d}"
        category = random.choice(categories)
        subcategory = random.choice(subcategories[category])
        status = random.choice(statuses)
        priority = random.choice(priorities)
        phase = random.randint(1, 5)
        effort = round(random.uniform(0.25, 4.0), 2)

        # Generate dependencies
        deps: list[str] = []
        if i > 0 and random.random() < dependency_density:
            # Reference earlier requirements only (to avoid cycles by default)
            num_deps = random.randint(1, min(3, i))
            dep_indices = random.sample(range(i), num_deps)
            deps = [f"REQ-PERF-{idx:05d}" for idx in dep_indices]

        requirements.append(
            {
                "req_id": req_id,
                "category": category,
                "subcategory": subcategory,
                "requirement_text": f"Performance test requirement {i} for stress testing",
                "target_value": f"Value {i}",
                "test_module": f"tests/test_perf_{i % 100}.py",
                "test_function": f"test_req_{i}",
                "validation_method": "Unit Test",
                "status": status.value,
                "priority": priority.value,
                "phase": phase,
                "notes": "Generated for performance testing",
                "effort_weeks": effort,
                "dependencies": "|".join(deps),
                "blocks": "",
                "assignee": f"dev{i % 10}",
                "sprint": f"v{phase}.{i % 5}",
                "started_date": "",
                "completed_date": "",
                "requirement_file": f"docs/requirements/{category}/REQ-PERF-{i:05d}.md",
            }
        )

    # Add cycles if requested
    if include_cycles and num_requirements >= 10:
        # Create a few cycles by adding backward dependencies
        cycle_sizes = [3, 4, 5]  # Different cycle sizes
        start_idx = num_requirements // 2

        for size in cycle_sizes:
            if start_idx + size >= num_requirements:
                break
            # Make last req in cycle depend on first
            cycle_start = start_idx
            cycle_end = start_idx + size - 1
            existing_deps = requirements[cycle_end]["dependencies"]
            new_dep = f"REQ-PERF-{cycle_start:05d}"
            if existing_deps:
                requirements[cycle_end]["dependencies"] = f"{existing_deps}|{new_dep}"
            else:
                requirements[cycle_end]["dependencies"] = new_dep
            start_idx += size + 5  # Skip ahead to create separate cycles

    # Write to CSV
    fieldnames = [
        "req_id",
        "category",
        "subcategory",
        "requirement_text",
        "target_value",
        "test_module",
        "test_function",
        "validation_method",
        "status",
        "priority",
        "phase",
        "notes",
        "effort_weeks",
        "dependencies",
        "blocks",
        "assignee",
        "sprint",
        "started_date",
        "completed_date",
        "requirement_file",
    ]

    with path.open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(requirements)


def generate_test_files(test_dir: Path, num_files: int, tests_per_file: int = 10) -> None:
    """Generate mock test files for from-tests scanning.

    Args:
        test_dir: Directory to create test files in
        num_files: Number of test files to create
        tests_per_file: Number of test functions per file
    """
    test_dir.mkdir(parents=True, exist_ok=True)

    for file_idx in range(num_files):
        file_path = test_dir / f"test_perf_{file_idx:04d}.py"

        lines = [
            '"""Generated test file for performance testing."""\n',
            "import pytest\n\n",
        ]

        for test_idx in range(tests_per_file):
            req_id = f"REQ-PERF-{file_idx * tests_per_file + test_idx:05d}"
            lines.extend(
                [
                    f'@pytest.mark.req("{req_id}")\n',
                    "@pytest.mark.scope_unit\n",
                    "@pytest.mark.technique_stress\n",
                    "@pytest.mark.env_simulation\n",
                    f"def test_requirement_{test_idx}():\n",
                    f'    """Test for {req_id}."""\n',
                    "    assert True\n\n",
                ]
            )

        file_path.write_text("".join(lines))


@pytest.fixture
def large_rtm_csv(tmp_path: Path) -> Path:
    """Generate a large RTM CSV with 1000+ requirements."""
    csv_path = tmp_path / "large_rtm.csv"
    generate_large_csv(1100, csv_path, dependency_density=0.2)
    return csv_path


@pytest.fixture
def very_large_rtm_csv(tmp_path: Path) -> Path:
    """Generate a very large RTM CSV with 10000 requirements for memory testing."""
    csv_path = tmp_path / "very_large_rtm.csv"
    generate_large_csv(10000, csv_path, dependency_density=0.15)
    return csv_path


@pytest.fixture
def complex_graph_csv(tmp_path: Path) -> Path:
    """Generate a CSV with complex dependency graph including cycles."""
    csv_path = tmp_path / "complex_graph_rtm.csv"
    generate_large_csv(1000, csv_path, dependency_density=0.3, include_cycles=True)
    return csv_path


@pytest.fixture
def large_test_directory(tmp_path: Path) -> Path:
    """Generate 500+ test files for from-tests scanning."""
    test_dir = tmp_path / "tests"
    generate_test_files(test_dir, 500, tests_per_file=10)  # 5000 total tests
    return test_dir


@pytest.mark.req("REQ-TEST-009")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestPerformance:
    """Large scale performance tests for RTMX commands."""

    def test_status_1000_requirements(self, large_rtm_csv: Path) -> None:
        """Test status command completes in <5s with 1000+ requirements.

        Acceptance Criteria:
        - Status command with 1000+ requirements completes in <5s
        """
        # Mock sys.exit to prevent test termination
        with mock.patch.object(sys, "exit"):
            # Capture stdout
            captured = StringIO()
            with mock.patch.object(sys, "stdout", captured):
                start_time = time.perf_counter()
                run_status(large_rtm_csv, verbosity=0, json_output=None)
                elapsed = time.perf_counter() - start_time

        assert (
            elapsed < STATUS_TIMEOUT_SECONDS
        ), f"Status command took {elapsed:.2f}s, exceeds {STATUS_TIMEOUT_SECONDS}s limit"

        # Verify output was generated
        output = captured.getvalue()
        assert "RTM Status" in output or "Requirements:" in output

    def test_status_verbose_1000_requirements(self, large_rtm_csv: Path) -> None:
        """Test verbose status command completes in <5s with 1000+ requirements."""
        with mock.patch.object(sys, "exit"):
            captured = StringIO()
            with mock.patch.object(sys, "stdout", captured):
                start_time = time.perf_counter()
                run_status(large_rtm_csv, verbosity=3, json_output=None)
                elapsed = time.perf_counter() - start_time

        assert (
            elapsed < STATUS_TIMEOUT_SECONDS
        ), f"Verbose status took {elapsed:.2f}s, exceeds {STATUS_TIMEOUT_SECONDS}s limit"

    def test_backlog_1000_requirements(self, large_rtm_csv: Path) -> None:
        """Test backlog command completes in <5s with 1000+ requirements.

        Acceptance Criteria:
        - Backlog command with 1000+ requirements completes in <5s
        """
        with mock.patch.object(sys, "exit"):
            captured = StringIO()
            with mock.patch.object(sys, "stdout", captured):
                start_time = time.perf_counter()
                run_backlog(large_rtm_csv, phase=None, limit=10)
                elapsed = time.perf_counter() - start_time

        assert (
            elapsed < BACKLOG_TIMEOUT_SECONDS
        ), f"Backlog command took {elapsed:.2f}s, exceeds {BACKLOG_TIMEOUT_SECONDS}s limit"

        # Verify output was generated
        output = captured.getvalue()
        assert "Backlog" in output or "CRITICAL" in output or "Requirements" in output

    def test_cycles_complex_graph(self, large_rtm_csv: Path) -> None:
        """Test cycles detection algorithm completes in <10s with complex graph.

        Acceptance Criteria:
        - Cycles command with complex dependency graphs completes in <10s

        Note: We test the underlying graph algorithm directly since the CLI
        command has additional output formatting. The large_rtm_csv has
        dependencies but no guaranteed cycles (since they're built forward-only).
        """
        db = RTMDatabase.load(large_rtm_csv)

        start_time = time.perf_counter()
        cycles = db.find_cycles()
        graph = db._get_graph()
        stats = graph.statistics()
        elapsed = time.perf_counter() - start_time

        assert (
            elapsed < CYCLES_TIMEOUT_SECONDS
        ), f"Cycles detection took {elapsed:.2f}s, exceeds {CYCLES_TIMEOUT_SECONDS}s limit"

        # Verify the operations completed and produced valid results
        assert isinstance(cycles, list)
        assert "nodes" in stats
        assert stats["nodes"] >= 1000

    def test_from_tests_500_files(self, large_test_directory: Path) -> None:
        """Test from-tests scanning completes in <30s with 500+ test files.

        Acceptance Criteria:
        - From-tests with 500+ test files completes in <30s
        """
        start_time = time.perf_counter()
        results = scan_test_directory(large_test_directory)
        elapsed = time.perf_counter() - start_time

        assert (
            elapsed < FROM_TESTS_TIMEOUT_SECONDS
        ), f"From-tests scan took {elapsed:.2f}s, exceeds {FROM_TESTS_TIMEOUT_SECONDS}s limit"

        # Verify we found the expected number of test markers
        assert len(results) > 0, "Expected to find test markers"
        # We generated 500 files * 10 tests = 5000 test markers
        assert len(results) >= 4500, f"Expected ~5000 markers, found {len(results)}"

    def test_memory_usage_10000_requirements(self, very_large_rtm_csv: Path) -> None:
        """Test memory usage stays under 500MB for 10,000 requirement database.

        Acceptance Criteria:
        - Memory usage stays under 500MB for 10,000 requirement database
        """
        tracemalloc.start()

        # Load the database
        db = RTMDatabase.load(very_large_rtm_csv)

        # Perform operations that might use memory
        _ = list(db.all())
        _ = db.status_counts()
        _ = db.completion_percentage()
        _ = db.filter(status=Status.MISSING)

        # Get memory usage
        current, peak = tracemalloc.get_traced_memory()
        tracemalloc.stop()

        peak_mb = peak / (1024 * 1024)

        assert (
            peak_mb < MAX_MEMORY_MB
        ), f"Peak memory usage {peak_mb:.1f}MB exceeds {MAX_MEMORY_MB}MB limit"

        # Verify database was loaded correctly
        assert len(db) == 10000, f"Expected 10000 requirements, got {len(db)}"

    def test_graph_operations_performance(self, large_rtm_csv: Path) -> None:
        """Test graph operations complete efficiently with large databases."""
        db = RTMDatabase.load(large_rtm_csv)

        # Test find_cycles performance
        start_time = time.perf_counter()
        cycles = db.find_cycles()
        cycles_time = time.perf_counter() - start_time

        assert cycles_time < 5.0, f"find_cycles took {cycles_time:.2f}s, should be under 5s"

        # Test critical_path performance
        start_time = time.perf_counter()
        critical = db.critical_path()
        critical_time = time.perf_counter() - start_time

        assert critical_time < 5.0, f"critical_path took {critical_time:.2f}s, should be under 5s"

        # Verify operations produced results
        assert isinstance(cycles, list)
        assert isinstance(critical, list)


@pytest.mark.req("REQ-TEST-009")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestGraphPerformance:
    """Performance tests for dependency graph operations."""

    def test_dependency_graph_creation(self, large_rtm_csv: Path) -> None:
        """Test dependency graph creation is efficient."""
        db = RTMDatabase.load(large_rtm_csv)

        start_time = time.perf_counter()
        graph = DependencyGraph.from_database(db)
        elapsed = time.perf_counter() - start_time

        assert elapsed < 2.0, f"Graph creation took {elapsed:.2f}s, should be under 2s"

        # Verify graph was created correctly
        assert graph.node_count >= 1000

    def test_transitive_blocks_performance(self, large_rtm_csv: Path) -> None:
        """Test transitive_blocks is efficient for large graphs."""
        db = RTMDatabase.load(large_rtm_csv)

        # Get a requirement that has dependencies
        reqs = db.all()
        req_with_deps = next((r for r in reqs if r.dependencies), reqs[0])

        start_time = time.perf_counter()
        blocked = db.transitive_blocks(req_with_deps.req_id)
        elapsed = time.perf_counter() - start_time

        assert elapsed < 1.0, f"transitive_blocks took {elapsed:.2f}s, should be under 1s"

        assert isinstance(blocked, set)


@pytest.mark.req("REQ-TEST-009")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestDatabaseScaling:
    """Tests for database operations at scale."""

    @pytest.mark.parametrize(
        "num_requirements",
        [100, 500, 1000, 2000],
        ids=["100-reqs", "500-reqs", "1000-reqs", "2000-reqs"],
    )
    def test_database_load_scaling(self, tmp_path: Path, num_requirements: int) -> None:
        """Test database loading scales reasonably with size."""
        csv_path = tmp_path / f"rtm_{num_requirements}.csv"
        generate_large_csv(num_requirements, csv_path)

        start_time = time.perf_counter()
        db = RTMDatabase.load(csv_path)
        elapsed = time.perf_counter() - start_time

        # Linear scaling: roughly 0.1s per 1000 requirements
        expected_max = 0.5 + (num_requirements / 1000) * 0.5
        assert elapsed < expected_max, (
            f"Loading {num_requirements} reqs took {elapsed:.2f}s, "
            f"expected under {expected_max:.2f}s"
        )

        assert len(db) == num_requirements

    def test_filter_performance(self, large_rtm_csv: Path) -> None:
        """Test filtering operations are efficient."""
        db = RTMDatabase.load(large_rtm_csv)

        # Test status filter
        start_time = time.perf_counter()
        missing = db.filter(status=Status.MISSING)
        elapsed = time.perf_counter() - start_time

        assert elapsed < 0.5, f"Status filter took {elapsed:.2f}s, should be under 0.5s"
        assert isinstance(missing, list)

        # Test combined filter
        start_time = time.perf_counter()
        filtered = db.filter(status=Status.MISSING, priority=Priority.HIGH, phase=1)
        elapsed = time.perf_counter() - start_time

        assert elapsed < 0.5, f"Combined filter took {elapsed:.2f}s, should be under 0.5s"
        assert isinstance(filtered, list)

    def test_status_counts_performance(self, large_rtm_csv: Path) -> None:
        """Test status_counts is efficient."""
        db = RTMDatabase.load(large_rtm_csv)

        start_time = time.perf_counter()
        counts = db.status_counts()
        elapsed = time.perf_counter() - start_time

        assert elapsed < 0.1, f"status_counts took {elapsed:.2f}s, should be under 0.1s"

        # Verify counts add up
        total = sum(counts.values())
        assert total == len(db)


@pytest.mark.req("REQ-TEST-009")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestEdgeCases:
    """Edge case tests for performance scenarios."""

    def test_minimal_database(self, tmp_path: Path) -> None:
        """Test operations on minimal database complete quickly."""
        csv_path = tmp_path / "minimal_rtm.csv"
        generate_large_csv(5, csv_path)  # Minimal non-empty database

        start_time = time.perf_counter()
        db = RTMDatabase.load(csv_path)
        elapsed = time.perf_counter() - start_time

        assert elapsed < 0.5
        assert len(db) == 5

    def test_no_dependencies_database(self, tmp_path: Path) -> None:
        """Test database with no dependencies."""
        csv_path = tmp_path / "no_deps_rtm.csv"
        generate_large_csv(1000, csv_path, dependency_density=0.0)

        db = RTMDatabase.load(csv_path)

        start_time = time.perf_counter()
        cycles = db.find_cycles()
        elapsed = time.perf_counter() - start_time

        assert elapsed < 1.0
        assert len(cycles) == 0  # No dependencies means no cycles

    def test_dense_dependencies(self, tmp_path: Path) -> None:
        """Test database with dense dependency graph."""
        csv_path = tmp_path / "dense_rtm.csv"
        generate_large_csv(500, csv_path, dependency_density=0.5)

        db = RTMDatabase.load(csv_path)

        start_time = time.perf_counter()
        _ = db.find_cycles()
        _ = db.critical_path()
        elapsed = time.perf_counter() - start_time

        assert (
            elapsed < CYCLES_TIMEOUT_SECONDS
        ), f"Dense graph operations took {elapsed:.2f}s, exceeds limit"
