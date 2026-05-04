# REQ-PLAN-014: Category-Driven Version Increment Policy

## Metadata
- **Category**: PLAN
- **Subcategory**: Versioning
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-005, REQ-PLAN-003

## Requirement

RTMX shall support a `version_policy` configuration section that maps
requirement categories to semver increment levels (major, minor, patch).
The `rtmx release gate` command shall use this policy to validate that
the proposed version tag matches the highest-impact category present in
the release scope, and recommend the correct version bump when it does not.

## Rationale

The versioning policy in CLAUDE.md defines major/minor/patch rules in prose,
but enforcement is manual. A team completing a requirement in category CLI
(new command = minor bump) alongside one in category BENCH (test improvement
= patch) should be told "this release requires at least a minor bump" before
they tag v0.3.1 when they should tag v0.4.0.

By making the version strategy traceable to requirement categories, the
increment decision is auditable: "v0.4.0 is a minor bump because
REQ-PLAN-005 (category PLAN, subcategory Release) introduced a new command."

## Design

### Configuration

```yaml
# .rtmx/config.yaml or rtmx.yaml
rtmx:
  version_policy:
    # Map categories to their maximum semver impact.
    # When a release contains completed requirements from multiple
    # categories, the highest-impact category determines the minimum
    # version bump.
    #
    # Levels: major, minor, patch, none
    # - major: breaking change to public surface
    # - minor: new feature, backward compatible
    # - patch: bug fix, no new features
    # - none: no version impact (documentation, internal tooling)

    categories:
      CLI:        minor    # new/changed commands or flags
      DATA:       major    # CSV format or config schema changes
      SYNC:       minor    # sync protocol changes
      ADAPT:      minor    # adapter interface changes
      OUTPUT:     minor    # output format changes
      MCP:        minor    # MCP protocol changes
      PLUGIN:     minor    # plugin framework changes
      AUTH:       minor    # authentication changes
      PLAN:       minor    # planning commands
      VERIFY:     minor    # verification changes
      GRAPH:      patch    # algorithm improvements
      PARITY:     patch    # parity fixes
      MIGRATION:  patch    # migration tooling
      E2E:        patch    # test infrastructure
      CI:         none     # CI/CD pipeline
      TEST:       none     # test coverage
      BENCH:      none     # benchmark infrastructure
      SECURITY:   patch    # security hardening
      DIST:       patch    # distribution changes
      RELEASE:    none     # release process
      ORCH:       minor    # orchestration commands
      INTEGRITY:  patch    # integrity checks
      LANG:       minor    # language extension support

    # Subcategory overrides (optional, more specific wins)
    subcategories:
      CLI/Commands:    minor   # new commands
      CLI/Foundation:  major   # foundational CLI changes
      DATA/Config:     major   # config schema changes
      DATA/CSV:        major   # database format changes
      SECURITY/ZeroTrust: minor  # new security features
      PLAN/Release:    minor   # new release commands
      PLAN/Display:    patch   # display-only changes
      PLAN/Automation: patch   # internal automation

    # Default for unmapped categories
    default: patch
```

### Go Implementation

```go
// internal/config/config.go

type VersionPolicyConfig struct {
    Categories    map[string]string `yaml:"categories"`
    Subcategories map[string]string `yaml:"subcategories"`
    Default       string            `yaml:"default"`
}

// IncrementLevel returns the semver increment level for a category/subcategory.
func (v *VersionPolicyConfig) IncrementLevel(category, subcategory string) string {
    // Subcategory override wins (more specific)
    key := category + "/" + subcategory
    if level, ok := v.Subcategories[key]; ok {
        return level
    }
    // Category level
    if level, ok := v.Categories[category]; ok {
        return level
    }
    // Default
    if v.Default != "" {
        return v.Default
    }
    return "patch"
}
```

### Semver Parsing

```go
// internal/version/semver.go

type Version struct {
    Major      int
    Minor      int
    Patch      int
    Prerelease string
}

// Parse parses a version string like "v1.2.3" or "v1.2.3-rc1".
func Parse(s string) (Version, error)

// BumpMajor returns v+1.0.0
func (v Version) BumpMajor() Version

// BumpMinor returns v.major+1.0
func (v Version) BumpMinor() Version

// BumpPatch returns v.major.minor+1
func (v Version) BumpPatch() Version

// String returns "vMAJOR.MINOR.PATCH"
func (v Version) String() string
```

No external dependency needed -- semver parsing is ~30 lines of standard
library code (strings.Split, strconv.Atoi).

### Gate Integration

`rtmx release gate <version>` gains version policy validation:

1. Load version policy from config
2. For each COMPLETE requirement in the release scope, determine its
   increment level from category/subcategory
3. Compute the highest increment level across all requirements
4. Parse the proposed version tag and the previous version tag
   (from `git describe --tags --abbrev=0 HEAD~1` or the previous release)
5. Verify the actual bump matches or exceeds the required bump
6. If mismatch: warn or fail depending on strictness config

```
$ rtmx release gate v0.3.1

Release Gate: v0.3.1
  Requirements: 5 total, 5 complete

  Version policy check:
    REQ-PLAN-005 (PLAN/Release)  -> minor
    REQ-PLAN-003 (PLAN/Query)    -> minor
    REQ-PLAN-007 (PLAN/Release)  -> minor
    REQ-PLAN-010 (PLAN/Automation) -> patch
    REQ-E2E-005 (E2E/Onboarding) -> patch

  Highest impact: minor
  Proposed bump: v0.2.7 -> v0.3.1 = patch

  WARN: Release contains minor-level changes but version bump is patch.
  Recommended: v0.3.0 (minor bump from v0.2.7)
```

### Backward Compatibility Constraint

The version policy enforces backward compatibility by default. A
requirement that would break backward compatibility (mapped to `major`)
triggers an elevated review signal in the gate output. The config
supports a `backward_compatible` flag that, when true, causes the gate
to **fail** on any release containing major-level requirements -- forcing
the team to either redesign for compatibility or explicitly override.

```yaml
rtmx:
  version_policy:
    backward_compatible: true   # default: true
```

When `backward_compatible: true`:
- Requirements mapped to `major` cause gate failure unless the version
  tag is itself a major bump (e.g., v2.0.0)
- Gate output lists the specific requirements and categories that would
  break compatibility, with the field/command/format they affect
- Override: `rtmx release gate v2.0.0 --allow-breaking` explicitly
  acknowledges the breaking change

This ensures backward compatibility is the default posture. Breaking
changes require a deliberate major version bump, not an accidental one.

### Strictness Levels

```yaml
rtmx:
  version_policy:
    enforcement: warn    # warn | enforce | off
```

- `off`: no version policy check (default for backwards compatibility)
- `warn`: print warning but allow the tag (exit 0)
- `enforce`: fail the gate if bump is insufficient (exit 1)

### CLI Output

```
$ rtmx release gate v0.4.0 --json
{
  "version": "v0.4.0",
  "passed": true,
  "total": 5,
  "complete": 5,
  "version_policy": {
    "required_bump": "minor",
    "actual_bump": "minor",
    "previous_version": "v0.3.0",
    "compliant": true,
    "category_impacts": [
      {"category": "PLAN", "subcategory": "Release", "level": "minor", "count": 3},
      {"category": "PLAN", "subcategory": "Automation", "level": "patch", "count": 1},
      {"category": "E2E", "subcategory": "Onboarding", "level": "patch", "count": 1}
    ]
  }
}
```

## Acceptance Criteria

1. `version_policy.categories` maps category names to increment levels
2. `version_policy.subcategories` provides override specificity
3. `version_policy.default` applies to unmapped categories (default: patch)
4. `version_policy.enforcement` controls warn/enforce/off behavior
5. `rtmx release gate` includes version policy check in output
6. `rtmx release gate --json` includes category_impacts in JSON output
7. Insufficient version bump produces clear diagnostic with recommendation
8. `enforcement: enforce` causes gate failure on insufficient bump
9. `enforcement: off` skips policy check entirely (backwards compatible)
10. No external semver dependency (standard library only)
11. Previous version determined from git tags (`git describe --tags`)
12. `backward_compatible: true` (default) fails gate on major-level requirements
    unless the version tag is itself a major bump
13. `--allow-breaking` flag overrides backward compatibility constraint
14. Gate output identifies specific requirements that would break compatibility

## Files to Create/Modify

- `internal/config/config.go` -- Add VersionPolicyConfig struct and field
- `internal/version/semver.go` -- Semver parsing and bump logic (new)
- `internal/version/semver_test.go` -- Semver tests (new)
- `internal/cmd/release.go` -- Integrate policy check into gate
- `internal/cmd/release_test.go` -- Gate policy tests
- `.rtmx/config.yaml` -- Add version_policy section

## Effort Estimate

1.5 weeks (config struct + semver parser + gate integration + tests)
