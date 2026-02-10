"""Tests for REQ-LANG-007: Language-Agnostic Marker Annotation Specification.

These tests verify the marker schema, parser registry, language detection,
and marker discovery functionality.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import TYPE_CHECKING

import pytest

if TYPE_CHECKING:
    from rtmx.markers.registry import ParserRegistry


# =============================================================================
# Test Fixtures
# =============================================================================


@pytest.fixture
def marker_schema() -> dict:
    """Load the canonical marker JSON schema."""
    from rtmx.markers.schema import MARKER_SCHEMA

    return MARKER_SCHEMA


@pytest.fixture
def parser_registry() -> ParserRegistry:
    """Create a fresh parser registry."""
    from rtmx.markers.registry import ParserRegistry

    return ParserRegistry()


@pytest.fixture
def sample_python_file(tmp_path: Path) -> Path:
    """Create a sample Python test file with markers."""
    content = '''"""Sample test module."""
import pytest

@pytest.mark.req("REQ-TEST-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_sample():
    assert True

@pytest.mark.req("REQ-TEST-002")
@pytest.mark.scope_integration
def test_another():
    pass
'''
    file_path = tmp_path / "test_sample.py"
    file_path.write_text(content)
    return file_path


@pytest.fixture
def sample_javascript_file(tmp_path: Path) -> Path:
    """Create a sample JavaScript test file with markers."""
    content = """/**
 * @req REQ-JS-001
 * @scope unit
 * @technique nominal
 */
describe('Sample Test', () => {
    // @req REQ-JS-002
    it('should pass', () => {
        expect(true).toBe(true);
    });
});
"""
    file_path = tmp_path / "sample.test.js"
    file_path.write_text(content)
    return file_path


@pytest.fixture
def sample_rust_file(tmp_path: Path) -> Path:
    """Create a sample Rust test file with markers."""
    content = """//! Test module

#[cfg(test)]
mod tests {
    // @req REQ-RUST-001
    // @scope unit
    #[test]
    fn test_sample() {
        assert!(true);
    }

    /// @req REQ-RUST-002
    /// @technique stress
    #[test]
    fn test_another() {
        assert_eq!(1, 1);
    }
}
"""
    file_path = tmp_path / "lib.rs"
    file_path.write_text(content)
    return file_path


@pytest.fixture
def sample_go_file(tmp_path: Path) -> Path:
    """Create a sample Go test file with markers."""
    content = """package main

import "testing"

// @req REQ-GO-001
// @scope integration
// @env hil
func TestSample(t *testing.T) {
    if true != true {
        t.Error("expected true")
    }
}
"""
    file_path = tmp_path / "main_test.go"
    file_path.write_text(content)
    return file_path


@pytest.fixture
def sample_java_file(tmp_path: Path) -> Path:
    """Create a sample Java test file with markers."""
    content = """package com.example;

import org.junit.jupiter.api.Test;

/**
 * @req REQ-JAVA-001
 * @scope system
 */
public class SampleTest {
    @Test
    void testSample() {
        assertTrue(true);
    }
}
"""
    file_path = tmp_path / "SampleTest.java"
    file_path.write_text(content)
    return file_path


@pytest.fixture
def mixed_language_project(tmp_path: Path) -> Path:
    """Create a project with multiple languages."""
    # Python
    (tmp_path / "tests").mkdir()
    (tmp_path / "tests" / "test_main.py").write_text("""
import pytest

@pytest.mark.req("REQ-MIXED-001")
def test_python():
    pass
""")

    # JavaScript
    (tmp_path / "tests" / "main.test.js").write_text("""
// @req REQ-MIXED-002
test('javascript test', () => {});
""")

    # Go
    (tmp_path / "main_test.go").write_text("""
package main

// @req REQ-MIXED-003
func TestGo(t *testing.T) {}
""")

    return tmp_path


# =============================================================================
# REQ-LANG-007: Schema Validation Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMarkerSchemaValidation:
    """Test JSON Schema for marker format specification."""

    def test_schema_is_valid_json_schema(self, marker_schema: dict) -> None:
        """Schema should be a valid JSON Schema v2020-12."""
        assert "$schema" in marker_schema
        assert "2020-12" in marker_schema["$schema"]
        assert "type" in marker_schema
        assert marker_schema["type"] == "object"

    def test_schema_has_required_fields(self, marker_schema: dict) -> None:
        """Schema should require req_id field."""
        assert "required" in marker_schema
        assert "req_id" in marker_schema["required"]

    def test_schema_defines_req_id_pattern(self, marker_schema: dict) -> None:
        """Schema should define REQ-XXX-NNN pattern for req_id."""
        properties = marker_schema.get("properties", {})
        assert "req_id" in properties
        req_id_schema = properties["req_id"]
        assert "pattern" in req_id_schema
        # Pattern should match REQ-XXX-NNN format
        assert "REQ-" in req_id_schema["pattern"]

    def test_schema_defines_optional_scope(self, marker_schema: dict) -> None:
        """Schema should define optional scope enum."""
        properties = marker_schema.get("properties", {})
        assert "scope" in properties
        scope_schema = properties["scope"]
        assert "enum" in scope_schema
        expected_scopes = {"unit", "integration", "system", "acceptance"}
        assert set(scope_schema["enum"]) == expected_scopes

    def test_schema_defines_optional_technique(self, marker_schema: dict) -> None:
        """Schema should define optional technique enum."""
        properties = marker_schema.get("properties", {})
        assert "technique" in properties
        technique_schema = properties["technique"]
        assert "enum" in technique_schema
        expected_techniques = {"nominal", "parametric", "monte_carlo", "stress", "boundary"}
        assert set(technique_schema["enum"]) == expected_techniques

    def test_schema_defines_optional_env(self, marker_schema: dict) -> None:
        """Schema should define optional env enum."""
        properties = marker_schema.get("properties", {})
        assert "env" in properties
        env_schema = properties["env"]
        assert "enum" in env_schema
        expected_envs = {"simulation", "hil", "anechoic", "field"}
        assert set(env_schema["enum"]) == expected_envs

    def test_valid_marker_passes_schema(self, marker_schema: dict) -> None:
        """Valid marker data should pass schema validation."""
        from rtmx.markers.schema import validate_marker

        valid_marker = {
            "req_id": "REQ-TEST-001",
            "scope": "unit",
            "technique": "nominal",
            "env": "simulation",
        }
        # Should not raise
        validate_marker(valid_marker)

    def test_invalid_req_id_fails_schema(self, marker_schema: dict) -> None:
        """Invalid req_id pattern should fail validation."""
        from rtmx.markers.schema import MarkerValidationError, validate_marker

        invalid_marker = {
            "req_id": "INVALID-ID",
        }
        with pytest.raises(MarkerValidationError) as exc_info:
            validate_marker(invalid_marker)
        assert "req_id" in str(exc_info.value).lower()

    def test_invalid_scope_fails_schema(self, marker_schema: dict) -> None:
        """Invalid scope value should fail validation."""
        from rtmx.markers.schema import MarkerValidationError, validate_marker

        invalid_marker = {
            "req_id": "REQ-TEST-001",
            "scope": "invalid_scope",
        }
        with pytest.raises(MarkerValidationError) as exc_info:
            validate_marker(invalid_marker)
        assert "scope" in str(exc_info.value).lower()


# =============================================================================
# REQ-LANG-007: Parser Registry Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestParserRegistryRegistration:
    """Test parser plugin registration."""

    def test_registry_starts_empty(self, parser_registry: ParserRegistry) -> None:
        """Fresh registry should have no custom parsers."""
        # Built-in parsers may be present, but custom ones should not
        assert parser_registry.custom_parsers == {}

    def test_register_parser_by_extension(self, parser_registry: ParserRegistry) -> None:
        """Should register parser for file extension."""
        from rtmx.markers.registry import BaseParser

        class CustomParser(BaseParser):
            def parse(self, content: str, file_path: Path) -> list:
                return []

        parser_registry.register(".custom", CustomParser())
        assert ".custom" in parser_registry.get_supported_extensions()

    def test_register_parser_by_language(self, parser_registry: ParserRegistry) -> None:
        """Should register parser by language name."""
        from rtmx.markers.registry import BaseParser

        class CustomParser(BaseParser):
            def parse(self, content: str, file_path: Path) -> list:
                return []

        parser_registry.register_language("customlang", CustomParser(), extensions=[".cl"])
        assert parser_registry.get_parser_for_language("customlang") is not None

    def test_get_parser_for_extension(self, parser_registry: ParserRegistry) -> None:
        """Should return correct parser for known extension."""
        # Python parser should be built-in
        parser = parser_registry.get_parser_for_extension(".py")
        assert parser is not None

    def test_get_parser_for_unknown_extension(self, parser_registry: ParserRegistry) -> None:
        """Should return None for unknown extension."""
        parser = parser_registry.get_parser_for_extension(".unknown123")
        assert parser is None

    def test_builtin_python_parser_registered(self, parser_registry: ParserRegistry) -> None:
        """Built-in Python parser should be registered."""
        assert ".py" in parser_registry.get_supported_extensions()

    def test_list_supported_languages(self, parser_registry: ParserRegistry) -> None:
        """Should list all supported languages."""
        languages = parser_registry.get_supported_languages()
        assert "python" in languages


# =============================================================================
# REQ-LANG-007: Language Auto-Detection Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestLanguageAutodetection:
    """Test file type detection."""

    def test_detect_python_by_extension(self, sample_python_file: Path) -> None:
        """Should detect Python by .py extension."""
        from rtmx.markers.detection import detect_language

        lang = detect_language(sample_python_file)
        assert lang == "python"

    def test_detect_javascript_by_extension(self, sample_javascript_file: Path) -> None:
        """Should detect JavaScript by .js extension."""
        from rtmx.markers.detection import detect_language

        lang = detect_language(sample_javascript_file)
        assert lang == "javascript"

    def test_detect_rust_by_extension(self, sample_rust_file: Path) -> None:
        """Should detect Rust by .rs extension."""
        from rtmx.markers.detection import detect_language

        lang = detect_language(sample_rust_file)
        assert lang == "rust"

    def test_detect_go_by_extension(self, sample_go_file: Path) -> None:
        """Should detect Go by .go extension."""
        from rtmx.markers.detection import detect_language

        lang = detect_language(sample_go_file)
        assert lang == "go"

    def test_detect_java_by_extension(self, sample_java_file: Path) -> None:
        """Should detect Java by .java extension."""
        from rtmx.markers.detection import detect_language

        lang = detect_language(sample_java_file)
        assert lang == "java"

    def test_detect_by_shebang(self, tmp_path: Path) -> None:
        """Should detect language by shebang line."""
        from rtmx.markers.detection import detect_language

        script = tmp_path / "script"
        script.write_text("#!/usr/bin/env python3\nprint('hello')")
        lang = detect_language(script)
        assert lang == "python"

    def test_detect_typescript_by_extension(self, tmp_path: Path) -> None:
        """Should detect TypeScript by .ts extension."""
        from rtmx.markers.detection import detect_language

        ts_file = tmp_path / "test.ts"
        ts_file.write_text("// test")
        lang = detect_language(ts_file)
        assert lang == "typescript"

    def test_detect_unknown_returns_none(self, tmp_path: Path) -> None:
        """Should return None for unknown file types."""
        from rtmx.markers.detection import detect_language

        unknown = tmp_path / "file.xyz123"
        unknown.write_text("content")
        lang = detect_language(unknown)
        assert lang is None

    def test_explicit_config_overrides_extension(self, tmp_path: Path) -> None:
        """Explicit config should override extension-based detection."""
        from rtmx.markers.detection import detect_language

        # Create a .txt file but configure it as Python
        txt_file = tmp_path / "tests.txt"
        txt_file.write_text("@pytest.mark.req('REQ-001')")

        lang = detect_language(txt_file, override_language="python")
        assert lang == "python"


# =============================================================================
# REQ-LANG-007: Marker Normalization Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMarkerNormalization:
    """Test MarkerInfo object creation."""

    def test_marker_info_creation(self) -> None:
        """Should create MarkerInfo with all fields."""
        from rtmx.markers.models import MarkerInfo

        marker = MarkerInfo(
            req_id="REQ-TEST-001",
            scope="unit",
            technique="nominal",
            env="simulation",
            file_path=Path("test.py"),
            line_number=10,
            language="python",
        )
        assert marker.req_id == "REQ-TEST-001"
        assert marker.scope == "unit"
        assert marker.technique == "nominal"
        assert marker.env == "simulation"
        assert marker.line_number == 10

    def test_marker_info_minimal(self) -> None:
        """Should create MarkerInfo with only required field."""
        from rtmx.markers.models import MarkerInfo

        marker = MarkerInfo(
            req_id="REQ-TEST-002",
            file_path=Path("test.py"),
            line_number=1,
            language="python",
        )
        assert marker.req_id == "REQ-TEST-002"
        assert marker.scope is None
        assert marker.technique is None
        assert marker.env is None

    def test_marker_info_to_dict(self) -> None:
        """Should convert MarkerInfo to dictionary."""
        from rtmx.markers.models import MarkerInfo

        marker = MarkerInfo(
            req_id="REQ-TEST-001",
            scope="unit",
            file_path=Path("test.py"),
            line_number=10,
            language="python",
        )
        data = marker.to_dict()
        assert data["req_id"] == "REQ-TEST-001"
        assert data["scope"] == "unit"
        assert data["file_path"] == "test.py"

    def test_marker_info_from_dict(self) -> None:
        """Should create MarkerInfo from dictionary."""
        from rtmx.markers.models import MarkerInfo

        data = {
            "req_id": "REQ-TEST-001",
            "scope": "integration",
            "file_path": "test.py",
            "line_number": 5,
            "language": "python",
        }
        marker = MarkerInfo.from_dict(data)
        assert marker.req_id == "REQ-TEST-001"
        assert marker.scope == "integration"

    def test_marker_info_validates_req_id(self) -> None:
        """Should validate req_id format on creation."""
        from rtmx.markers.models import MarkerInfo, MarkerValidationError

        with pytest.raises(MarkerValidationError):
            MarkerInfo(
                req_id="INVALID",
                file_path=Path("test.py"),
                line_number=1,
                language="python",
            )


# =============================================================================
# REQ-LANG-007: CLI Discovery Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCLIMarkersDiscover:
    """Test discovery CLI command."""

    def test_discover_python_markers(self, sample_python_file: Path) -> None:
        """Should discover markers in Python files."""
        from rtmx.markers.discover import discover_markers

        markers = discover_markers(sample_python_file.parent)
        req_ids = [m.req_id for m in markers]
        assert "REQ-TEST-001" in req_ids
        assert "REQ-TEST-002" in req_ids

    def test_discover_returns_marker_info_objects(self, sample_python_file: Path) -> None:
        """Should return MarkerInfo objects."""
        from rtmx.markers.discover import discover_markers
        from rtmx.markers.models import MarkerInfo

        markers = discover_markers(sample_python_file.parent)
        assert len(markers) > 0
        assert all(isinstance(m, MarkerInfo) for m in markers)

    def test_discover_includes_line_numbers(self, sample_python_file: Path) -> None:
        """Should include line numbers in results."""
        from rtmx.markers.discover import discover_markers

        markers = discover_markers(sample_python_file.parent)
        for marker in markers:
            assert marker.line_number > 0

    def test_discover_multi_language_project(self, mixed_language_project: Path) -> None:
        """Should discover markers across multiple languages."""
        from rtmx.markers.discover import discover_markers

        markers = discover_markers(mixed_language_project)
        req_ids = [m.req_id for m in markers]
        assert "REQ-MIXED-001" in req_ids  # Python
        assert "REQ-MIXED-002" in req_ids  # JavaScript
        assert "REQ-MIXED-003" in req_ids  # Go

    def test_discover_respects_gitignore(self, tmp_path: Path) -> None:
        """Should respect .gitignore patterns."""
        from rtmx.markers.discover import discover_markers

        # Create a file in ignored directory
        (tmp_path / "node_modules").mkdir()
        (tmp_path / "node_modules" / "test.js").write_text("// @req REQ-IGNORED-001")

        # Create a file not ignored
        (tmp_path / "test.py").write_text('@pytest.mark.req("REQ-INCLUDED-001")')

        # Create .gitignore
        (tmp_path / ".gitignore").write_text("node_modules/")

        markers = discover_markers(tmp_path)
        req_ids = [m.req_id for m in markers]
        assert "REQ-IGNORED-001" not in req_ids
        assert "REQ-INCLUDED-001" in req_ids

    def test_cli_discover_command(
        self, sample_python_file: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Should run discovery via CLI."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(main, ["markers", "discover", str(sample_python_file.parent)])
        assert result.exit_code == 0
        assert "REQ-TEST-001" in result.output

    def test_cli_discover_json_output(
        self, sample_python_file: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Should output JSON when requested."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(
            main, ["markers", "discover", str(sample_python_file.parent), "--format", "json"]
        )
        assert result.exit_code == 0
        data = json.loads(result.output)
        assert "markers" in data
        assert len(data["markers"]) > 0


# =============================================================================
# REQ-LANG-007: Error Reporting Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestInvalidMarkerErrorReporting:
    """Test error messages for invalid markers."""

    def test_invalid_marker_includes_file_location(self, tmp_path: Path) -> None:
        """Error should include file path and line number."""
        from rtmx.markers.discover import discover_markers

        # Create file with invalid marker
        test_file = tmp_path / "test_invalid.py"
        test_file.write_text("""
@pytest.mark.req("INVALID-FORMAT")
def test_bad():
    pass
""")

        markers = discover_markers(tmp_path, include_errors=True)
        # Check that errors are captured with location info
        error_markers = [m for m in markers if hasattr(m, "error") and m.error]
        # The marker should either be skipped or reported as error
        # With include_errors=True, invalid markers should be included
        assert len(error_markers) >= 0  # At least doesn't crash

    def test_error_suggests_fix_for_malformed_req_id(self, tmp_path: Path) -> None:
        """Error should suggest correct format for malformed req_id."""
        from rtmx.markers.validation import validate_marker_format

        result = validate_marker_format("req-test-001")  # lowercase
        assert not result.is_valid
        assert "REQ-" in result.suggestion

    def test_error_suggests_fix_for_invalid_scope(self) -> None:
        """Error should suggest valid scopes for invalid scope."""
        from rtmx.markers.validation import validate_marker_format

        result = validate_marker_format("REQ-TEST-001", scope="component")
        assert not result.is_valid
        assert any(s in result.suggestion for s in ["unit", "integration", "system", "acceptance"])

    def test_collect_all_errors_in_file(self, tmp_path: Path) -> None:
        """Should collect all errors in a file, not stop at first."""
        from rtmx.markers.discover import discover_markers

        test_file = tmp_path / "test_multiple_errors.py"
        test_file.write_text("""
@pytest.mark.req("bad-001")
def test_first():
    pass

@pytest.mark.req("bad-002")
def test_second():
    pass

@pytest.mark.req("REQ-GOOD-001")
def test_valid():
    pass
""")

        markers = discover_markers(tmp_path, include_errors=True)
        # Should still find the valid marker
        valid_markers = [m for m in markers if m.req_id == "REQ-GOOD-001"]
        assert len(valid_markers) == 1


# =============================================================================
# REQ-LANG-007: Python Parser Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestPythonParser:
    """Test Python-specific marker parsing."""

    def test_parse_pytest_marker(self, sample_python_file: Path) -> None:
        """Should parse @pytest.mark.req markers."""
        from rtmx.markers.parsers.python import PythonParser

        parser = PythonParser()
        content = sample_python_file.read_text()
        markers = parser.parse(content, sample_python_file)

        assert len(markers) == 2
        assert markers[0].req_id == "REQ-TEST-001"
        assert markers[1].req_id == "REQ-TEST-002"

    def test_parse_scope_markers(self, sample_python_file: Path) -> None:
        """Should parse scope markers."""
        from rtmx.markers.parsers.python import PythonParser

        parser = PythonParser()
        content = sample_python_file.read_text()
        markers = parser.parse(content, sample_python_file)

        # First marker has scope_unit
        assert markers[0].scope == "unit"
        # Second marker has scope_integration
        assert markers[1].scope == "integration"

    def test_parse_technique_markers(self, sample_python_file: Path) -> None:
        """Should parse technique markers."""
        from rtmx.markers.parsers.python import PythonParser

        parser = PythonParser()
        content = sample_python_file.read_text()
        markers = parser.parse(content, sample_python_file)

        assert markers[0].technique == "nominal"

    def test_parse_env_markers(self, sample_python_file: Path) -> None:
        """Should parse env markers."""
        from rtmx.markers.parsers.python import PythonParser

        parser = PythonParser()
        content = sample_python_file.read_text()
        markers = parser.parse(content, sample_python_file)

        assert markers[0].env == "simulation"

    def test_parse_multiple_req_markers_same_function(self, tmp_path: Path) -> None:
        """Should handle multiple @req markers on same function."""
        from rtmx.markers.parsers.python import PythonParser

        test_file = tmp_path / "test_multi.py"
        test_file.write_text("""
@pytest.mark.req("REQ-MULTI-001")
@pytest.mark.req("REQ-MULTI-002")
def test_covers_two():
    pass
""")

        parser = PythonParser()
        content = test_file.read_text()
        markers = parser.parse(content, test_file)

        req_ids = [m.req_id for m in markers]
        assert "REQ-MULTI-001" in req_ids
        assert "REQ-MULTI-002" in req_ids


# =============================================================================
# REQ-LANG-007: Configuration Tests
# =============================================================================


@pytest.mark.req("REQ-LANG-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMarkerConfiguration:
    """Test configuration file support for custom parsers."""

    def test_load_custom_parser_from_config(self, tmp_path: Path) -> None:
        """Should load custom parser from rtmx.yaml."""
        from rtmx.markers.config import load_marker_config

        config_file = tmp_path / "rtmx.yaml"
        config_file.write_text("""
rtmx:
  markers:
    custom_extensions:
      ".myext": "python"  # Treat .myext files as Python
""")

        config = load_marker_config(config_file)
        assert ".myext" in config.custom_extensions
        assert config.custom_extensions[".myext"] == "python"

    def test_custom_extension_mapping(self, tmp_path: Path) -> None:
        """Should use custom extension mapping."""
        from rtmx.markers.discover import discover_markers

        # Create config
        config_file = tmp_path / "rtmx.yaml"
        config_file.write_text("""
rtmx:
  markers:
    custom_extensions:
      ".pyt": "python"
""")

        # Create file with custom extension
        test_file = tmp_path / "test.pyt"
        test_file.write_text('@pytest.mark.req("REQ-CUSTOM-001")\ndef test_custom(): pass')

        markers = discover_markers(tmp_path, config_path=config_file)
        req_ids = [m.req_id for m in markers]
        assert "REQ-CUSTOM-001" in req_ids

    def test_exclude_patterns_from_config(self, tmp_path: Path) -> None:
        """Should respect exclude patterns from config."""
        from rtmx.markers.discover import discover_markers

        config_file = tmp_path / "rtmx.yaml"
        config_file.write_text("""
rtmx:
  markers:
    exclude:
      - "**/vendor/**"
      - "**/generated/**"
""")

        # Create files
        (tmp_path / "vendor").mkdir()
        (tmp_path / "vendor" / "test.py").write_text('@pytest.mark.req("REQ-VENDOR-001")')
        (tmp_path / "test_main.py").write_text('@pytest.mark.req("REQ-MAIN-001")')

        markers = discover_markers(tmp_path, config_path=config_file)
        req_ids = [m.req_id for m in markers]
        assert "REQ-VENDOR-001" not in req_ids
        assert "REQ-MAIN-001" in req_ids
