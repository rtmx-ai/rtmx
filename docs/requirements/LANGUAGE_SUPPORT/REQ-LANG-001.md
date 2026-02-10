# REQ-LANG-001: Jest/Vitest Plugin for JavaScript/TypeScript

## Status: MISSING
## Priority: HIGH
## Phase: 14

## Description
System shall provide Jest and Vitest plugins for JavaScript/TypeScript projects that enable requirement traceability through **native test framework APIs as the primary mechanism**, with JSDoc comments as a deprecated fallback for legacy codebases.

## Marker Strategy: Native APIs First

**Primary (Recommended)**: Use test framework extension APIs that provide first-class requirement tracking with IDE support, type safety, and runtime validation.

**Secondary (Deprecated Fallback)**: JSDoc comments for legacy codebases where modifying test code is impractical.

## Acceptance Criteria

### Native API Integration (PRIMARY)
- [ ] Vitest: Custom test function via `test.extend` pattern:
  ```typescript
  import { rtmxTest } from '@rtmx/vitest';

  const test = rtmxTest.extend({
    req: 'REQ-AUTH-001',
    scope: 'integration',
    technique: 'nominal',
  });

  test('login succeeds with valid credentials', async ({ req }) => {
    // Test implementation
  });
  ```
- [ ] Jest: Custom `describe.rtmx` and `test.rtmx` wrappers:
  ```typescript
  import { rtmx } from '@rtmx/jest';

  rtmx.describe('REQ-AUTH-001', { scope: 'integration' }, () => {
    rtmx.test('login succeeds', async () => {
      // Test implementation
    });
  });
  ```
- [ ] TypeScript types provide autocompletion for requirement IDs
- [ ] Runtime validation ensures requirement ID format matches `^REQ-[A-Z]+-[0-9]+$`
- [ ] Test metadata accessible in reporters for RTM output generation

### Companion Package Distribution
- [ ] npm package `@rtmx/vitest` published to npmjs.com
- [ ] npm package `@rtmx/jest` published to npmjs.com
- [ ] npm package `@rtmx/core` with shared types and utilities
- [ ] Zero production dependencies (devDependencies only for testing)

### Reporter Integration
- [ ] Custom Vitest reporter outputs RTM-compatible JSON results
- [ ] Custom Jest reporter outputs RTM-compatible JSON results
- [ ] Test results include requirement coverage mapping
- [ ] Reporter hooks: `onTestResult` (Jest), `onFinished` (Vitest)

### CLI Integration
- [ ] `rtmx from-jest <results.json>` imports Jest test results into RTM database
- [ ] `rtmx from-vitest <results.json>` imports Vitest test results into RTM database
- [ ] Marker extraction from source files via tree-sitter (for migration assistance)

### Legacy Fallback (DEPRECATED)
- [ ] JSDoc marker format: `/** @rtmx REQ-XXX-NNN */` parsed before test functions
- [ ] Inline marker format: `// rtmx:req=REQ-XXX-NNN` parsed within test bodies
- [ ] Deprecation warnings emitted when JSDoc markers detected
- [ ] Migration tool: `rtmx migrate-markers --from=comments --to=api` rewrites tests

## Technical Notes

### Why Native APIs Over Comments

| Aspect | Native API | Comments |
|--------|------------|----------|
| Type safety | Full TypeScript support | None |
| IDE support | Autocompletion, go-to-definition | None |
| Runtime validation | Validates at test execution | Requires separate parsing |
| Refactoring | Works with rename refactors | Breaks on refactors |
| Discoverability | Explicit in test signature | Hidden in comments |
| Maintenance | Single source of truth | Can drift from code |

### Vitest Extension Pattern

```typescript
// @rtmx/vitest/src/extend.ts
import { test as base, expect } from 'vitest';

interface RtmxFixtures {
  req: string;
  scope?: 'unit' | 'integration' | 'system' | 'acceptance';
  technique?: 'nominal' | 'parametric' | 'monte_carlo' | 'stress' | 'boundary';
  env?: 'simulation' | 'hil' | 'anechoic' | 'field';
}

export const rtmxTest = base.extend<RtmxFixtures>({
  req: ['', { option: true }],
  scope: [undefined, { option: true }],
  technique: [undefined, { option: true }],
  env: [undefined, { option: true }],
});

// Usage captures metadata in test context
rtmxTest.beforeEach(({ req, scope, technique, env }) => {
  // Register requirement association for reporter
});
```

### Jest Global Extension

```typescript
// @rtmx/jest/src/globals.ts
declare global {
  namespace jest {
    interface Rtmx {
      describe(req: string, opts: RtmxOptions, fn: () => void): void;
      test(name: string, fn: () => void | Promise<void>): void;
    }
  }
  const rtmx: jest.Rtmx;
}

// Implementation wraps describe/test with metadata injection
```

### Migration from Comments

```bash
# Analyze current comment-based markers
rtmx markers discover --format json > current-markers.json

# Generate migration script
rtmx migrate-markers --from=comments --to=api --dry-run

# Apply migration
rtmx migrate-markers --from=comments --to=api
```

### Support for Parameterized Tests
- `describe.each` / `test.each` patterns with requirement inheritance
- ESM and CommonJS module format support
- JSON output format aligned with RTMX marker specification (REQ-LANG-007)

## Test Cases
1. `tests/test_lang_jest.py::test_native_api_marker_extraction` - Parse rtmx.test() markers
2. `tests/test_lang_jest.py::test_vitest_extend_marker_extraction` - Parse test.extend() markers
3. `tests/test_lang_jest.py::test_extended_marker_attributes` - Parse scope/technique/env
4. `tests/test_lang_jest.py::test_jest_reporter_output` - Validate reporter JSON format
5. `tests/test_lang_jest.py::test_vitest_reporter_output` - Validate Vitest reporter format
6. `tests/test_lang_jest.py::test_parameterized_test_handling` - Test each() variants
7. `tests/test_lang_jest.py::test_typescript_type_definitions` - Validate .d.ts files
8. `tests/test_lang_jest.py::test_from_jest_import` - CLI import command
9. `tests/test_lang_jest.py::test_jsdoc_deprecation_warning` - Deprecation emitted for comments
10. `tests/test_lang_jest.py::test_migrate_markers_command` - Migration tool works

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec
- REQ-DIST-002: Standalone binary CLI (for marker extraction without Node.js)

## Blocks
- REQ-DIST-001: TypeScript port (uses same @rtmx packages)

## Effort
3.0 weeks
