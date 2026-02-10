"""Comment-based marker parser.

Generic parser that extracts requirement markers from comments.
Works with any language that uses // or # or /* */ comments.
"""

from __future__ import annotations

import re
from pathlib import Path

from rtmx.markers.models import MarkerInfo, MarkerValidationError
from rtmx.markers.registry import BaseParser


class CommentParser(BaseParser):
    """Generic comment-based marker parser.

    Extracts markers from comments in various formats:
    - // @req REQ-XXX-NNN (C-style single line)
    - # @req REQ-XXX-NNN (shell/Python style)
    - /* @req REQ-XXX-NNN */ (C-style block)
    - /** @req REQ-XXX-NNN */ (JSDoc/Javadoc style)
    - /// @req REQ-XXX-NNN (Rust doc comments)

    Also supports:
    - @scope unit|integration|system|acceptance
    - @technique nominal|parametric|monte_carlo|stress|boundary
    - @env simulation|hil|anechoic|field
    """

    # Pattern for @req in comments
    REQ_PATTERN = re.compile(
        r"(?://|#|/\*+|\*|///)\s*@req\s+([A-Z]+-[A-Z]+-[0-9]+|REQ-[A-Z]+-[0-9]+)",
        re.IGNORECASE,
    )

    # Patterns for scope, technique, and env
    SCOPE_PATTERN = re.compile(r"@scope\s+(unit|integration|system|acceptance)", re.IGNORECASE)
    TECHNIQUE_PATTERN = re.compile(
        r"@technique\s+(nominal|parametric|monte_carlo|stress|boundary)", re.IGNORECASE
    )
    ENV_PATTERN = re.compile(r"@env\s+(simulation|hil|anechoic|field)", re.IGNORECASE)

    def get_language(self) -> str:
        """Get the language this parser handles."""
        return "generic"

    def parse(self, content: str, file_path: Path) -> list[MarkerInfo]:
        """Parse markers from source code comments.

        Args:
            content: Source code content.
            file_path: Path to the source file.

        Returns:
            List of MarkerInfo objects.
        """
        markers: list[MarkerInfo] = []
        lines = content.splitlines()

        # Detect language from file extension for MarkerInfo
        language = self._detect_language_from_path(file_path)

        for i, line in enumerate(lines, start=1):
            req_match = self.REQ_PATTERN.search(line)
            if req_match:
                req_id = req_match.group(1).upper()

                # Ensure proper REQ- prefix
                if not req_id.startswith("REQ-"):
                    req_id = f"REQ-{req_id}"

                # Look for metadata in surrounding context (previous 10 lines)
                context_start = max(0, i - 10)
                context = "\n".join(lines[context_start:i])

                scope_match = self.SCOPE_PATTERN.search(context)
                technique_match = self.TECHNIQUE_PATTERN.search(context)
                env_match = self.ENV_PATTERN.search(context)

                # Also check current and next few lines
                context_after = "\n".join(lines[i - 1 : min(len(lines), i + 3)])
                if not scope_match:
                    scope_match = self.SCOPE_PATTERN.search(context_after)
                if not technique_match:
                    technique_match = self.TECHNIQUE_PATTERN.search(context_after)
                if not env_match:
                    env_match = self.ENV_PATTERN.search(context_after)

                try:
                    marker = MarkerInfo(
                        req_id=req_id,
                        file_path=file_path,
                        line_number=i,
                        language=language,
                        scope=scope_match.group(1).lower() if scope_match else None,
                        technique=(technique_match.group(1).lower() if technique_match else None),
                        env=env_match.group(1).lower() if env_match else None,
                    )
                    markers.append(marker)
                except MarkerValidationError as e:
                    error_marker = MarkerInfo.error_marker(
                        raw_req_id=req_id,
                        file_path=file_path,
                        line_number=i,
                        language=language,
                        error_message=str(e),
                    )
                    markers.append(error_marker)

        return markers

    def _detect_language_from_path(self, file_path: Path) -> str:
        """Detect language from file path.

        Args:
            file_path: Path to the source file.

        Returns:
            Language name.
        """
        ext = file_path.suffix.lower()
        ext_map = {
            ".js": "javascript",
            ".mjs": "javascript",
            ".cjs": "javascript",
            ".jsx": "javascript",
            ".ts": "typescript",
            ".tsx": "typescript",
            ".go": "go",
            ".rs": "rust",
            ".java": "java",
            ".kt": "kotlin",
            ".cs": "csharp",
            ".rb": "ruby",
            ".c": "c",
            ".cpp": "cpp",
            ".h": "c",
            ".hpp": "cpp",
        }
        return ext_map.get(ext, "unknown")
