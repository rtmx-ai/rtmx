# REQ-GO-089: Interactive rtmx link and flagless bootstrap picker

## Status: COMPLETE
## Priority: P0
## Phase: 22
## External ID: rtmx-ai/REQ-MONO-044b

## Description

Go CLI SHALL provide `rtmx link` and flagless `rtmx bootstrap` (TTY) as an
interactive provider picker that runs the MONO-040 bootstrap pipeline.

## Acceptance Criteria

- [x] `rtmx link` command registered; help documents providers
- [x] Picker model tests (select → invoke bootstrap source) without network
- [x] Non-TTY bootstrap without `--from-*` errors clearly
- [x] Skip and `--dry-run` paths tested

## Dependencies

- REQ-GO-087
- REQ-GO-088
- rtmx-ai/REQ-MONO-044b
