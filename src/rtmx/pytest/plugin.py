"""RTMX pytest plugin.

Provides pytest integration for requirement traceability markers.
"""

from __future__ import annotations

import pytest


def pytest_configure(config: pytest.Config) -> None:
    """Register RTMX markers with pytest."""
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


def pytest_collection_modifyitems(
    session: pytest.Session,
    config: pytest.Config,
    items: list[pytest.Item],
) -> None:
    """Collect requirement markers for reporting."""
    # This hook can be extended to collect requirement coverage data
    pass


def pytest_report_header(config: pytest.Config) -> list[str]:
    """Add RTMX header to pytest output."""
    return ["RTMX requirement traceability markers enabled"]
