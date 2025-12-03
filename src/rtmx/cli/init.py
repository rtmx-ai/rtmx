"""RTMX init command.

Initialize RTM structure in a new project.
"""

from __future__ import annotations

import sys
from pathlib import Path

from rtmx.formatting import Colors


def run_init(force: bool = False) -> None:
    """Run init command.

    Creates the RTM directory structure and sample files.

    Args:
        force: If True, overwrite existing files
    """
    cwd = Path.cwd()

    # Files to create
    rtm_csv = cwd / "docs" / "rtm_database.csv"
    requirements_dir = cwd / "docs" / "requirements"
    config_file = cwd / "rtmx.yaml"

    # Check for existing files
    if not force:
        existing = []
        if rtm_csv.exists():
            existing.append(str(rtm_csv))
        if config_file.exists():
            existing.append(str(config_file))

        if existing:
            print(f"{Colors.YELLOW}Warning: The following files already exist:{Colors.RESET}")
            for f in existing:
                print(f"  {f}")
            print()
            print(f"{Colors.DIM}Use --force to overwrite{Colors.RESET}")
            sys.exit(1)

    # Create directories
    print(f"Creating RTM structure in {cwd}")
    print()

    rtm_csv.parent.mkdir(parents=True, exist_ok=True)
    requirements_dir.mkdir(parents=True, exist_ok=True)

    # Create sample RTM database
    sample_rtm = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-EX-001,EXAMPLE,SAMPLE,Sample requirement for demonstration,Target value here,tests/test_example.py,test_sample,Unit Test,MISSING,MEDIUM,1,This is a sample requirement,1.0,,,developer,v0.1,,,docs/requirements/EXAMPLE/REQ-EX-001.md
"""

    with rtm_csv.open("w") as f:
        f.write(sample_rtm)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Created {rtm_csv}")

    # Create sample requirement file
    sample_req_file = requirements_dir / "EXAMPLE" / "REQ-EX-001.md"
    sample_req_file.parent.mkdir(parents=True, exist_ok=True)

    sample_req_content = """# REQ-EX-001: Sample Requirement

## Description
This is a sample requirement demonstrating the RTMX requirement file format.

## Target
**Metric**: Target value here

## Acceptance Criteria
- [ ] Achieves target value
- [ ] Test implemented and passing
- [ ] Documentation complete

## Implementation
- **Status**: MISSING
- **Phase**: 1
- **Priority**: MEDIUM

## Validation
- **Test**: tests/test_example.py::test_sample
- **Method**: Unit Test

## Dependencies
None

## Notes
This is a sample requirement. Replace with your actual requirements.
"""

    with sample_req_file.open("w") as f:
        f.write(sample_req_content)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Created {sample_req_file}")

    # Create config file
    config_content = """# RTMX Configuration
# See https://github.com/iotactical/rtm for documentation

rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core
  pytest:
    marker_prefix: "req"
    register_markers: true
"""

    with config_file.open("w") as f:
        f.write(config_content)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Created {config_file}")

    print()
    print(f"{Colors.GREEN}✓ RTM initialized successfully!{Colors.RESET}")
    print()
    print("Next steps:")
    print(f"  1. Edit {rtm_csv} to add your requirements")
    print(f"  2. Create requirement spec files in {requirements_dir}")
    print("  3. Run 'rtmx status' to see progress")
    print()
    print("Makefile targets (optional):")
    print("  rtmx makefile >> Makefile")
