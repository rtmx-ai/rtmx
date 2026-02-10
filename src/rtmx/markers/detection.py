"""Language detection for source files.

Detects programming language from file extension, shebang, or configuration.
"""

from __future__ import annotations

from pathlib import Path

# Extension to language mapping
EXTENSION_MAP: dict[str, str] = {
    # Python
    ".py": "python",
    ".pyw": "python",
    ".pyi": "python",
    # JavaScript/TypeScript
    ".js": "javascript",
    ".mjs": "javascript",
    ".cjs": "javascript",
    ".jsx": "javascript",
    ".ts": "typescript",
    ".tsx": "typescript",
    ".mts": "typescript",
    ".cts": "typescript",
    # Go
    ".go": "go",
    # Rust
    ".rs": "rust",
    # Java/Kotlin
    ".java": "java",
    ".kt": "kotlin",
    ".kts": "kotlin",
    # C#
    ".cs": "csharp",
    # Ruby
    ".rb": "ruby",
    # C/C++
    ".c": "c",
    ".h": "c",
    ".cpp": "cpp",
    ".hpp": "cpp",
    ".cc": "cpp",
    ".cxx": "cpp",
    # Swift
    ".swift": "swift",
    # PHP
    ".php": "php",
    # Shell
    ".sh": "shell",
    ".bash": "shell",
    ".zsh": "shell",
}

# Shebang to language mapping
SHEBANG_MAP: dict[str, str] = {
    "python": "python",
    "python3": "python",
    "python2": "python",
    "node": "javascript",
    "nodejs": "javascript",
    "ruby": "ruby",
    "perl": "perl",
    "bash": "shell",
    "sh": "shell",
    "zsh": "shell",
    "php": "php",
}


def detect_language(
    file_path: Path,
    override_language: str | None = None,
) -> str | None:
    """Detect the programming language of a source file.

    Detection priority:
    1. Explicit override (from config or parameter)
    2. Shebang line
    3. File extension
    4. Content heuristics (not implemented yet)

    Args:
        file_path: Path to the source file.
        override_language: Explicit language override.

    Returns:
        Language name (lowercase) or None if unknown.
    """
    # Priority 1: Explicit override
    if override_language:
        return override_language.lower()

    # Priority 2: Shebang line
    shebang_lang = _detect_from_shebang(file_path)
    if shebang_lang:
        return shebang_lang

    # Priority 3: File extension
    ext_lang = _detect_from_extension(file_path)
    if ext_lang:
        return ext_lang

    # Priority 4: Content heuristics (future)
    return None


def _detect_from_extension(file_path: Path) -> str | None:
    """Detect language from file extension.

    Args:
        file_path: Path to the source file.

    Returns:
        Language name or None.
    """
    suffix = file_path.suffix.lower()
    return EXTENSION_MAP.get(suffix)


def _detect_from_shebang(file_path: Path) -> str | None:
    """Detect language from shebang line.

    Args:
        file_path: Path to the source file.

    Returns:
        Language name or None.
    """
    try:
        with open(file_path, encoding="utf-8", errors="ignore") as f:
            first_line = f.readline().strip()
    except OSError:
        return None

    if not first_line.startswith("#!"):
        return None

    # Parse shebang: #!/usr/bin/env python3 or #!/usr/bin/python
    shebang = first_line[2:].strip()

    # Handle /usr/bin/env interpreter
    if "env " in shebang:
        parts = shebang.split()
        if len(parts) >= 2:
            interpreter = parts[1]
        else:
            return None
    else:
        # Direct path like #!/usr/bin/python
        interpreter = shebang.split("/")[-1].split()[0]

    # Remove version numbers (python3 -> python)
    interpreter_base = interpreter.rstrip("0123456789.")

    return SHEBANG_MAP.get(interpreter_base) or SHEBANG_MAP.get(interpreter)


def get_supported_extensions() -> list[str]:
    """Get list of supported file extensions.

    Returns:
        List of file extensions (with dots).
    """
    return sorted(EXTENSION_MAP.keys())


def get_supported_languages() -> list[str]:
    """Get list of supported languages.

    Returns:
        List of language names.
    """
    return sorted(set(EXTENSION_MAP.values()))


def get_extensions_for_language(language: str) -> list[str]:
    """Get file extensions associated with a language.

    Args:
        language: Language name.

    Returns:
        List of file extensions.
    """
    return [ext for ext, lang in EXTENSION_MAP.items() if lang == language.lower()]
