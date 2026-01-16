# REQ-CRDT-001: CRDT layer shall use pycrdt library

## Status: COMPLETE
## Priority: HIGH
## Phase: 9

## Description
CRDT layer shall use pycrdt library for conflict-free replicated data types, enabling real-time collaborative editing of requirements.

## Acceptance Criteria
- [x] pycrdt package is an optional dependency (`pip install rtmx[sync]`)
- [x] `is_sync_available()` function checks for pycrdt availability
- [x] `require_sync()` raises ImportError with helpful message if pycrdt not installed
- [x] RTMDocument class wraps pycrdt Y.Doc
- [x] All CRDT operations work when pycrdt is installed

## Test Cases
- `tests/test_crdt.py::TestRequirementToYmap` - Verifies requirement conversion to Y.Map
- `tests/test_crdt.py::TestYmapToRequirement` - Verifies Y.Map conversion back to requirement
- `tests/test_crdt.py::TestRTMDocument` - Verifies Y.Doc wrapper operations

## Notes
pycrdt provides Python bindings for Yrs, a Rust implementation of Yjs CRDTs. This is the modern successor to y-py and provides better performance and maintenance.

## Dependencies
- None (foundation requirement)

## Blocks
- REQ-CRDT-002: Requirement Y.Map structures
- REQ-CRDT-003: CSV sync
