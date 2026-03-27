# REQ-LANG-005: Rust Testing Integration

## Metadata
- **Category**: LANG
- **Subcategory**: Rust
- **Priority**: MEDIUM
- **Phase**: 18
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007

## Requirement

RTMX shall provide Rust testing integration via procedural macros that enable requirement markers on test functions.

## Rationale

Rust is increasingly used for systems programming, embedded systems, and performance-critical applications where requirements traceability is essential.

## Design

### Installation

```toml
# Cargo.toml
[dev-dependencies]
rtmx = "0.1"
```

### Marker Syntax

```rust
use rtmx::req;

#[test]
#[req("REQ-AUTH-001")]
fn test_login_success() {
    // test implementation
}

#[test]
#[req("REQ-AUTH-002", scope = "integration", technique = "boundary")]
fn test_login_invalid_password() {
    // test implementation
}

// Multiple requirements
#[test]
#[req("REQ-AUTH-001")]
#[req("REQ-AUDIT-001")]
fn test_login_audited() {
    // test implementation
}
```

### Alternative: Macro-based

```rust
use rtmx::test_req;

test_req!("REQ-AUTH-001", test_login_success, {
    // test implementation
});
```

### Output Integration

```bash
# Run with RTMX output
cargo test 2>&1 | rtmx from-tests --lang rust

# Or configure in .cargo/config.toml
[env]
RTMX_OUTPUT = "rtmx-results.json"
```

## Implementation

This requirement has two parts:

### Part 1: CLI-Side Rust Marker Scanning (Implemented)

The RTMX Go CLI scans Rust test files for requirement markers via `extractRustMarkersFromFile()` in `internal/cmd/from_tests.go`. Three marker styles are recognized:

1. **Attribute macro**: `#[req("REQ-ID")]` -- the primary style, used with the `rtmx` crate
2. **Comment marker**: `// rtmx:req REQ-ID` -- works without any crate dependency
3. **Function call**: `rtmx::req("REQ-ID")` -- runtime marker inside test body

File matching rules in `scanTestDirectory()`:
- `*_test.rs` files anywhere in the tree (unit test convention)
- Any `.rs` file inside a `tests/` directory (Rust integration test convention)

The scanner tracks `mod` blocks to produce qualified function names (e.g., `tests::test_inside_mod`), and handles `async fn` and `pub fn` test functions.

Tests: `internal/cmd/from_tests_rust_test.go`

### Part 2: Rust Proc-Macro Crate (Specification Only)

The `rtmx` crate will be published to crates.io from a separate repository (`rtmx-ai/rtmx-rs`). The crate provides:

- `#[rtmx::req("REQ-ID")]` attribute macro via a proc-macro sub-crate
- Optional metadata: `#[req("REQ-ID", scope = "integration", technique = "boundary")]`
- JSON results output compatible with `rtmx from-go` import format
- Integration with `cargo test` via either:
  - A custom test harness that wraps libtest and emits RTMX JSON after the run
  - A post-processor: `cargo test 2>&1 | rtmx from-tests --lang rust`

#### Crate Architecture

```
rtmx/                        # crates.io workspace
├── Cargo.toml               # Workspace root
├── rtmx/                    # Main crate (re-exports macro, provides runtime)
│   ├── Cargo.toml
│   └── src/
│       ├── lib.rs           # Re-export #[req], provide req() function
│       └── output.rs        # JSON results writer (RTMX_OUTPUT env var)
├── rtmx-macros/             # Proc macro crate
│   ├── Cargo.toml
│   └── src/
│       └── lib.rs           # #[req] attribute macro implementation
└── examples/
    └── basic.rs
```

#### Proc Macro Design

The `#[req("REQ-ID")]` attribute macro:
1. Registers the requirement ID and metadata at compile time
2. Injects a call to `rtmx::__register_marker()` at the start of the test function
3. Does NOT modify test behavior -- the original `#[test]` attribute drives execution
4. At process exit (via `atexit` handler or custom harness), writes collected markers + results to `RTMX_OUTPUT` path if set

#### Runtime `req()` Function

For cases where the attribute macro cannot be used (e.g., dynamic test generation):

```rust
#[test]
fn test_dynamic() {
    rtmx::req("REQ-DYN-001");
    // test body
}
```

This registers the marker at runtime. The CLI scanner also detects this pattern statically.

## Acceptance Criteria

1. `rtmx` crate available on crates.io (Part 2 -- separate repo)
2. `#[req("REQ-XXX")]` attribute macro works on test functions (Part 2 -- separate repo)
3. Test results output compatible JSON format (Part 2 -- separate repo)
4. `rtmx verify --command "cargo test"` correctly updates status (Part 2 -- separate repo)
5. Works with standard `#[test]` attribute (Part 2 -- separate repo)
6. CLI `from-tests` and `markers` commands discover Rust markers (Part 1 -- implemented)
7. CLI scans `*_test.rs` and `tests/*.rs` files (Part 1 -- implemented)
8. CLI recognizes all three marker styles (Part 1 -- implemented)

## Test Strategy

- Table-driven unit tests for `extractRustMarkersFromFile()` (Part 1 -- implemented)
- Directory scanning tests for Rust file matching (Part 1 -- implemented)
- Unit tests for macro expansion (Part 2 -- separate repo)
- Integration tests with cargo test (Part 2 -- separate repo)
- Cross-platform testing (Linux, macOS, Windows)

## Package Structure

```
rtmx/                    # crates.io package (separate repo: rtmx-ai/rtmx-rs)
├── Cargo.toml
├── rtmx/
│   └── src/
│       ├── lib.rs           # Main library
│       └── output.rs        # JSON output
├── rtmx-macros/             # Proc macro crate
│   ├── Cargo.toml
│   └── src/lib.rs           # #[req] implementation
└── examples/
    └── basic.rs
```

## References

- Rust procedural macros: https://doc.rust-lang.org/reference/procedural-macros.html
- cargo test output format
- REQ-LANG-007 marker specification
- CLI implementation: `internal/cmd/from_tests.go` (extractRustMarkersFromFile, isRustTestFile)
- CLI tests: `internal/cmd/from_tests_rust_test.go`
