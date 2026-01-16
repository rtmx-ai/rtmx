# REQ-LANG-007: Language-Agnostic Marker Annotation Specification

## Status: MISSING
## Priority: HIGH
## Phase: 14

## Description
System shall define a language-agnostic specification for requirement markers that can be consistently parsed across all supported languages, enabling unified RTM tracking regardless of the implementation language.

## Acceptance Criteria
- [ ] JSON Schema defines canonical marker format with required fields (req_id, scope, technique, env)
- [ ] Schema validates marker patterns across all supported languages
- [ ] Parser registry architecture allows registration of language-specific parsers
- [ ] Language auto-detection from file extensions and shebang lines
- [ ] Marker extraction API returns normalized `MarkerInfo` objects
- [ ] CLI command `rtmx markers discover <path>` scans codebase for markers in all languages
- [ ] Configuration file (`.rtmx.yaml`) allows custom parser registration
- [ ] Error reporting identifies invalid markers with file location and fix suggestions

## Technical Notes
- JSON Schema v2020-12 for marker format specification
- Tree-sitter for language-agnostic AST parsing where applicable
- Plugin architecture using Python entry points for parser registration
- Marker format must support:
  - Requirement ID: `REQ-[A-Z]+-[0-9]+` pattern
  - Optional scope: `unit`, `integration`, `system`, `acceptance`
  - Optional technique: `nominal`, `parametric`, `monte_carlo`, `stress`, `boundary`
  - Optional environment: `simulation`, `hil`, `anechoic`, `field`
- Language detection priority: explicit config > shebang > extension > content heuristics
- Caching layer for parsed markers to improve performance on large codebases

## Test Cases
1. `tests/test_lang_spec.py::test_marker_schema_validation` - Validate JSON Schema correctness
2. `tests/test_lang_spec.py::test_parser_registry_registration` - Test parser plugin registration
3. `tests/test_lang_spec.py::test_language_autodetection` - Test file type detection
4. `tests/test_lang_spec.py::test_marker_normalization` - Test MarkerInfo object creation
5. `tests/test_lang_spec.py::test_cli_markers_discover` - Test discovery CLI command
6. `tests/test_lang_spec.py::test_invalid_marker_error_reporting` - Test error messages

## Dependencies
- None (foundation requirement for language support)

## Blocks
- REQ-LANG-001: Jest/Vitest JavaScript/TypeScript plugin
- REQ-LANG-002: JUnit Java/Kotlin extension
- REQ-LANG-003: Go testing integration
- REQ-LANG-004: Rust test framework support
- REQ-LANG-005: NUnit/xUnit C# extension
- REQ-LANG-006: RSpec Ruby integration

## Effort
2.0 weeks
