# REQ-AGENT-001: Expanded AI Agent Config Support

## Metadata
- **Category**: ADAPT
- **Subcategory**: Agents
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-029
- **Blocks**: REQ-GO-047

## Requirement

`rtmx install --agents` shall support config injection for all major CLI/IDE AI agents, expanding beyond the current Claude Code, Cursor, and GitHub Copilot support.

## Rationale

The AI-assisted development ecosystem has expanded significantly. RTMX should provide out-of-the-box integration with all major agents to reduce friction for teams adopting requirements traceability in agentic workflows.

## Currently Supported (3)

| Agent | Config File | Format |
|-------|------------|--------|
| Claude Code | `CLAUDE.md`, `.claude/CLAUDE.md` | Markdown |
| Cursor | `.cursorrules` | Markdown |
| GitHub Copilot | `.github/copilot-instructions.md` | Markdown |

## New Agents to Support (7)

| Agent | Config File | Format | Priority |
|-------|------------|--------|----------|
| Cline | `.clinerules` | Markdown | HIGH |
| Gemini CLI | `GEMINI.md` | Markdown | HIGH |
| Windsurf/Cascade | `.windsurfrules` | Markdown | HIGH |
| Aider | `.aider.conf.yml` | YAML | MEDIUM |
| Amazon Q Developer | `.amazonq/rules` | Markdown | MEDIUM |
| Zed Editor | `.zed/settings.json` (assistant.instructions) | JSON | MEDIUM |
| Continue.dev | `.continue/config.yaml` (system message) | YAML | LOW |

## Design

### Agent Registry

```go
type AgentConfig struct {
    Name        string   // Display name
    ID          string   // CLI identifier
    ConfigPaths []string // Possible config file locations (first found wins)
    Format      string   // "markdown", "yaml", "json"
    InjectFn    func(existing string, rtmxContext string) string
}

var agents = []AgentConfig{
    {Name: "Claude Code",    ID: "claude",    ConfigPaths: []string{"CLAUDE.md", ".claude/CLAUDE.md"}, Format: "markdown"},
    {Name: "Cursor",         ID: "cursor",    ConfigPaths: []string{".cursorrules"}, Format: "markdown"},
    {Name: "GitHub Copilot", ID: "copilot",   ConfigPaths: []string{".github/copilot-instructions.md"}, Format: "markdown"},
    {Name: "Cline",          ID: "cline",     ConfigPaths: []string{".clinerules"}, Format: "markdown"},
    {Name: "Gemini CLI",     ID: "gemini",    ConfigPaths: []string{"GEMINI.md"}, Format: "markdown"},
    {Name: "Windsurf",       ID: "windsurf",  ConfigPaths: []string{".windsurfrules"}, Format: "markdown"},
    {Name: "Aider",          ID: "aider",     ConfigPaths: []string{".aider.conf.yml"}, Format: "yaml"},
    {Name: "Amazon Q",       ID: "amazonq",   ConfigPaths: []string{".amazonq/rules"}, Format: "markdown"},
    {Name: "Zed",            ID: "zed",       ConfigPaths: []string{".zed/settings.json"}, Format: "json"},
    {Name: "Continue.dev",   ID: "continue",  ConfigPaths: []string{".continue/config.yaml"}, Format: "yaml"},
}
```

### CLI Interface

```bash
rtmx install --agents              # Install all detected agents
rtmx install --agents claude       # Install specific agent
rtmx install --agents cline,gemini # Install multiple
rtmx install --agents --all        # Install all (create configs if missing)
rtmx install --agents --list       # List supported agents and detection status
```

### Context Template

All agents receive the same RTMX context, adapted to their format:

**Markdown agents** (Claude, Cursor, Copilot, Cline, Gemini, Windsurf, Amazon Q):
```markdown
## RTMX Requirements Traceability

This project uses RTMX for requirements traceability.

### Quick Commands
- `rtmx status` - Show completion status
- `rtmx backlog` - Show prioritized backlog
- `rtmx verify --update` - Run tests and update status

### Test Markers
Every test must link to a requirement:
- Go: `rtmx.Req(t, "REQ-XXX-NNN")`
- Python: `@pytest.mark.req("REQ-XXX-NNN")`
```

**YAML agents** (Aider, Continue.dev):
```yaml
# RTMX Requirements Traceability
rtmx:
  commands:
    status: "rtmx status"
    backlog: "rtmx backlog"
    verify: "rtmx verify --update"
  markers:
    go: 'rtmx.Req(t, "REQ-XXX-NNN")'
    python: '@pytest.mark.req("REQ-XXX-NNN")'
```

**JSON agents** (Zed):
```json
{
  "assistant": {
    "instructions": "This project uses RTMX for requirements traceability. Use rtmx status, rtmx backlog, rtmx verify --update. Mark tests with rtmx.Req() or @pytest.mark.req()."
  }
}
```

## Acceptance Criteria

1. `rtmx install --agents` detects and configures all supported agents
2. `rtmx install --agents --list` shows 10 agents with detection status
3. Each agent's config file is correctly formatted (markdown/yaml/json)
4. Existing config content is preserved (RTMX section appended/updated)
5. `--dry-run` shows what would be changed without modifying files
6. Backup created before modification (unless `--skip-backup`)
7. Idempotent: running twice doesn't duplicate the RTMX section

## Files to Modify

- `internal/cmd/install.go` - Expand agent registry, add new agents
- `internal/cmd/install_test.go` - Tests for each agent format
- Consider extracting to `internal/agents/` package if install.go gets too large

## Test Strategy

- Unit test per agent: verify config generation for each format
- Integration test: `rtmx install --agents --list` shows all 10
- Idempotency test: install twice, verify no duplication
- Backup test: verify backup created before modification
