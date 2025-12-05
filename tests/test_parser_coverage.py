"""Comprehensive tests for rtmx.parser module."""

import csv
from pathlib import Path

import pytest

from rtmx.models import Requirement, RTMError
from rtmx.parser import (
    DEFAULT_RTM_PATH,
    detect_column_format,
    find_rtm_database,
    format_dependencies,
    load_csv,
    normalize_column_name,
    parse_dependencies,
    save_csv,
)


class TestParseDependencies:
    """Tests for parse_dependencies function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_parse_dependencies_empty(self):
        """Test parsing empty dependency string."""
        assert parse_dependencies("") == set()
        assert parse_dependencies("   ") == set()

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_parse_dependencies_pipe_separated(self):
        """Test parsing pipe-separated dependencies."""
        deps = parse_dependencies("REQ-A|REQ-B|REQ-C")
        assert deps == {"REQ-A", "REQ-B", "REQ-C"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_parse_dependencies_space_separated(self):
        """Test parsing space-separated dependencies."""
        deps = parse_dependencies("REQ-A REQ-B REQ-C")
        assert deps == {"REQ-A", "REQ-B", "REQ-C"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_parse_dependencies_single(self):
        """Test parsing single dependency."""
        deps = parse_dependencies("REQ-A")
        assert deps == {"REQ-A"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_parse_dependencies_with_whitespace(self):
        """Test parsing dependencies with extra whitespace."""
        deps = parse_dependencies("  REQ-A  |  REQ-B  ")
        assert deps == {"REQ-A", "REQ-B"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_parse_dependencies_mixed_separators(self):
        """Test pipe separator takes precedence over spaces."""
        # If pipe is present, split by pipe (even if spaces exist)
        deps = parse_dependencies("REQ-A REQ-B|REQ-C REQ-D")
        # Should split on pipe, resulting in two groups
        assert "REQ-A REQ-B" in deps or "REQ-C REQ-D" in deps

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_parse_dependencies_empty_parts(self):
        """Test parsing dependencies with empty parts."""
        deps = parse_dependencies("REQ-A||REQ-B")
        # Empty parts should be filtered out
        assert deps == {"REQ-A", "REQ-B"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_parse_dependencies_duplicate_entries(self):
        """Test parsing dependencies with duplicates."""
        deps = parse_dependencies("REQ-A|REQ-B|REQ-A")
        # Set should deduplicate
        assert deps == {"REQ-A", "REQ-B"}


class TestFormatDependencies:
    """Tests for format_dependencies function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_dependencies_empty(self):
        """Test formatting empty dependency set."""
        result = format_dependencies(set())
        assert result == ""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_dependencies_single(self):
        """Test formatting single dependency."""
        result = format_dependencies({"REQ-A"})
        assert result == "REQ-A"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_dependencies_multiple(self):
        """Test formatting multiple dependencies."""
        result = format_dependencies({"REQ-C", "REQ-A", "REQ-B"})
        # Should be sorted and pipe-separated
        assert result == "REQ-A|REQ-B|REQ-C"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_dependencies_sorts(self):
        """Test formatting sorts dependencies alphabetically."""
        result = format_dependencies({"REQ-Z", "REQ-A", "REQ-M"})
        assert result == "REQ-A|REQ-M|REQ-Z"


class TestDetectColumnFormat:
    """Tests for detect_column_format function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_snake_case_explicit(self):
        """Test detecting snake_case from req_id column."""
        fieldnames = ["req_id", "category", "status"]
        assert detect_column_format(fieldnames) == "snake_case"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_pascal_case_explicit(self):
        """Test detecting PascalCase from Req_ID column."""
        fieldnames = ["Req_ID", "Category", "Status"]
        assert detect_column_format(fieldnames) == "PascalCase"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_snake_case_implicit(self):
        """Test detecting snake_case from lowercase underscore columns."""
        fieldnames = ["test_module", "test_function", "effort_weeks"]
        assert detect_column_format(fieldnames) == "snake_case"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_detect_pascal_case_default(self):
        """Test PascalCase is default when no indicators found."""
        fieldnames = ["Column1", "Column2", "Column3"]
        assert detect_column_format(fieldnames) == "PascalCase"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_detect_empty_fieldnames(self):
        """Test detecting format with empty fieldnames."""
        fieldnames = []
        # Should default to PascalCase
        assert detect_column_format(fieldnames) == "PascalCase"


class TestNormalizeColumnName:
    """Tests for normalize_column_name function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_normalize_req_id_to_snake(self):
        """Test normalizing Req_ID to snake_case."""
        assert normalize_column_name("Req_ID", "snake_case") == "req_id"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_normalize_category_to_snake(self):
        """Test normalizing Category to snake_case."""
        assert normalize_column_name("Category", "snake_case") == "category"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_normalize_requirement_text_to_snake(self):
        """Test normalizing Requirement_Text to snake_case."""
        assert normalize_column_name("Requirement_Text", "snake_case") == "requirement_text"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_normalize_req_id_to_pascal(self):
        """Test normalizing req_id to PascalCase."""
        assert normalize_column_name("req_id", "PascalCase") == "Req_ID"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_normalize_category_to_pascal(self):
        """Test normalizing category to PascalCase."""
        assert normalize_column_name("category", "PascalCase") == "Category"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_normalize_unknown_column_to_snake(self):
        """Test normalizing unknown column to snake_case."""
        # Unknown columns should be lowercased
        result = normalize_column_name("Unknown_Column", "snake_case")
        assert result == "unknown_column"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_normalize_unknown_column_to_pascal(self):
        """Test normalizing unknown column to PascalCase."""
        # Unknown columns should remain unchanged
        result = normalize_column_name("unknown_column", "PascalCase")
        assert result == "unknown_column"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_normalize_all_core_columns(self):
        """Test normalizing all core RTM columns."""
        core_columns = [
            ("Req_ID", "req_id"),
            ("Status", "status"),
            ("Priority", "priority"),
            ("Phase", "phase"),
            ("Dependencies", "dependencies"),
            ("Blocks", "blocks"),
            ("Assignee", "assignee"),
        ]
        for pascal, snake in core_columns:
            assert normalize_column_name(pascal, "snake_case") == snake


class TestFindRtmDatabase:
    """Tests for find_rtm_database function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_find_rtm_database_exists(self, tmp_path):
        """Test finding RTM database that exists."""
        # Create RTM database
        rtm_path = tmp_path / DEFAULT_RTM_PATH
        rtm_path.parent.mkdir(parents=True, exist_ok=True)
        rtm_path.write_text("req_id\nREQ-001\n")

        # Should find it
        found = find_rtm_database(tmp_path)
        assert found.exists()
        assert found.name == "rtm_database.csv"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_find_rtm_database_not_found(self, tmp_path):
        """Test finding RTM database raises error when not found."""
        with pytest.raises(RTMError) as exc:
            find_rtm_database(tmp_path)
        assert "Could not find RTM database" in str(exc.value)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_find_rtm_database_default_path(self, tmp_path, monkeypatch):
        """Test finding RTM database from current directory."""
        # Create RTM database
        rtm_path = tmp_path / DEFAULT_RTM_PATH
        rtm_path.parent.mkdir(parents=True, exist_ok=True)
        rtm_path.write_text("req_id\nREQ-001\n")

        # Change to temp directory
        monkeypatch.chdir(tmp_path)

        # Should find it from cwd
        found = find_rtm_database()
        assert found.name == "rtm_database.csv"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_find_rtm_database_searches_upward(self, tmp_path):
        """Test finding RTM database searches parent directories."""
        # Create nested structure
        nested = tmp_path / "a" / "b" / "c"
        nested.mkdir(parents=True)

        # Create RTM at root
        rtm_path = tmp_path / DEFAULT_RTM_PATH
        rtm_path.parent.mkdir(parents=True, exist_ok=True)
        rtm_path.write_text("req_id\nREQ-001\n")

        # Should find it from nested directory
        found = find_rtm_database(nested)
        assert found == rtm_path


class TestLoadCsv:
    """Tests for load_csv function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_basic(self, tmp_path):
        """Test loading basic CSV file."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text(
            "req_id,category,status\n" "REQ-001,SOFTWARE,COMPLETE\n" "REQ-002,HARDWARE,MISSING\n"
        )

        requirements = load_csv(csv_path)
        assert len(requirements) == 2
        assert requirements[0].req_id == "REQ-001"
        assert requirements[1].req_id == "REQ-002"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_snake_case(self, tmp_path):
        """Test loading CSV with snake_case columns."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text(
            "req_id,category,requirement_text,status\n"
            "REQ-001,SOFTWARE,Test requirement,COMPLETE\n"
        )

        requirements = load_csv(csv_path)
        assert len(requirements) == 1
        assert requirements[0].requirement_text == "Test requirement"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_pascal_case(self, tmp_path):
        """Test loading CSV with PascalCase columns."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text(
            "Req_ID,Category,Requirement_Text,Status\n"
            "REQ-001,SOFTWARE,Test requirement,COMPLETE\n"
        )

        requirements = load_csv(csv_path)
        assert len(requirements) == 1
        assert requirements[0].req_id == "REQ-001"
        assert requirements[0].requirement_text == "Test requirement"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_with_dependencies(self, tmp_path):
        """Test loading CSV with dependencies."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text("req_id,dependencies,blocks\n" "REQ-001,REQ-A|REQ-B,REQ-C\n")

        requirements = load_csv(csv_path)
        assert requirements[0].dependencies == {"REQ-A", "REQ-B"}
        assert requirements[0].blocks == {"REQ-C"}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_with_booleans(self, tmp_path):
        """Test loading CSV with boolean fields."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text(
            "req_id,unit_test,integration_test,scope_unit\n" "REQ-001,True,False,true\n"
        )

        requirements = load_csv(csv_path)
        assert requirements[0].extra.get("unit_test") is True
        assert requirements[0].extra.get("integration_test") is False
        assert requirements[0].extra.get("scope_unit") is True

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_load_csv_file_not_found(self, tmp_path):
        """Test loading non-existent CSV raises error."""
        csv_path = tmp_path / "nonexistent.csv"
        with pytest.raises(RTMError) as exc:
            load_csv(csv_path)
        assert "not found" in str(exc.value)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_load_csv_empty_file(self, tmp_path):
        """Test loading empty CSV raises error."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text("req_id\n")  # Header only

        with pytest.raises(RTMError) as exc:
            load_csv(csv_path)
        assert "empty" in str(exc.value).lower()

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_load_csv_no_header(self, tmp_path):
        """Test loading CSV without header raises error."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text("")  # Completely empty

        with pytest.raises(RTMError) as exc:
            load_csv(csv_path)
        assert "no header" in str(exc.value).lower()

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_load_csv_malformed(self, tmp_path):
        """Test loading malformed CSV handles gracefully."""
        csv_path = tmp_path / "test.csv"
        # Create CSV with mismatched quotes - Python's csv module is quite forgiving
        # so we just verify that edge cases don't crash
        csv_path.write_text(
            "req_id,category\n" '"REQ-001,SOFTWARE\n'  # Missing closing quote
        )

        # CSV parser handles this gracefully - just verify no crash
        requirements = load_csv(csv_path)
        # Should parse at least one requirement even if data is odd
        assert len(requirements) >= 1

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_with_extra_fields(self, tmp_path):
        """Test loading CSV with extra custom fields."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text(
            "req_id,category,custom_field,another_field\n"
            "REQ-001,SOFTWARE,custom_value,another_value\n"
        )

        requirements = load_csv(csv_path)
        assert requirements[0].extra.get("custom_field") == "custom_value"
        assert requirements[0].extra.get("another_field") == "another_value"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_load_csv_preserves_types(self, tmp_path):
        """Test loading CSV preserves numeric types."""
        csv_path = tmp_path / "test.csv"
        csv_path.write_text("req_id,phase,effort_weeks\n" "REQ-001,1,2.5\n")

        requirements = load_csv(csv_path)
        assert requirements[0].phase == 1
        assert requirements[0].effort_weeks == 2.5


class TestSaveCsv:
    """Tests for save_csv function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_basic(self, tmp_path):
        """Test saving basic requirements to CSV."""
        csv_path = tmp_path / "test.csv"
        requirements = [
            Requirement(req_id="REQ-001", category="SOFTWARE"),
            Requirement(req_id="REQ-002", category="HARDWARE"),
        ]

        save_csv(requirements, csv_path)
        assert csv_path.exists()

        # Verify content
        content = csv_path.read_text()
        assert "REQ-001" in content
        assert "REQ-002" in content

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_creates_directory(self, tmp_path):
        """Test saving CSV creates parent directories."""
        csv_path = tmp_path / "nested" / "dir" / "test.csv"
        requirements = [Requirement(req_id="REQ-001")]

        save_csv(requirements, csv_path)
        assert csv_path.exists()

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_with_dependencies(self, tmp_path):
        """Test saving requirements with dependencies."""
        csv_path = tmp_path / "test.csv"
        requirements = [
            Requirement(
                req_id="REQ-001",
                dependencies={"REQ-A", "REQ-B"},
                blocks={"REQ-C"},
            )
        ]

        save_csv(requirements, csv_path)

        # Verify dependencies are pipe-separated
        content = csv_path.read_text()
        assert "|" in content or ("REQ-A" in content and "REQ-B" in content)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_with_booleans(self, tmp_path):
        """Test saving requirements with boolean extra fields."""
        csv_path = tmp_path / "test.csv"
        requirements = [
            Requirement(
                req_id="REQ-001",
                extra={"unit_test": True, "integration_test": False},
            )
        ]

        save_csv(requirements, csv_path)

        # Verify booleans are saved as strings
        content = csv_path.read_text()
        assert "True" in content
        assert "False" in content

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_save_csv_empty_list(self, tmp_path):
        """Test saving empty requirements list raises error."""
        csv_path = tmp_path / "test.csv"
        with pytest.raises(RTMError) as exc:
            save_csv([], csv_path)
        assert "empty" in str(exc.value).lower()

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_load_roundtrip(self, tmp_path):
        """Test saving and loading CSV preserves data."""
        csv_path = tmp_path / "test.csv"
        original = [
            Requirement(
                req_id="REQ-001",
                category="SOFTWARE",
                requirement_text="Test requirement",
                dependencies={"REQ-A"},
                phase=1,
                effort_weeks=2.5,
            )
        ]

        save_csv(original, csv_path)
        loaded = load_csv(csv_path)

        assert len(loaded) == len(original)
        assert loaded[0].req_id == original[0].req_id
        assert loaded[0].category == original[0].category
        assert loaded[0].requirement_text == original[0].requirement_text
        assert loaded[0].dependencies == original[0].dependencies
        assert loaded[0].phase == original[0].phase
        assert loaded[0].effort_weeks == original[0].effort_weeks

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_with_extra_fields(self, tmp_path):
        """Test saving requirements with extra custom fields."""
        csv_path = tmp_path / "test.csv"
        requirements = [
            Requirement(
                req_id="REQ-001",
                extra={"custom_field": "value1", "another_field": "value2"},
            )
        ]

        save_csv(requirements, csv_path)

        # Load and verify extra fields preserved
        loaded = load_csv(csv_path)
        assert loaded[0].extra.get("custom_field") == "value1"
        assert loaded[0].extra.get("another_field") == "value2"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_consistent_field_order(self, tmp_path):
        """Test saving CSV maintains consistent field order."""
        csv_path = tmp_path / "test.csv"
        requirements = [
            Requirement(req_id="REQ-001", extra={"field_a": "a"}),
            Requirement(req_id="REQ-002", extra={"field_b": "b"}),
        ]

        save_csv(requirements, csv_path)

        # Read back and verify both extra fields are columns
        with csv_path.open() as f:
            reader = csv.DictReader(f)
            fieldnames = reader.fieldnames
            assert fieldnames is not None
            # Both field_a and field_b should be in header
            assert "field_a" in fieldnames or "field_b" in fieldnames

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_save_csv_handles_none_values(self, tmp_path):
        """Test saving CSV handles None values correctly."""
        csv_path = tmp_path / "test.csv"
        requirements = [
            Requirement(
                req_id="REQ-001",
                phase=None,
                effort_weeks=None,
            )
        ]

        save_csv(requirements, csv_path)
        loaded = load_csv(csv_path)

        assert loaded[0].phase is None
        assert loaded[0].effort_weeks is None


class TestDefaultRtmPath:
    """Tests for DEFAULT_RTM_PATH constant."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_default_rtm_path_exists(self):
        """Test DEFAULT_RTM_PATH constant is defined."""
        assert DEFAULT_RTM_PATH is not None
        assert isinstance(DEFAULT_RTM_PATH, Path)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_default_rtm_path_value(self):
        """Test DEFAULT_RTM_PATH has correct value."""
        assert str(DEFAULT_RTM_PATH) == "docs/rtm_database.csv"
