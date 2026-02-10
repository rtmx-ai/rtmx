# REQ-CLAUDE-004: Claude.ai Web MCP Extension

## Status: NOT_STARTED
## Priority: MEDIUM
## Phase: 19
## Effort: 2.0 weeks

## Description

RTMX shall provide a browser extension or Claude.ai integration that enables MCP-style RTM access from the Claude.ai web interface. This allows non-CLI users to benefit from requirements traceability when using Claude.ai for code review, documentation, or planning.

## Rationale

Many users interact with Claude through claude.ai rather than Claude Code:
- Code reviewers who don't have local checkouts
- Product managers discussing requirements
- Stakeholders reviewing project status
- Mobile users on the go

A web-accessible RTM viewer/editor expands RTMX's reach beyond CLI-native developers.

## Acceptance Criteria

- [ ] Browser extension connects to rtmx-sync for authenticated RTM access
- [ ] Claude.ai sidebar shows project RTM status when discussing code
- [ ] Extension detects GitHub/GitLab URLs and fetches associated RTM
- [ ] User can query requirements via natural language in Claude.ai
- [ ] Read-only mode for users without write access
- [ ] OAuth authentication via Zitadel (same as rtmx-sync)
- [ ] Works in Chrome, Firefox, and Safari
- [ ] Extension respects privacy settings from rtmx-sync

## Technical Notes

### Architecture Options

**Option A: Browser Extension**
```
┌──────────────────┐
│    Claude.ai     │
│   (browser tab)  │
└────────┬─────────┘
         │ postMessage
┌────────▼─────────┐
│  RTMX Extension  │
│ (content script) │
└────────┬─────────┘
         │ HTTPS
┌────────▼─────────┐
│   rtmx-sync API  │
└──────────────────┘
```

**Option B: Claude.ai MCP Integration (if supported)**
```
┌──────────────────┐
│    Claude.ai     │
│  (MCP-enabled)   │
└────────┬─────────┘
         │ MCP Protocol
┌────────▼─────────┐
│  rtmx-sync MCP   │
│    endpoint      │
└──────────────────┘
```

### Extension Manifest (Chrome)

```json
{
  "manifest_version": 3,
  "name": "RTMX for Claude.ai",
  "version": "0.1.0",
  "description": "Requirements traceability for Claude.ai conversations",
  "permissions": ["storage", "identity"],
  "host_permissions": ["https://claude.ai/*", "https://*.rtmx.ai/*"],
  "content_scripts": [{
    "matches": ["https://claude.ai/*"],
    "js": ["content.js"],
    "css": ["sidebar.css"]
  }],
  "background": {
    "service_worker": "background.js"
  }
}
```

### API Endpoints

Extension communicates with rtmx-sync:
- `GET /api/v1/projects` - List accessible projects
- `GET /api/v1/projects/{id}/requirements` - Fetch RTM
- `GET /api/v1/requirements/{id}` - Single requirement detail
- `POST /api/v1/requirements/{id}/status` - Update status (if permitted)

### URL Detection

Extension detects repository context from:
1. GitHub/GitLab URLs in conversation
2. Code blocks with file paths
3. User-provided project context

## Gherkin Specification

```gherkin
@REQ-CLAUDE-004 @scope_system @technique_nominal
Feature: Claude.ai Web MCP Extension
  As a Claude.ai user
  I want to access RTM information while chatting
  So that I can discuss requirements without switching to terminal

  Background:
    Given the RTMX browser extension is installed
    And I am logged into rtmx-sync

  Scenario: Extension shows RTM sidebar
    Given I open claude.ai
    When I click the RTMX extension icon
    Then a sidebar appears showing my projects
    And I can select a project to view its RTM

  Scenario: Detect GitHub URL context
    Given I paste a GitHub URL "https://github.com/rtmx-ai/rtmx"
    When the extension detects the URL
    Then it fetches the associated RTM from rtmx-sync
    And displays relevant requirements in the sidebar

  Scenario: Query requirements via Claude
    Given the RTM sidebar shows project "rtmx"
    When I ask Claude "What requirements are blocking REQ-AUTH-001?"
    Then Claude can reference the RTM data from the extension
    And provides accurate blocking information

  Scenario: Read-only for unauthorized users
    Given I have read-only access to project "acme-corp"
    When I view the RTM sidebar
    Then status update controls are disabled
    And I see "Read-only access" indicator
```

## Test Cases

1. `tests/test_web_extension.py::test_extension_oauth_flow`
2. `tests/test_web_extension.py::test_project_list_fetch`
3. `tests/test_web_extension.py::test_github_url_detection`
4. `tests/test_web_extension.py::test_sidebar_rendering`
5. `tests/test_web_extension.py::test_readonly_mode`
6. `tests/test_web_extension.py::test_cross_browser_compatibility`

## Files to Create

- `extension/` (new directory) - Browser extension source
- `extension/manifest.json` - Chrome manifest
- `extension/manifest-firefox.json` - Firefox manifest
- `extension/content.js` - Content script
- `extension/background.js` - Service worker
- `extension/sidebar.html` - Sidebar UI
- `extension/sidebar.css` - Sidebar styles
- `extension/api.js` - rtmx-sync API client

## Dependencies

- REQ-ZT-001: Zitadel OIDC integration (authentication)
- REQ-ZT-002: rtmx-sync dark service (API backend)

## Blocks

- None (leaf requirement)
