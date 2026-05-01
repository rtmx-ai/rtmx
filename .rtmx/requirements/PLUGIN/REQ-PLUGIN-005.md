# REQ-PLUGIN-005: Schema Plugin Framework for Custom RTM Columns

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Schema
- **Priority**: P0
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-GO-008
- **Blocks**: REQ-PLUGIN-006, REQ-PLUGIN-007

## Requirement

RTMX shall provide a schema plugin framework that allows teams to define,
distribute, and enforce custom RTM column sets for domain-specific
certification and compliance workflows. Schema plugins declare typed columns
with validation rules, integrate with `rtmx health` for schema-aware checks,
and are loadable via configuration without modifying core RTMX code.

## Rationale

The core 20-column schema serves general-purpose requirements traceability.
Domain-specific certification programs (DO-178C for avionics, ISO 26262 for
automotive safety, FedRAMP for cloud security, CMMC for defense cyber) each
require additional structured fields that do not belong in core. Today, teams
can add arbitrary columns via the `Extra map[string]string` field on the
Requirement struct, but there is no validation, no type enforcement, no
documentation generation, and no way to share a schema definition across
projects or organizations.

The existing `schema: core | phoenix` config option and the Python-era
`register_schema()` API in docs/schema.md describe the intended design
surface but are not implemented in the Go CLI. This requirement bridges
that gap with a concrete, minimal plugin framework.

## Design

### Schema Definition File

A schema plugin is a YAML file that declares additional columns, their types,
validation rules, and display behavior. Schema files live in a well-known
location and are referenced by name in rtmx.yaml.

```yaml
# .rtmx/schemas/do178c.yaml (or distributed as a package)
schema:
  name: do178c
  version: "1.0.0"
  description: "DO-178C airborne software certification columns"
  extends: core

  columns:
    - name: dal_level
      type: enum
      values: [A, B, C, D, E]
      required: true
      description: "Design Assurance Level per DO-178C Table A-1"
      display:
        show_in_status: true
        show_in_backlog: true

    - name: verification_method
      type: enum
      values: [review, analysis, test]
      required: true
      description: "Verification method per DO-178C Table A-3 through A-7"

    - name: independence_level
      type: enum
      values: [independent, non_independent, N/A]
      required: false
      description: "Independence of verification activity"

    - name: sw_level_objective
      type: string
      required: false
      description: "Specific DO-178C objective reference (e.g., A-3.1)"

    - name: tool_qualification
      type: enum
      values: [TQL-1, TQL-2, TQL-3, TQL-4, TQL-5, N/A]
      required: false
      description: "Tool qualification level if tool-qualified test"

  health_checks:
    - name: dal_coverage
      description: "Every DAL A/B requirement must have independence_level set"
      rule: |
        if dal_level in ["A", "B"] then independence_level != ""

    - name: verification_complete
      description: "COMPLETE requirements must have verification_method set"
      rule: |
        if status == "COMPLETE" then verification_method != ""

  groups:
    - name: certification
      columns: [dal_level, verification_method, independence_level, sw_level_objective]
      description: "Certification-related fields"

    - name: tooling
      columns: [tool_qualification]
      description: "Tool qualification fields"
```

### Configuration

```yaml
# rtmx.yaml or .rtmx/config.yaml
rtmx:
  database: .rtmx/database.csv
  schema: core
  schema_plugins:
    - do178c                    # looked up in .rtmx/schemas/
    - ./vendor/iso26262.yaml    # explicit path
    - rtmx-schema-fedramp       # future: installable package
```

### Plugin Resolution

Schema plugins are resolved in order:
1. `.rtmx/schemas/<name>.yaml` (project-local)
2. `~/.config/rtmx/schemas/<name>.yaml` (user-global)
3. Explicit file path (relative or absolute)
4. Future: installable packages via `rtmx plugin install <name>`

### Go Implementation

```go
// internal/schema/schema.go

// ColumnType defines the type of a schema column.
type ColumnType string

const (
    ColumnString  ColumnType = "string"
    ColumnEnum    ColumnType = "enum"
    ColumnBool    ColumnType = "bool"
    ColumnFloat   ColumnType = "float"
    ColumnInteger ColumnType = "integer"
    ColumnDate    ColumnType = "date"
)

// ColumnDef defines a single column in a schema plugin.
type ColumnDef struct {
    Name        string     `yaml:"name"`
    Type        ColumnType `yaml:"type"`
    Values      []string   `yaml:"values,omitempty"`  // for enum type
    Required    bool       `yaml:"required"`
    Description string     `yaml:"description"`
    Display     DisplayOpts `yaml:"display,omitempty"`
}

// SchemaPlugin defines a loadable schema extension.
type SchemaPlugin struct {
    Name        string       `yaml:"name"`
    Version     string       `yaml:"version"`
    Description string       `yaml:"description"`
    Extends     string       `yaml:"extends"`
    Columns     []ColumnDef  `yaml:"columns"`
    HealthChecks []HealthRule `yaml:"health_checks,omitempty"`
    Groups      []ColumnGroup `yaml:"groups,omitempty"`
}

// Registry holds loaded schema plugins.
type Registry struct {
    plugins map[string]*SchemaPlugin
}

// Validate checks a requirement against all loaded schema plugins.
func (r *Registry) Validate(req *database.Requirement) []ValidationError
```

### Integration Points

1. **Database loading** (`internal/database/csv.go`):
   - After loading, validate `Extra` fields against loaded schema plugins
   - Type-check enum values, required fields, date formats

2. **Health checks** (`internal/cmd/health.go`):
   - Schema-aware health rules run alongside existing checks
   - Report violations per-requirement with schema source attribution

3. **Status/backlog display** (`internal/cmd/status.go`, `backlog.go`):
   - Schema columns with `display.show_in_status: true` appear in status output
   - Columns grouped by schema group for organized display

4. **Init/setup** (`internal/cmd/init.go`, `setup.go`):
   - `rtmx init --schema do178c` creates database with extended columns
   - `rtmx setup` detects schema plugins and validates configuration

5. **Documentation** (`internal/cmd/docs.go`):
   - `rtmx docs --schema` generates column documentation from loaded plugins

6. **JSON output**:
   - `--json` output includes schema plugin columns as first-class fields
   - Schema metadata included in `rtmx config --json`

### CLI Commands

```bash
rtmx schema list                # List available schema plugins
rtmx schema info do178c         # Show schema details and column definitions
rtmx schema validate            # Validate database against loaded schemas
rtmx schema init do178c         # Create .rtmx/schemas/do178c.yaml template
```

### Phased Delivery

**Phase 1 (REQ-PLUGIN-005)**: Schema definition, loading, validation, health integration.
This is the framework. No built-in domain schemas ship with core.

**Phase 2 (REQ-PLUGIN-006)**: Built-in schemas: phoenix (already documented),
do178c, iso26262. These ship as `.yaml` files in the rtmx distribution.

**Phase 3 (REQ-PLUGIN-007)**: `rtmx plugin install` for community schemas.
Package registry, versioning, dependency resolution.

## Acceptance Criteria

1. Schema plugins are defined as YAML files with typed column definitions
2. `rtmx.yaml` `schema_plugins` field loads one or more schema plugins
3. `rtmx health` reports schema validation errors (missing required fields,
   invalid enum values, type mismatches)
4. `rtmx schema list` shows available plugins with source location
5. `rtmx schema validate` checks all requirements against loaded schemas
6. Custom columns in the CSV are validated against their schema definitions
7. Columns without a schema definition are preserved (backwards compatible)
8. `rtmx schema info <name>` displays column definitions and health rules
9. Schema plugin loading fails gracefully with clear error messages
10. Existing `schema: core | phoenix` config remains supported
11. Extra field round-trip is preserved (no data loss)
12. All new code has >90% test coverage

## Files to Create/Modify

- `internal/schema/schema.go` -- Schema plugin types and registry
- `internal/schema/loader.go` -- YAML loading and resolution
- `internal/schema/validator.go` -- Requirement validation against schemas
- `internal/schema/schema_test.go` -- Unit tests
- `internal/cmd/schema.go` -- CLI commands (list, info, validate, init)
- `internal/cmd/schema_test.go` -- Command tests
- `internal/cmd/health.go` -- Integrate schema validation into health checks
- `internal/database/csv.go` -- Hook schema validation into database loading
- `internal/config/config.go` -- Add schema_plugins config field
- `docs/schema.md` -- Update with plugin framework documentation

## Effort Estimate

3 weeks (schema types + loader + validator + CLI + health integration + tests)

## Dependencies and Blocking

- **Depends on**: REQ-GO-008 (config management -- already COMPLETE)
- **Blocks**: REQ-PLUGIN-006 (built-in domain schemas), REQ-PLUGIN-007 (plugin install)
- **Related**: Phoenix schema documentation in docs/schema.md describes
  the column set that would become the first built-in schema plugin
