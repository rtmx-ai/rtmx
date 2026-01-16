# REQ-CRDT-002: Requirements shall be represented as Y.Map structures

## Status: COMPLETE
## Priority: HIGH
## Phase: 9

## Description
Each requirement shall be represented as a Y.Map structure within the CRDT document, enabling field-level conflict resolution during collaborative editing.

## Acceptance Criteria
- [x] Each requirement is stored as a Y.Map in the document's requirements map
- [x] All Requirement model fields are mapped to Y.Map entries
- [x] `requirement_to_ymap()` converts Requirement to Y.Map-compatible dict
- [x] `ymap_to_requirement()` converts Y.Map back to Requirement
- [x] Round-trip conversion preserves all data
- [x] Optional fields (phase, effort_weeks) handle None correctly
- [x] Set fields (dependencies, blocks) serialize as pipe-delimited strings
- [x] Extra fields are preserved in conversion

## Test Cases
- `tests/test_crdt.py::TestRequirementToYmap::test_basic_fields`
- `tests/test_crdt.py::TestRequirementToYmap::test_dependencies_serialization`
- `tests/test_crdt.py::TestYmapToRequirement::test_basic_fields`
- `tests/test_crdt.py::TestYmapToRequirement::test_dependencies_parsing`
- `tests/test_crdt.py::TestRoundTrip::test_roundtrip_preserves_data`

## Notes
Each requirement field uses Last-Writer-Wins (LWW) semantics by default in Y.Map. This means concurrent edits to the same field will resolve to the most recent write.

## Dependencies
- REQ-CRDT-001: pycrdt library

## Blocks
- REQ-CRDT-005: Offline operations
- REQ-CRDT-006: Concurrent merge
