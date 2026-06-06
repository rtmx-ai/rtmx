# REQ-ADAPT-014: Wire Orphaned Adapters into Config and Sync Command

## Metadata
- **Category**: ADAPT
- **Subcategory**: Wiring
- **Priority**: P0
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-001, REQ-ADAPT-004, REQ-ADAPT-007, REQ-ADAPT-010, REQ-ADAPT-012
- **Blocks**: REQ-ADAPT-015

## Requirement

The five orphaned adapters (Asana, Monday, GitLab, Slack, Webhook) shall be
wired into the configuration system and sync command so they are usable from
the CLI. Currently these adapters have full library code and tests but no
config wiring, no CLI entry point, and no user-facing command.

## Design

### Config Changes (config.go)

Add the five missing adapter types to `AdaptersConfig`:

```go
type AdaptersConfig struct {
    GitHub  GitHubConfig  `yaml:"github"`
    Jira    JiraConfig    `yaml:"jira"`
    Asana   AsanaConfig   `yaml:"asana"`
    Monday  MondayConfig  `yaml:"monday"`
    GitLab  GitLabConfig  `yaml:"gitlab"`
    Slack   SlackConfig   `yaml:"slack"`
    Webhook WebhookConfig `yaml:"webhook"`
}
```

### Sync Command Changes (sync.go)

Extend `getAdapter()` switch to handle all service types:

```go
case "asana":
    return adapters.NewAsanaAdapter(&cfg.RTMX.Adapters.Asana)
case "monday":
    return adapters.NewMondayAdapter(&cfg.RTMX.Adapters.Monday)
case "gitlab":
    return adapters.NewGitLabAdapter(&cfg.RTMX.Adapters.GitLab)
```

Slack and Webhook are notification-only (no ServiceAdapter), so they get
separate wiring in the serve command rather than sync.

## Acceptance Criteria

1. `AdaptersConfig` struct includes all 7 adapter config types.
2. `getAdapter()` handles github, jira, asana, monday, gitlab.
3. `rtmx sync --service asana` creates and uses AsanaAdapter.
4. `rtmx sync --service monday` creates and uses MondayAdapter.
5. `rtmx sync --service gitlab` creates and uses GitLabAdapter.
6. Slack/Webhook wiring in serve command for notification routes.
7. `rtmx.yaml` with adapter sections parses correctly (not silently ignored).
8. Existing GitHub/Jira sync tests continue to pass.

## Files to Create/Modify

- `internal/config/config.go` -- Add fields to AdaptersConfig
- `internal/cmd/sync.go` -- Extend getAdapter() switch
- `internal/cmd/serve.go` -- Wire Slack/Webhook notification adapters

## Effort Estimate

0.5 weeks
