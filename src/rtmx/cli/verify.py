"""Closed-loop requirement verification.

Runs tests and automatically updates RTM status based on results.

This is the core of RTMX's value proposition: requirements are verified
by tests, and status is automatically updated when tests pass in CI.
"""

from __future__ import annotations

import contextlib
import json
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

import click

from rtmx.config import load_config
from rtmx.models import RTMDatabase, Status


@dataclass
class VerificationResult:
    """Result of verifying a single requirement."""

    req_id: str
    tests_total: int
    tests_passed: int
    tests_failed: int
    tests_skipped: int
    previous_status: Status
    new_status: Status
    updated: bool

    @property
    def all_passed(self) -> bool:
        """Check if all tests passed."""
        return self.tests_total > 0 and self.tests_failed == 0 and self.tests_passed > 0


def run_tests_with_coverage(test_path: str | None = None) -> dict:
    """Run pytest and collect requirement coverage data.

    Returns:
        Dictionary with requirement coverage from RTMX plugin
    """
    # Build pytest command
    cmd = [
        sys.executable,
        "-m",
        "pytest",
        "--tb=no",  # No tracebacks for cleaner output
        "-q",  # Quiet mode
    ]

    if test_path:
        cmd.append(test_path)

    # Run pytest and capture output
    # The RTMX plugin prints coverage to terminal, but we need structured data
    # So we'll run with --collect-only first to get test->requirement mapping,
    # then run actual tests and parse results

    # For now, use a simpler approach: run pytest with JSON output plugin
    # or parse the terminal output from RTMX plugin

    # Run tests and get exit code
    result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        cwd=Path.cwd(),
    )

    # Parse RTMX coverage from output
    coverage = parse_rtmx_coverage(result.stdout + result.stderr)
    coverage["exit_code"] = result.returncode

    return coverage


def parse_rtmx_coverage(output: str) -> dict:
    """Parse RTMX requirement coverage from pytest output.

    The RTMX plugin prints:
        RTMX Requirement Coverage
        Requirements with tests: N
          Passing: X  Failing: Y  Skipped: Z

    We need more detail, so we'll enhance the plugin output.
    For now, use from-tests to get the mapping and pytest results.
    """
    # This is a simplified parser - we'll enhance the plugin for better data
    coverage: dict = {"requirements": {}, "summary": {}}

    lines = output.split("\n")
    for i, line in enumerate(lines):
        if "RTMX Requirement Coverage" in line:
            # Found the summary section
            for j in range(i + 1, min(i + 5, len(lines))):
                if "Passing:" in lines[j]:
                    parts = lines[j].split()
                    for k, part in enumerate(parts):
                        if part == "Passing:":
                            coverage["summary"]["passing"] = int(parts[k + 1])
                        elif part == "Failing:":
                            coverage["summary"]["failing"] = int(parts[k + 1])
                        elif part == "Skipped:":
                            coverage["summary"]["skipped"] = int(parts[k + 1])
            break

    return coverage


def get_requirement_test_results(test_path: str | None = None) -> dict[str, dict]:
    """Get detailed test results per requirement.

    Uses pytest with JSON output to get structured results.
    """
    import tempfile

    # Create temp file for JSON report
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as f:
        json_path = f.name

    try:
        # Run pytest with JSON report
        cmd = [
            sys.executable,
            "-m",
            "pytest",
            f"--json-report-file={json_path}",
            "--json-report",
            "-q",
        ]

        if test_path:
            cmd.append(test_path)

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            cwd=Path.cwd(),
        )

        # Check if pytest-json-report is available
        if "unrecognized arguments: --json-report" in result.stderr:
            # Fallback: use marker-based collection
            return get_requirement_test_results_fallback(test_path)

        # Parse JSON report
        json_path_obj = Path(json_path)
        if json_path_obj.exists():
            with open(json_path_obj) as f:
                report = json.load(f)
            return parse_json_report(report)

    except Exception:
        pass
    finally:
        Path(json_path).unlink(missing_ok=True)

    return get_requirement_test_results_fallback(test_path)


def get_requirement_test_results_fallback(test_path: str | None = None) -> dict[str, dict]:
    """Fallback method using pytest collection and execution."""
    results: dict[str, dict] = {}

    # First, collect tests and their requirement markers
    cmd = [
        sys.executable,
        "-m",
        "pytest",
        "--collect-only",
        "-q",
    ]
    if test_path:
        cmd.append(test_path)

    subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        cwd=Path.cwd(),
    )

    # Now run tests and get results per requirement using the plugin
    cmd = [
        sys.executable,
        "-m",
        "pytest",
        "-vv",  # Extra verbose to see individual test results
        "--tb=no",
        "--no-cov",  # Disable coverage to keep output clean
    ]
    if test_path:
        cmd.append(test_path)

    test_result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        cwd=Path.cwd(),
    )

    # Parse verbose output to map tests to pass/fail
    test_outcomes = parse_verbose_pytest_output(test_result.stdout)

    # Use from-tests to get requirement->test mapping
    from rtmx.cli.from_tests import extract_markers_from_file, scan_test_directory

    test_path_obj = Path(test_path) if test_path else Path("tests")
    if test_path_obj.is_file():
        # For a single file, extract markers from just that file
        markers = extract_markers_from_file(test_path_obj)
    else:
        # For a directory, scan all test files
        markers = scan_test_directory(test_path_obj)

    # Group markers by requirement ID
    markers_by_req: dict[str, list] = {}
    for m in markers:
        if m.req_id not in markers_by_req:
            markers_by_req[m.req_id] = []
        markers_by_req[m.req_id].append(m)

    # Build results per requirement
    cwd = Path.cwd()
    for req_id, test_list in markers_by_req.items():
        results[req_id] = {
            "total": 0,
            "passed": 0,
            "failed": 0,
            "skipped": 0,
            "tests": [],
        }
        for test_info in test_list:
            # Convert absolute path to relative for nodeid matching
            test_file = Path(test_info.test_file)
            with contextlib.suppress(ValueError):
                test_file = test_file.relative_to(cwd)
            test_nodeid = f"{test_file}::{test_info.test_function}"
            outcome = test_outcomes.get(test_nodeid, "unknown")

            results[req_id]["total"] += 1
            results[req_id]["tests"].append({"nodeid": test_nodeid, "outcome": outcome})

            if outcome == "passed":
                results[req_id]["passed"] += 1
            elif outcome == "failed":
                results[req_id]["failed"] += 1
            elif outcome == "skipped":
                results[req_id]["skipped"] += 1

    return results


def parse_verbose_pytest_output(output: str) -> dict[str, str]:
    """Parse pytest verbose output to get test outcomes."""
    outcomes: dict[str, str] = {}

    for line in output.split("\n"):
        line = line.strip()
        if " PASSED" in line or " FAILED" in line or " SKIPPED" in line:
            # Format: tests/test_foo.py::test_bar PASSED [  2%]
            # Remove the percentage indicator if present
            if "[" in line:
                line = line.split("[")[0].strip()
            # Now split to get nodeid and outcome
            parts = line.rsplit(" ", 1)
            if len(parts) == 2:
                nodeid = parts[0].strip()
                outcome = parts[1].strip().lower()
                outcomes[nodeid] = outcome

    return outcomes


def parse_json_report(report: dict) -> dict[str, dict]:
    """Parse pytest-json-report output."""
    results: dict[str, dict] = {}

    for test in report.get("tests", []):
        nodeid = test.get("nodeid", "")
        outcome = test.get("outcome", "unknown")
        markers = test.get("markers", [])

        # Find req marker
        for marker in markers:
            if isinstance(marker, dict) and marker.get("name") == "req":
                req_id = marker.get("args", [None])[0]
                if req_id:
                    if req_id not in results:
                        results[req_id] = {
                            "total": 0,
                            "passed": 0,
                            "failed": 0,
                            "skipped": 0,
                            "tests": [],
                        }

                    results[req_id]["total"] += 1
                    results[req_id]["tests"].append({"nodeid": nodeid, "outcome": outcome})

                    if outcome == "passed":
                        results[req_id]["passed"] += 1
                    elif outcome == "failed":
                        results[req_id]["failed"] += 1
                    elif outcome == "skipped":
                        results[req_id]["skipped"] += 1

    return results


def determine_new_status(
    req_results: dict,
    current_status: Status,
    _require_all_pass: bool = True,
) -> Status:
    """Determine new status based on test results.

    Args:
        req_results: Test results for the requirement
        current_status: Current RTM status
        require_all_pass: If True, all non-skipped tests must pass for COMPLETE

    Returns:
        New status to set

    Status update rules:
        - All tests pass (none fail, some may skip) → COMPLETE
        - Any test fails → Downgrade COMPLETE to PARTIAL, or keep current
        - No tests pass, none fail → Keep current status
        - Skipped tests don't affect status (they're intentional exclusions)
    """
    total = req_results.get("total", 0)
    passed = req_results.get("passed", 0)
    failed = req_results.get("failed", 0)
    # Note: skipped tests don't affect status determination

    if total == 0:
        # No tests - keep current status
        return current_status

    if failed > 0:
        # Any failures - regression
        if current_status == Status.COMPLETE:
            # Downgrade to PARTIAL
            return Status.PARTIAL
        return current_status

    # No failures - check if any tests actually ran and passed
    if passed > 0:
        # Tests ran and passed (skipped tests don't affect this)
        return Status.COMPLETE

    # All tests skipped or unknown - keep current status
    return current_status


def run_verify(
    test_path: str | None = None,
    update: bool = False,
    dry_run: bool = False,
    verbose: bool = False,
) -> list[VerificationResult]:
    """Run verification and optionally update RTM.

    Args:
        test_path: Path to tests (default: tests/)
        update: Whether to update RTM database
        dry_run: Show what would change without updating
        verbose: Show detailed output

    Returns:
        List of verification results
    """
    config = load_config()
    db = RTMDatabase.load(config.database)

    click.echo("Running tests and collecting requirement coverage...")

    # Get test results per requirement
    req_results = get_requirement_test_results(test_path)

    results: list[VerificationResult] = []
    updates_needed: list[tuple[str, Status]] = []

    for req_id, test_data in sorted(req_results.items()):
        # Check if requirement exists in RTM
        if not db.exists(req_id):
            if verbose:
                click.echo(f"  {req_id}: Not in RTM database (skipping)")
            continue

        req = db.get(req_id)
        current_status = req.status
        new_status = determine_new_status(test_data, current_status)

        result = VerificationResult(
            req_id=req_id,
            tests_total=test_data["total"],
            tests_passed=test_data["passed"],
            tests_failed=test_data["failed"],
            tests_skipped=test_data["skipped"],
            previous_status=current_status,
            new_status=new_status,
            updated=new_status != current_status,
        )
        results.append(result)

        if result.updated:
            updates_needed.append((req_id, new_status))

    # Print results
    click.echo()
    click.echo("Verification Results:")
    click.echo("-" * 60)

    passing = [r for r in results if r.all_passed]
    failing = [r for r in results if r.tests_failed > 0]
    to_update = [r for r in results if r.updated]

    if passing:
        click.echo(click.style(f"  PASSING: {len(passing)} requirements", fg="green"))
    if failing:
        click.echo(click.style(f"  FAILING: {len(failing)} requirements", fg="red"))

    if to_update:
        click.echo()
        click.echo("Status changes:")
        for r in to_update:
            status_change = f"{r.previous_status.value} → {r.new_status.value}"
            if r.new_status == Status.COMPLETE:
                click.echo(click.style(f"  {r.req_id}: {status_change}", fg="green"))
            elif r.new_status == Status.PARTIAL:
                click.echo(click.style(f"  {r.req_id}: {status_change}", fg="yellow"))
            else:
                click.echo(f"  {r.req_id}: {status_change}")

    # Update RTM if requested
    if update and updates_needed and not dry_run:
        click.echo()
        click.echo("Updating RTM database...")
        for req_id, new_status in updates_needed:
            db.update(req_id, status=new_status)
        db.save()
        click.echo(click.style(f"  Updated {len(updates_needed)} requirement(s)", fg="green"))
    elif dry_run and updates_needed:
        click.echo()
        click.echo(click.style("Dry run - no changes made", fg="yellow"))
    elif update and not updates_needed:
        click.echo()
        click.echo("No status changes needed")

    return results


@click.command("verify")
@click.argument("test_path", required=False)
@click.option("--update", is_flag=True, help="Update RTM database with results")
@click.option("--dry-run", is_flag=True, help="Show changes without updating")
@click.option("-v", "--verbose", is_flag=True, help="Verbose output")
def verify_cmd(
    test_path: str | None,
    update: bool,
    dry_run: bool,
    verbose: bool,
) -> None:
    """Verify requirements by running tests and updating status.

    This is closed-loop verification: tests are run, and RTM status
    is automatically updated based on pass/fail results.

    \b
    Examples:
        rtmx verify                    # Run all tests, show results
        rtmx verify --update           # Run tests and update RTM
        rtmx verify tests/unit/ --update  # Verify specific tests
        rtmx verify --dry-run          # Show what would change

    \b
    Status update rules:
        - All tests pass → COMPLETE
        - Some tests pass, none fail → PARTIAL
        - Any test fails → Keep current (or downgrade COMPLETE to PARTIAL)
        - No tests → Keep current status
    """
    results = run_verify(
        test_path=test_path,
        update=update,
        dry_run=dry_run,
        verbose=verbose,
    )

    # Exit with error if any tests failed
    if any(r.tests_failed > 0 for r in results):
        raise SystemExit(1)
