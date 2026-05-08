# REQ-LAUNCH-001: README rewritten as Show HN landing page

## Status: MISSING
## Priority: HIGH
## Phase: 14 (Distribution)

## Description
The rtmx README shall be rewritten as a landing page optimized for the
Show HN URL post. When a visitor clicks through from Hacker News, the README
is the pitch. It must communicate value and make starring effortless.

The current README reads as a reference document (installation, migration,
architecture). It needs to function as a landing page first, reference
second.

### Required sections (in order)
1. Hero: logo + one-line pitch + workflow GIF (docs/assets/rtmx-workflow.gif)
2. Install: brew/scoop/go install commands, prominent
3. What it does: 5-6 commands with one-line descriptions
4. The AI workflow: dev-loop diagram (Mermaid) showing next --> code --> verify
5. Why CSV in git: csv-diff diagram (Mermaid) showing a PR diff
6. MCP integration: mcp-architecture diagram (Mermaid)
7. Self-referential: "RTMX manages its own 219 requirements"
8. Links: docs (rtmx.ai), blog post ("Read the backstory"), contributing

### Sections to move or remove
- Migration guide: move to docs/MIGRATION.md, link from README
- Development/architecture: move to CONTRIBUTING.md or keep at bottom
  below a fold

## Acceptance Criteria
- [ ] Hero section with workflow GIF visible without scrolling
- [ ] brew install command within first screenful
- [ ] Mermaid diagrams embedded (GitHub renders natively) or linked images
- [ ] Migration guide moved out of README
- [ ] Development section moved below feature content
- [ ] Link to rtmx.ai/blog/show-hn-rtmx blog post

## Test Cases
- `test/launch_test.go::TestReadmeLaunchReady`

## External References
- rtmx.ai/REQ-SITE-029: Show HN launch checklist (coordinates blog + media + README)
- rtmx.ai/REQ-SITE-028: Media assets produced as code (provides the GIFs and diagrams)
- docs/show-hn-v0.1.0.md (in rtmx.ai repo): full launch playbook

## Notes
The README is the single most important asset for the 100-star goal. A HN
visitor decides to star within 10 seconds of landing on the repo. The hero
GIF and install command must be above the fold.
