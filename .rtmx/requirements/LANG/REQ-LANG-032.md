# REQ-LANG-032: Reset class context for module-level functions in the pytest scanner

## Metadata
- **Category**: LANG
- **Subcategory**: Python
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-LANG-004

## Requirement

The Python marker scanner shall attribute a requirement marker to the correct
enclosing scope: a test method inside a class is recorded as `Class::method`,
while a module-level test function — including one defined AFTER a test class —
is recorded with its bare function name. The scanner leaves a class context once
a non-blank, non-comment line appears at or below the class definition's
indentation.

## Rationale

The scanner tracked the current class name but never cleared it, so once any test
class was seen, every subsequent function — even module-level ones after the class
block — was qualified as `Class::function`. That stale qualification did not match
the JUnit test case (whose class-name-derived key carries the real class, or none
for a module-level test), so those tests silently failed to join and their
requirements never promoted. This is common in files that mix a test class with
module-level test functions (e.g. the Phoenix radar RTM's
`packages/signal-processing/tests/test_range_vs_rcs.py`, where the SW-DSP
module-level tests follow a documentation test class).

## Acceptance Criteria

- [ ] A test method inside a class is recorded as `Class::method`.
- [ ] A module-level test function defined after a class is recorded with a bare,
      unqualified function name.
- [ ] Consecutive classes each scope only their own methods.
- [ ] Verified by `internal/cmd/from_tests_test.go::TestExtractMarkersModuleFuncAfterClass`.
