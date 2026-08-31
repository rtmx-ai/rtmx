# REQ-GO-081: CSV Line Ending Normalization

## Metadata
- **Category**: DATA
- **Subcategory**: CSV
- **Priority**: HIGH
- **Phase**: 24
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-007
- **Blocks**: REQ-INT-005

## Requirement

The Go CLI CSV reader and writer shall implement the line-ending contract
defined in REQ-INT-005:

- `ReadCSV` accepts CRLF-terminated input via `encoding/csv` and
  `normalizeCellValue` for field-level `\r` cleanup.
- `WriteCSV` sets `writer.UseCRLF = false` so saved databases always
  use Unix line endings.

## Rationale

REQ-GO-007 guarantees Python/Go round-trip parity for well-formed LF
files. This requirement closes the gap for CRLF inputs encountered on
Windows and in managed Sync deployments where editors differ.

## Acceptance Criteria

1. `TestCSVCrlfInput` passes against an inline CRLF fixture
2. `TestWriteCSVEmitsLF` confirms LF-only output
3. Existing `TestCSVRoundTrip` remains green

## Files to Create/Modify

- `internal/database/csv.go`
- `internal/database/database_test.go`

## Effort Estimate

0.25 weeks
