"""Python marker parser.

Parses pytest markers from Python test files.
"""

from __future__ import annotations

import ast
import re
from pathlib import Path

from rtmx.markers.models import MarkerInfo, MarkerValidationError
from rtmx.markers.registry import BaseParser


class PythonParser(BaseParser):
    """Parser for Python pytest markers.

    Extracts requirement markers from Python test files that use pytest markers:
    - @pytest.mark.req("REQ-XXX-NNN")
    - @pytest.mark.scope_unit, scope_integration, etc.
    - @pytest.mark.technique_nominal, technique_parametric, etc.
    - @pytest.mark.env_simulation, env_hil, etc.
    """

    # Pattern for req marker in decorator
    REQ_PATTERN = re.compile(r'@pytest\.mark\.req\s*\(\s*["\']([^"\']+)["\']\s*\)', re.MULTILINE)

    # Patterns for scope, technique, and env markers
    SCOPE_PATTERN = re.compile(r"@pytest\.mark\.scope_(unit|integration|system|acceptance)")
    TECHNIQUE_PATTERN = re.compile(
        r"@pytest\.mark\.technique_(nominal|parametric|monte_carlo|stress|boundary)"
    )
    ENV_PATTERN = re.compile(r"@pytest\.mark\.env_(simulation|hil|anechoic|field)")

    def get_language(self) -> str:
        """Get the language this parser handles."""
        return "python"

    def parse(self, content: str, file_path: Path) -> list[MarkerInfo]:
        """Parse markers from Python source code.

        Uses a combination of AST parsing and regex for robustness.

        Args:
            content: Python source code.
            file_path: Path to the source file.

        Returns:
            List of MarkerInfo objects.
        """
        markers: list[MarkerInfo] = []

        # Try AST-based parsing first for accuracy
        try:
            tree = ast.parse(content)
            markers = self._parse_with_ast(tree, content, file_path)
        except SyntaxError:
            # Fall back to regex-based parsing for invalid Python
            markers = self._parse_with_regex(content, file_path)

        return markers

    def _parse_with_ast(self, tree: ast.Module, content: str, file_path: Path) -> list[MarkerInfo]:
        """Parse markers using AST.

        Args:
            tree: Parsed AST.
            content: Original source code (for line lookup).
            file_path: Path to the source file.

        Returns:
            List of MarkerInfo objects.
        """
        markers: list[MarkerInfo] = []
        lines = content.splitlines()

        for node in ast.walk(tree):
            if isinstance(node, ast.FunctionDef | ast.AsyncFunctionDef):
                # Check decorators for markers
                function_markers = self._extract_markers_from_decorators(node, lines, file_path)
                markers.extend(function_markers)

        return markers

    def _extract_markers_from_decorators(
        self,
        node: ast.FunctionDef | ast.AsyncFunctionDef,
        lines: list[str],
        file_path: Path,
    ) -> list[MarkerInfo]:
        """Extract markers from function decorators.

        Args:
            node: Function definition AST node.
            lines: Source code lines.
            file_path: Path to the source file.

        Returns:
            List of MarkerInfo objects.
        """
        markers: list[MarkerInfo] = []

        # Collect req IDs and metadata from decorators
        req_ids: list[tuple[str, int]] = []  # (req_id, line_number)
        scope: str | None = None
        technique: str | None = None
        env: str | None = None

        for decorator in node.decorator_list:
            decorator_line = decorator.lineno

            # Get the source text for this decorator
            if decorator_line <= len(lines):
                # Look at lines around the decorator for full marker
                start_line = max(0, decorator_line - 1)
                end_line = min(len(lines), decorator_line + 1)
                decorator_text = "\n".join(lines[start_line:end_line])

                # Check for @pytest.mark.req
                req_match = self.REQ_PATTERN.search(decorator_text)
                if req_match:
                    req_ids.append((req_match.group(1), decorator_line))

                # Check for scope marker
                scope_match = self.SCOPE_PATTERN.search(decorator_text)
                if scope_match:
                    scope = scope_match.group(1)

                # Check for technique marker
                technique_match = self.TECHNIQUE_PATTERN.search(decorator_text)
                if technique_match:
                    technique = technique_match.group(1)

                # Check for env marker
                env_match = self.ENV_PATTERN.search(decorator_text)
                if env_match:
                    env = env_match.group(1)

        # Create MarkerInfo for each req ID found
        for req_id, line_number in req_ids:
            try:
                marker = MarkerInfo(
                    req_id=req_id,
                    file_path=file_path,
                    line_number=line_number,
                    language="python",
                    scope=scope,
                    technique=technique,
                    env=env,
                    function_name=node.name,
                )
                markers.append(marker)
            except MarkerValidationError as e:
                # Create error marker for invalid req_id
                error_marker = MarkerInfo.error_marker(
                    raw_req_id=req_id,
                    file_path=file_path,
                    line_number=line_number,
                    language="python",
                    error_message=str(e),
                )
                markers.append(error_marker)

        return markers

    def _parse_with_regex(self, content: str, file_path: Path) -> list[MarkerInfo]:
        """Parse markers using regex (fallback for invalid Python).

        Args:
            content: Python source code.
            file_path: Path to the source file.

        Returns:
            List of MarkerInfo objects.
        """
        markers: list[MarkerInfo] = []
        lines = content.splitlines()

        for i, line in enumerate(lines, start=1):
            req_match = self.REQ_PATTERN.search(line)
            if req_match:
                req_id = req_match.group(1)

                # Look for scope/technique/env in surrounding lines
                context_start = max(0, i - 5)
                context_end = min(len(lines), i + 5)
                context = "\n".join(lines[context_start:context_end])

                scope_match = self.SCOPE_PATTERN.search(context)
                technique_match = self.TECHNIQUE_PATTERN.search(context)
                env_match = self.ENV_PATTERN.search(context)

                try:
                    marker = MarkerInfo(
                        req_id=req_id,
                        file_path=file_path,
                        line_number=i,
                        language="python",
                        scope=scope_match.group(1) if scope_match else None,
                        technique=technique_match.group(1) if technique_match else None,
                        env=env_match.group(1) if env_match else None,
                    )
                    markers.append(marker)
                except MarkerValidationError as e:
                    error_marker = MarkerInfo.error_marker(
                        raw_req_id=req_id,
                        file_path=file_path,
                        line_number=i,
                        language="python",
                        error_message=str(e),
                    )
                    markers.append(error_marker)

        return markers
