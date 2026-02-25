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

## Acceptance Criteria

1. `rtmx` crate available on crates.io
2. `#[req("REQ-XXX")]` attribute macro works on test functions
3. Test results output compatible JSON format
4. `rtmx verify --command "cargo test"` correctly updates status
5. Works with standard `#[test]` attribute

## Test Strategy

- Unit tests for macro expansion
- Integration tests with cargo test
- Cross-platform testing (Linux, macOS, Windows)

## Package Structure

```
rtmx/                    # crates.io package
├── Cargo.toml
├── src/
│   ├── lib.rs           # Main library
│   └── output.rs        # JSON output
├── rtmx-macros/         # Proc macro crate
│   ├── Cargo.toml
│   └── src/lib.rs       # #[req] implementation
└── examples/
    └── basic.rs
```

## References

- Rust procedural macros
- cargo test output format
- REQ-LANG-007 marker specification
