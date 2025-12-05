"""Additional comprehensive tests for rtmx.models module to increase coverage.

This test suite focuses on uncovered lines in models.py:
- Import variations and edge cases
- Property accessors with various data states
- Exception handling paths
- RTMDatabase edge cases
- Graph operations delegation
- Validation delegation
"""

from pathlib import Path

import pytest

from rtmx import Priority, Requirement, RTMDatabase, Status
from rtmx.models import RequirementNotFoundError, RTMError, RTMValidationError


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestModuleExceptions:
    """Tests for module-level exception classes."""

    def test_rtm_error_is_exception(self):
        """Test RTMError is an Exception subclass."""
        err = RTMError("base error")
        assert isinstance(err, Exception)
        assert str(err) == "base error"

    def test_rtm_error_can_be_raised(self):
        """Test RTMError can be raised and caught."""
        with pytest.raises(RTMError):
            raise RTMError("test error")

    def test_requirement_not_found_error_is_rtm_error(self):
        """Test RequirementNotFoundError inherits from RTMError."""
        err = RequirementNotFoundError("not found")
        assert isinstance(err, RTMError)
        assert isinstance(err, Exception)

    def test_requirement_not_found_error_can_be_raised(self):
        """Test RequirementNotFoundError can be raised and caught."""
        with pytest.raises(RequirementNotFoundError):
            raise RequirementNotFoundError("REQ-X not found")

    def test_rtm_validation_error_is_rtm_error(self):
        """Test RTMValidationError inherits from RTMError."""
        err = RTMValidationError("validation failed")
        assert isinstance(err, RTMError)
        assert isinstance(err, Exception)

    def test_rtm_validation_error_can_be_raised(self):
        """Test RTMValidationError can be raised and caught."""
        with pytest.raises(RTMValidationError):
            raise RTMValidationError("validation failed")


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementProperties:
    """Tests for Requirement property accessors."""

    def test_id_property_returns_req_id(self):
        """Test id property returns req_id value."""
        req = Requirement(req_id="REQ-PROP-001")
        assert req.id == "REQ-PROP-001"

    def test_text_property_returns_requirement_text(self):
        """Test text property returns requirement_text value."""
        req = Requirement(
            req_id="REQ-PROP-002",
            requirement_text="Test text",
        )
        assert req.text == "Test text"

    def test_text_property_returns_empty_string_when_not_set(self):
        """Test text property returns empty string when requirement_text not set."""
        req = Requirement(req_id="REQ-PROP-003")
        assert req.text == ""

    def test_rationale_property_returns_notes_when_no_extra(self):
        """Test rationale returns notes when extra.rationale not present."""
        req = Requirement(
            req_id="REQ-PROP-004",
            notes="Rationale from notes",
        )
        assert req.rationale == "Rationale from notes"

    def test_rationale_property_returns_empty_when_no_notes(self):
        """Test rationale returns empty string when no notes."""
        req = Requirement(req_id="REQ-PROP-005")
        # Should access extra dict which returns empty string for missing key
        assert req.rationale == ""

    def test_acceptance_property_returns_target_value_when_no_extra(self):
        """Test acceptance returns target_value when extra.acceptance not present."""
        req = Requirement(
            req_id="REQ-PROP-006",
            target_value="≥95%",
        )
        assert req.acceptance == "≥95%"

    def test_acceptance_property_returns_empty_when_no_target_value(self):
        """Test acceptance returns empty string when no target_value."""
        req = Requirement(req_id="REQ-PROP-007")
        # Should access extra dict which returns empty string for missing key
        assert req.acceptance == ""


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementIsBlocked:
    """Tests for Requirement.is_blocked method edge cases."""

    def test_is_blocked_with_empty_dependencies(self):
        """Test is_blocked returns False when no dependencies."""
        req = Requirement(req_id="REQ-BLOCK-001")
        db = RTMDatabase([req])
        assert req.is_blocked(db) is False

    def test_is_blocked_with_missing_dependency_ignored(self):
        """Test is_blocked ignores missing dependencies (RequirementNotFoundError)."""
        req = Requirement(
            req_id="REQ-BLOCK-002",
            dependencies={"REQ-NONEXISTENT"},
        )
        db = RTMDatabase([req])
        # Should not raise error, should return False
        assert req.is_blocked(db) is False

    def test_is_blocked_with_partial_dependency(self):
        """Test is_blocked returns True when dependency is PARTIAL."""
        dep = Requirement(req_id="REQ-DEP", status=Status.PARTIAL)
        req = Requirement(
            req_id="REQ-BLOCKED",
            dependencies={"REQ-DEP"},
        )
        db = RTMDatabase([dep, req])
        assert req.is_blocked(db) is True

    def test_is_blocked_with_not_started_dependency(self):
        """Test is_blocked returns True when dependency is NOT_STARTED."""
        dep = Requirement(req_id="REQ-DEP", status=Status.NOT_STARTED)
        req = Requirement(
            req_id="REQ-BLOCKED",
            dependencies={"REQ-DEP"},
        )
        db = RTMDatabase([dep, req])
        assert req.is_blocked(db) is True


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementIsComplete:
    """Tests for Requirement.is_complete method."""

    def test_is_complete_with_complete_status(self):
        """Test is_complete returns True only for COMPLETE status."""
        req = Requirement(req_id="REQ-COMP-001", status=Status.COMPLETE)
        assert req.is_complete() is True

    def test_is_complete_with_partial_status(self):
        """Test is_complete returns False for PARTIAL status."""
        req = Requirement(req_id="REQ-COMP-002", status=Status.PARTIAL)
        assert req.is_complete() is False

    def test_is_complete_with_not_started_status(self):
        """Test is_complete returns False for NOT_STARTED status."""
        req = Requirement(req_id="REQ-COMP-003", status=Status.NOT_STARTED)
        assert req.is_complete() is False


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseExists:
    """Tests for RTMDatabase.exists method."""

    def test_exists_returns_true_for_existing(self):
        """Test exists returns True for existing requirement."""
        req = Requirement(req_id="REQ-EXISTS-001")
        db = RTMDatabase([req])
        assert db.exists("REQ-EXISTS-001") is True

    def test_exists_returns_false_for_nonexistent(self):
        """Test exists returns False for non-existent requirement."""
        db = RTMDatabase([])
        assert db.exists("REQ-NONEXISTENT") is False


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseUpdate:
    """Tests for RTMDatabase.update edge cases."""

    def test_update_nonexistent_requirement_raises_error(self):
        """Test updating non-existent requirement raises RequirementNotFoundError."""
        db = RTMDatabase([])
        with pytest.raises(RequirementNotFoundError):
            db.update("REQ-NONEXISTENT", status=Status.COMPLETE)

    def test_update_with_status_enum(self):
        """Test update with Status enum (not string)."""
        req = Requirement(req_id="REQ-UPDATE-001", status=Status.MISSING)
        db = RTMDatabase([req])
        db.update("REQ-UPDATE-001", status=Status.COMPLETE)
        assert db.get("REQ-UPDATE-001").status == Status.COMPLETE

    def test_update_with_priority_enum(self):
        """Test update with Priority enum (not string)."""
        req = Requirement(req_id="REQ-UPDATE-002", priority=Priority.MEDIUM)
        db = RTMDatabase([req])
        db.update("REQ-UPDATE-002", priority=Priority.P0)
        assert db.get("REQ-UPDATE-002").priority == Priority.P0

    def test_update_with_dependencies_set(self):
        """Test update with dependencies as set (not string)."""
        req = Requirement(req_id="REQ-UPDATE-003")
        db = RTMDatabase([req])
        db.update("REQ-UPDATE-003", dependencies={"REQ-A", "REQ-B"})
        assert db.get("REQ-UPDATE-003").dependencies == {"REQ-A", "REQ-B"}

    def test_update_with_blocks_set(self):
        """Test update with blocks as set (not string)."""
        req = Requirement(req_id="REQ-UPDATE-004")
        db = RTMDatabase([req])
        db.update("REQ-UPDATE-004", blocks={"REQ-X", "REQ-Y"})
        assert db.get("REQ-UPDATE-004").blocks == {"REQ-X", "REQ-Y"}

    def test_update_known_field(self):
        """Test update of known field updates the requirement attribute."""
        req = Requirement(req_id="REQ-UPDATE-005", notes="old notes")
        db = RTMDatabase([req])
        db.update("REQ-UPDATE-005", notes="new notes")
        assert db.get("REQ-UPDATE-005").notes == "new notes"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseRemove:
    """Tests for RTMDatabase.remove method."""

    def test_remove_returns_removed_requirement(self):
        """Test remove returns the removed requirement."""
        req = Requirement(req_id="REQ-REMOVE-001")
        db = RTMDatabase([req])
        removed = db.remove("REQ-REMOVE-001")
        assert removed == req
        assert removed.req_id == "REQ-REMOVE-001"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseAdd:
    """Tests for RTMDatabase.add method."""

    def test_add_updates_length(self):
        """Test add increases database length."""
        db = RTMDatabase([])
        assert len(db) == 0
        db.add(Requirement(req_id="REQ-ADD-001"))
        assert len(db) == 1


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseFilterEdgeCases:
    """Tests for RTMDatabase.filter edge cases."""

    def test_filter_returns_empty_list_when_no_matches(self):
        """Test filter returns empty list when no requirements match."""
        req = Requirement(req_id="REQ-FILTER-001", status=Status.COMPLETE)
        db = RTMDatabase([req])
        results = db.filter(status=Status.MISSING)
        assert results == []

    def test_filter_with_no_criteria_returns_all(self):
        """Test filter with no criteria returns all requirements."""
        req1 = Requirement(req_id="REQ-FILTER-002")
        req2 = Requirement(req_id="REQ-FILTER-003")
        db = RTMDatabase([req1, req2])
        results = db.filter()
        assert len(results) == 2


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseAll:
    """Tests for RTMDatabase.all method."""

    def test_all_returns_list(self):
        """Test all returns a list of requirements."""
        req1 = Requirement(req_id="REQ-ALL-001")
        req2 = Requirement(req_id="REQ-ALL-002")
        db = RTMDatabase([req1, req2])
        all_reqs = db.all()
        assert isinstance(all_reqs, list)
        assert len(all_reqs) == 2

    def test_all_returns_empty_list_for_empty_db(self):
        """Test all returns empty list for empty database."""
        db = RTMDatabase([])
        assert db.all() == []


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseContains:
    """Tests for RTMDatabase.__contains__ method."""

    def test_contains_with_in_operator(self):
        """Test __contains__ works with 'in' operator."""
        req = Requirement(req_id="REQ-CONTAINS-001")
        db = RTMDatabase([req])
        assert "REQ-CONTAINS-001" in db

    def test_not_contains_with_in_operator(self):
        """Test __contains__ returns False for non-existent."""
        db = RTMDatabase([])
        assert "REQ-NONEXISTENT" not in db


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseGraphDelegation:
    """Tests for RTMDatabase graph delegation methods."""

    def test_get_graph_creates_graph(self):
        """Test _get_graph creates DependencyGraph."""
        req = Requirement(req_id="REQ-GRAPH-001")
        db = RTMDatabase([req])
        graph = db._get_graph()
        assert graph is not None
        assert db._graph is graph

    def test_find_cycles_delegates_to_graph(self):
        """Test find_cycles delegates to DependencyGraph."""
        req = Requirement(req_id="REQ-CYCLE-001")
        db = RTMDatabase([req])
        cycles = db.find_cycles()
        assert isinstance(cycles, list)

    def test_transitive_blocks_delegates_to_graph(self):
        """Test transitive_blocks delegates to DependencyGraph."""
        req = Requirement(req_id="REQ-TB-001")
        db = RTMDatabase([req])
        blocked = db.transitive_blocks("REQ-TB-001")
        assert isinstance(blocked, set)

    def test_critical_path_delegates_to_graph(self):
        """Test critical_path delegates to DependencyGraph."""
        req = Requirement(req_id="REQ-CP-001")
        db = RTMDatabase([req])
        path = db.critical_path()
        assert isinstance(path, list)


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseValidationDelegation:
    """Tests for RTMDatabase validation delegation methods."""

    def test_validate_delegates_to_validation_module(self):
        """Test validate delegates to validation.validate_schema."""
        req = Requirement(req_id="REQ-VAL-001")
        db = RTMDatabase([req])
        errors = db.validate()
        assert isinstance(errors, list)

    def test_check_reciprocity_delegates_to_validation_module(self):
        """Test check_reciprocity delegates to validation module."""
        req = Requirement(req_id="REQ-RECIP-001")
        db = RTMDatabase([req])
        issues = db.check_reciprocity()
        assert isinstance(issues, list)

    def test_fix_reciprocity_delegates_to_validation_module(self):
        """Test fix_reciprocity delegates to validation module."""
        req = Requirement(req_id="REQ-FIX-001")
        db = RTMDatabase([req])
        count = db.fix_reciprocity()
        assert isinstance(count, int)
        assert count >= 0


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseStatusCounts:
    """Tests for RTMDatabase.status_counts edge cases."""

    def test_status_counts_initializes_all_statuses(self):
        """Test status_counts includes all Status enum values."""
        db = RTMDatabase([])
        counts = db.status_counts()
        # All statuses should be present
        for status in Status:
            assert status in counts
            assert counts[status] == 0


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabasePath:
    """Tests for RTMDatabase.path property."""

    def test_path_property_returns_none_when_not_loaded(self):
        """Test path property returns None when not loaded from file."""
        db = RTMDatabase([])
        assert db.path is None

    def test_path_property_returns_path_when_provided(self):
        """Test path property returns provided path."""
        path = Path("/tmp/test.csv")
        db = RTMDatabase([], path=path)
        assert db.path == path


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementFromDictEdgeCases:
    """Tests for Requirement.from_dict edge cases."""

    def test_from_dict_with_phase_as_string_phase_header(self):
        """Test from_dict handles phase value being the string 'phase'."""
        data = {
            "req_id": "REQ-PHASE-001",
            "phase": "phase",  # Header value, should be ignored
        }
        req = Requirement.from_dict(data)
        assert req.phase is None

    def test_from_dict_with_empty_effort_weeks(self):
        """Test from_dict handles empty effort_weeks."""
        data = {
            "req_id": "REQ-EFFORT-001",
            "effort_weeks": "",
        }
        req = Requirement.from_dict(data)
        assert req.effort_weeks is None

    def test_from_dict_with_none_effort_weeks(self):
        """Test from_dict handles None effort_weeks."""
        data = {
            "req_id": "REQ-EFFORT-002",
            "effort_weeks": None,
        }
        req = Requirement.from_dict(data)
        assert req.effort_weeks is None

    def test_from_dict_with_valid_float_effort(self):
        """Test from_dict handles valid float effort_weeks."""
        data = {
            "req_id": "REQ-EFFORT-003",
            "effort_weeks": "2.5",
        }
        req = Requirement.from_dict(data)
        assert req.effort_weeks == 2.5

    def test_from_dict_with_type_error_in_phase(self):
        """Test from_dict handles TypeError in phase conversion."""
        data = {
            "req_id": "REQ-PHASE-002",
            "phase": {"invalid": "dict"},  # Will cause TypeError
        }
        req = Requirement.from_dict(data)
        assert req.phase is None

    def test_from_dict_with_type_error_in_effort(self):
        """Test from_dict handles TypeError in effort conversion."""
        data = {
            "req_id": "REQ-EFFORT-004",
            "effort_weeks": {"invalid": "dict"},  # Will cause TypeError
        }
        req = Requirement.from_dict(data)
        assert req.effort_weeks is None


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementToDictEdgeCases:
    """Tests for Requirement.to_dict edge cases."""

    def test_to_dict_preserves_extra_fields(self):
        """Test to_dict includes extra fields."""
        req = Requirement(
            req_id="REQ-EXTRA-001",
            extra={"custom1": "value1", "custom2": "value2"},
        )
        data = req.to_dict()
        assert data["custom1"] == "value1"
        assert data["custom2"] == "value2"

    def test_to_dict_sorts_dependencies(self):
        """Test to_dict sorts dependencies alphabetically."""
        req = Requirement(
            req_id="REQ-SORT-001",
            dependencies={"REQ-Z", "REQ-A", "REQ-M"},
        )
        data = req.to_dict()
        # Should be sorted: REQ-A|REQ-M|REQ-Z
        assert data["dependencies"] == "REQ-A|REQ-M|REQ-Z"

    def test_to_dict_sorts_blocks(self):
        """Test to_dict sorts blocks alphabetically."""
        req = Requirement(
            req_id="REQ-SORT-002",
            blocks={"REQ-Z", "REQ-A", "REQ-M"},
        )
        data = req.to_dict()
        # Should be sorted: REQ-A|REQ-M|REQ-Z
        assert data["blocks"] == "REQ-A|REQ-M|REQ-Z"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseSaveEdgeCases:
    """Tests for RTMDatabase.save edge cases."""

    def test_save_updates_path_attribute(self, tmp_path: Path):
        """Test save updates the path attribute."""
        req = Requirement(req_id="REQ-SAVE-001")
        db = RTMDatabase([req])
        save_path = tmp_path / "test.csv"
        db.save(save_path)
        assert db.path == save_path


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseIteration:
    """Tests for RTMDatabase iteration protocol."""

    def test_iter_returns_iterator(self):
        """Test __iter__ returns an iterator."""
        req = Requirement(req_id="REQ-ITER-001")
        db = RTMDatabase([req])
        iterator = iter(db)
        assert hasattr(iterator, "__next__")

    def test_iter_empty_database(self):
        """Test iterating empty database yields no items."""
        db = RTMDatabase([])
        items = list(db)
        assert items == []


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCompletionPercentageEdgeCases:
    """Tests for RTMDatabase.completion_percentage edge cases."""

    def test_completion_percentage_with_only_partial(self):
        """Test completion_percentage with only PARTIAL requirements."""
        req1 = Requirement(req_id="REQ-PCT-001", status=Status.PARTIAL)
        req2 = Requirement(req_id="REQ-PCT-002", status=Status.PARTIAL)
        db = RTMDatabase([req1, req2])
        # Both partial = 50% complete
        assert db.completion_percentage() == 50.0

    def test_completion_percentage_rounds_correctly(self):
        """Test completion_percentage calculation precision."""
        # 1 complete, 2 partial, 1 missing = (1 + 1) / 4 = 0.5 = 50%
        reqs = [
            Requirement(req_id="REQ-1", status=Status.COMPLETE),
            Requirement(req_id="REQ-2", status=Status.PARTIAL),
            Requirement(req_id="REQ-3", status=Status.PARTIAL),
            Requirement(req_id="REQ-4", status=Status.MISSING),
        ]
        db = RTMDatabase(reqs)
        assert db.completion_percentage() == 50.0


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseLoadWithNone:
    """Tests for RTMDatabase.load with None path."""

    def test_load_with_none_searches_for_default(self, monkeypatch, tmp_path: Path):
        """Test load with None path searches for default database."""
        # Create a test database file
        test_db_path = tmp_path / "rtm_database.csv"
        req = Requirement(req_id="REQ-DEFAULT-001")
        db = RTMDatabase([req])
        db.save(test_db_path)

        # Mock find_rtm_database to return our test path
        def mock_find():
            return test_db_path

        monkeypatch.setattr("rtmx.parser.find_rtm_database", mock_find)

        # Load with None should use find_rtm_database
        loaded_db = RTMDatabase.load(None)
        assert len(loaded_db) == 1
        assert "REQ-DEFAULT-001" in loaded_db


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementHasTestEdgeCases:
    """Tests for Requirement.has_test edge cases."""

    def test_has_test_with_empty_test_module(self):
        """Test has_test returns False with empty test_module."""
        req = Requirement(
            req_id="REQ-TEST-EDGE-001",
            test_module="",
            test_function="test_something",
        )
        assert req.has_test() is False

    def test_has_test_with_empty_test_function(self):
        """Test has_test returns False with empty test_function."""
        req = Requirement(
            req_id="REQ-TEST-EDGE-002",
            test_module="tests/test_foo.py",
            test_function="",
        )
        assert req.has_test() is False

    def test_has_test_with_both_missing_string(self):
        """Test has_test returns False when both are 'MISSING'."""
        req = Requirement(
            req_id="REQ-TEST-EDGE-003",
            test_module="MISSING",
            test_function="MISSING",
        )
        assert req.has_test() is False


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementDefaults:
    """Tests for Requirement default values."""

    def test_requirement_with_only_req_id(self):
        """Test Requirement with only req_id uses defaults."""
        req = Requirement(req_id="REQ-DEFAULT-001")
        assert req.req_id == "REQ-DEFAULT-001"
        assert req.category == ""
        assert req.subcategory == ""
        assert req.requirement_text == ""
        assert req.target_value == ""
        assert req.test_module == ""
        assert req.test_function == ""
        assert req.validation_method == ""
        assert req.status == Status.MISSING
        assert req.priority == Priority.MEDIUM
        assert req.phase is None
        assert req.notes == ""
        assert req.effort_weeks is None
        assert req.dependencies == set()
        assert req.blocks == set()
        assert req.assignee == ""
        assert req.sprint == ""
        assert req.started_date == ""
        assert req.completed_date == ""
        assert req.requirement_file == ""
        assert req.external_id == ""
        assert req.extra == {}


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestStatusEnumMembers:
    """Tests for Status enum member access."""

    def test_status_enum_has_all_members(self):
        """Test Status enum has all expected members."""
        assert hasattr(Status, "COMPLETE")
        assert hasattr(Status, "PARTIAL")
        assert hasattr(Status, "MISSING")
        assert hasattr(Status, "NOT_STARTED")

    def test_status_enum_values(self):
        """Test Status enum values are correct."""
        assert Status.COMPLETE.value == "COMPLETE"
        assert Status.PARTIAL.value == "PARTIAL"
        assert Status.MISSING.value == "MISSING"
        assert Status.NOT_STARTED.value == "NOT_STARTED"

    def test_status_from_string_with_mixed_separators(self):
        """Test from_string handles mixed hyphens and spaces."""
        # Both get converted to underscores
        assert Status.from_string("NOT-STARTED") == Status.NOT_STARTED
        assert Status.from_string("NOT STARTED") == Status.NOT_STARTED


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestPriorityEnumMembers:
    """Tests for Priority enum member access."""

    def test_priority_enum_has_all_members(self):
        """Test Priority enum has all expected members."""
        assert hasattr(Priority, "P0")
        assert hasattr(Priority, "HIGH")
        assert hasattr(Priority, "MEDIUM")
        assert hasattr(Priority, "LOW")

    def test_priority_enum_values(self):
        """Test Priority enum values are correct."""
        assert Priority.P0.value == "P0"
        assert Priority.HIGH.value == "HIGH"
        assert Priority.MEDIUM.value == "MEDIUM"
        assert Priority.LOW.value == "LOW"

    def test_priority_from_string_with_whitespace(self):
        """Test from_string handles extra whitespace."""
        assert Priority.from_string("  P0  ") == Priority.P0
        assert Priority.from_string("  HIGH  ") == Priority.HIGH
        assert Priority.from_string("  MEDIUM  ") == Priority.MEDIUM
        assert Priority.from_string("  LOW  ") == Priority.LOW


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseFilterCombinations:
    """Tests for RTMDatabase.filter with various combinations."""

    def test_filter_by_category_and_phase(self):
        """Test filtering by both category and phase."""
        req1 = Requirement(req_id="REQ-F1", category="SOFTWARE", phase=1)
        req2 = Requirement(req_id="REQ-F2", category="SOFTWARE", phase=2)
        req3 = Requirement(req_id="REQ-F3", category="TESTING", phase=1)
        db = RTMDatabase([req1, req2, req3])

        results = db.filter(category="SOFTWARE", phase=1)
        assert len(results) == 1
        assert results[0].req_id == "REQ-F1"

    def test_filter_by_status_and_has_test(self):
        """Test filtering by status and has_test."""
        req1 = Requirement(
            req_id="REQ-F4",
            status=Status.COMPLETE,
            test_module="tests/test_a.py",
            test_function="test_a",
        )
        req2 = Requirement(
            req_id="REQ-F5",
            status=Status.COMPLETE,
        )
        req3 = Requirement(
            req_id="REQ-F6",
            status=Status.MISSING,
            test_module="tests/test_b.py",
            test_function="test_b",
        )
        db = RTMDatabase([req1, req2, req3])

        results = db.filter(status=Status.COMPLETE, has_test=True)
        assert len(results) == 1
        assert results[0].req_id == "REQ-F4"

    def test_filter_by_priority_and_subcategory(self):
        """Test filtering by priority and subcategory."""
        req1 = Requirement(req_id="REQ-F7", priority=Priority.HIGH, subcategory="UI")
        req2 = Requirement(req_id="REQ-F8", priority=Priority.HIGH, subcategory="API")
        req3 = Requirement(req_id="REQ-F9", priority=Priority.LOW, subcategory="UI")
        db = RTMDatabase([req1, req2, req3])

        results = db.filter(priority=Priority.HIGH, subcategory="UI")
        assert len(results) == 1
        assert results[0].req_id == "REQ-F7"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseInitialization:
    """Tests for RTMDatabase initialization."""

    def test_init_with_empty_list(self):
        """Test initialization with empty list."""
        db = RTMDatabase([])
        assert len(db) == 0
        assert db.path is None
        assert db._graph is None

    def test_init_with_requirements_list(self):
        """Test initialization with requirements list."""
        req1 = Requirement(req_id="REQ-INIT-001")
        req2 = Requirement(req_id="REQ-INIT-002")
        db = RTMDatabase([req1, req2])
        assert len(db) == 2
        assert "REQ-INIT-001" in db
        assert "REQ-INIT-002" in db

    def test_init_with_path(self):
        """Test initialization with path parameter."""
        path = Path("/tmp/test_db.csv")
        db = RTMDatabase([], path=path)
        assert db.path == path

    def test_init_creates_id_mapping(self):
        """Test initialization creates internal ID mapping."""
        req = Requirement(req_id="REQ-MAP-001")
        db = RTMDatabase([req])
        # Internal dict should map req_id to requirement
        assert db._requirements["REQ-MAP-001"] == req


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementFromDictAllFields:
    """Tests for Requirement.from_dict with all field variations."""

    def test_from_dict_with_all_capitalized_fields(self):
        """Test from_dict with all capitalized field names."""
        data = {
            "Req_ID": "REQ-CAP-001",
            "Category": "CAT",
            "Subcategory": "SUB",
            "Requirement_Text": "Text",
            "Target_Value": "Target",
            "Test_Module": "tests/test.py",
            "Test_Function": "test_func",
            "Validation_Method": "Unit",
            "Status": "COMPLETE",
            "Priority": "HIGH",
            "Phase": "1",
            "Notes": "Notes",
            "Effort_Weeks": "2.5",
            "Dependencies": "REQ-A",
            "Blocks": "REQ-B",
            "Assignee": "alice",
            "Sprint": "v1.0",
            "Started_Date": "2025-01-01",
            "Completed_Date": "2025-01-15",
            "Requirement_File": "docs/req.md",
            "External_ID": "GH-123",
        }
        req = Requirement.from_dict(data)
        assert req.req_id == "REQ-CAP-001"
        assert req.category == "CAT"
        assert req.subcategory == "SUB"
        assert req.requirement_text == "Text"
        assert req.target_value == "Target"
        assert req.test_module == "tests/test.py"
        assert req.test_function == "test_func"
        assert req.validation_method == "Unit"
        assert req.status == Status.COMPLETE
        assert req.priority == Priority.HIGH
        assert req.phase == 1
        assert req.notes == "Notes"
        assert req.effort_weeks == 2.5
        assert req.dependencies == {"REQ-A"}
        assert req.blocks == {"REQ-B"}
        assert req.assignee == "alice"
        assert req.sprint == "v1.0"
        assert req.started_date == "2025-01-01"
        assert req.completed_date == "2025-01-15"
        assert req.requirement_file == "docs/req.md"
        assert req.external_id == "GH-123"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMDatabaseUpdateDependenciesAndBlocks:
    """Tests for updating dependencies and blocks as strings vs sets."""

    def test_update_dependencies_as_empty_string(self):
        """Test update with empty dependencies string."""
        req = Requirement(req_id="REQ-UPD-DEPS-001", dependencies={"REQ-OLD"})
        db = RTMDatabase([req])
        db.update("REQ-UPD-DEPS-001", dependencies="")
        assert db.get("REQ-UPD-DEPS-001").dependencies == set()

    def test_update_blocks_as_empty_string(self):
        """Test update with empty blocks string."""
        req = Requirement(req_id="REQ-UPD-BLOCKS-001", blocks={"REQ-OLD"})
        db = RTMDatabase([req])
        db.update("REQ-UPD-BLOCKS-001", blocks="")
        assert db.get("REQ-UPD-BLOCKS-001").blocks == set()
