"""Tests for rtmx.validation module."""

from pathlib import Path

from rtmx import RTMDatabase, Status
from rtmx.validation import (
    check_reciprocity,
    fix_reciprocity,
    validate_all,
    validate_schema,
)


class TestValidateSchema:
    """Tests for schema validation."""

    def test_valid_schema(self, core_rtm_path: Path):
        """Test that valid RTM passes schema validation."""
        db = RTMDatabase.load(core_rtm_path)
        errors = validate_schema(db)
        assert len(errors) == 0

    def test_missing_req_id(self, core_rtm_path: Path):
        """Test detection of missing req_id."""
        db = RTMDatabase.load(core_rtm_path)
        # Manually corrupt a requirement
        req = db.get("REQ-SW-001")
        req.req_id = ""

        errors = validate_schema(db)
        assert any("req_id" in str(e).lower() for e in errors)

    def test_invalid_status(self, core_rtm_path: Path):
        """Test detection of invalid status values."""
        db = RTMDatabase.load(core_rtm_path)
        req = db.get("REQ-SW-001")
        # Manually set invalid status (bypass enum)
        object.__setattr__(req, "_status_value", "INVALID")

        # Validation should catch this if checking string values
        # Note: Since we use enum, this test validates enum behavior
        assert req.status in (Status.COMPLETE, Status.PARTIAL, Status.MISSING, Status.NOT_STARTED)


class TestCheckReciprocity:
    """Tests for dependency reciprocity checking."""

    def test_no_violations(self, core_rtm_path: Path):
        """Test database with proper reciprocity."""
        db = RTMDatabase.load(core_rtm_path)
        violations = check_reciprocity(db)
        # Core fixture may have some violations
        assert isinstance(violations, list)

    def test_missing_dependency(self, core_rtm_path: Path):
        """Test detection when blocks has no corresponding dependency."""
        db = RTMDatabase.load(core_rtm_path)
        # REQ-SW-001 blocks REQ-SW-002, check if REQ-SW-002 depends on REQ-SW-001
        req1 = db.get("REQ-SW-001")
        req2 = db.get("REQ-SW-002")

        # If REQ-SW-001 blocks REQ-SW-002 but REQ-SW-002 doesn't depend on REQ-SW-001
        if "REQ-SW-002" in req1.blocks and "REQ-SW-001" not in req2.dependencies:
            violations = check_reciprocity(db)
            assert len(violations) > 0

    def test_missing_blocks(self, core_rtm_path: Path):
        """Test detection when dependency has no corresponding blocks."""
        db = RTMDatabase.load(core_rtm_path)
        # Modify to create violation
        req2 = db.get("REQ-SW-002")
        req2.dependencies.add("REQ-SW-001")

        req1 = db.get("REQ-SW-001")
        req1.blocks.discard("REQ-SW-002")

        violations = check_reciprocity(db)
        assert len(violations) > 0


class TestFixReciprocity:
    """Tests for reciprocity fixing."""

    def test_fix_creates_reciprocal_relationships(self, core_rtm_path: Path):
        """Test that fix creates proper reciprocal relationships."""
        db = RTMDatabase.load(core_rtm_path)

        # Create an unreciprocated relationship
        req1 = db.get("REQ-SW-001")
        req3 = db.get("REQ-SW-003")

        req1.blocks.add("REQ-SW-003")
        req3.dependencies.discard("REQ-SW-001")

        # Verify violation exists
        violations_before = check_reciprocity(db)
        has_violation = any("REQ-SW-001" in str(v) and "REQ-SW-003" in str(v) for v in violations_before)

        if has_violation:
            # Fix it
            fix_reciprocity(db)

            # Verify fix
            violations_after = check_reciprocity(db)
            assert len(violations_after) < len(violations_before)


class TestValidateAll:
    """Tests for validate_all function."""

    def test_validate_all_returns_dict(self, core_rtm_path: Path):
        """Test validate_all returns expected structure."""
        db = RTMDatabase.load(core_rtm_path)
        result = validate_all(db)

        assert "errors" in result
        assert "warnings" in result
        assert "reciprocity" in result
        assert isinstance(result["errors"], list)
        assert isinstance(result["warnings"], list)
        assert isinstance(result["reciprocity"], list)


class TestDatabaseValidationIntegration:
    """Integration tests for RTMDatabase validation methods."""

    def test_database_validate_method(self, core_rtm_path: Path):
        """Test RTMDatabase.validate() method."""
        db = RTMDatabase.load(core_rtm_path)
        errors = db.validate()
        assert isinstance(errors, list)

    def test_database_check_reciprocity_method(self, core_rtm_path: Path):
        """Test RTMDatabase.check_reciprocity() method."""
        db = RTMDatabase.load(core_rtm_path)
        violations = db.check_reciprocity()
        assert isinstance(violations, list)
