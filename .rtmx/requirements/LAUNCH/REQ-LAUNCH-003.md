# REQ-LAUNCH-003: VHS GIF Generation from Terminal Demos

## Metadata
- **Category**: LAUNCH
- **Subcategory**: ShowHN
- **Priority**: HIGH
- **Phase**: 26
- **Status**: MISSING
- **Effort**: 0.5 weeks
- **Dependencies**: REQ-LAUNCH-001 (README rewritten with GIF placeholders)
- **Blocks**: REQ-LAUNCH-002

## Requirement

RTMX shall generate terminal GIF animations from VHS tape files and embed them
in the README. The workflow GIF shall demonstrate the core status/next/verify
loop in under 30 seconds. The agent-loop GIF shall demonstrate AI-driven
requirement completion.

## Rationale

REQ-LAUNCH-001 restructured the README as a landing page with a commented-out
hero GIF placeholder. The tape files exist at `docs/tapes/workflow.tape` and
`docs/tapes/agent-loop.tape` but no GIFs have been generated. Show HN posts
with a workflow GIF in the first 3 seconds get significantly more engagement.

## VHS Tape Inventory

| Tape | Output | Duration | Purpose |
|------|--------|----------|---------|
| `docs/tapes/workflow.tape` | `docs/assets/rtmx-workflow.gif` | ~15s | Hero GIF: status -> next -> verify -> status |
| `docs/tapes/agent-loop.tape` | `docs/assets/rtmx-agent-loop.gif` | ~15s | AI agent picking and completing a requirement |

## Implementation

1. Install VHS: `brew install charmbracelet/tap/vhs`
2. Ensure ttyd and ffmpeg are available (VHS dependencies)
3. Run `vhs docs/tapes/workflow.tape` to generate workflow GIF
4. Run `vhs docs/tapes/agent-loop.tape` to generate agent-loop GIF
5. Uncomment and update GIF references in README.md
6. Add Makefile target: `make gifs`
7. Add CI step to regenerate GIFs on tape file changes (optional)

## Acceptance Criteria

1. VHS installed and tape files produce valid GIF output
2. `docs/assets/rtmx-workflow.gif` exists and shows status/next/verify flow
3. `docs/assets/rtmx-agent-loop.gif` exists and shows agent requirement loop
4. README.md embeds the hero GIF (uncommented, visible)
5. GIFs are under 5MB each for fast loading
6. `make gifs` target generates both GIFs
7. GIFs render correctly on GitHub (dark and light themes)

## Verification

Test validates tape files exist with correct Output directives, docs/assets/
directory exists with generated GIFs, and README.md references them without
HTML comments.

## Files to Create/Modify

- `docs/assets/rtmx-workflow.gif` -- generated hero GIF
- `docs/assets/rtmx-agent-loop.gif` -- generated agent loop GIF
- `README.md` -- uncomment GIF embed
- `Makefile` -- add `gifs` target
- `docs/tapes/workflow.tape` -- may need tuning for timing/clarity
- `docs/tapes/agent-loop.tape` -- may need tuning for timing/clarity

## Notes

- VHS (charmbracelet/vhs) renders tape files to GIF/MP4/WebM
- Tapes must run against a real project with requirements to show real output
- Consider optimizing GIF size with gifsicle if over 5MB
- The workflow.tape currently uses the project's own database for dogfooding
