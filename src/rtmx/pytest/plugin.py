"""RTMX pytest plugin.

Provides pytest integration for requirement traceability markers.

The plugin:
- Registers requirement markers with pytest
- Collects requirement coverage data during test runs
- Generates coverage reports showing which requirements have passing tests
- Outputs RTMX results JSON for cross-language verification via --rtmx-output
"""

from __future__ import annotations

import json
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import datetime, timezone
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


@dataclass
class TestResultRecord:
    """A single test result in RTMX results JSON format."""

    req_id: str
    test_name: str
    test_file: str
    line: int = 0
    scope: str = ""
    technique: str = ""
    env: str = ""
    passed: bool = True
    duration_ms: float = 0.0
    error: str = ""
    timestamp: str = ""

    def to_dict(self) -> dict[str, Any]:
        """Serialize to RTMX results JSON format."""
        marker: dict[str, Any] = {
            "req_id": self.req_id,
            "test_name": self.test_name,
            "test_file": self.test_file,
        }
        if self.line > 0:
            marker["line"] = self.line
        if self.scope:
            marker["scope"] = self.scope
        if self.technique:
            marker["technique"] = self.technique
        if self.env:
            marker["env"] = self.env

        result: dict[str, Any] = {
            "marker": marker,
            "passed": self.passed,
        }
        if self.duration_ms > 0:
            result["duration_ms"] = round(self.duration_ms, 3)
        if self.error:
            result["error"] = self.error
        if self.timestamp:
            result["timestamp"] = self.timestamp
        return result


class RTMXPlugin:
    """RTMX pytest plugin instance."""

    def __init__(self) -> None:
        self.coverage: dict[str, RequirementCoverage] = defaultdict(lambda: RequirementCoverage(""))
        self._current_req_ids: list[str] = []
        self.results: list[TestResultRecord] = []
        self.config: pytest.Config | None = None

    def record_test(self, item: pytest.Item, outcome: str, report: Any = None) -> None:
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

                # Record detailed result for --rtmx-output
                if outcome != "skipped":
                    try:
                        fspath, lineno, _ = item.location
                        record = TestResultRecord(
                            req_id=req_id,
                            test_name=item.name,
                            test_file=str(fspath),
                            line=lineno + 1 if lineno is not None else 0,
                            scope=self._extract_marker_value(item, "scope_"),
                            technique=self._extract_marker_value(item, "technique_"),
                            env=self._extract_marker_value(item, "env_"),
                            passed=outcome == "passed",
                            duration_ms=(report.duration * 1000) if report else 0.0,
                            error=(report.longreprtext if report and outcome == "failed" else ""),
                            timestamp=datetime.now(tz=timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
                        )
                        self.results.append(record)
                    except (TypeError, AttributeError):
                        pass  # Mock objects in tests may not have location

    @staticmethod
    def _extract_marker_value(item: pytest.Item, prefix: str) -> str:
        """Extract value from a prefixed marker (e.g., scope_unit -> unit)."""
        for marker in item.iter_markers():
            if marker.name.startswith(prefix):
                return marker.name[len(prefix) :]
        return ""

    def write_results(self, path: str) -> None:
        """Write results to RTMX results JSON file."""
        data = [r.to_dict() for r in self.results]
        with open(path, "w") as f:
            json.dump(data, f, indent=2)

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


def pytest_addoption(parser: pytest.Parser) -> None:
    """Register --rtmx-output option."""
    group = parser.getgroup("rtmx", "RTMX requirement traceability")
    group.addoption(
        "--rtmx-output",
        action="store",
        default=None,
        metavar="FILE",
        help="Write RTMX results JSON to FILE for use with `rtmx verify --results`",
    )


def pytest_configure(config: pytest.Config) -> None:
    """Register RTMX markers with pytest."""
    global _plugin
    _plugin = RTMXPlugin()
    _plugin.config = config

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
        _plugin.record_test(item, report.outcome, report)


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
    terminalreporter.write_line(f"Requirements with tests: {summary['total_requirements']}")
    terminalreporter.write_line(
        f"  Passing: {summary['passing']}  "
        f"Failing: {summary['failing']}  "
        f"Skipped: {summary['skipped']}"
    )


def pytest_sessionfinish(session: pytest.Session, exitstatus: int) -> None:  # noqa: ARG001
    """Write RTMX results JSON if --rtmx-output was provided."""
    if _plugin is None or _plugin.config is None:
        return
    output_path = _plugin.config.getoption("rtmx_output", default=None)
    if not output_path:
        return
    _plugin.write_results(output_path)


def pytest_report_header(config: pytest.Config) -> list[str]:  # noqa: ARG001
    """Add RTMX header to pytest output."""
    return ["RTMX requirement traceability markers enabled"]
