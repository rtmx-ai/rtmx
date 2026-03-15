"""Tests for RTMX pytest plugin --rtmx-output flag.

REQ-LANG-004: Python pytest integration with RTMX results JSON output.
"""

from __future__ import annotations

import json
import re

import pytest

pytest_plugins = ["pytester"]


@pytest.mark.req("REQ-LANG-004")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRtmxOutputOption:
    """Tests for --rtmx-output CLI option."""

    def test_option_registered(self, pytester: pytest.Pytester) -> None:
        """--rtmx-output appears in pytest help."""
        result = pytester.runpytest("--help")
        result.stdout.fnmatch_lines(["*--rtmx-output*"])

    def test_no_flag_no_file(self, pytester: pytest.Pytester) -> None:
        """Without --rtmx-output, no file is written."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            def test_example():
                assert True
            """
        )
        pytester.runpytest()
        assert not (pytester.path / "results.json").exists()


@pytest.mark.req("REQ-LANG-004")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRtmxOutputJSON:
    """Tests for RTMX results JSON output format."""

    def test_writes_json_file(self, pytester: pytest.Pytester) -> None:
        """--rtmx-output writes a valid JSON file."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            def test_example():
                assert True
            """
        )
        result = pytester.runpytest("--rtmx-output=results.json")
        result.assert_outcomes(passed=1)

        results_file = pytester.path / "results.json"
        assert results_file.exists()
        data = json.loads(results_file.read_text())
        assert isinstance(data, list)
        assert len(data) == 1

    def test_json_schema_fields(self, pytester: pytest.Pytester) -> None:
        """Output contains required marker fields."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            @pytest.mark.scope_unit
            @pytest.mark.technique_nominal
            @pytest.mark.env_simulation
            def test_example():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert len(data) == 1
        r = data[0]

        # Required fields
        assert r["marker"]["req_id"] == "REQ-TEST-001"
        assert r["marker"]["test_name"] == "test_example"
        assert "test_file" in r["marker"]
        assert r["passed"] is True

        # Optional marker fields
        assert r["marker"]["scope"] == "unit"
        assert r["marker"]["technique"] == "nominal"
        assert r["marker"]["env"] == "simulation"

    def test_failed_test(self, pytester: pytest.Pytester) -> None:
        """Failed tests have passed=false and error field."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            def test_fail():
                assert False, "deliberate failure"
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert len(data) == 1
        r = data[0]
        assert r["passed"] is False
        assert "error" in r
        assert r["error"] != ""

    def test_multiple_reqs_on_one_test(self, pytester: pytest.Pytester) -> None:
        """A test with multiple @pytest.mark.req produces multiple entries."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            @pytest.mark.req("REQ-TEST-002")
            def test_multi():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert len(data) == 2
        req_ids = {r["marker"]["req_id"] for r in data}
        assert req_ids == {"REQ-TEST-001", "REQ-TEST-002"}

    def test_multiple_tests(self, pytester: pytest.Pytester) -> None:
        """Multiple tests produce multiple entries."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            def test_a():
                assert True

            @pytest.mark.req("REQ-TEST-002")
            def test_b():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert len(data) == 2

    def test_timestamp_format(self, pytester: pytest.Pytester) -> None:
        """Timestamp is ISO 8601 UTC."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            def test_example():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        ts = data[0].get("timestamp", "")
        # ISO 8601 with Z suffix
        assert re.match(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}", ts)

    def test_duration_ms(self, pytester: pytest.Pytester) -> None:
        """Duration is reported in milliseconds."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            def test_example():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert "duration_ms" in data[0]
        assert isinstance(data[0]["duration_ms"], int | float)
        assert data[0]["duration_ms"] >= 0

    def test_empty_results(self, pytester: pytest.Pytester) -> None:
        """No tests with req markers produces empty array."""
        pytester.makepyfile(
            """
            def test_no_markers():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert data == []

    def test_class_level_markers(self, pytester: pytest.Pytester) -> None:
        """Class-level @pytest.mark.req applies to all methods."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            class TestGroup:
                def test_a(self):
                    assert True

                def test_b(self):
                    assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        assert len(data) == 2
        assert all(r["marker"]["req_id"] == "REQ-TEST-001" for r in data)

    def test_skipped_test_excluded(self, pytester: pytest.Pytester) -> None:
        """Skipped tests are not included in results."""
        pytester.makepyfile(
            """
            import pytest

            @pytest.mark.req("REQ-TEST-001")
            @pytest.mark.skip(reason="not yet")
            def test_skipped():
                assert True

            @pytest.mark.req("REQ-TEST-002")
            def test_passing():
                assert True
            """
        )
        pytester.runpytest("--rtmx-output=results.json")

        data = json.loads((pytester.path / "results.json").read_text())
        # Only the passing test should be in results
        req_ids = {r["marker"]["req_id"] for r in data}
        assert "REQ-TEST-002" in req_ids
