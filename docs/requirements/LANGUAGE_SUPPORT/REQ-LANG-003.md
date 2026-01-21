# REQ-LANG-003: Go Testing Integration

## Status: MISSING
## Priority: MEDIUM
## Phase: 14

## Description
System shall provide Go testing integration that enables requirement traceability through comment markers, with support for standard `go test` workflow and subtest patterns.

## Acceptance Criteria
- [ ] Go module `github.com/rtmx-ai/rtmx-go` published and importable
- [ ] Comment marker format: `// rtmx:req=REQ-XXX-NNN` parsed before test functions
- [ ] Extended markers support: `// rtmx:req=REQ-XXX-NNN,scope=unit,technique=nominal,env=simulation`
- [ ] `t.Run()` subtest support with inherited and overridden requirement markers
- [ ] Tree-sitter Go parser extracts markers from test file AST
- [ ] `rtmx-go` CLI tool wraps `go test -json` and enriches output with requirement data
- [ ] JSON output compatible with RTMX marker specification
- [ ] Table-driven test support extracts markers from test case definitions
- [ ] `rtmx from-go <results.json>` imports Go test results into RTM database
- [ ] Integration with `testing.T` for runtime marker registration via helper functions

## Technical Notes
- Use `tree-sitter-go` for AST parsing of test files
- Comment markers must be immediately preceding `func Test*` declarations
- Subtest marker inheritance: child inherits parent unless explicitly overridden
- `go test -json` provides structured output; enrich with marker metadata
- Table-driven tests: scan struct literals for `rtmx` field tags or preceding comments
- Consider `testify` suite support for `suite.Suite` based tests
- Helper function `rtmx.Req(t, "REQ-XXX-NNN")` for runtime registration
- Output format: JSON Lines for streaming large test suites

## Test Cases
1. `tests/test_lang_go.py::test_comment_marker_extraction` - Parse comment markers
2. `tests/test_lang_go.py::test_extended_marker_attributes` - Parse scope/technique/env
3. `tests/test_lang_go.py::test_subtest_marker_inheritance` - t.Run() inheritance
4. `tests/test_lang_go.py::test_table_driven_test_support` - Table test extraction
5. `tests/test_lang_go.py::test_go_test_json_enrichment` - JSON output enrichment
6. `tests/test_lang_go.py::test_runtime_marker_registration` - Helper function support
7. `tests/test_lang_go.py::test_testify_suite_support` - Testify compatibility
8. `tests/test_lang_go.py::test_from_go_import` - CLI import command

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec

## Blocks
- None

## Effort
2.5 weeks
