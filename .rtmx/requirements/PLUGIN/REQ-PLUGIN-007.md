# REQ-PLUGIN-007: Community Schema Plugin Distribution

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-005
- **Blocks**: (none)

## Requirement

RTMX shall support installing community-contributed schema plugins via
`rtmx plugin install <name>`. Plugins are versioned, resolvable from a
registry or Git URL, and installed to the user-global or project-local
schemas directory.

## Acceptance Criteria

1. `rtmx plugin install rtmx-schema-fedramp` installs from registry/Git
2. `rtmx plugin list` shows installed plugins with version and source
3. `rtmx plugin remove <name>` uninstalls cleanly
4. Version conflicts between plugins are detected and reported
5. Installed plugins are available in `rtmx schema list`

## Effort Estimate

2 weeks (registry resolution + Git clone + version management + tests)
