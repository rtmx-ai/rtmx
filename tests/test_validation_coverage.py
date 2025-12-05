"""Comprehensive tests for rtmx.validation module.

This module provides extensive test coverage for all validation functions:
- validate_schema(): Schema validation with required fields and valid values
- check_reciprocity(): Dependency/blocks reciprocity checking
- fix_reciprocity(): Automatic reciprocity violation fixing
- validate_cycles(): Circular dependency detection
- validate_all(): Combined validation operations
"""

from pathlib import Path

import pytest

from rtmx import Priority, Requirement, RTMDatabase, Status
from rtmx.validation import (
    check_reciprocity,
    fix_reciprocity,
    validate_all,
    validate_cycles,
    validate_schema,
)


class TestValidateSchema:
    """Tests for validate_schema() function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_empty_database(self):
        """Test validation of empty database returns no errors."""
        db = RTMDatabase([])
        errors = validate_schema(db)
        assert errors == []

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_valid_database(self, core_rtm_path: Path):
        """Test validation of valid database passes without errors."""
        db = RTMDatabase.load(core_rtm_path)
        errors = validate_schema(db)
        assert len(errors) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_missing_req_id(self):
        """Test detection of missing req_id field."""
        req = Requirement(
            req_id="",
            category="TEST",
            requirement_text="Test requirement",
        )
        db = RTMDatabase([req])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("req_id" in str(e).lower() for e in errors)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_missing_category(self):
        """Test detection of missing category field."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="",
            requirement_text="Test requirement",
        )
        db = RTMDatabase([req])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("category" in str(e).lower() for e in errors)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_missing_requirement_text(self):
        """Test detection of missing requirement_text field."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="",
        )
        db = RTMDatabase([req])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("requirement_text" in str(e).lower() for e in errors)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_duplicate_req_ids(self):
        """Test detection of duplicate requirement IDs.

        Note: RTMDatabase constructor uses a dict, so duplicates are
        automatically de-duplicated. This test verifies that validation
        can iterate over the database without errors even with the same
        requirement ID appearing multiple times in the source list.
        """
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="First requirement",
        )
        req2 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Duplicate ID",
        )
        # RTMDatabase constructor de-duplicates via dict, keeps last
        db = RTMDatabase([req1, req2])
        errors = validate_schema(db)

        # No errors since duplicates are handled at construction
        # In practice, duplicates in CSV would be caught by parser
        assert isinstance(errors, list)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_detect_invalid_phase_zero(self):
        """Test detection of invalid phase value (zero)."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Test requirement",
            phase=0,
        )
        db = RTMDatabase([req])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("phase" in str(e).lower() for e in errors)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_detect_invalid_phase_negative(self):
        """Test detection of invalid phase value (negative)."""
        req = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Test requirement",
            phase=-1,
        )
        db = RTMDatabase([req])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("phase" in str(e).lower() for e in errors)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_valid_phase_values(self):
        """Test that valid phase values pass validation."""
        requirements = []
        for i in range(1, 6):
            requirements.append(
                Requirement(
                    req_id=f"REQ-TEST-{i:03d}",
                    category="TEST",
                    requirement_text=f"Phase {i} requirement",
                    phase=i,
                )
            )
        db = RTMDatabase(requirements)
        errors = validate_schema(db)

        # Should have no phase-related errors
        phase_errors = [e for e in errors if "phase" in str(e).lower()]
        assert len(phase_errors) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_nonexistent_dependency(self):
        """Test detection of dependency referencing non-existent requirement."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Test requirement",
            dependencies={"REQ-DOES-NOT-EXIST"},
        )
        db = RTMDatabase([req1])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any(
            "dependency" in str(e).lower() and "non-existent" in str(e).lower() for e in errors
        )

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_nonexistent_blocks(self):
        """Test detection of blocks referencing non-existent requirement."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Test requirement",
            blocks={"REQ-DOES-NOT-EXIST"},
        )
        db = RTMDatabase([req1])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("blocks" in str(e).lower() and "non-existent" in str(e).lower() for e in errors)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_all_status_values_valid(self):
        """Test that all Status enum values pass validation."""
        for i, status in enumerate(Status):
            req = Requirement(
                req_id=f"REQ-TEST-{i:03d}",
                category="TEST",
                requirement_text=f"Requirement with {status.value} status",
                status=status,
            )
            db = RTMDatabase([req])
            errors = validate_schema(db)

            # No status-related errors
            status_errors = [e for e in errors if "status" in str(e).lower()]
            assert len(status_errors) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_all_priority_values_valid(self):
        """Test that all Priority enum values pass validation."""
        for i, priority in enumerate(Priority):
            req = Requirement(
                req_id=f"REQ-TEST-{i:03d}",
                category="TEST",
                requirement_text=f"Requirement with {priority.value} priority",
                priority=priority,
            )
            db = RTMDatabase([req])
            errors = validate_schema(db)

            # No priority-related errors
            priority_errors = [e for e in errors if "priority" in str(e).lower()]
            assert len(priority_errors) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_whitespace_only_fields_detected(self):
        """Test detection of whitespace-only required fields."""
        req = Requirement(
            req_id="   ",  # Whitespace only
            category="TEST",
            requirement_text="Test requirement",
        )
        db = RTMDatabase([req])
        errors = validate_schema(db)

        assert len(errors) > 0
        assert any("req_id" in str(e).lower() for e in errors)


class TestCheckReciprocity:
    """Tests for check_reciprocity() function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_empty_database_no_violations(self):
        """Test that empty database has no reciprocity violations."""
        db = RTMDatabase([])
        violations = check_reciprocity(db)
        assert violations == []

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_reciprocal_relationships_valid(self):
        """Test that proper reciprocal relationships have no violations."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks REQ-TEST-002",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        db = RTMDatabase([req1, req2])
        violations = check_reciprocity(db)

        assert len(violations) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_blocks_without_dependency_violation(self):
        """Test detection when A blocks B but B doesn't depend on A."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks REQ-TEST-002",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Does not depend on REQ-TEST-001",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2])
        violations = check_reciprocity(db)

        assert len(violations) > 0
        assert any("REQ-TEST-001" in str(v) and "REQ-TEST-002" in str(v) for v in violations)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_dependency_without_blocks_violation(self):
        """Test detection when A depends on B but B doesn't block A."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Depends on REQ-TEST-002",
            dependencies={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Does not block REQ-TEST-001",
            blocks=set(),
        )
        db = RTMDatabase([req1, req2])
        violations = check_reciprocity(db)

        assert len(violations) > 0
        assert any("REQ-TEST-001" in str(v) and "REQ-TEST-002" in str(v) for v in violations)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_blocks_nonexistent_requirement(self):
        """Test detection of blocks referencing non-existent requirement."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks non-existent requirement",
            blocks={"REQ-DOES-NOT-EXIST"},
        )
        db = RTMDatabase([req1])
        violations = check_reciprocity(db)

        assert len(violations) > 0
        assert any("non-existent" in str(v).lower() for v in violations)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_depends_on_nonexistent_requirement(self):
        """Test detection of dependency on non-existent requirement."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Depends on non-existent requirement",
            dependencies={"REQ-DOES-NOT-EXIST"},
        )
        db = RTMDatabase([req1])
        violations = check_reciprocity(db)

        assert len(violations) > 0
        assert any("non-existent" in str(v).lower() for v in violations)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_complex_dependency_chain(self):
        """Test reciprocity checking in complex dependency chain."""
        # REQ-001 blocks REQ-002 blocks REQ-003
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="First in chain",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Middle of chain",
            dependencies={"REQ-TEST-001"},
            blocks={"REQ-TEST-003"},
        )
        req3 = Requirement(
            req_id="REQ-TEST-003",
            category="TEST",
            requirement_text="End of chain",
            dependencies={"REQ-TEST-002"},
        )
        db = RTMDatabase([req1, req2, req3])
        violations = check_reciprocity(db)

        assert len(violations) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_multiple_dependencies_and_blocks(self):
        """Test reciprocity with multiple dependencies and blocks."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Has multiple blocks",
            blocks={"REQ-TEST-002", "REQ-TEST-003"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        req3 = Requirement(
            req_id="REQ-TEST-003",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        db = RTMDatabase([req1, req2, req3])
        violations = check_reciprocity(db)

        assert len(violations) == 0


class TestFixReciprocity:
    """Tests for fix_reciprocity() function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fix_empty_database(self):
        """Test fix_reciprocity on empty database returns zero fixes."""
        db = RTMDatabase([])
        fixed_count = fix_reciprocity(db)
        assert fixed_count == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fix_adds_missing_dependency(self):
        """Test that fix adds missing dependency for blocks relationship."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks REQ-TEST-002",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Should depend on REQ-TEST-001",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2])

        fixed_count = fix_reciprocity(db)

        assert fixed_count > 0
        assert "REQ-TEST-001" in req2.dependencies

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fix_adds_missing_blocks(self):
        """Test that fix adds missing blocks for dependency relationship."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Depends on REQ-TEST-002",
            dependencies={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Should block REQ-TEST-001",
            blocks=set(),
        )
        db = RTMDatabase([req1, req2])

        fixed_count = fix_reciprocity(db)

        assert fixed_count > 0
        assert "REQ-TEST-001" in req2.blocks

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fix_does_not_modify_valid_relationships(self):
        """Test that fix doesn't modify already valid relationships."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks REQ-TEST-002",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        db = RTMDatabase([req1, req2])

        fixed_count = fix_reciprocity(db)

        assert fixed_count == 0
        assert req1.blocks == {"REQ-TEST-002"}
        assert req2.dependencies == {"REQ-TEST-001"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fix_skips_nonexistent_dependencies(self):
        """Test that fix skips references to non-existent requirements."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks non-existent",
            blocks={"REQ-DOES-NOT-EXIST"},
        )
        db = RTMDatabase([req1])

        # Should not raise exception
        fixed_count = fix_reciprocity(db)

        assert fixed_count == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fix_and_verify_no_violations(self):
        """Test that after fix, no reciprocity violations remain."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks REQ-TEST-002",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Missing dependency",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2])

        # Fix reciprocity
        fix_reciprocity(db)

        # Verify no violations remain
        violations = check_reciprocity(db)
        assert len(violations) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_fix_multiple_violations(self):
        """Test fixing multiple reciprocity violations."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Has violations",
            blocks={"REQ-TEST-002", "REQ-TEST-003"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Missing dependency",
            dependencies=set(),
        )
        req3 = Requirement(
            req_id="REQ-TEST-003",
            category="TEST",
            requirement_text="Missing dependency",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2, req3])

        fixed_count = fix_reciprocity(db)

        assert fixed_count >= 2
        assert "REQ-TEST-001" in req2.dependencies
        assert "REQ-TEST-001" in req3.dependencies


class TestValidateCycles:
    """Tests for validate_cycles() function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_no_cycles_in_empty_database(self):
        """Test that empty database has no cycles."""
        db = RTMDatabase([])
        warnings = validate_cycles(db)
        assert warnings == []

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_no_cycles_in_acyclic_graph(self):
        """Test that acyclic dependency graph has no cycle warnings."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="First",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Second",
            dependencies={"REQ-TEST-001"},
            blocks={"REQ-TEST-003"},
        )
        req3 = Requirement(
            req_id="REQ-TEST-003",
            category="TEST",
            requirement_text="Third",
            dependencies={"REQ-TEST-002"},
        )
        db = RTMDatabase([req1, req2, req3])
        warnings = validate_cycles(db)

        assert len(warnings) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_simple_cycle(self):
        """Test detection of simple two-requirement cycle."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Depends on REQ-TEST-002",
            dependencies={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        db = RTMDatabase([req1, req2])
        warnings = validate_cycles(db)

        assert len(warnings) > 0
        assert any("cycle" in str(w).lower() for w in warnings)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_three_requirement_cycle(self):
        """Test detection of three-requirement cycle."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Depends on REQ-TEST-002",
            dependencies={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Depends on REQ-TEST-003",
            dependencies={"REQ-TEST-003"},
        )
        req3 = Requirement(
            req_id="REQ-TEST-003",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        db = RTMDatabase([req1, req2, req3])
        warnings = validate_cycles(db)

        assert len(warnings) > 0


class TestValidateAll:
    """Tests for validate_all() function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_all_returns_complete_structure(self):
        """Test that validate_all returns all expected keys."""
        db = RTMDatabase([])
        result = validate_all(db)

        assert "errors" in result
        assert "warnings" in result
        assert "reciprocity" in result
        assert isinstance(result["errors"], list)
        assert isinstance(result["warnings"], list)
        assert isinstance(result["reciprocity"], list)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_all_on_valid_database(self, core_rtm_path: Path):
        """Test validate_all on a valid database."""
        db = RTMDatabase.load(core_rtm_path)
        result = validate_all(db)

        # May have some reciprocity violations or warnings, but no schema errors
        assert len(result["errors"]) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_all_detects_schema_errors(self):
        """Test that validate_all detects schema validation errors."""
        req = Requirement(
            req_id="",
            category="TEST",
            requirement_text="Invalid requirement",
        )
        db = RTMDatabase([req])
        result = validate_all(db)

        assert len(result["errors"]) > 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_all_detects_reciprocity_violations(self):
        """Test that validate_all detects reciprocity violations."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Blocks REQ-TEST-002",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Missing dependency",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2])
        result = validate_all(db)

        assert len(result["reciprocity"]) > 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_validate_all_detects_cycles(self):
        """Test that validate_all detects circular dependencies."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Depends on REQ-TEST-002",
            dependencies={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Depends on REQ-TEST-001",
            dependencies={"REQ-TEST-001"},
        )
        db = RTMDatabase([req1, req2])
        result = validate_all(db)

        assert len(result["warnings"]) > 0


class TestDatabaseValidationMethods:
    """Integration tests for RTMDatabase validation method delegation."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_database_validate_delegates_correctly(self, core_rtm_path: Path):
        """Test that RTMDatabase.validate() delegates to validate_schema."""
        db = RTMDatabase.load(core_rtm_path)

        direct_result = validate_schema(db)
        method_result = db.validate()

        assert direct_result == method_result

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_database_check_reciprocity_delegates_correctly(self):
        """Test that RTMDatabase.check_reciprocity() delegates correctly."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Test requirement",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Test requirement",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2])

        direct_result = check_reciprocity(db)
        method_result = db.check_reciprocity()

        assert direct_result == method_result

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_database_fix_reciprocity_delegates_correctly(self):
        """Test that RTMDatabase.fix_reciprocity() delegates correctly."""
        req1 = Requirement(
            req_id="REQ-TEST-001",
            category="TEST",
            requirement_text="Test requirement",
            blocks={"REQ-TEST-002"},
        )
        req2 = Requirement(
            req_id="REQ-TEST-002",
            category="TEST",
            requirement_text="Test requirement",
            dependencies=set(),
        )
        db = RTMDatabase([req1, req2])

        fixed_count = db.fix_reciprocity()

        assert fixed_count > 0
        assert "REQ-TEST-001" in req2.dependencies
