"""RTMX from-tests command.

Scan test files for requirement markers and report coverage.
"""

from __future__ import annotations

import ast
import contextlib
import sys
from dataclasses import dataclass, field
from pathlib import Path

from rtmx.formatting import Colors
from rtmx.parser import find_rtm_database, load_csv


@dataclass
class TestRequirement:
    """A requirement marker found in a test file."""

    req_id: str
    test_file: str
    test_function: str
    line_number: int
    markers: list[str] = field(default_factory=list)


def extract_markers_from_file(file_path: Path) -> list[TestRequirement]:
    """Extract requirement markers from a Python test file.

    Args:
        file_path: Path to the test file

    Returns:
        List of TestRequirement objects found in the file
    """
    results: list[TestRequirement] = []

    try:
        source = file_path.read_text()
        tree = ast.parse(source, filename=str(file_path))
    except (SyntaxError, UnicodeDecodeError):
        return results

    for node in ast.walk(tree):
        # Check functions and async functions
        if isinstance(node, ast.FunctionDef | ast.AsyncFunctionDef):
            req_markers = _extract_req_markers(node)
            other_markers = _extract_other_markers(node)

            for req_id in req_markers:
                results.append(
                    TestRequirement(
                        req_id=req_id,
                        test_file=str(file_path),
                        test_function=node.name,
                        line_number=node.lineno,
                        markers=other_markers,
                    )
                )

        # Check classes for class-level markers
        elif isinstance(node, ast.ClassDef):
            req_markers = _extract_req_markers(node)
            other_markers = _extract_other_markers(node)

            if req_markers:
                # Find all test methods in the class
                for item in node.body:
                    if isinstance(
                        item, ast.FunctionDef | ast.AsyncFunctionDef
                    ) and item.name.startswith("test_"):
                        for req_id in req_markers:
                            results.append(
                                TestRequirement(
                                    req_id=req_id,
                                    test_file=str(file_path),
                                    test_function=f"{node.name}::{item.name}",
                                    line_number=item.lineno,
                                    markers=other_markers,
                                )
                            )

    return results


def _extract_req_markers(node: ast.FunctionDef | ast.AsyncFunctionDef | ast.ClassDef) -> list[str]:
    """Extract requirement IDs from @pytest.mark.req() decorators."""
    req_ids: list[str] = []

    for decorator in node.decorator_list:
        # Handle @pytest.mark.req("REQ-ID")
        if isinstance(decorator, ast.Call) and _is_req_marker(decorator.func):
            for arg in decorator.args:
                if isinstance(arg, ast.Constant) and isinstance(arg.value, str):
                    req_ids.append(arg.value)

    return req_ids


def _extract_other_markers(
    node: ast.FunctionDef | ast.AsyncFunctionDef | ast.ClassDef,
) -> list[str]:
    """Extract other RTM-related markers (scope_, technique_, env_)."""
    markers: list[str] = []
    rtm_prefixes = ("scope_", "technique_", "env_")

    for decorator in node.decorator_list:
        # Handle @pytest.mark.scope_unit style
        if (
            isinstance(decorator, ast.Attribute)
            and isinstance(decorator.value, ast.Attribute)
            and decorator.attr.startswith(rtm_prefixes)
        ):
            markers.append(decorator.attr)

    return markers


def _is_req_marker(node: ast.expr) -> bool:
    """Check if a node represents pytest.mark.req."""
    return (
        isinstance(node, ast.Attribute)
        and node.attr == "req"
        and isinstance(node.value, ast.Attribute)
        and node.value.attr == "mark"
    )


def scan_test_directory(test_dir: Path, pattern: str = "test_*.py") -> list[TestRequirement]:
    """Scan a directory for test files and extract requirement markers.

    Args:
        test_dir: Directory to scan
        pattern: Glob pattern for test files

    Returns:
        List of all TestRequirement objects found
    """
    results: list[TestRequirement] = []

    for test_file in test_dir.rglob(pattern):
        results.extend(extract_markers_from_file(test_file))

    return results


def run_from_tests(
    test_path: str | None = None,
    rtm_csv: str | None = None,
    show_all: bool = False,
    show_missing: bool = False,
    update: bool = False,
) -> None:
    """Run from-tests command.

    Scans test files for @pytest.mark.req() markers and reports coverage.

    Args:
        test_path: Path to test directory or file
        rtm_csv: Path to RTM database CSV
        show_all: Show all markers found
        show_missing: Show requirements not in database
        update: Update RTM database with test information
    """
    # Determine test path
    path = Path(test_path) if test_path else Path.cwd() / "tests"

    if not path.exists():
        print(f"{Colors.RED}Error: Test path does not exist: {path}{Colors.RESET}")
        sys.exit(1)
        return  # Unreachable, but needed for mocked sys.exit in tests

    # Scan for markers
    print(f"Scanning {path} for requirement markers...")
    print()

    markers = extract_markers_from_file(path) if path.is_file() else scan_test_directory(path)

    if not markers:
        print(f"{Colors.YELLOW}No requirement markers found.{Colors.RESET}")
        return

    # Group by requirement
    by_req: dict[str, list[TestRequirement]] = {}
    for m in markers:
        if m.req_id not in by_req:
            by_req[m.req_id] = []
        by_req[m.req_id].append(m)

    print(f"Found {len(markers)} test(s) linked to {len(by_req)} requirement(s)")
    print()

    # Load RTM database if available
    db_reqs: set[str] = set()
    db_path: Path | None = None
    db_path = Path(rtm_csv) if rtm_csv else find_rtm_database()

    if db_path and db_path.exists():
        requirements = load_csv(db_path)
        db_reqs = {req.req_id for req in requirements}
        print(f"RTM database: {db_path} ({len(db_reqs)} requirements)")
    else:
        print(f"{Colors.YELLOW}No RTM database found{Colors.RESET}")

    print()

    # Show markers
    if show_all:
        print(f"{Colors.BOLD}All Requirements with Tests:{Colors.RESET}")
        print("-" * 60)

        for req_id in sorted(by_req.keys()):
            tests = by_req[req_id]
            in_db = "✓" if req_id in db_reqs else "✗"
            status_color = Colors.GREEN if req_id in db_reqs else Colors.YELLOW

            print(
                f"{status_color}{in_db}{Colors.RESET} {Colors.BOLD}{req_id}{Colors.RESET} ({len(tests)} test(s))"
            )

            for t in tests:
                marker_str = f" [{', '.join(t.markers)}]" if t.markers else ""
                print(f"    {t.test_file}::{t.test_function}{marker_str}")

        print()

    # Show requirements not in database
    if show_missing or not show_all:
        not_in_db = {req_id for req_id in by_req if req_id not in db_reqs}

        if not_in_db:
            print(f"{Colors.YELLOW}Requirements in tests but not in RTM database:{Colors.RESET}")
            for req_id in sorted(not_in_db):
                tests = by_req[req_id]
                print(f"  {Colors.BOLD}{req_id}{Colors.RESET} ({len(tests)} test(s))")
            print()

    # Show requirements in database without tests
    if db_reqs and (show_missing or not show_all):
        no_tests = db_reqs - set(by_req.keys())
        if no_tests:
            print(f"{Colors.YELLOW}Requirements in RTM database without tests:{Colors.RESET}")
            for req_id in sorted(no_tests):
                print(f"  {Colors.DIM}{req_id}{Colors.RESET}")
            print()

    # Summary
    print(f"{Colors.BOLD}Summary:{Colors.RESET}")
    tested = len(set(by_req.keys()) & db_reqs)
    print(f"  Requirements with tests: {tested}/{len(db_reqs) if db_reqs else '?'}")
    print(f"  Tests linked to requirements: {len(markers)}")

    # Update database if requested
    if update and db_path and db_path.exists():
        from rtmx import RTMDatabase

        db = RTMDatabase.load(db_path)
        updated = 0

        for req_id, tests in by_req.items():
            if db.exists(req_id) and tests:
                rel_path = Path(tests[0].test_file)
                with contextlib.suppress(ValueError):
                    rel_path = rel_path.relative_to(Path.cwd())

                db.update(
                    req_id,
                    test_module=str(rel_path),
                    test_function=tests[0].test_function,
                )
                updated += 1

        if updated > 0:
            db.save()
            print()
            print(f"{Colors.GREEN}✓ Updated {updated} requirement(s) in RTM database{Colors.RESET}")
