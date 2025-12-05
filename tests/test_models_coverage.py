"""Comprehensive tests for rtmx.models module - achieving high coverage.

This test suite covers all classes and methods in models.py:
- Status enum and from_string variations
- Priority enum and from_string variations
- Requirement dataclass (from_dict, to_dict, properties, methods)
- RTMDatabase class (CRUD, filtering, statistics, graph operations)
"""

from pathlib import Path

import pytest

from rtmx import Priority, Requirement, RTMDatabase, Status
from rtmx.models import RequirementNotFoundError, RTMError, RTMValidationError


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestStatusEnum:
    """Tests for Status enum."""

    def test_status_from_string_with_hyphens(self):
        """Test parsing status with hyphens."""
        assert Status.from_string("NOT-STARTED") == Status.NOT_STARTED
        assert Status.from_string("not-started") == Status.NOT_STARTED

    def test_status_from_string_with_spaces(self):
        """Test parsing status with spaces."""
        assert Status.from_string("NOT STARTED") == Status.NOT_STARTED
        # Multiple spaces become multiple underscores, which won't match
        # So this defaults to MISSING
        assert Status.from_string("  NOT  STARTED  ") == Status.MISSING

    def test_status_from_string_mixed_case(self):
        """Test parsing status with mixed case."""
        assert Status.from_string("Complete") == Status.COMPLETE
        assert Status.from_string("pArTiAl") == Status.PARTIAL

    def test_status_from_string_defaults_to_missing(self):
        """Test unknown status defaults to MISSING."""
        assert Status.from_string("INVALID") == Status.MISSING
        assert Status.from_string("") == Status.MISSING
        assert Status.from_string("   ") == Status.MISSING


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestPriorityEnum:
    """Tests for Priority enum."""

    def test_priority_from_string_critical(self):
        """Test parsing CRITICAL priority."""
        assert Priority.from_string("CRITICAL") == Priority.P0
        assert Priority.from_string("critical") == Priority.P0
        assert Priority.from_string("  CRITICAL  ") == Priority.P0

    def test_priority_from_string_p0(self):
        """Test parsing P0 priority."""
        assert Priority.from_string("P0") == Priority.P0
        assert Priority.from_string("p0") == Priority.P0

    def test_priority_from_string_low(self):
        """Test parsing LOW priority."""
        assert Priority.from_string("LOW") == Priority.LOW
        assert Priority.from_string("low") == Priority.LOW

    def test_priority_from_string_defaults_to_medium(self):
        """Test unknown priority defaults to MEDIUM."""
        assert Priority.from_string("INVALID") == Priority.MEDIUM
        assert Priority.from_string("") == Priority.MEDIUM
        assert Priority.from_string("P5") == Priority.MEDIUM


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementDataclass:
    """Tests for Requirement dataclass."""

    def test_requirement_is_complete_true(self):
        """Test is_complete returns True for COMPLETE status."""
        req = Requirement(req_id="REQ-TEST-001", status=Status.COMPLETE)
        assert req.is_complete() is True

    def test_requirement_is_complete_false(self):
        """Test is_complete returns False for non-COMPLETE status."""
        req = Requirement(req_id="REQ-TEST-001", status=Status.PARTIAL)
        assert req.is_complete() is False

        req2 = Requirement(req_id="REQ-TEST-002", status=Status.MISSING)
        assert req2.is_complete() is False

    def test_requirement_is_blocked_true(self):
        """Test is_blocked returns True when dependencies are incomplete."""
        req1 = Requirement(req_id="REQ-A", status=Status.COMPLETE)
        req2 = Requirement(req_id="REQ-B", status=Status.MISSING)
        req3 = Requirement(
            req_id="REQ-C",
            dependencies={"REQ-A", "REQ-B"},
        )
        db = RTMDatabase([req1, req2, req3])

        assert req3.is_blocked(db) is True

    def test_requirement_is_blocked_false(self):
        """Test is_blocked returns False when all dependencies are complete."""
        req1 = Requirement(req_id="REQ-A", status=Status.COMPLETE)
        req2 = Requirement(req_id="REQ-B", status=Status.COMPLETE)
        req3 = Requirement(
            req_id="REQ-C",
            dependencies={"REQ-A", "REQ-B"},
        )
        db = RTMDatabase([req1, req2, req3])

        assert req3.is_blocked(db) is False

    def test_requirement_is_blocked_missing_dependency(self):
        """Test is_blocked handles missing dependencies gracefully."""
        req = Requirement(
            req_id="REQ-A",
            dependencies={"REQ-MISSING"},
        )
        db = RTMDatabase([req])

        # Should not raise, should return False for missing deps
        assert req.is_blocked(db) is False

    def test_requirement_id_property(self):
        """Test id property alias for req_id."""
        req = Requirement(req_id="REQ-TEST-001")
        assert req.id == "REQ-TEST-001"
        assert req.id == req.req_id

    def test_requirement_text_property(self):
        """Test text property alias for requirement_text."""
        req = Requirement(
            req_id="REQ-TEST-001",
            requirement_text="Test requirement",
        )
        assert req.text == "Test requirement"
        assert req.text == req.requirement_text

    def test_requirement_rationale_from_notes(self):
        """Test rationale property returns notes when extra.rationale not set."""
        req = Requirement(
            req_id="REQ-TEST-001",
            notes="This is the rationale",
        )
        assert req.rationale == "This is the rationale"

    def test_requirement_rationale_from_extra(self):
        """Test rationale property returns extra.rationale when set."""
        req = Requirement(
            req_id="REQ-TEST-001",
            notes="Notes",
            extra={"rationale": "Extra rationale"},
        )
        assert req.rationale == "Extra rationale"

    def test_requirement_acceptance_from_target_value(self):
        """Test acceptance property returns target_value when extra.acceptance not set."""
        req = Requirement(
            req_id="REQ-TEST-001",
            target_value="≥90%",
        )
        assert req.acceptance == "≥90%"

    def test_requirement_acceptance_from_extra(self):
        """Test acceptance property returns extra.acceptance when set."""
        req = Requirement(
            req_id="REQ-TEST-001",
            target_value="Target",
            extra={"acceptance": "Extra acceptance"},
        )
        assert req.acceptance == "Extra acceptance"

    def test_requirement_to_dict_complete(self):
        """Test to_dict with all fields populated."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="SOFTWARE",
            subcategory="ALGORITHM",
            requirement_text="Test requirement",
            target_value="100%",
            test_module="tests/test_foo.py",
            test_function="test_bar",
            validation_method="Unit Test",
            status=Status.COMPLETE,
            priority=Priority.HIGH,
            phase=1,
            notes="Test notes",
            effort_weeks=2.5,
            dependencies={"REQ-A", "REQ-B"},
            blocks={"REQ-C"},
            assignee="alice",
            sprint="v0.1",
            started_date="2025-01-01",
            completed_date="2025-01-15",
            requirement_file="docs/requirements/REQ-TEST-001.md",
            external_id="GH-123",
            extra={"custom_field": "custom_value"},
        )
        data = req.to_dict()

        assert data["req_id"] == "REQ-TEST-001"
        assert data["category"] == "SOFTWARE"
        assert data["subcategory"] == "ALGORITHM"
        assert data["requirement_text"] == "Test requirement"
        assert data["target_value"] == "100%"
        assert data["test_module"] == "tests/test_foo.py"
        assert data["test_function"] == "test_bar"
        assert data["validation_method"] == "Unit Test"
        assert data["status"] == "COMPLETE"
        assert data["priority"] == "HIGH"
        assert data["phase"] == 1
        assert data["notes"] == "Test notes"
        assert data["effort_weeks"] == 2.5
        assert "REQ-A" in data["dependencies"]
        assert "REQ-B" in data["dependencies"]
        assert data["blocks"] == "REQ-C"
        assert data["assignee"] == "alice"
        assert data["sprint"] == "v0.1"
        assert data["started_date"] == "2025-01-01"
        assert data["completed_date"] == "2025-01-15"
        assert data["requirement_file"] == "docs/requirements/REQ-TEST-001.md"
        assert data["external_id"] == "GH-123"
        assert data["custom_field"] == "custom_value"

    def test_requirement_to_dict_with_none_values(self):
        """Test to_dict handles None values for phase and effort_weeks."""
        req = Requirement(
            req_id="REQ-TEST-001",
            phase=None,
            effort_weeks=None,
        )
        data = req.to_dict()

        assert data["phase"] == ""
        assert data["effort_weeks"] == ""

    def test_requirement_from_dict_capitalized_keys(self):
        """Test from_dict handles capitalized keys (CSV header style)."""
        data = {
            "Req_ID": "REQ-TEST-001",
            "Category": "SOFTWARE",
            "Subcategory": "ALGORITHM",
            "Requirement_Text": "Test requirement",
            "Status": "COMPLETE",
            "Priority": "HIGH",
            "Phase": "2",
            "Dependencies": "REQ-A|REQ-B",
            "Blocks": "REQ-C",
        }
        req = Requirement.from_dict(data)

        assert req.req_id == "REQ-TEST-001"
        assert req.category == "SOFTWARE"
        assert req.subcategory == "ALGORITHM"
        assert req.requirement_text == "Test requirement"
        assert req.status == Status.COMPLETE
        assert req.priority == Priority.HIGH
        assert req.phase == 2
        assert req.dependencies == {"REQ-A", "REQ-B"}
        assert req.blocks == {"REQ-C"}

    def test_requirement_from_dict_with_extra_fields(self):
        """Test from_dict captures unknown fields in extra."""
        data = {
            "req_id": "REQ-TEST-001",
            "custom_field": "custom_value",
            "another_field": "another_value",
        }
        req = Requirement.from_dict(data)

        assert req.req_id == "REQ-TEST-001"
        assert req.extra["custom_field"] == "custom_value"
        assert req.extra["another_field"] == "another_value"

    def test_requirement_from_dict_invalid_phase(self):
        """Test from_dict handles invalid phase values."""
        data = {
            "req_id": "REQ-TEST-001",
            "phase": "invalid",
        }
        req = Requirement.from_dict(data)

        assert req.phase is None

    def test_requirement_from_dict_invalid_effort(self):
        """Test from_dict handles invalid effort_weeks values."""
        data = {
            "req_id": "REQ-TEST-001",
            "effort_weeks": "invalid",
        }
        req = Requirement.from_dict(data)

        assert req.effort_weeks is None

    def test_requirement_from_dict_empty_phase(self):
        """Test from_dict handles empty phase values."""
        data = {
            "req_id": "REQ-TEST-001",
            "phase": "",
        }
        req = Requirement.from_dict(data)

        assert req.phase is None


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseCRUD:
    """Tests for RTMDatabase CRUD operations."""

    def test_database_init(self):
        """Test database initialization."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B")
        db = RTMDatabase([req1, req2])

        assert len(db) == 2
        assert "REQ-A" in db
        assert "REQ-B" in db

    def test_database_add(self):
        """Test adding a requirement."""
        db = RTMDatabase([])
        req = Requirement(req_id="REQ-NEW")

        db.add(req)

        assert len(db) == 1
        assert "REQ-NEW" in db
        assert db.get("REQ-NEW") == req

    def test_database_add_duplicate_raises_error(self):
        """Test adding duplicate requirement raises error."""
        req1 = Requirement(req_id="REQ-A")
        db = RTMDatabase([req1])

        req2 = Requirement(req_id="REQ-A")
        with pytest.raises(RTMError, match="already exists"):
            db.add(req2)

    def test_database_remove(self):
        """Test removing a requirement."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        removed = db.remove("REQ-A")

        assert removed == req
        assert len(db) == 0
        assert "REQ-A" not in db

    def test_database_remove_nonexistent_raises_error(self):
        """Test removing non-existent requirement raises error."""
        db = RTMDatabase([])

        with pytest.raises(RequirementNotFoundError):
            db.remove("REQ-MISSING")

    def test_database_update_string_status(self):
        """Test updating status with string value."""
        req = Requirement(req_id="REQ-A", status=Status.MISSING)
        db = RTMDatabase([req])

        db.update("REQ-A", status="COMPLETE")

        assert db.get("REQ-A").status == Status.COMPLETE

    def test_database_update_string_priority(self):
        """Test updating priority with string value."""
        req = Requirement(req_id="REQ-A", priority=Priority.MEDIUM)
        db = RTMDatabase([req])

        db.update("REQ-A", priority="HIGH")

        assert db.get("REQ-A").priority == Priority.HIGH

    def test_database_update_dependencies(self):
        """Test updating dependencies with string value."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        db.update("REQ-A", dependencies="REQ-B|REQ-C")

        assert db.get("REQ-A").dependencies == {"REQ-B", "REQ-C"}

    def test_database_update_blocks(self):
        """Test updating blocks with string value."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        db.update("REQ-A", blocks="REQ-B|REQ-C")

        assert db.get("REQ-A").blocks == {"REQ-B", "REQ-C"}

    def test_database_update_extra_field(self):
        """Test updating non-existent field adds to extra."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        db.update("REQ-A", custom_field="custom_value")

        assert db.get("REQ-A").extra["custom_field"] == "custom_value"

    def test_database_update_invalidates_graph(self):
        """Test update invalidates cached dependency graph."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        # Access graph to cache it
        db._get_graph()
        assert db._graph is not None

        # Update should invalidate
        db.update("REQ-A", status=Status.COMPLETE)
        assert db._graph is None

    def test_database_all(self):
        """Test getting all requirements."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B")
        db = RTMDatabase([req1, req2])

        all_reqs = db.all()

        assert len(all_reqs) == 2
        assert req1 in all_reqs
        assert req2 in all_reqs

    def test_database_contains(self):
        """Test __contains__ method."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        assert "REQ-A" in db
        assert "REQ-B" not in db


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseFiltering:
    """Tests for RTMDatabase filtering operations."""

    def test_filter_by_priority(self):
        """Test filtering by priority."""
        req1 = Requirement(req_id="REQ-A", priority=Priority.HIGH)
        req2 = Requirement(req_id="REQ-B", priority=Priority.LOW)
        req3 = Requirement(req_id="REQ-C", priority=Priority.HIGH)
        db = RTMDatabase([req1, req2, req3])

        high_priority = db.filter(priority=Priority.HIGH)

        assert len(high_priority) == 2
        assert all(r.priority == Priority.HIGH for r in high_priority)

    def test_filter_by_subcategory(self):
        """Test filtering by subcategory."""
        req1 = Requirement(req_id="REQ-A", subcategory="ALGORITHM")
        req2 = Requirement(req_id="REQ-B", subcategory="UI")
        req3 = Requirement(req_id="REQ-C", subcategory="ALGORITHM")
        db = RTMDatabase([req1, req2, req3])

        algorithms = db.filter(subcategory="ALGORITHM")

        assert len(algorithms) == 2
        assert all(r.subcategory == "ALGORITHM" for r in algorithms)

    def test_filter_by_has_test_true(self):
        """Test filtering by has_test=True."""
        req1 = Requirement(
            req_id="REQ-A",
            test_module="tests/test_a.py",
            test_function="test_a",
        )
        req2 = Requirement(req_id="REQ-B")
        db = RTMDatabase([req1, req2])

        with_tests = db.filter(has_test=True)

        assert len(with_tests) == 1
        assert with_tests[0].req_id == "REQ-A"

    def test_filter_by_has_test_false(self):
        """Test filtering by has_test=False."""
        req1 = Requirement(
            req_id="REQ-A",
            test_module="tests/test_a.py",
            test_function="test_a",
        )
        req2 = Requirement(req_id="REQ-B")
        db = RTMDatabase([req1, req2])

        without_tests = db.filter(has_test=False)

        assert len(without_tests) == 1
        assert without_tests[0].req_id == "REQ-B"

    def test_filter_multiple_criteria(self):
        """Test filtering with multiple criteria."""
        req1 = Requirement(
            req_id="REQ-A",
            category="SOFTWARE",
            status=Status.COMPLETE,
            priority=Priority.HIGH,
        )
        req2 = Requirement(
            req_id="REQ-B",
            category="SOFTWARE",
            status=Status.MISSING,
            priority=Priority.HIGH,
        )
        req3 = Requirement(
            req_id="REQ-C",
            category="TESTING",
            status=Status.COMPLETE,
            priority=Priority.HIGH,
        )
        db = RTMDatabase([req1, req2, req3])

        filtered = db.filter(
            category="SOFTWARE",
            status=Status.COMPLETE,
            priority=Priority.HIGH,
        )

        assert len(filtered) == 1
        assert filtered[0].req_id == "REQ-A"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseStatistics:
    """Tests for RTMDatabase statistics operations."""

    def test_status_counts_all_statuses(self):
        """Test status_counts includes all status types."""
        req1 = Requirement(req_id="REQ-A", status=Status.COMPLETE)
        req2 = Requirement(req_id="REQ-B", status=Status.PARTIAL)
        req3 = Requirement(req_id="REQ-C", status=Status.MISSING)
        req4 = Requirement(req_id="REQ-D", status=Status.NOT_STARTED)
        db = RTMDatabase([req1, req2, req3, req4])

        counts = db.status_counts()

        assert counts[Status.COMPLETE] == 1
        assert counts[Status.PARTIAL] == 1
        assert counts[Status.MISSING] == 1
        assert counts[Status.NOT_STARTED] == 1

    def test_status_counts_empty_database(self):
        """Test status_counts on empty database."""
        db = RTMDatabase([])

        counts = db.status_counts()

        assert counts[Status.COMPLETE] == 0
        assert counts[Status.PARTIAL] == 0
        assert counts[Status.MISSING] == 0
        assert counts[Status.NOT_STARTED] == 0

    def test_completion_percentage_all_complete(self):
        """Test completion_percentage with all complete requirements."""
        req1 = Requirement(req_id="REQ-A", status=Status.COMPLETE)
        req2 = Requirement(req_id="REQ-B", status=Status.COMPLETE)
        db = RTMDatabase([req1, req2])

        pct = db.completion_percentage()

        assert pct == 100.0

    def test_completion_percentage_none_complete(self):
        """Test completion_percentage with no complete requirements."""
        req1 = Requirement(req_id="REQ-A", status=Status.MISSING)
        req2 = Requirement(req_id="REQ-B", status=Status.NOT_STARTED)
        db = RTMDatabase([req1, req2])

        pct = db.completion_percentage()

        assert pct == 0.0

    def test_completion_percentage_partial_counts_half(self):
        """Test completion_percentage counts PARTIAL as 50%."""
        req1 = Requirement(req_id="REQ-A", status=Status.PARTIAL)
        req2 = Requirement(req_id="REQ-B", status=Status.PARTIAL)
        db = RTMDatabase([req1, req2])

        pct = db.completion_percentage()

        assert pct == 50.0

    def test_completion_percentage_mixed(self):
        """Test completion_percentage with mixed statuses."""
        # 1 COMPLETE, 1 PARTIAL, 2 MISSING
        # (1 + 0.5) / 4 * 100 = 37.5%
        req1 = Requirement(req_id="REQ-A", status=Status.COMPLETE)
        req2 = Requirement(req_id="REQ-B", status=Status.PARTIAL)
        req3 = Requirement(req_id="REQ-C", status=Status.MISSING)
        req4 = Requirement(req_id="REQ-D", status=Status.NOT_STARTED)
        db = RTMDatabase([req1, req2, req3, req4])

        pct = db.completion_percentage()

        assert pct == 37.5

    def test_completion_percentage_empty_database(self):
        """Test completion_percentage on empty database."""
        db = RTMDatabase([])

        pct = db.completion_percentage()

        assert pct == 0.0


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabasePersistence:
    """Tests for RTMDatabase save/load operations."""

    def test_save_to_path(self, tmp_path: Path):
        """Test saving database to specified path."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="SOFTWARE",
            status=Status.COMPLETE,
        )
        db = RTMDatabase([req])

        save_path = tmp_path / "test_rtm.csv"
        db.save(save_path)

        assert save_path.exists()
        assert db.path == save_path

    def test_save_without_path_raises_error(self):
        """Test saving database without path raises error."""
        req = Requirement(req_id="REQ-TEST-001")
        db = RTMDatabase([req], path=None)

        with pytest.raises(RTMError, match="No save path specified"):
            db.save()

    def test_save_uses_original_path(self, tmp_path: Path):
        """Test save uses original load path when no path specified."""
        req = Requirement(req_id="REQ-TEST-001")
        original_path = tmp_path / "original.csv"
        db = RTMDatabase([req], path=original_path)

        db.save()

        assert original_path.exists()

    def test_load_nonexistent_file_raises_error(self, tmp_path: Path):
        """Test loading non-existent file raises error."""
        nonexistent = tmp_path / "nonexistent.csv"

        with pytest.raises(RTMError, match="RTM database not found"):
            RTMDatabase.load(nonexistent)

    def test_path_property(self, tmp_path: Path):
        """Test path property returns source file path."""
        path = tmp_path / "test.csv"
        db = RTMDatabase([], path=path)

        assert db.path == path

    def test_round_trip_save_load(self, tmp_path: Path):
        """Test saving and loading preserves data."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="SOFTWARE",
            requirement_text="Test requirement",
            status=Status.PARTIAL,
            priority=Priority.HIGH,
            dependencies={"REQ-A", "REQ-B"},
        )
        db = RTMDatabase([req])

        save_path = tmp_path / "test_rtm.csv"
        db.save(save_path)

        loaded_db = RTMDatabase.load(save_path)

        assert len(loaded_db) == 1
        loaded_req = loaded_db.get("REQ-TEST-001")
        assert loaded_req.req_id == "REQ-TEST-001"
        assert loaded_req.category == "SOFTWARE"
        assert loaded_req.requirement_text == "Test requirement"
        assert loaded_req.status == Status.PARTIAL
        assert loaded_req.priority == Priority.HIGH
        assert loaded_req.dependencies == {"REQ-A", "REQ-B"}


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseGraphOperations:
    """Tests for RTMDatabase graph delegation operations."""

    def test_find_cycles_no_cycles(self):
        """Test find_cycles returns empty list when no cycles exist."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B", dependencies={"REQ-A"})
        db = RTMDatabase([req1, req2])

        cycles = db.find_cycles()

        assert cycles == []

    def test_transitive_blocks(self):
        """Test transitive_blocks returns all blocked requirements.

        Note: The graph is built from dependencies, not blocks.
        If A blocks B, then B should depend on A for the graph to work.
        """
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B", dependencies={"REQ-A"})
        req3 = Requirement(req_id="REQ-C", dependencies={"REQ-B"})
        db = RTMDatabase([req1, req2, req3])

        # REQ-A blocks REQ-B (which depends on it)
        # REQ-B blocks REQ-C (which depends on it)
        # So REQ-A transitively blocks both REQ-B and REQ-C
        blocked = db.transitive_blocks("REQ-A")

        assert "REQ-B" in blocked
        assert "REQ-C" in blocked

    def test_critical_path(self):
        """Test critical_path returns requirements sorted by blocking count."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B", dependencies={"REQ-A"})
        req3 = Requirement(req_id="REQ-C", dependencies={"REQ-B"})
        db = RTMDatabase([req1, req2, req3])

        path = db.critical_path()

        # Critical path returns only nodes that block at least one other
        # REQ-A blocks 2 (REQ-B and REQ-C transitively)
        # REQ-B blocks 1 (REQ-C)
        # REQ-C blocks 0 (not included)
        assert len(path) == 2
        assert path[0] == "REQ-A"  # Blocks the most
        assert path[1] == "REQ-B"  # Blocks one

    def test_graph_caching(self):
        """Test dependency graph is cached."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        # First call creates graph
        graph1 = db._get_graph()
        # Second call returns same instance
        graph2 = db._get_graph()

        assert graph1 is graph2

    def test_add_invalidates_graph(self):
        """Test add invalidates cached graph."""
        req1 = Requirement(req_id="REQ-A")
        db = RTMDatabase([req1])

        # Cache the graph
        db._get_graph()
        assert db._graph is not None

        # Add should invalidate
        req2 = Requirement(req_id="REQ-B")
        db.add(req2)

        assert db._graph is None

    def test_remove_invalidates_graph(self):
        """Test remove invalidates cached graph."""
        req = Requirement(req_id="REQ-A")
        db = RTMDatabase([req])

        # Cache the graph
        db._get_graph()
        assert db._graph is not None

        # Remove should invalidate
        db.remove("REQ-A")

        assert db._graph is None


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseValidation:
    """Tests for RTMDatabase validation delegation operations."""

    def test_validate(self, core_rtm_path: Path):
        """Test validate delegates to validation module."""
        db = RTMDatabase.load(core_rtm_path)

        errors = db.validate()

        # Should return a list (may be empty or have errors)
        assert isinstance(errors, list)

    def test_check_reciprocity(self, core_rtm_path: Path):
        """Test check_reciprocity delegates to validation module."""
        db = RTMDatabase.load(core_rtm_path)

        issues = db.check_reciprocity()

        # Should return a list of tuples
        assert isinstance(issues, list)

    def test_fix_reciprocity(self, core_rtm_path: Path):
        """Test fix_reciprocity delegates to validation module."""
        db = RTMDatabase.load(core_rtm_path)

        fixed_count = db.fix_reciprocity()

        # Should return an integer
        assert isinstance(fixed_count, int)
        assert fixed_count >= 0


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMExceptions:
    """Tests for RTM exception hierarchy."""

    def test_rtm_error_base_exception(self):
        """Test RTMError is base exception."""
        err = RTMError("Test error")
        assert isinstance(err, Exception)
        assert str(err) == "Test error"

    def test_requirement_not_found_error(self):
        """Test RequirementNotFoundError inherits from RTMError."""
        err = RequirementNotFoundError("REQ-TEST not found")
        assert isinstance(err, RTMError)
        assert isinstance(err, Exception)
        assert str(err) == "REQ-TEST not found"

    def test_rtm_validation_error(self):
        """Test RTMValidationError inherits from RTMError."""
        err = RTMValidationError("Validation failed")
        assert isinstance(err, RTMError)
        assert isinstance(err, Exception)
        assert str(err) == "Validation failed"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseIterationProtocol:
    """Tests for RTMDatabase iteration and collection protocol."""

    def test_len(self):
        """Test __len__ returns number of requirements."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B")
        req3 = Requirement(req_id="REQ-C")
        db = RTMDatabase([req1, req2, req3])

        assert len(db) == 3

    def test_len_empty(self):
        """Test __len__ on empty database."""
        db = RTMDatabase([])

        assert len(db) == 0

    def test_iter(self):
        """Test __iter__ yields all requirements."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B")
        db = RTMDatabase([req1, req2])

        reqs = list(db)

        assert len(reqs) == 2
        assert req1 in reqs
        assert req2 in reqs

    def test_iter_order_independent(self):
        """Test iteration produces all requirements regardless of order."""
        reqs = [Requirement(req_id=f"REQ-{i}") for i in range(10)]
        db = RTMDatabase(reqs)

        iterated = list(db)

        assert len(iterated) == 10
        for req in reqs:
            assert req in iterated

    def test_get_error_message_helpful(self):
        """Test get error message shows available requirement IDs."""
        req1 = Requirement(req_id="REQ-A")
        req2 = Requirement(req_id="REQ-B")
        req3 = Requirement(req_id="REQ-C")
        db = RTMDatabase([req1, req2, req3])

        with pytest.raises(RequirementNotFoundError) as exc_info:
            db.get("REQ-MISSING")

        error_msg = str(exc_info.value)
        assert "REQ-MISSING" in error_msg
        assert "Available:" in error_msg
