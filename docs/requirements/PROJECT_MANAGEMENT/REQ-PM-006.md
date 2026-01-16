# REQ-PM-006: Release management

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 17
## Estimated Effort: 2.0 weeks

## Description

System shall provide release management capabilities for grouping requirements into versioned releases. This includes release planning, release notes generation, and git tag integration for coordinating requirement completion with software releases.

## Acceptance Criteria

- [ ] `rtmx release list` displays all releases with status and requirement counts
- [ ] `rtmx release create <version> --name <name>` creates new release
- [ ] `rtmx release assign <version> <req_id>...` assigns requirements to release
- [ ] `rtmx release unassign <version> <req_id>...` removes requirements from release
- [ ] `rtmx release show <version>` displays release details with requirements
- [ ] `rtmx release notes <version>` generates release notes markdown
- [ ] `rtmx release ship <version>` marks release as shipped with timestamp
- [ ] Release status: `planning`, `in_progress`, `ready`, `shipped`
- [ ] Release notes template is configurable in `rtmx.yaml`
- [ ] Git tag integration: `rtmx release tag <version>` creates git tag
- [ ] Release validation: warn if assigned requirements are not COMPLETE
- [ ] Release can be linked to epics (include all epic's requirements)
- [ ] Version format validation (semver by default, configurable)
- [ ] `rtmx release diff <v1> <v2>` shows changes between releases
- [ ] Export release notes as JSON for CI/CD integration

## Test Cases

- `tests/test_release.py::test_release_create` - Create new release
- `tests/test_release.py::test_release_list` - List all releases
- `tests/test_release.py::test_release_assign` - Assign requirements to release
- `tests/test_release.py::test_release_unassign` - Remove requirements from release
- `tests/test_release.py::test_release_show` - Display release details
- `tests/test_release.py::test_release_notes_generation` - Generate release notes
- `tests/test_release.py::test_release_ship` - Ship release with timestamp
- `tests/test_release.py::test_release_git_tag` - Create git tag for release
- `tests/test_release.py::test_release_validation` - Warn on incomplete requirements
- `tests/test_release.py::test_release_diff` - Diff between releases
- `tests/test_release.py::test_release_epic_link` - Link epic to release

## Technical Notes

### Release CSV Schema

```csv
version,name,status,created_at,shipped_at,git_tag,description
v0.5.0,Sprint Planning Release,shipped,2024-01-15T10:00:00Z,2024-01-29T17:00:00Z,v0.5.0,Added sprint and velocity features
v0.6.0,Custom Workflows,in_progress,2024-01-30T10:00:00Z,,Custom status workflows
```

### Release-Requirement Junction

```csv
version,req_id,added_at
v0.5.0,REQ-PM-001,2024-01-15T10:00:00Z
v0.5.0,REQ-PM-002,2024-01-15T10:00:00Z
v0.6.0,REQ-PM-004,2024-01-30T10:00:00Z
```

### Release Notes Template

```yaml
release:
  version_format: semver  # or custom regex
  notes_template: |
    # Release {version}: {name}

    Released: {shipped_at}

    ## Features
    {%- for req in features %}
    - {req.description} ({req.req_id})
    {%- endfor %}

    ## Bug Fixes
    {%- for req in fixes %}
    - {req.description} ({req.req_id})
    {%- endfor %}

    ## Breaking Changes
    {%- for req in breaking %}
    - {req.description} ({req.req_id})
    {%- endfor %}
  categories:
    features: [FEATURES, API, CLI]
    fixes: [BUGFIX, HOTFIX]
    breaking: [BREAKING]
```

### CLI Examples

```bash
$ rtmx release create v0.6.0 --name "Custom Workflows"
Created release v0.6.0

$ rtmx release assign v0.6.0 REQ-PM-004 REQ-PM-007
Assigned 2 requirements to v0.6.0

$ rtmx release show v0.6.0
Release: v0.6.0 - Custom Workflows
Status: in_progress
Requirements: 2 total (1 COMPLETE, 1 IN_PROGRESS)

  REQ-PM-004  COMPLETE     Custom status workflows
  REQ-PM-007  IN_PROGRESS  Custom fields schema

$ rtmx release notes v0.6.0
# Release v0.6.0: Custom Workflows

## Features
- Custom status workflows with state machine (REQ-PM-004)
- User-defined custom fields schema (REQ-PM-007)

$ rtmx release ship v0.6.0
Warning: 1 requirement is not COMPLETE
Ship anyway? [y/N]: y
Release v0.6.0 shipped at 2024-02-15T17:00:00Z

$ rtmx release tag v0.6.0
Created git tag v0.6.0
```

### Git Tag Integration

```python
def create_release_tag(version: str, message: str | None = None) -> None:
    """Create annotated git tag for release."""
    release = get_release(version)
    tag_message = message or f"Release {version}: {release.name}"
    subprocess.run([
        "git", "tag", "-a", version, "-m", tag_message
    ], check=True)
```

## Dependencies

- REQ-PM-005: Epic/Initiative hierarchy (releases can include epics)

## Blocks

None - this is a leaf requirement for release workflows.
