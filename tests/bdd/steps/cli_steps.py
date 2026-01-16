"""CLI-specific step definitions for RTMX BDD tests.

These steps extend the common patterns with CLI-specific behavior.
"""

from __future__ import annotations

from typing import Any

from pytest_bdd import given, parsers, then


@given("verbose output is enabled")
def enable_verbose(bdd_context: dict[str, Any]) -> None:
    """Enable verbose output for subsequent commands."""
    bdd_context["verbose"] = True


@given(parsers.parse("the output format is {format_name}"))
def set_output_format(bdd_context: dict[str, Any], format_name: str) -> None:
    """Set the expected output format for subsequent commands."""
    bdd_context["output_format"] = format_name


@then("the output should be valid JSON")
def output_is_json(bdd_context: dict[str, Any]) -> None:
    """Assert the command output is valid JSON."""
    import json

    stdout = bdd_context.get("stdout", "")
    try:
        json.loads(stdout)
    except json.JSONDecodeError as e:
        raise AssertionError(f"Output is not valid JSON: {e}\nOutput: {stdout}") from e


@then(parsers.parse("the output should show {percentage}% completion"))
def output_shows_percentage(bdd_context: dict[str, Any], percentage: str) -> None:
    """Assert the output shows specific completion percentage."""
    stdout = bdd_context.get("stdout", "")
    stderr = bdd_context.get("stderr", "")
    combined = stdout + stderr

    # Look for percentage pattern
    assert f"{percentage}%" in combined, f"Expected '{percentage}%' in output. Got:\n{combined}"


@then("I should see the completion percentage")
def see_completion_percentage(bdd_context: dict[str, Any]) -> None:
    """Assert the output contains a completion percentage."""
    stdout = bdd_context.get("stdout", "")
    stderr = bdd_context.get("stderr", "")
    combined = stdout + stderr

    # Look for any percentage pattern
    import re

    pattern = r"\d+(\.\d+)?%"
    assert re.search(pattern, combined), f"No percentage found in output:\n{combined}"


@then(parsers.parse('I should see "{text}"'))
def should_see_text(bdd_context: dict[str, Any], text: str) -> None:
    """Assert the output contains specific text (alias for output_contains)."""
    stdout = bdd_context.get("stdout", "")
    stderr = bdd_context.get("stderr", "")
    combined = stdout + stderr

    assert text in combined, f"Expected '{text}' in output. Got:\n{combined}"
