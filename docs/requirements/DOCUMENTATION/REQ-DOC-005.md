# REQ-DOC-005: Typst Whitepaper Template

## Requirement

Typst whitepaper template shall match rtmx.ai dark theme with RTMX branding.

## Description

Create a professional Typst template for RTMX whitepapers that maintains visual consistency with the rtmx.ai website. The template should be reusable for all RTMX documentation products.

## Acceptance Criteria

- [ ] Template uses dark background (#1e293b) matching rtmx.ai
- [ ] Sky blue accent color (#0ea5e9) for headings and highlights
- [ ] Green (#22c55e) and amber (#f59e0b) as secondary accents
- [ ] RTMX logo in header/footer
- [ ] Professional typography suitable for technical documentation
- [ ] Code blocks with syntax highlighting
- [ ] Table styling consistent with website
- [ ] PDF output with proper metadata
- [ ] A4 and Letter page size support

## Theme Colors (from rtmx.ai)

| Element | Color | Hex |
|---------|-------|-----|
| Background | Slate 800 | #1e293b |
| Primary | Sky 500 | #0ea5e9 |
| Success | Green 500 | #22c55e |
| Warning | Amber 500 | #f59e0b |
| Text | Slate 100 | #f1f5f9 |
| Muted | Slate 400 | #94a3b8 |

## Files to Create

- `templates/whitepaper.typ` - Main template
- `templates/rtmx-theme.typ` - Theme configuration
- `assets/rtmx-logo-light.svg` - Logo for dark backgrounds

## Dependencies

- REQ-SITE-006: Website dark theme (color reference)

## Blocks

- REQ-DOC-006: RTMX 101 whitepaper
- REQ-DOC-007: Patterns whitepaper
