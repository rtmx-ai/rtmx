# REQ-DIST-004: PyPI Package (pip)

## Metadata
- **Category**: DIST
- **Subcategory**: PyPI
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043, REQ-LANG-004
- **Effort**: 2 weeks

## Requirement

RTMX shall be installable via pip, with the package providing the pytest plugin with cross-language verification output (`--rtmx-output`).

## Rationale

The original rtmx is a Python package on PyPI (v0.0.6). We need to ship the new `--rtmx-output` feature that enables cross-language verification with the Go CLI. This maintains backward compatibility for existing users while enabling the new distributed verification workflow.

## Current State

- rtmx v0.0.6 is on PyPI (published)
- rtmx v0.0.7 is local with `--rtmx-output` feature (not yet released)
- Release workflow exists at `.github/workflows/release.yml` (tag-triggered)
- Python CLI commands work independently (no Go binary required)
- Go binary is available separately via `brew install` or GitHub releases

## Phased Approach

### Phase A: Ship --rtmx-output (this requirement)
- Bump version to 0.0.7
- Update CHANGELOG.md
- Tag and release to PyPI
- Verify `pip install rtmx` works and pytest plugin loads

### Phase B: Bundle Go binary in wheels (future requirement)
- Platform-specific wheels with Go binary
- Zero-dependency installation of full RTMX ecosystem
- Requires cibuildwheel CI configuration

## Acceptance Criteria

1. `pip install rtmx` installs version >= 0.0.7 from PyPI
2. `python -c "import rtmx"` succeeds
3. `@pytest.mark.req("REQ-XXX")` markers work on test functions
4. `pytest --rtmx-output=results.json` produces valid RTMX results JSON
5. Go CLI's `rtmx verify --results results.json` successfully consumes the output
6. `rtmx` CLI command available after install (Python implementation)
7. Backward compatible: existing v0.0.6 users upgrade without breaking changes
8. Supports Python 3.9+

## Files to Modify

In the `rtmx` (Python) repository:
- `pyproject.toml` - Bump version to 0.0.7
- `CHANGELOG.md` - Add v0.0.7 release notes
- `src/rtmx/__init__.py` - Verify version matches

## Dependencies

| Requirement | Status | Relationship |
|-------------|--------|-------------|
| REQ-GO-043 | COMPLETE | GoReleaser produces Go binary (separate install) |
| REQ-LANG-004 | COMPLETE | pytest plugin with --rtmx-output implemented |

## Test Strategy

- Pre-release: `pip install -e .` and verify all acceptance criteria locally
- Post-release: `pip install rtmx==0.0.7` in clean virtualenv and verify
- CI: Existing tests pass in release workflow

## Release Process

```bash
# 1. Verify all tests pass
make test

# 2. Update version and changelog
# pyproject.toml: version = "0.0.7"
# CHANGELOG.md: Add release notes

# 3. Commit and tag
git commit -m "chore: Bump version to v0.0.7"
git tag v0.0.7
git push origin main --tags

# 4. Release workflow publishes to PyPI automatically

# 5. Verify
pip install rtmx==0.0.7
```

## References

- PyPI: https://pypi.org/project/rtmx/
- Release workflow: `.github/workflows/release.yml`
- REQ-LANG-004 implementation: `src/rtmx/pytest/plugin.py`
