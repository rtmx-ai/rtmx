"""Template rendering for requirement specification scaffolding (REQ-DX-003).

This module provides Jinja2-based template rendering for auto-generating
requirement specification files from RTM database entries.

Features:
- Default template with standard sections
- Custom template support via .rtmx/templates/requirement.md.j2
- Batch scaffolding for all requirements
- Safe handling of existing files (skip by default, --force to overwrite)
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, TypedDict


class ScaffoldDict(TypedDict):
    """Typed dictionary for scaffold operation results."""

    created: int
    skipped: int
    would_create: int
    errors: int
    files: list[Path]


# Default template for requirement specification files
DEFAULT_TEMPLATE = """# {{ req_id }}: {{ requirement_text[:60] }}{% if requirement_text|length > 60 %}...{% endif %}

## Status: {{ status }}
## Priority: {{ priority }}
## Phase: {{ phase }}

## Description
{{ requirement_text }}

## Acceptance Criteria
- [ ] {{ target_value or "TBD" }}

## Test Cases
{% if test_module and test_function -%}
- `{{ test_module }}::{{ test_function }}`
{% else -%}
- No test cases defined yet
{% endif %}

## Notes
{{ notes or "None" }}
"""


@dataclass
class ScaffoldResult:
    """Result of scaffold operation."""

    success: bool = True
    specs_created: int = 0
    specs_skipped: int = 0
    would_create: int = 0
    errors: list[str] = field(default_factory=list)
    files: list[Path] = field(default_factory=list)


def _get_jinja_env():
    """Get Jinja2 environment, lazy import to handle optional dependency."""
    try:
        from jinja2 import Environment
    except ImportError as e:
        raise ImportError(
            "Jinja2 is required for template rendering. Install with: pip install rtmx[agents]"
        ) from e

    return Environment(autoescape=False)


def render_template(requirement: dict[str, Any], template: str | None = None) -> str:
    """Render a requirement specification using Jinja2 template.

    Args:
        requirement: Dictionary containing requirement fields
        template: Optional custom template string. Uses DEFAULT_TEMPLATE if None.

    Returns:
        Rendered markdown string
    """
    env = _get_jinja_env()
    template_str = template if template else DEFAULT_TEMPLATE
    tmpl = env.from_string(template_str)

    # Ensure all expected fields have values (avoid undefined errors)
    context = {
        "req_id": requirement.get("req_id", ""),
        "requirement_text": requirement.get("requirement_text", ""),
        "status": requirement.get("status", "MISSING"),
        "priority": requirement.get("priority", "MEDIUM"),
        "phase": requirement.get("phase", ""),
        "target_value": requirement.get("target_value", ""),
        "test_module": requirement.get("test_module", ""),
        "test_function": requirement.get("test_function", ""),
        "notes": requirement.get("notes", ""),
        "category": requirement.get("category", ""),
        "subcategory": requirement.get("subcategory", ""),
    }

    return tmpl.render(**context)


def load_custom_template(project_path: Path) -> str:
    """Load custom template from .rtmx/templates/requirement.md.j2.

    Args:
        project_path: Path to project root

    Returns:
        Custom template string if found and valid, otherwise DEFAULT_TEMPLATE
    """
    template_path = project_path / ".rtmx" / "templates" / "requirement.md.j2"

    if template_path.exists():
        content = template_path.read_text()
        # Only strip and check for empty - don't strip the returned content
        if content.strip():
            return content

    return DEFAULT_TEMPLATE


def scaffold_requirement_spec(
    requirement: dict[str, Any],
    requirements_dir: Path,
    force: bool = False,
    template: str | None = None,
) -> bool:
    """Create a specification file for a single requirement.

    Args:
        requirement: Dictionary containing requirement fields
        requirements_dir: Base directory for requirement specs
        force: If True, overwrite existing files
        template: Optional custom template string

    Returns:
        True if file was created, False if skipped
    """
    req_id = requirement.get("req_id", "")
    category = requirement.get("category", "GENERAL")

    # Create category subdirectory
    category_dir = requirements_dir / category
    category_dir.mkdir(parents=True, exist_ok=True)

    # Determine spec file path
    spec_path = category_dir / f"{req_id}.md"

    # Skip if exists and not forcing
    if spec_path.exists() and not force:
        return False

    # Render and write
    content = render_template(requirement, template)
    spec_path.write_text(content)

    return True


def scaffold_all_specs(
    project_path: Path,
    force: bool = False,
    dry_run: bool = False,
) -> ScaffoldDict:
    """Scaffold specification files for all requirements in the RTM database.

    Args:
        project_path: Path to project root
        force: If True, overwrite existing spec files
        dry_run: If True, don't create files, just report what would happen

    Returns:
        Dictionary with:
        - created: Number of files created
        - skipped: Number of files skipped (already exist)
        - would_create: Number of files that would be created (dry_run only)
        - errors: Number of errors encountered
        - files: List of created file paths
    """
    from rtmx.config import load_config
    from rtmx.models import RTMDatabase

    result: ScaffoldDict = {
        "created": 0,
        "skipped": 0,
        "would_create": 0,
        "errors": 0,
        "files": [],
    }

    # Load config and database
    config_path = project_path / "rtmx.yaml"
    if not config_path.exists():
        config_path = project_path / ".rtmx.yaml"

    config = load_config(config_path if config_path.exists() else None)

    db_path = project_path / config.database
    if not db_path.exists():
        result["errors"] = 1
        return result

    db = RTMDatabase.load(db_path)

    # Load custom template if available
    template = load_custom_template(project_path)

    # Determine requirements directory
    requirements_dir = project_path / config.requirements_dir

    # Process each requirement
    for req in db:
        req_dict = req.to_dict()
        category = req_dict.get("category", "GENERAL")
        req_id = req_dict.get("req_id", "")

        spec_path = requirements_dir / category / f"{req_id}.md"

        if dry_run:
            if not spec_path.exists() or force:
                result["would_create"] += 1
                result["files"].append(spec_path)
            else:
                result["skipped"] += 1
        else:
            if scaffold_requirement_spec(req_dict, requirements_dir, force, template):
                result["created"] += 1
                result["files"].append(spec_path)
            else:
                result["skipped"] += 1

    return result


def run_scaffold(
    project_path: Path,
    force: bool = False,
    dry_run: bool = False,
) -> ScaffoldResult:
    """CLI-friendly wrapper for scaffold_all_specs.

    Args:
        project_path: Path to project root
        force: If True, overwrite existing spec files
        dry_run: If True, don't create files, just report what would happen

    Returns:
        ScaffoldResult with operation details
    """
    from rtmx.formatting import Colors, header

    print(header("Scaffold Requirement Specs", "="))
    print()

    if dry_run:
        print(f"{Colors.YELLOW}DRY RUN - no files will be created{Colors.RESET}")
    if force:
        print(f"{Colors.YELLOW}FORCE MODE - existing files will be overwritten{Colors.RESET}")
    print()

    result = ScaffoldResult()

    try:
        scaffold_result = scaffold_all_specs(project_path, force, dry_run)

        result.specs_created = scaffold_result["created"]
        result.specs_skipped = scaffold_result["skipped"]
        result.would_create = scaffold_result["would_create"]
        result.files = scaffold_result["files"]

        if scaffold_result["errors"] > 0:
            result.success = False
            result.errors.append("Failed to load RTM database")

        # Print results
        if dry_run:
            print(f"Would create: {result.would_create} spec files")
            print(f"Would skip: {result.specs_skipped} existing files")
        else:
            print(f"{Colors.GREEN}Created: {result.specs_created} spec files{Colors.RESET}")
            print(f"Skipped: {result.specs_skipped} existing files")

        if result.files:
            print()
            print("Files:")
            for f in result.files[:10]:  # Show first 10
                print(f"  - {f}")
            if len(result.files) > 10:
                print(f"  ... and {len(result.files) - 10} more")

    except Exception as e:
        result.success = False
        result.errors.append(str(e))
        print(f"{Colors.RED}Error: {e}{Colors.RESET}")

    print()
    return result
