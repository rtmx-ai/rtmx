# REQ-INT-005: RTM Database Line Ending Contract

## Metadata
- **Category**: INTEGRITY
- **Subcategory**: DataQuality
- **Priority**: HIGH
- **Phase**: 24
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-007
- **Blocks**: (none)

## Requirement

RTM CSV databases shall use a deterministic line-ending contract across
the monorepo and all vendored projects:

1. **Read tolerance**: Parsers shall accept files with CRLF or LF record
   terminators and normalize stray carriage returns in field values.
2. **Write canonical form**: Writers shall emit LF-only line endings
   (`UseCRLF = false`).
3. **Repository hygiene**: Root `.gitattributes` shall pin `*.csv`,
   `*.yaml`, `*.feature`, and related RTM artifacts to `eol=lf`.

## Rationale

Windows editors and some Git configurations introduce CRLF into RTM
databases. Tolerant read avoids silent parse failures; canonical LF on
write and in version control prevents cross-platform diff noise and
status-field corruption in live managed Sync rooms.

## Acceptance Criteria

1. `ReadCSV` parses a minimal CRLF fixture without error
2. `WriteCSV` output contains LF but not CRLF record terminators
3. Root `.gitattributes` declares LF for RTM text artifacts
4. Field normalization strips trailing `\r` after whitespace trim

## Files to Create/Modify

- `.gitattributes` (monorepo root)
- `internal/database/csv.go`
- `internal/database/database_test.go`

## Effort Estimate

0.25 weeks
