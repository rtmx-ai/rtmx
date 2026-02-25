# REQ-DIST-004: PyPI Package (pip)

## Metadata
- **Category**: DIST
- **Subcategory**: PyPI
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043, REQ-LANG-004

## Requirement

RTMX shall be installable via pip, with the package providing the pytest plugin and ensuring the Go binary is available.

## Rationale

The original rtmx was a Python package. Maintaining PyPI distribution ensures:
1. Backward compatibility for existing users
2. pytest plugin integration
3. Familiar installation path for Python developers

## Design

### Installation

```bash
# Install from PyPI
pip install rtmx

# With optional dependencies
pip install rtmx[dev]  # Include test utilities
```

### Package Behavior

```python
# On import, ensure Go binary is available
import rtmx

# If Go binary not found, guide user:
# "RTMX CLI not found. Install via: brew install rtmx-ai/tap/rtmx"
# Or offer to download automatically
```

### Package Structure

```
rtmx/
├── pyproject.toml
├── src/rtmx/
│   ├── __init__.py       # Version, CLI detection
│   ├── pytest/
│   │   ├── __init__.py
│   │   └── plugin.py     # pytest plugin
│   ├── markers/
│   │   └── __init__.py   # Marker utilities
│   └── _binary.py        # Go binary management
└── tests/
```

### Binary Management Options

1. **Bundled wheels** (platform-specific wheels include binary)
   ```
   rtmx-0.1.0-py3-none-manylinux_x86_64.whl  # Includes linux-amd64 binary
   rtmx-0.1.0-py3-none-macosx_arm64.whl      # Includes darwin-arm64 binary
   ```

2. **Post-install download** (pure Python wheel, downloads on first use)
   ```python
   def ensure_binary():
       if not binary_exists():
           download_and_install()
   ```

3. **External dependency** (require separate installation)
   ```python
   def ensure_binary():
       if not binary_exists():
           raise RuntimeError("Install rtmx CLI: brew install rtmx-ai/tap/rtmx")
   ```

### Recommended: Option 1 (Bundled Wheels)

Platform-specific wheels provide the best UX:
- No network required after pip install
- Works in air-gapped environments
- Consistent behavior across platforms

```toml
# pyproject.toml
[tool.cibuildwheel]
build = "cp39-* cp310-* cp311-* cp312-*"
archs = ["x86_64", "arm64"]
```

## Acceptance Criteria

1. `pip install rtmx` installs pytest plugin
2. `python -c "import rtmx; rtmx.version()"` shows version
3. pytest markers work: `@pytest.mark.req("REQ-XXX")`
4. `rtmx` CLI command available after install
5. Works offline (binary bundled in wheel)
6. Supports Python 3.9+

## Test Strategy

- CI matrix: Python 3.9-3.12 × linux/macos/windows × x64/arm64
- pytest plugin integration tests
- Binary invocation tests

## Backward Compatibility

Existing rtmx users should be able to upgrade seamlessly:
```bash
pip install --upgrade rtmx
# All existing @pytest.mark.req() markers continue to work
```

## References

- Original rtmx package
- cibuildwheel for platform-specific wheels
- maturin pattern (Rust binaries in Python wheels)
