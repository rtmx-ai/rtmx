# REQ-GO-090: init TTY handoff to rtmx link

## Status: COMPLETE
## Priority: HIGH
## Phase: 22
## External ID: rtmx-ai/REQ-MONO-044c

## Description

After interactive `rtmx init` succeeds, offer the link picker when stdin is a
TTY and `--yes` is unset. Non-interactive init never prompts.

## Acceptance Criteria

- [x] TTY handoff tested with fake linker
- [x] `--yes` / non-TTY skip tested
- [x] Init files remain if link/bootstrap fails

## Dependencies

- REQ-GO-089
- rtmx-ai/REQ-MONO-044c
