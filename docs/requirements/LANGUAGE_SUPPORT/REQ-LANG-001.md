# REQ-LANG-001: Jest/Vitest Plugin for JavaScript/TypeScript

## Status: MISSING
## Priority: HIGH
## Phase: 14

## Description
System shall provide a Jest and Vitest plugin for JavaScript/TypeScript projects that enables requirement traceability through JSDoc comments and inline markers, with full integration into the RTMX ecosystem.

## Acceptance Criteria
- [ ] npm package `@rtmx/jest` published to npmjs.com
- [ ] npm package `@rtmx/vitest` published to npmjs.com
- [ ] JSDoc marker format: `/** @rtmx REQ-XXX-NNN */` parsed before test functions
- [ ] Inline marker format: `// rtmx:req=REQ-XXX-NNN` parsed within test bodies
- [ ] Extended markers support scope/technique/env: `// rtmx:req=REQ-XXX-NNN,scope=unit,technique=nominal`
- [ ] Tree-sitter TypeScript/JavaScript parser extracts markers from AST
- [ ] Custom Jest reporter outputs RTM-compatible JSON results
- [ ] Custom Vitest reporter outputs RTM-compatible JSON results
- [ ] Test results include requirement coverage mapping
- [ ] `rtmx from-jest <results.json>` imports Jest test results into RTM database
- [ ] `rtmx from-vitest <results.json>` imports Vitest test results into RTM database
- [ ] TypeScript type definitions included for marker helpers

## Technical Notes
- Use `tree-sitter-javascript` and `tree-sitter-typescript` for robust parsing
- JSDoc extraction requires walking comment nodes attached to function declarations
- Inline comments require scanning within function body scope
- Jest reporter hook: `onTestResult` for capturing requirement associations
- Vitest reporter hook: `onFinished` with test metadata
- JSON output format aligned with RTMX marker specification (REQ-LANG-007)
- Consider `describe.each` and `it.each` parameterized test handling
- Support for ESM and CommonJS module formats

## Test Cases
1. `tests/test_lang_jest.py::test_jsdoc_marker_extraction` - Parse JSDoc markers
2. `tests/test_lang_jest.py::test_inline_marker_extraction` - Parse inline comments
3. `tests/test_lang_jest.py::test_extended_marker_attributes` - Parse scope/technique/env
4. `tests/test_lang_jest.py::test_jest_reporter_output` - Validate reporter JSON format
5. `tests/test_lang_jest.py::test_vitest_reporter_output` - Validate Vitest reporter format
6. `tests/test_lang_jest.py::test_parameterized_test_handling` - Test each() variants
7. `tests/test_lang_jest.py::test_typescript_type_definitions` - Validate .d.ts files
8. `tests/test_lang_jest.py::test_from_jest_import` - CLI import command

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec

## Blocks
- None

## Effort
3.0 weeks
