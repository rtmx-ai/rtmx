# REQ-PLUGIN-003d: Gastown Integration

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Gastown
- **Priority**: MEDIUM
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003, REQ-MCP-005
- **Blocks**: (none)

## Requirement
`rtmx install --gastown` shall provide plugin integration for Steve Yegge's Gastown project, pending architecture investigation.

## Design
Integration method depends on Gastown's plugin architecture (MCP, HTTP, or CLI). This requirement includes the investigation phase.

## Acceptance Criteria
1. Gastown architecture investigated and integration method chosen.
2. rtmx install --gastown generates appropriate config.
3. Gastown can discover RTMX as a tool provider.

## Files to Create/Modify
- internal/cmd/install.go
- templates/gastown/
