"""RTMX pytest plugin.

Provides pytest integration for requirement traceability markers.

The plugin:
- Registers requirement markers with pytest
- Collects requirement coverage data during test runs
- Generates coverage reports showing which requirements have passing tests
"""

from __future__ import annotations

from collections import defaultdict
from dataclasses import dataclass, field
from typing import Any

import pytest


@dataclass
class RequirementCoverage:
    """Tracks test coverage for a single requirement."""

    req_id: str
    tests: list[str] = field(default_factory=list)
    passed: int = 0
    failed: int = 0
    skipped: int = 0

    @property
    def total(self) -> int:
        return self.passed + self.failed + self.skipped

    @property
    def status(self) -> str:
        if self.total == 0:
            return "MISSING"
        if self.failed > 0:
            return "FAILING"
        if self.passed > 0:
            return "PASSING"
        return "SKIPPED"


class RTMXPlugin:
    """RTMX pytest plugin instance."""

    def __init__(self) -> None:
        self.coverage: dict[str, RequirementCoverage] = defaultdict(
            lambda: RequirementCoverage("")
        )
        self._current_req_ids: list[str] = []

    def record_test(self, item: pytest.Item, outcome: str) -> None:
        """Record test outcome for requirements."""
        # Get requirement markers
        for marker in item.iter_markers("req"):
            if marker.args:
                req_id = str(marker.args[0])
                if req_id not in self.coverage:
                    self.coverage[req_id] = RequirementCoverage(req_id)

                cov = self.coverage[req_id]
                cov.tests.append(item.nodeid)

                if outcome == "passed":
                    cov.passed += 1
                elif outcome == "failed":
                    cov.failed += 1
                elif outcome == "skipped":
                    cov.skipped += 1

    def get_coverage_report(self) -> dict[str, Any]:
        """Generate coverage report data."""
        return {
            "requirements": {
                req_id: {
                    "tests": cov.tests,
                    "passed": cov.passed,
                    "failed": cov.failed,
                    "skipped": cov.skipped,
                    "status": cov.status,
                }
                for req_id, cov in self.coverage.items()
            },
            "summary": {
                "total_requirements": len(self.coverage),
                "passing": sum(1 for c in self.coverage.values() if c.status == "PASSING"),
                "failing": sum(1 for c in self.coverage.values() if c.status == "FAILING"),
                "skipped": sum(1 for c in self.coverage.values() if c.status == "SKIPPED"),
            },
        }


# Global plugin instance
_plugin: RTMXPlugin | None = None


def get_plugin() -> RTMXPlugin | None:
    """Get the global plugin instance."""
    return _plugin


def pytest_configure(config: pytest.Config) -> None:
    """Register RTMX markers with pytest."""
    global _plugin
    _plugin = RTMXPlugin()

    # Core requirement marker
    config.addinivalue_line(
        "markers",
        "req(id): Link test to RTM requirement ID",
    )

    # Scope markers (REQ-VAL-001)
    config.addinivalue_line(
        "markers",
        "scope_unit: Single component isolation test, <1ms typical",
    )
    config.addinivalue_line(
        "markers",
        "scope_integration: Multiple components interacting, <100ms typical",
    )
    config.addinivalue_line(
        "markers",
        "scope_system: Entire system end-to-end, <1s typical",
    )

    # Technique markers (REQ-VAL-001)
    config.addinivalue_line(
        "markers",
        "technique_nominal: Typical operating parameters, happy path",
    )
    config.addinivalue_line(
        "markers",
        "technique_parametric: Systematic parameter space exploration",
    )
    config.addinivalue_line(
        "markers",
        "technique_monte_carlo: Random scenarios, 1-10s typical",
    )
    config.addinivalue_line(
        "markers",
        "technique_stress: Boundary/edge cases, extreme conditions",
    )

    # Environment markers (REQ-VAL-001)
    config.addinivalue_line(
        "markers",
        "env_simulation: Pure software, synthetic signals",
    )
    config.addinivalue_line(
        "markers",
        "env_hil: Real hardware, controlled signals",
    )
    config.addinivalue_line(
        "markers",
        "env_anechoic: RF characterization",
    )
    config.addinivalue_line(
        "markers",
        "env_static_field: Outdoor, stationary targets",
    )
    config.addinivalue_line(
        "markers",
        "env_dynamic_field: Outdoor, moving targets",
    )


@pytest.hookimpl(tryfirst=True, hookwrapper=True)
def pytest_runtest_makereport(item: pytest.Item, call: pytest.CallInfo) -> Any:  # noqa: ARG001
    """Record test outcomes for requirement coverage."""
    outcome = yield
    report = outcome.get_result()

    # Only record on call phase (not setup/teardown)
    if report.when == "call" and _plugin is not None:
        _plugin.record_test(item, report.outcome)


def pytest_terminal_summary(
    terminalreporter: Any,
    exitstatus: int,  # noqa: ARG001
    config: pytest.Config,  # noqa: ARG001
) -> None:
    """Print requirement coverage summary."""
    if _plugin is None or not _plugin.coverage:
        return

    report = _plugin.get_coverage_report()
    summary = report["summary"]

    if summary["total_requirements"] == 0:
        return

    terminalreporter.write_sep("=", "RTMX Requirement Coverage")
    terminalreporter.write_line(
        f"Requirements with tests: {summary['total_requirements']}"
    )
    terminalreporter.write_line(
        f"  Passing: {summary['passing']}  "
        f"Failing: {summary['failing']}  "
        f"Skipped: {summary['skipped']}"
    )


def pytest_report_header(config: pytest.Config) -> list[str]:  # noqa: ARG001
    """Add RTMX header to pytest output."""
    return ["RTMX requirement traceability markers enabled"]
