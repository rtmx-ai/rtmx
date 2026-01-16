"""Common step definitions shared across BDD scenarios.

These patterns are designed to be portable across BDD frameworks.
When porting to other languages (JS, Go, Rust), translate these
patterns using the target language's BDD runner.
"""

from __future__ import annotations

import subprocess
from pathlib import Path
from typing import Any

from pytest_bdd import given, parsers, then, when


@given("an empty directory")
def empty_directory(project_dir: Path, bdd_context: dict[str, Any]) -> None:
    """Set up an empty project directory without RTMX initialization.

    Used for testing init/setup commands that create the project structure.
    """
    bdd_context["project_dir"] = project_dir
    # Directory is already empty from tmp_path fixture


@given("an initialized RTMX project")
def initialized_project(
    project_dir: Path, rtmx_yaml_content: str, bdd_context: dict[str, Any]
) -> None:
    """Create an initialized RTMX project with config and database."""
    # Create rtmx.yaml
    config_file = project_dir / "rtmx.yaml"
    config_file.write_text(rtmx_yaml_content)

    # Create docs directory and empty database with headers
    docs_dir = project_dir / "docs"
    docs_dir.mkdir(exist_ok=True)

    db_file = docs_dir / "rtm_database.csv"
    header = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id,\n"
    db_file.write_text(header)

    bdd_context["project_dir"] = project_dir
    bdd_context["db_file"] = db_file


@given(parsers.parse("the RTM database has {count:d} requirements"))
def rtm_has_requirements(bdd_context: dict[str, Any], count: int) -> None:
    """Populate the RTM database with the specified number of requirements."""
    db_file = bdd_context.get("db_file")
    if db_file is None:
        raise RuntimeError(
            "Database file not initialized. Use 'Given an initialized RTMX project' first."
        )

    # Read existing content (header)
    content = db_file.read_text()

    # Add requirements
    for i in range(1, count + 1):
        req_line = f"REQ-TEST-{i:03d},TEST,BDD,Test requirement {i},Target {i},tests/test_bdd.py,test_{i},Unit Test,MISSING,MEDIUM,1,Test note {i},0.5,,,,,,,docs/requirements/TEST/REQ-TEST-{i:03d}.md,,\n"
        content += req_line

    db_file.write_text(content)
    bdd_context["requirement_count"] = count


@given(parsers.parse("{complete:d} requirements are COMPLETE"))
def requirements_are_complete(bdd_context: dict[str, Any], complete: int) -> None:
    """Mark the specified number of requirements as COMPLETE."""
    db_file = bdd_context.get("db_file")
    if db_file is None:
        raise RuntimeError("Database file not initialized.")

    content = db_file.read_text()
    lines = content.split("\n")

    # Update status for first N requirements
    updated_lines = [lines[0]]  # Keep header
    completed_count = 0
    for line in lines[1:]:
        if line.strip() and completed_count < complete:
            # Replace MISSING with COMPLETE
            line = line.replace(",MISSING,", ",COMPLETE,")
            completed_count += 1
        updated_lines.append(line)

    db_file.write_text("\n".join(updated_lines))
    bdd_context["complete_count"] = complete


@given(parsers.parse("{complete:d} of {total:d} requirements are COMPLETE"))
def n_of_m_requirements_complete(bdd_context: dict[str, Any], complete: int, total: int) -> None:
    """Set up database with specific complete/total ratio."""
    db_file = bdd_context.get("db_file")
    if db_file is None:
        raise RuntimeError("Database file not initialized.")

    # Read existing header
    content = db_file.read_text()
    lines = content.split("\n")
    header = lines[0]

    # Create requirements with specified complete/incomplete ratio
    new_content = header + "\n"
    for i in range(1, total + 1):
        status = "COMPLETE" if i <= complete else "MISSING"
        req_line = f"REQ-TEST-{i:03d},TEST,BDD,Test requirement {i},Target {i},tests/test_bdd.py,test_{i},Unit Test,{status},MEDIUM,1,Test note {i},0.5,,,,,,,docs/requirements/TEST/REQ-TEST-{i:03d}.md,,\n"
        new_content += req_line

    db_file.write_text(new_content)
    bdd_context["requirement_count"] = total
    bdd_context["complete_count"] = complete


@when(parsers.parse('I run "{command}"'))
def run_command(bdd_context: dict[str, Any], command: str) -> None:
    """Execute a CLI command in the project directory.

    This pattern is portable: the command string is parsed and executed.
    Other language implementations would use equivalent subprocess mechanisms.
    """
    project_dir = bdd_context.get("project_dir")
    if project_dir is None:
        raise RuntimeError("Project directory not set.")

    # Parse command - handle "rtmx <subcommand>" pattern
    parts = command.split()
    cmd = ["python", "-m", "rtmx", *parts[1:]] if parts[0] == "rtmx" else parts

    result = subprocess.run(
        cmd,
        cwd=project_dir,
        capture_output=True,
        text=True,
    )

    bdd_context["result"] = result
    bdd_context["stdout"] = result.stdout
    bdd_context["stderr"] = result.stderr
    bdd_context["exit_code"] = result.returncode


@then("the command should succeed")
def command_succeeds(bdd_context: dict[str, Any]) -> None:
    """Assert the command exited with code 0."""
    result = bdd_context.get("result")
    if result is None:
        raise RuntimeError("No command result. Use 'When I run' first.")

    assert result.returncode == 0, f"Command failed with code {result.returncode}:\n{result.stderr}"


@then(parsers.parse("the exit code should be {code:d}"))
def exit_code_is(bdd_context: dict[str, Any], code: int) -> None:
    """Assert the command exited with specific code."""
    result = bdd_context.get("result")
    if result is None:
        raise RuntimeError("No command result. Use 'When I run' first.")

    assert result.returncode == code, (
        f"Expected exit code {code}, got {result.returncode}\n"
        f"stdout: {result.stdout}\nstderr: {result.stderr}"
    )


@then(parsers.parse('I should see "{text}" in the output'))
def output_contains(bdd_context: dict[str, Any], text: str) -> None:
    """Assert the command output contains specific text."""
    stdout = bdd_context.get("stdout", "")
    stderr = bdd_context.get("stderr", "")
    combined = stdout + stderr

    assert text in combined, f"Expected '{text}' in output. Got:\n{combined}"


@then(parsers.parse('the output should contain "{text}"'))
def output_should_contain(bdd_context: dict[str, Any], text: str) -> None:
    """Alternative pattern for output assertion."""
    output_contains(bdd_context, text)


@then("the command should fail")
def command_fails(bdd_context: dict[str, Any]) -> None:
    """Assert the command exited with non-zero code."""
    result = bdd_context.get("result")
    if result is None:
        raise RuntimeError("No command result. Use 'When I run' first.")

    assert result.returncode != 0, f"Expected command to fail, but it succeeded:\n{result.stdout}"
