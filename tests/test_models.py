"""Tests for rtmx.models module."""

from pathlib import Path

import pytest

from rtmx import Priority, Requirement, RTMDatabase, Status
from rtmx.models import RequirementNotFoundError, RTMError


class TestStatus:
    """Tests for Status enum."""

    def test_from_string_complete(self):
        """Test parsing COMPLETE status."""
        assert Status.from_string("COMPLETE") == Status.COMPLETE
        assert Status.from_string("complete") == Status.COMPLETE
        assert Status.from_string("  COMPLETE  ") == Status.COMPLETE

    def test_from_string_partial(self):
        """Test parsing PARTIAL status."""
        assert Status.from_string("PARTIAL") == Status.PARTIAL

    def test_from_string_missing(self):
        """Test parsing MISSING status."""
        assert Status.from_string("MISSING") == Status.MISSING
        assert Status.from_string("NOT_STARTED") == Status.NOT_STARTED

    def test_from_string_unknown(self):
        """Test parsing unknown status defaults to MISSING."""
        assert Status.from_string("UNKNOWN") == Status.MISSING
        assert Status.from_string("") == Status.MISSING


class TestPriority:
    """Tests for Priority enum."""

    def test_from_string_p0(self):
        """Test parsing P0 priority."""
        assert Priority.from_string("P0") == Priority.P0
        assert Priority.from_string("CRITICAL") == Priority.P0

    def test_from_string_high(self):
        """Test parsing HIGH priority."""
        assert Priority.from_string("HIGH") == Priority.HIGH
        assert Priority.from_string("high") == Priority.HIGH

    def test_from_string_medium(self):
        """Test parsing MEDIUM priority."""
        assert Priority.from_string("MEDIUM") == Priority.MEDIUM

    def test_from_string_unknown(self):
        """Test parsing unknown priority defaults to MEDIUM."""
        assert Priority.from_string("UNKNOWN") == Priority.MEDIUM


class TestRequirement:
    """Tests for Requirement dataclass."""

    def test_from_dict_minimal(self):
        """Test creating requirement from minimal dict."""
        data = {"req_id": "REQ-TEST-001"}
        req = Requirement.from_dict(data)

        assert req.req_id == "REQ-TEST-001"
        assert req.status == Status.MISSING
        assert req.priority == Priority.MEDIUM
        assert req.dependencies == set()
        assert req.blocks == set()

    def test_from_dict_full(self):
        """Test creating requirement from full dict."""
        data = {
            "req_id": "REQ-SW-001",
            "category": "SOFTWARE",
            "subcategory": "ALGORITHM",
            "requirement_text": "Implement feature X",
            "status": "COMPLETE",
            "priority": "HIGH",
            "phase": "1",
            "dependencies": "REQ-A|REQ-B",
            "blocks": "REQ-C",
        }
        req = Requirement.from_dict(data)

        assert req.req_id == "REQ-SW-001"
        assert req.category == "SOFTWARE"
        assert req.status == Status.COMPLETE
        assert req.priority == Priority.HIGH
        assert req.phase == 1
        assert req.dependencies == {"REQ-A", "REQ-B"}
        assert req.blocks == {"REQ-C"}

    def test_has_test_true(self):
        """Test has_test returns True when test is specified."""
        req = Requirement(
            req_id="REQ-TEST",
            test_module="tests/test_foo.py",
            test_function="test_bar",
        )
        assert req.has_test() is True

    def test_has_test_false(self):
        """Test has_test returns False when test is missing."""
        req = Requirement(req_id="REQ-TEST")
        assert req.has_test() is False

        req2 = Requirement(
            req_id="REQ-TEST",
            test_module="MISSING",
            test_function="MISSING",
        )
        assert req2.has_test() is False

    def test_to_dict(self):
        """Test converting requirement to dict."""
        req = Requirement(
            req_id="REQ-TEST",
            category="TEST",
            status=Status.COMPLETE,
            dependencies={"REQ-A", "REQ-B"},
        )
        data = req.to_dict()

        assert data["req_id"] == "REQ-TEST"
        assert data["category"] == "TEST"
        assert data["status"] == "COMPLETE"
        assert "REQ-A" in data["dependencies"]
        assert "REQ-B" in data["dependencies"]


class TestRTMDatabase:
    """Tests for RTMDatabase class."""

    def test_load_from_fixture(self, core_rtm_path: Path):
        """Test loading database from fixture."""
        db = RTMDatabase.load(core_rtm_path)

        assert len(db) == 8
        assert "REQ-SW-001" in db
        assert "REQ-DOC-001" in db

    def test_get_existing(self, core_rtm_path: Path):
        """Test getting existing requirement."""
        db = RTMDatabase.load(core_rtm_path)
        req = db.get("REQ-SW-001")

        assert req.req_id == "REQ-SW-001"
        assert req.category == "SOFTWARE"
        assert req.status == Status.COMPLETE

    def test_get_nonexistent(self, core_rtm_path: Path):
        """Test getting non-existent requirement raises error."""
        db = RTMDatabase.load(core_rtm_path)

        with pytest.raises(RequirementNotFoundError):
            db.get("REQ-NONEXISTENT")

    def test_exists(self, core_rtm_path: Path):
        """Test exists method."""
        db = RTMDatabase.load(core_rtm_path)

        assert db.exists("REQ-SW-001") is True
        assert db.exists("REQ-NONEXISTENT") is False

    def test_filter_by_status(self, core_rtm_path: Path):
        """Test filtering by status."""
        db = RTMDatabase.load(core_rtm_path)

        complete = db.filter(status=Status.COMPLETE)
        assert len(complete) == 1
        assert complete[0].req_id == "REQ-SW-001"

        missing = db.filter(status=Status.MISSING)
        assert len(missing) > 0

    def test_filter_by_category(self, core_rtm_path: Path):
        """Test filtering by category."""
        db = RTMDatabase.load(core_rtm_path)

        software = db.filter(category="SOFTWARE")
        assert len(software) == 3
        assert all(r.category == "SOFTWARE" for r in software)

    def test_filter_by_phase(self, core_rtm_path: Path):
        """Test filtering by phase."""
        db = RTMDatabase.load(core_rtm_path)

        phase1 = db.filter(phase=1)
        assert len(phase1) > 0
        assert all(r.phase == 1 for r in phase1)

    def test_update(self, core_rtm_path: Path):
        """Test updating requirement."""
        db = RTMDatabase.load(core_rtm_path)

        db.update("REQ-SW-002", status=Status.COMPLETE, notes="Updated")
        req = db.get("REQ-SW-002")

        assert req.status == Status.COMPLETE
        assert req.notes == "Updated"

    def test_status_counts(self, core_rtm_path: Path):
        """Test status counts."""
        db = RTMDatabase.load(core_rtm_path)
        counts = db.status_counts()

        assert Status.COMPLETE in counts
        assert Status.PARTIAL in counts
        assert Status.MISSING in counts
        assert sum(counts.values()) == len(db)

    def test_completion_percentage(self, core_rtm_path: Path):
        """Test completion percentage calculation."""
        db = RTMDatabase.load(core_rtm_path)
        pct = db.completion_percentage()

        assert 0 <= pct <= 100

    def test_iteration(self, core_rtm_path: Path):
        """Test iterating over database."""
        db = RTMDatabase.load(core_rtm_path)

        count = 0
        for req in db:
            assert isinstance(req, Requirement)
            count += 1

        assert count == len(db)
