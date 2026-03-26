# REQ-DIST-003: npm Wrapper Package

## Metadata
- **Category**: DIST
- **Subcategory**: npm
- **Priority**: MEDIUM
- **Phase**: 18
- **Status**: MISSING
- **Dependencies**: REQ-GO-043, REQ-LANG-006

## Requirement

RTMX shall be installable via npm, with the package downloading and installing the appropriate Go binary for the platform.

## Rationale

JavaScript developers expect to install tools via npm. A wrapper package provides seamless installation while leveraging the Go binary for actual CLI functionality.

## Design

### Installation

```bash
# Global install
npm install -g rtmx

# Or as dev dependency for project-specific version
npm install --save-dev rtmx
```

### Package Structure

```
rtmx/
├── package.json
├── bin/
│   └── rtmx              # Shell wrapper that finds Go binary
├── scripts/
│   ├── install.js        # postinstall: download Go binary
│   └── uninstall.js      # preuninstall: cleanup
├── lib/
│   └── binary.js         # Binary download/verification logic
└── README.md
```

### postinstall Script

```javascript
// scripts/install.js
const os = require('os');
const { downloadBinary } = require('../lib/binary');

const platform = os.platform();  // 'darwin', 'linux', 'win32'
const arch = os.arch();          // 'x64', 'arm64'

const binaryUrl = `https://github.com/rtmx-ai/rtmx/releases/download/v${version}/rtmx_${version}_${platform}_${arch}.tar.gz`;

await downloadBinary(binaryUrl);
```

### Binary Verification

```javascript
// Verify SHA256 checksum
const expected = checksums[`rtmx_${version}_${platform}_${arch}`];
const actual = crypto.createHash('sha256').update(binary).digest('hex');
if (actual !== expected) {
  throw new Error('Binary checksum mismatch');
}
```

### Execution

```bash
# When user runs: npx rtmx status
# bin/rtmx finds the downloaded Go binary and executes it
```

## Acceptance Criteria

1. `npm install -g rtmx` installs working CLI
2. `npx rtmx version` shows correct version
3. Works on macOS (Intel + Apple Silicon), Linux (x64 + arm64), Windows
4. Binary integrity verified via checksum
5. Graceful error if platform not supported
6. No runtime dependency on Go

## Test Strategy

- CI matrix testing all supported platforms
- npm pack + local install testing
- Version update testing

## Prior Art

- esbuild (Go binary distributed via npm)
- sass (Dart binary distributed via npm)
- @playwright/test (browser binaries downloaded on install)

## References

- npm postinstall scripts
- node-pre-gyp pattern
