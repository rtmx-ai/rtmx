# REQ-DIST-001: TypeScript/npm Port

## Status: MISSING
## Priority: MEDIUM
## Phase: 12

## Description
TypeScript port shall provide npm package with core rtmx functionality for JavaScript/TypeScript ecosystems.

## Acceptance Criteria
- [ ] Package published to npm as `@rtmx-ai/rtmx`
- [ ] TypeScript interfaces match Python Requirement model
- [ ] Zod schemas provide runtime validation
- [ ] CSV parser reads/writes rtm_database.csv
- [ ] YAML parser loads rtmx.yaml config
- [ ] Commander.js CLI provides status, backlog, health commands
- [ ] Graph algorithms (Tarjan's SCC, topological sort, critical path) implemented
- [ ] Vitest/Jest plugin provides requirement markers
- [ ] Documentation with usage examples

## Architecture

```
@rtmx-ai/rtmx/
├── src/
│   ├── models/
│   │   ├── requirement.ts      # Requirement interface + Zod schema
│   │   ├── database.ts         # RTMDatabase class
│   │   └── config.ts           # RTMXConfig from rtmx.yaml
│   ├── graph/
│   │   ├── dependency.ts       # Tarjan's SCC, topological sort
│   │   └── critical-path.ts    # Critical path analysis
│   ├── cli/
│   │   ├── index.ts            # Commander.js CLI
│   │   ├── status.ts
│   │   ├── backlog.ts
│   │   └── health.ts
│   ├── adapters/
│   │   ├── github.ts           # GitHub Issues sync
│   │   └── jira.ts             # Jira sync
│   └── testing/
│       └── vitest-plugin.ts    # Vitest/Jest requirement markers
├── package.json
└── tsconfig.json
```

## Technology Mapping

| Python | TypeScript |
|--------|------------|
| Pydantic models | Zod schemas + TypeScript interfaces |
| Click CLI | Commander.js |
| pytest plugin | Vitest/Jest plugin |
| csv module | papaparse |
| PyYAML | js-yaml |
| tabulate | cli-table3 |

## Test Cases
- `tests/test_distribution.py::test_npm_package`

## Notes
Enables rtmx usage in TypeScript/Next.js projects without Python runtime dependency.
Complements Python package rather than replacing it.

## References
- ADR-006: npm Distribution Trade Study (recommended Python-only for core, TypeScript for interop)
