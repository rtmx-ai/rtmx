# REQ-LANG-004: Rust Test Framework Support

## Status: MISSING
## Priority: MEDIUM
## Phase: 14

## Description
System shall provide Rust test framework support through a companion crate that enables requirement traceability via procedural macro attributes, published to crates.io for the Rust ecosystem.

## Acceptance Criteria
- [ ] Crate `rtmx` published to crates.io
- [ ] Procedural macro crate `rtmx-macros` published to crates.io
- [ ] Attribute macro: `#[rtmx::req("REQ-XXX-NNN")]` marks test functions
- [ ] Extended attributes: `#[rtmx::req("REQ-XXX-NNN", scope = "unit", technique = "nominal")]`
- [ ] Attribute works with `#[test]`, `#[tokio::test]`, and `#[rstest]` test macros
- [ ] Compile-time validation of requirement ID format via proc macro
- [ ] Runtime registry captures requirement associations during test execution
- [ ] Custom Cargo test harness option with RTM JSON output
- [ ] Tree-sitter Rust parser extracts markers from AST for static analysis
- [ ] `cargo rtmx` subcommand via cargo plugin for marker extraction
- [ ] `rtmx from-rust <results.json>` imports Rust test results into RTM database
- [ ] Documentation with `rustdoc` examples for all public APIs

## Technical Notes
- Proc macro crate uses `syn` 2.x and `quote` for AST manipulation
- Attribute macro transforms test function to register requirement before execution
- Compile-time regex validation: `^REQ-[A-Z]+-[0-9]+$`
- Runtime registry: thread-local storage or `lazy_static` for test requirement map
- Custom test harness: implement `libtest_mimic` for RTM-aware test runner
- Consider `trybuild` for proc macro compile-fail tests
- Support `#[ignore]` and `#[should_panic]` test attributes
- Integration with `cargo nextest` output format

## Test Cases
1. `tests/test_lang_rust.py::test_req_attribute_extraction` - Parse #[rtmx::req] attributes
2. `tests/test_lang_rust.py::test_extended_attributes` - Parse scope/technique/env
3. `tests/test_lang_rust.py::test_compile_time_validation` - Invalid req ID rejection
4. `tests/test_lang_rust.py::test_tokio_test_compatibility` - Async test support
5. `tests/test_lang_rust.py::test_rstest_compatibility` - rstest parameterized support
6. `tests/test_lang_rust.py::test_runtime_registry_capture` - Registry functionality
7. `tests/test_lang_rust.py::test_cargo_subcommand` - cargo rtmx integration
8. `tests/test_lang_rust.py::test_custom_harness_output` - JSON output format
9. `tests/test_lang_rust.py::test_from_rust_import` - CLI import command

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec

## Blocks
- None

## Effort
3.0 weeks
