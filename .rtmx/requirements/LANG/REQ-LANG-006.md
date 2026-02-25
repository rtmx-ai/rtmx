# REQ-LANG-006: JavaScript/TypeScript Testing Integration

## Metadata
- **Category**: LANG
- **Subcategory**: JavaScript
- **Priority**: MEDIUM
- **Phase**: 18
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007

## Requirement

RTMX shall provide JavaScript and TypeScript testing integration supporting Jest, Mocha, and Vitest test frameworks.

## Rationale

JavaScript/TypeScript dominates web development and increasingly backend services. Supporting the major test frameworks enables RTMX adoption across the JS ecosystem.

## Design

### Installation

```bash
npm install --save-dev @rtmx/jest      # Jest integration
npm install --save-dev @rtmx/mocha     # Mocha integration
npm install --save-dev @rtmx/vitest    # Vitest integration
```

### Jest Integration

```typescript
import { req } from '@rtmx/jest';

describe('Authentication', () => {
  req('REQ-AUTH-001');

  it('should login successfully', () => {
    // test implementation
  });
});

// Or per-test
describe('Authentication', () => {
  it('should login successfully', () => {
    req('REQ-AUTH-001', { scope: 'integration' });
    // test implementation
  });
});
```

### Mocha Integration

```typescript
import { req } from '@rtmx/mocha';

describe('Authentication', function() {
  req(this, 'REQ-AUTH-001');

  it('should login successfully', function() {
    // test implementation
  });
});
```

### Vitest Integration

```typescript
import { req } from '@rtmx/vitest';
import { describe, it } from 'vitest';

describe('Authentication', () => {
  req('REQ-AUTH-001');

  it('should login successfully', () => {
    // test implementation
  });
});
```

### TypeScript Decorators (Experimental)

```typescript
import { Req } from '@rtmx/decorators';

class AuthTests {
  @Req('REQ-AUTH-001')
  testLoginSuccess() {
    // test implementation
  }
}
```

### Output Integration

```bash
# Jest
npx jest --reporters=default --reporters=@rtmx/jest/reporter

# Mocha
npx mocha --reporter @rtmx/mocha/reporter

# Or via rtmx
rtmx verify --command "npm test"
```

## Acceptance Criteria

1. Packages available on npm
2. Jest, Mocha, and Vitest integrations work
3. Test results output compatible JSON format
4. `rtmx verify --command "npm test"` correctly updates status
5. TypeScript types included
6. ESM and CommonJS module support

## Test Strategy

- Unit tests for each framework integration
- Integration tests with real test suites
- TypeScript compilation tests

## Package Structure

```
packages/
├── rtmx-jest/           # @rtmx/jest
│   ├── package.json
│   ├── src/
│   │   ├── index.ts
│   │   └── reporter.ts
│   └── tsconfig.json
├── rtmx-mocha/          # @rtmx/mocha
├── rtmx-vitest/         # @rtmx/vitest
└── rtmx-core/           # @rtmx/core (shared utilities)
```

## References

- Jest custom reporters
- Mocha reporters
- Vitest reporters
- REQ-LANG-007 marker specification
