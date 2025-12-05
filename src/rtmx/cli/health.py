"""RTMX health command.

Single health check command for CI/CD pipelines and integration validation.
"""

from __future__ import annotations

import json
import sys
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any

from rtmx.config import RTMXConfig, load_config
from rtmx.formatting import Colors, header


class HealthStatus(str, Enum):
    """Overall health status."""

    HEALTHY = "healthy"
    DEGRADED = "degraded"
    UNHEALTHY = "unhealthy"


class CheckResult(str, Enum):
    """Individual check result."""

    PASS = "pass"
    WARN = "warn"
    FAIL = "fail"
    SKIP = "skip"


@dataclass
class Check:
    """Individual health check result."""

    name: str
    result: CheckResult
    message: str
    blocking: bool = True
    details: dict[str, Any] | None = None

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary."""
        d: dict[str, Any] = {
            "name": self.name,
            "result": self.result.value,
            "message": self.message,
            "blocking": self.blocking,
        }
        if self.details:
            d["details"] = self.details
        return d


@dataclass
class HealthReport:
    """Complete health check report."""

    status: HealthStatus
    checks: list[Check] = field(default_factory=list)
    summary: dict[str, int] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary."""
        return {
            "status": self.status.value,
            "summary": self.summary,
            "checks": [c.to_dict() for c in self.checks],
        }


def check_rtm_exists(config: RTMXConfig) -> Check:
    """Check if RTM database exists and is readable."""
    db_path = Path(config.database)

    if not db_path.exists():
        return Check(
            name="rtm_exists",
            result=CheckResult.FAIL,
            message=f"RTM database not found: {db_path}",
            blocking=True,
        )

    # Try to read file
    try:
        content = db_path.read_text()
        if not content.strip():
            return Check(
                name="rtm_exists",
                result=CheckResult.FAIL,
                message="RTM database is empty",
                blocking=True,
            )
    except Exception as e:
        return Check(
            name="rtm_exists",
            result=CheckResult.FAIL,
            message=f"Cannot read RTM database: {e}",
            blocking=True,
        )

    return Check(
        name="rtm_exists",
        result=CheckResult.PASS,
        message=f"RTM database found: {db_path}",
        blocking=True,
    )


def check_rtm_loads(config: RTMXConfig) -> Check:
    """Check if RTM database loads without errors."""
    from rtmx.models import RTMDatabase, RTMError

    try:
        db = RTMDatabase.load(Path(config.database))
        req_count = len(db)
        return Check(
            name="rtm_loads",
            result=CheckResult.PASS,
            message=f"RTM database loaded: {req_count} requirements",
            blocking=True,
            details={"requirement_count": req_count},
        )
    except RTMError as e:
        return Check(
            name="rtm_loads",
            result=CheckResult.FAIL,
            message=f"RTM database load failed: {e}",
            blocking=True,
        )


def check_schema_valid(config: RTMXConfig) -> Check:
    """Check if RTM schema is valid."""
    from rtmx.models import RTMDatabase, RTMError
    from rtmx.validation import validate_schema

    try:
        db = RTMDatabase.load(Path(config.database))
        errors = validate_schema(db)

        if errors:
            return Check(
                name="schema_valid",
                result=CheckResult.FAIL,
                message=f"Schema validation failed: {len(errors)} errors",
                blocking=True,
                details={"errors": errors[:10]},  # First 10 errors
            )

        return Check(
            name="schema_valid",
            result=CheckResult.PASS,
            message="Schema validation passed",
            blocking=True,
        )
    except RTMError as e:
        return Check(
            name="schema_valid",
            result=CheckResult.FAIL,
            message=f"Schema validation error: {e}",
            blocking=True,
        )


def check_reciprocity(config: RTMXConfig) -> Check:
    """Check dependency reciprocity."""
    from rtmx.models import RTMDatabase, RTMError
    from rtmx.validation import check_reciprocity as validate_reciprocity

    try:
        db = RTMDatabase.load(Path(config.database))
        violations = validate_reciprocity(db)

        if violations:
            return Check(
                name="reciprocity",
                result=CheckResult.WARN,
                message=f"Reciprocity violations: {len(violations)}",
                blocking=False,
                details={"violations": [str(v) for v in violations[:10]]},
            )

        return Check(
            name="reciprocity",
            result=CheckResult.PASS,
            message="Reciprocity check passed",
            blocking=False,
        )
    except RTMError as e:
        return Check(
            name="reciprocity",
            result=CheckResult.FAIL,
            message=f"Reciprocity check error: {e}",
            blocking=False,
        )


def check_cycles(config: RTMXConfig) -> Check:
    """Check for circular dependencies."""
    from rtmx.models import RTMDatabase, RTMError

    try:
        db = RTMDatabase.load(Path(config.database))
        cycles = db.find_cycles()

        if cycles:
            total_in_cycles = sum(len(c) for c in cycles)
            return Check(
                name="cycles",
                result=CheckResult.WARN,
                message=f"Found {len(cycles)} circular dependency groups ({total_in_cycles} requirements)",
                blocking=False,
                details={
                    "cycle_count": len(cycles),
                    "requirements_in_cycles": total_in_cycles,
                },
            )

        return Check(
            name="cycles",
            result=CheckResult.PASS,
            message="No circular dependencies",
            blocking=False,
        )
    except RTMError as e:
        return Check(
            name="cycles",
            result=CheckResult.FAIL,
            message=f"Cycle detection error: {e}",
            blocking=False,
        )


def check_test_markers(config: RTMXConfig) -> Check:
    """Check test marker coverage."""
    from rtmx.cli.from_tests import extract_markers_from_file
    from rtmx.models import RTMDatabase, RTMError

    try:
        db = RTMDatabase.load(Path(config.database))

        # Find test files
        test_paths = [Path("tests"), Path("test")]
        test_dir = None
        for p in test_paths:
            if p.exists() and p.is_dir():
                test_dir = p
                break

        if not test_dir:
            return Check(
                name="test_markers",
                result=CheckResult.SKIP,
                message="No tests directory found",
                blocking=False,
            )

        # Scan test files
        markers_found: dict[str, list[str]] = {}
        test_files = list(test_dir.rglob("test_*.py"))

        for test_file in test_files:
            try:
                file_markers = extract_markers_from_file(test_file)
                for test_req in file_markers:
                    req_id = test_req.req_id
                    if req_id not in markers_found:
                        markers_found[req_id] = []
                    markers_found[req_id].append(f"{test_file}::{test_req.test_function}")
            except Exception:
                pass  # Skip files that can't be parsed

        # Compare with RTM
        rtm_ids = {req.req_id for req in db}
        tested_ids = set(markers_found.keys())

        missing_tests = rtm_ids - tested_ids
        orphan_markers = tested_ids - rtm_ids

        coverage = len(tested_ids & rtm_ids) / len(rtm_ids) * 100 if rtm_ids else 0

        if missing_tests:
            return Check(
                name="test_markers",
                result=CheckResult.WARN,
                message=f"Test coverage: {coverage:.1f}% ({len(missing_tests)} requirements without tests)",
                blocking=False,
                details={
                    "coverage_percent": coverage,
                    "tested_count": len(tested_ids & rtm_ids),
                    "missing_count": len(missing_tests),
                    "orphan_count": len(orphan_markers),
                },
            )

        return Check(
            name="test_markers",
            result=CheckResult.PASS,
            message=f"Test coverage: {coverage:.1f}%",
            blocking=False,
            details={
                "coverage_percent": coverage,
                "tested_count": len(tested_ids),
            },
        )
    except RTMError as e:
        return Check(
            name="test_markers",
            result=CheckResult.FAIL,
            message=f"Test marker check error: {e}",
            blocking=False,
        )


def check_agent_configs() -> Check:
    """Check if agent configs have RTMX sections."""
    agent_files = [
        ("CLAUDE.md", "## RTMX"),
        (".cursorrules", "RTMX"),
        (".github/copilot-instructions.md", "RTMX"),
    ]

    found = []
    missing = []

    for filename, marker in agent_files:
        path = Path(filename)
        if path.exists():
            content = path.read_text()
            if marker in content:
                found.append(filename)
            else:
                missing.append(filename)

    if not found and not missing:
        return Check(
            name="agent_configs",
            result=CheckResult.SKIP,
            message="No agent config files detected",
            blocking=False,
        )

    if missing:
        return Check(
            name="agent_configs",
            result=CheckResult.WARN,
            message=f"Agent configs without RTMX: {', '.join(missing)}",
            blocking=False,
            details={"configured": found, "missing": missing},
        )

    return Check(
        name="agent_configs",
        result=CheckResult.PASS,
        message=f"Agent configs configured: {', '.join(found)}",
        blocking=False,
        details={"configured": found},
    )


def check_config_valid() -> Check:
    """Check if rtmx.yaml is valid."""
    config_paths = [Path("rtmx.yaml"), Path(".rtmx.yaml")]

    for config_path in config_paths:
        if config_path.exists():
            try:
                load_config(config_path)
                return Check(
                    name="config_valid",
                    result=CheckResult.PASS,
                    message=f"Config valid: {config_path}",
                    blocking=True,
                )
            except Exception as e:
                return Check(
                    name="config_valid",
                    result=CheckResult.FAIL,
                    message=f"Config invalid: {e}",
                    blocking=True,
                )

    # No config file - use defaults
    return Check(
        name="config_valid",
        result=CheckResult.PASS,
        message="Using default configuration",
        blocking=True,
    )


def run_health_checks(
    config: RTMXConfig,
    checks_to_run: list[str] | None = None,
) -> HealthReport:
    """Run all health checks.

    Args:
        config: RTMX configuration
        checks_to_run: Optional list of specific checks to run

    Returns:
        HealthReport with all results
    """
    all_checks = [
        ("config_valid", check_config_valid),
        ("rtm_exists", lambda: check_rtm_exists(config)),
        ("rtm_loads", lambda: check_rtm_loads(config)),
        ("schema_valid", lambda: check_schema_valid(config)),
        ("reciprocity", lambda: check_reciprocity(config)),
        ("cycles", lambda: check_cycles(config)),
        ("test_markers", lambda: check_test_markers(config)),
        ("agent_configs", check_agent_configs),
    ]

    results: list[Check] = []

    for check_name, check_fn in all_checks:
        if checks_to_run and check_name not in checks_to_run:
            continue

        try:
            result = check_fn()
            results.append(result)
        except Exception as e:
            results.append(
                Check(
                    name=check_name,
                    result=CheckResult.FAIL,
                    message=f"Check failed with exception: {e}",
                    blocking=True,
                )
            )

    # Determine overall status
    has_blocking_fail = any(c.result == CheckResult.FAIL and c.blocking for c in results)
    has_warn = any(c.result == CheckResult.WARN for c in results)

    if has_blocking_fail:
        status = HealthStatus.UNHEALTHY
    elif has_warn:
        status = HealthStatus.DEGRADED
    else:
        status = HealthStatus.HEALTHY

    # Summary counts
    summary = {
        "total": len(results),
        "passed": sum(1 for c in results if c.result == CheckResult.PASS),
        "warnings": sum(1 for c in results if c.result == CheckResult.WARN),
        "failed": sum(1 for c in results if c.result == CheckResult.FAIL),
        "skipped": sum(1 for c in results if c.result == CheckResult.SKIP),
    }

    return HealthReport(status=status, checks=results, summary=summary)


def run_health(
    format_type: str = "terminal",
    strict: bool = False,
    checks: list[str] | None = None,
    config: RTMXConfig | None = None,
) -> None:
    """Run health check command.

    Args:
        format_type: Output format (terminal, json, ci)
        strict: Treat warnings as errors
        checks: Specific checks to run (None for all)
        config: Optional pre-loaded config
    """
    if config is None:
        config = load_config()

    report = run_health_checks(config, checks)

    # Adjust status for strict mode
    if strict and report.status == HealthStatus.DEGRADED:
        report.status = HealthStatus.UNHEALTHY

    if format_type == "json":
        print(json.dumps(report.to_dict(), indent=2))
    elif format_type == "ci":
        # Minimal output for CI
        print(f"status={report.status.value}")
        print(f"passed={report.summary['passed']}")
        print(f"failed={report.summary['failed']}")
        print(f"warnings={report.summary['warnings']}")
    else:
        # Terminal format
        print(header("RTMX Health Check", "="))
        print()

        for check in report.checks:
            if check.result == CheckResult.PASS:
                icon = f"{Colors.GREEN}PASS{Colors.RESET}"
            elif check.result == CheckResult.WARN:
                icon = f"{Colors.YELLOW}WARN{Colors.RESET}"
            elif check.result == CheckResult.FAIL:
                icon = f"{Colors.RED}FAIL{Colors.RESET}"
            else:
                icon = f"{Colors.DIM}SKIP{Colors.RESET}"

            blocking_marker = (
                " [blocking]" if check.blocking and check.result == CheckResult.FAIL else ""
            )
            print(f"  [{icon}] {check.name}: {check.message}{blocking_marker}")

        print()
        print(f"{'=' * 60}")

        if report.status == HealthStatus.HEALTHY:
            print(f"Status: {Colors.GREEN}HEALTHY{Colors.RESET}")
        elif report.status == HealthStatus.DEGRADED:
            print(f"Status: {Colors.YELLOW}DEGRADED{Colors.RESET} (warnings present)")
        else:
            print(f"Status: {Colors.RED}UNHEALTHY{Colors.RESET} (blocking errors)")

        print(
            f"Summary: {report.summary['passed']} passed, "
            f"{report.summary['warnings']} warnings, "
            f"{report.summary['failed']} failed, "
            f"{report.summary['skipped']} skipped"
        )

    # Exit code
    if report.status == HealthStatus.UNHEALTHY:
        sys.exit(2)
    elif report.status == HealthStatus.DEGRADED and strict:
        sys.exit(1)
    else:
        sys.exit(0)
