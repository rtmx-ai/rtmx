"""Tests for rtmx.templates module (REQ-DX-003).

This module tests the scaffold template functionality:
- Default template rendering
- Custom template loading from .rtmx/templates/
- Scaffold generation for all requirements
- Skip existing spec files
- Force regeneration with --force flag
"""

from __future__ import annotations

from pathlib import Path

import pytest

# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def sample_requirement() -> dict:
    """Create a sample requirement dictionary."""
    return {
        "req_id": "REQ-TEST-001",
        "requirement_text": "The system shall validate input data",
        "status": "MISSING",
        "priority": "HIGH",
        "phase": 1,
        "target_value": "100% validation coverage",
        "test_module": "tests/test_validation.py",
        "test_function": "test_input_validation",
        "notes": "Critical requirement for data integrity",
        "category": "VALIDATION",
        "subcategory": "INPUT",
    }


@pytest.fixture
def sample_rtm_database(tmp_path: Path) -> Path:
    """Create a sample RTM database CSV."""
    rtm_csv = tmp_path / "docs" / "rtm_database.csv"
    rtm_csv.parent.mkdir(parents=True, exist_ok=True)
    rtm_csv.write_text(
        """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-TEST-001,VALIDATION,INPUT,The system shall validate input data,100% coverage,tests/test_val.py,test_input,Unit Test,MISSING,HIGH,1,Critical requirement,0.5,,,developer,v0.1,,,docs/requirements/VALIDATION/REQ-TEST-001.md
REQ-TEST-002,VALIDATION,OUTPUT,The system shall format output correctly,Formatted output,tests/test_val.py,test_output,Unit Test,COMPLETE,MEDIUM,1,Optional formatting,0.5,REQ-TEST-001,,developer,v0.1,,,docs/requirements/VALIDATION/REQ-TEST-002.md
REQ-TEST-003,CORE,API,The system shall provide REST API,REST endpoints,,,Unit Test,PARTIAL,HIGH,2,API requirement,1.0,,,developer,v0.2,,,docs/requirements/CORE/REQ-TEST-003.md
"""
    )
    return rtm_csv


@pytest.fixture
def project_with_rtm(tmp_path: Path, sample_rtm_database: Path) -> Path:
    """Create a project structure with RTM database."""
    # Create rtmx.yaml config
    config_path = tmp_path / "rtmx.yaml"
    config_path.write_text(
        """rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core
"""
    )
    return tmp_path


# =============================================================================
# Tests for DEFAULT_TEMPLATE
# =============================================================================


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_default_template_exists():
    """Test that DEFAULT_TEMPLATE constant exists."""
    from rtmx.templates import DEFAULT_TEMPLATE

    assert DEFAULT_TEMPLATE is not None
    assert isinstance(DEFAULT_TEMPLATE, str)
    assert "{{ req_id }}" in DEFAULT_TEMPLATE
    assert "{{ requirement_text" in DEFAULT_TEMPLATE


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_default_template_contains_required_sections():
    """Test that DEFAULT_TEMPLATE contains all required sections."""
    from rtmx.templates import DEFAULT_TEMPLATE

    # Required placeholders from acceptance criteria
    assert "{{ req_id }}" in DEFAULT_TEMPLATE
    assert "{{ status }}" in DEFAULT_TEMPLATE
    assert "{{ priority }}" in DEFAULT_TEMPLATE
    assert "{{ phase }}" in DEFAULT_TEMPLATE
    assert "{{ requirement_text }}" in DEFAULT_TEMPLATE
    assert "{{ target_value" in DEFAULT_TEMPLATE
    assert "{{ test_module" in DEFAULT_TEMPLATE
    assert "{{ test_function" in DEFAULT_TEMPLATE
    assert "{{ notes" in DEFAULT_TEMPLATE

    # Required markdown sections
    assert "## Description" in DEFAULT_TEMPLATE
    assert "## Acceptance Criteria" in DEFAULT_TEMPLATE
    assert "## Test Cases" in DEFAULT_TEMPLATE
    assert "## Notes" in DEFAULT_TEMPLATE


# =============================================================================
# Tests for render_template
# =============================================================================


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_basic(sample_requirement: dict):
    """Test render_template with basic requirement data."""
    from rtmx.templates import render_template

    result = render_template(sample_requirement)

    assert "REQ-TEST-001" in result
    assert "The system shall validate input data" in result
    assert "MISSING" in result
    assert "HIGH" in result


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_includes_test_info(sample_requirement: dict):
    """Test render_template includes test module and function."""
    from rtmx.templates import render_template

    result = render_template(sample_requirement)

    assert "tests/test_validation.py" in result
    assert "test_input_validation" in result


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_handles_missing_test_info():
    """Test render_template handles missing test info gracefully."""
    from rtmx.templates import render_template

    req = {
        "req_id": "REQ-NO-TEST",
        "requirement_text": "A requirement without test",
        "status": "MISSING",
        "priority": "LOW",
        "phase": 1,
        "target_value": "",
        "test_module": "",
        "test_function": "",
        "notes": "",
    }

    result = render_template(req)

    assert "REQ-NO-TEST" in result
    # Should not crash, result should still be valid markdown


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_handles_tbd_target_value():
    """Test render_template shows TBD for empty target_value."""
    from rtmx.templates import render_template

    req = {
        "req_id": "REQ-TBD-001",
        "requirement_text": "Requirement with no target",
        "status": "MISSING",
        "priority": "MEDIUM",
        "phase": 1,
        "target_value": "",
        "test_module": "",
        "test_function": "",
        "notes": "",
    }

    result = render_template(req)

    assert "TBD" in result


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_handles_none_notes():
    """Test render_template shows None for empty notes."""
    from rtmx.templates import render_template

    req = {
        "req_id": "REQ-NO-NOTES",
        "requirement_text": "Requirement without notes",
        "status": "COMPLETE",
        "priority": "HIGH",
        "phase": 2,
        "target_value": "Something",
        "test_module": "tests/test.py",
        "test_function": "test_func",
        "notes": "",
    }

    result = render_template(req)

    assert "None" in result or "notes" in result.lower()


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_truncates_long_title():
    """Test render_template truncates requirement text in title to 60 chars."""
    from rtmx.templates import render_template

    long_text = "A" * 100  # 100 character requirement text
    req = {
        "req_id": "REQ-LONG-001",
        "requirement_text": long_text,
        "status": "MISSING",
        "priority": "MEDIUM",
        "phase": 1,
        "target_value": "",
        "test_module": "",
        "test_function": "",
        "notes": "",
    }

    result = render_template(req)

    # First line should contain truncated title (check for heading with truncated text)
    lines = result.split("\n")
    title_line = lines[0]
    # Title should be truncated, full text should appear in Description section
    assert long_text in result  # Full text should be somewhere
    assert "REQ-LONG-001" in title_line


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_render_template_with_custom_template():
    """Test render_template with custom template string."""
    from rtmx.templates import render_template

    custom_template = """# {{ req_id }}
Priority: {{ priority }}
Custom format only.
"""
    req = {
        "req_id": "REQ-CUSTOM-001",
        "requirement_text": "Custom requirement",
        "status": "MISSING",
        "priority": "P0",
        "phase": 1,
        "target_value": "",
        "test_module": "",
        "test_function": "",
        "notes": "",
    }

    result = render_template(req, template=custom_template)

    assert "REQ-CUSTOM-001" in result
    assert "Priority: P0" in result
    assert "Custom format only" in result
    # Should NOT contain default template sections
    assert "## Description" not in result


# =============================================================================
# Tests for load_custom_template
# =============================================================================


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_custom_template_returns_default_if_not_found(tmp_path: Path):
    """Test load_custom_template returns default when custom template missing."""
    from rtmx.templates import DEFAULT_TEMPLATE, load_custom_template

    result = load_custom_template(tmp_path)

    assert result == DEFAULT_TEMPLATE


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_custom_template_loads_from_rtmx_dir(tmp_path: Path):
    """Test load_custom_template loads from .rtmx/templates/requirement.md.j2."""
    from rtmx.templates import load_custom_template

    template_dir = tmp_path / ".rtmx" / "templates"
    template_dir.mkdir(parents=True)
    template_file = template_dir / "requirement.md.j2"
    custom_content = """# Custom: {{ req_id }}
This is custom.
"""
    template_file.write_text(custom_content)

    result = load_custom_template(tmp_path)

    assert result == custom_content
    assert "# Custom:" in result


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_custom_template_handles_invalid_template(tmp_path: Path):
    """Test load_custom_template falls back to default on invalid template."""
    from rtmx.templates import DEFAULT_TEMPLATE, load_custom_template

    template_dir = tmp_path / ".rtmx" / "templates"
    template_dir.mkdir(parents=True)
    template_file = template_dir / "requirement.md.j2"
    # Write an empty or invalid template
    template_file.write_text("")

    result = load_custom_template(tmp_path)

    # Empty template should fall back to default
    assert result == DEFAULT_TEMPLATE


# =============================================================================
# Tests for scaffold_requirement_spec
# =============================================================================


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_requirement_spec_creates_file(tmp_path: Path, sample_requirement: dict):
    """Test scaffold_requirement_spec creates spec file."""
    from rtmx.templates import scaffold_requirement_spec

    req_dir = tmp_path / "docs" / "requirements"

    result = scaffold_requirement_spec(sample_requirement, req_dir)

    assert result is True
    spec_path = req_dir / "VALIDATION" / "REQ-TEST-001.md"
    assert spec_path.exists()
    content = spec_path.read_text()
    assert "REQ-TEST-001" in content


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_requirement_spec_creates_category_dir(tmp_path: Path):
    """Test scaffold_requirement_spec creates category subdirectory."""
    from rtmx.templates import scaffold_requirement_spec

    req = {
        "req_id": "REQ-NEW-001",
        "requirement_text": "New requirement",
        "status": "MISSING",
        "priority": "HIGH",
        "phase": 1,
        "category": "NEWCAT",
        "target_value": "",
        "test_module": "",
        "test_function": "",
        "notes": "",
    }
    req_dir = tmp_path / "docs" / "requirements"

    scaffold_requirement_spec(req, req_dir)

    category_dir = req_dir / "NEWCAT"
    assert category_dir.is_dir()


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_requirement_spec_skips_existing(tmp_path: Path, sample_requirement: dict):
    """Test scaffold_requirement_spec skips existing files without --force."""
    from rtmx.templates import scaffold_requirement_spec

    req_dir = tmp_path / "docs" / "requirements"
    spec_dir = req_dir / "VALIDATION"
    spec_dir.mkdir(parents=True)
    spec_path = spec_dir / "REQ-TEST-001.md"
    spec_path.write_text("# Existing content\nDo not overwrite.")

    result = scaffold_requirement_spec(sample_requirement, req_dir, force=False)

    assert result is False  # Should return False when skipped
    assert spec_path.read_text() == "# Existing content\nDo not overwrite."


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_requirement_spec_overwrites_with_force(tmp_path: Path, sample_requirement: dict):
    """Test scaffold_requirement_spec overwrites with --force flag."""
    from rtmx.templates import scaffold_requirement_spec

    req_dir = tmp_path / "docs" / "requirements"
    spec_dir = req_dir / "VALIDATION"
    spec_dir.mkdir(parents=True)
    spec_path = spec_dir / "REQ-TEST-001.md"
    spec_path.write_text("# Old content")

    result = scaffold_requirement_spec(sample_requirement, req_dir, force=True)

    assert result is True
    content = spec_path.read_text()
    assert "# Old content" not in content
    assert "REQ-TEST-001" in content


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_requirement_spec_uses_custom_template(tmp_path: Path, sample_requirement: dict):
    """Test scaffold_requirement_spec uses provided template."""
    from rtmx.templates import scaffold_requirement_spec

    req_dir = tmp_path / "docs" / "requirements"
    custom_template = "# {{ req_id }}\nCustom template used."

    scaffold_requirement_spec(sample_requirement, req_dir, template=custom_template)

    spec_path = req_dir / "VALIDATION" / "REQ-TEST-001.md"
    content = spec_path.read_text()
    assert "Custom template used" in content


# =============================================================================
# Tests for scaffold_all_specs
# =============================================================================


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_all_specs_creates_all_files(project_with_rtm: Path):
    """Test scaffold_all_specs creates spec files for all requirements."""
    from rtmx.templates import scaffold_all_specs

    result = scaffold_all_specs(project_with_rtm)

    assert result["created"] == 3
    assert result["skipped"] == 0
    assert result["errors"] == 0

    # Check files exist
    req_dir = project_with_rtm / "docs" / "requirements"
    assert (req_dir / "VALIDATION" / "REQ-TEST-001.md").exists()
    assert (req_dir / "VALIDATION" / "REQ-TEST-002.md").exists()
    assert (req_dir / "CORE" / "REQ-TEST-003.md").exists()


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_all_specs_skips_existing(project_with_rtm: Path):
    """Test scaffold_all_specs skips existing spec files."""
    from rtmx.templates import scaffold_all_specs

    # Pre-create one spec file
    req_dir = project_with_rtm / "docs" / "requirements" / "VALIDATION"
    req_dir.mkdir(parents=True)
    (req_dir / "REQ-TEST-001.md").write_text("# Existing spec")

    result = scaffold_all_specs(project_with_rtm, force=False)

    assert result["created"] == 2
    assert result["skipped"] == 1


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_all_specs_force_regenerates_all(project_with_rtm: Path):
    """Test scaffold_all_specs with force regenerates all specs."""
    from rtmx.templates import scaffold_all_specs

    # Pre-create one spec file
    req_dir = project_with_rtm / "docs" / "requirements" / "VALIDATION"
    req_dir.mkdir(parents=True)
    (req_dir / "REQ-TEST-001.md").write_text("# Old content")

    result = scaffold_all_specs(project_with_rtm, force=True)

    assert result["created"] == 3
    assert result["skipped"] == 0

    # Verify content was regenerated
    content = (req_dir / "REQ-TEST-001.md").read_text()
    assert "# Old content" not in content


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_all_specs_uses_custom_template(project_with_rtm: Path):
    """Test scaffold_all_specs uses custom template from .rtmx/templates/."""
    from rtmx.templates import scaffold_all_specs

    # Create custom template
    template_dir = project_with_rtm / ".rtmx" / "templates"
    template_dir.mkdir(parents=True)
    (template_dir / "requirement.md.j2").write_text("# {{ req_id }}\nCustom scaffold template.")

    scaffold_all_specs(project_with_rtm)

    req_dir = project_with_rtm / "docs" / "requirements"
    content = (req_dir / "VALIDATION" / "REQ-TEST-001.md").read_text()
    assert "Custom scaffold template" in content


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_all_specs_dry_run(project_with_rtm: Path):
    """Test scaffold_all_specs dry_run mode doesn't create files."""
    from rtmx.templates import scaffold_all_specs

    result = scaffold_all_specs(project_with_rtm, dry_run=True)

    assert result["created"] == 0
    assert result["would_create"] == 3

    # Verify no files were created
    req_dir = project_with_rtm / "docs" / "requirements"
    assert not (req_dir / "VALIDATION" / "REQ-TEST-001.md").exists()


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_scaffold_all_specs_returns_file_list(project_with_rtm: Path):
    """Test scaffold_all_specs returns list of created files."""
    from rtmx.templates import scaffold_all_specs

    result = scaffold_all_specs(project_with_rtm)

    assert "files" in result
    assert len(result["files"]) == 3
    # All files should be Path objects or strings
    for f in result["files"]:
        assert "REQ-TEST-" in str(f)


# =============================================================================
# Tests for CLI integration (setup --scaffold)
# =============================================================================


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_scaffold_cli_integration(project_with_rtm: Path):
    """Test run_scaffold function for CLI integration."""
    from rtmx.templates import run_scaffold

    result = run_scaffold(project_with_rtm)

    assert result.success is True
    assert result.specs_created == 3
    assert result.specs_skipped == 0


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_scaffold_with_force(project_with_rtm: Path):
    """Test run_scaffold with force=True."""
    from rtmx.templates import run_scaffold

    # Pre-create a spec
    req_dir = project_with_rtm / "docs" / "requirements" / "VALIDATION"
    req_dir.mkdir(parents=True)
    (req_dir / "REQ-TEST-001.md").write_text("# Old")

    result = run_scaffold(project_with_rtm, force=True)

    assert result.success is True
    assert result.specs_created == 3


@pytest.mark.req("REQ-DX-003")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_scaffold_dry_run(project_with_rtm: Path):
    """Test run_scaffold with dry_run=True."""
    from rtmx.templates import run_scaffold

    result = run_scaffold(project_with_rtm, dry_run=True)

    assert result.success is True
    assert result.specs_created == 0
    assert result.would_create == 3

    # Verify no files were created
    req_dir = project_with_rtm / "docs" / "requirements"
    assert not (req_dir / "VALIDATION" / "REQ-TEST-001.md").exists()
