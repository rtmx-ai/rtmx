# ADR-006: npm Package Distribution Trade Study

## Status

Rejected

## Context

RTMX is currently distributed exclusively via PyPI as a Python package (`pip install rtmx`). This trade study evaluates whether to pursue npm package distribution to serve JavaScript/TypeScript ecosystems, enabling:

- Direct installation via `npm install rtmx`
- Integration with Node.js-based toolchains (Webpack, Vite, ESBuild)
- Compatibility with JavaScript-first CI/CD systems (GitHub Actions with Node)
- Broader accessibility for web developers and frontend teams

### RTMX Dependencies and Complexity

The current RTMX implementation has significant Python ecosystem dependencies:
- **pandas** (>=2.0): Data manipulation and CSV processing
- **pydantic** (>=2.0): Data validation and type coercion
- **click** (>=8.0): CLI framework with rich argument parsing
- **rich** (>=13.0): Terminal formatting and colors
- **tabulate** (>=0.9): Table rendering
- **pyyaml** (>=6.0): Configuration file parsing

Core functionality includes:
- CSV-based requirements database loading/saving
- Dependency graph analysis (cycle detection, critical path)
- Pytest plugin integration
- Adapters for GitHub, Jira, and MCP

## Decision

**We will NOT pursue npm package distribution at this time.**

The trade study concludes that none of the evaluated approaches provide an acceptable balance of effort, maintenance burden, and user experience for the current project maturity level (v0.0.2).

## Evaluation Matrix

| Approach | No Python Dep (25%) | Bundle Size <10MB (15%) | Startup <500ms (20%) | Feature Parity (25%) | Maintenance (15%) | Weighted Score |
|----------|---------------------|-------------------------|----------------------|----------------------|-------------------|----------------|
| **1. Node.js Wrapper** | 0 (requires Python) | 5 (tiny wrapper) | 2 (spawn overhead) | 5 (full parity) | 4 (low overhead) | 2.65 |
| **2. Pyodide/PyScript** | 5 (self-contained) | 0 (40MB+ base) | 0 (2-5s cold start) | 4 (most features) | 3 (version sync) | 2.45 |
| **3. Brython** | 5 (transpiled) | 2 (10-15MB) | 2 (transpile overhead) | 1 (limited stdlib) | 2 (compatibility) | 2.35 |
| **4. Manual JS Port** | 5 (native JS) | 5 (optimized) | 5 (<100ms) | 3 (feature subset) | 0 (dual codebase) | 3.20 |
| **5. py2wasm** | 5 (compiled) | 3 (varies) | 3 (wasm load) | 2 (limited support) | 1 (immature tools) | 2.70 |
| **6. python-shell** | 0 (requires Python) | 5 (tiny) | 2 (spawn) | 5 (full) | 4 (low) | 2.65 |

**Scoring: 0=Fail, 1=Poor, 2=Fair, 3=Good, 4=Very Good, 5=Excellent**

### Detailed Evaluation

#### 1. Native Python with Node.js Wrapper

**Mechanism**: npm package spawns `rtmx` CLI via child_process.

```javascript
// Conceptual implementation
const { spawn } = require('child_process');

function rtmx(args) {
  return new Promise((resolve, reject) => {
    const proc = spawn('rtmx', args);
    // Handle stdout/stderr...
  });
}
```

**Pros**:
- Trivial implementation effort (1-2 days)
- 100% feature parity guaranteed
- Single source of truth for logic

**Cons**:
- Users must install Python 3.10+ and `pip install rtmx`
- Process spawn adds ~200-500ms latency per invocation
- Complex cross-platform path resolution
- Dependency installation not managed by npm

**Verdict**: Defeats the purpose of npm distribution. Users need Python anyway.

#### 2. PyScript/Pyodide (Browser-based Python)

**Mechanism**: Bundle Pyodide (Python WASM interpreter) with RTMX.

**Research Findings** (from [pyodide.org](https://pyodide.org/en/stable/usage/downloading-and-deploying.html)):
- Core bundle: ~12MB compressed, expands to ~30MB
- Full distribution with pandas: 40MB+
- Cold start: 2-5 seconds
- Subsequent loads: ~2 seconds (cached)
- Filesystem access: Limited to virtual FS

**Pros**:
- No Python installation required
- Works in browser AND Node.js
- Active development (v0.29.0 as of 2025)

**Cons**:
- pandas alone adds ~15MB to bundle
- CLI startup time unacceptable (5+ seconds)
- No native filesystem access (critical for RTMX)
- Memory overhead significant (~100MB)

**Verdict**: Bundle size and startup time disqualify this approach for CLI tooling.

#### 3. Brython (Python-to-JS Transpilation)

**Mechanism**: Transpile Python source to JavaScript at build time.

**Research Findings** (from [brython.info](https://brython.info/)):
- brython.js: ~600KB
- brython_stdlib.js: ~5MB
- Transpilation adds runtime overhead
- Limited stdlib support (no pandas, pydantic unsupported)

**Pros**:
- Smaller than Pyodide
- No Python runtime required
- JavaScript-native execution

**Cons**:
- pandas not supported (critical dependency)
- pydantic not supported (validation layer)
- Poor startup for multi-module apps
- Active development concerns (7-year-old npm package)

**Verdict**: Core dependencies (pandas, pydantic) not supported. Non-starter.

#### 4. Manual JavaScript/TypeScript Port

**Mechanism**: Rewrite RTMX core in TypeScript.

```typescript
// Conceptual TypeScript implementation
interface Requirement {
  req_id: string;
  category: string;
  status: Status;
  // ...
}

class RTMDatabase {
  static async load(path: string): Promise<RTMDatabase> {
    const csv = await fs.readFile(path, 'utf-8');
    return this.parse(csv);
  }
}
```

**Effort Estimate**:
- Core models (Requirement, RTMDatabase): 2-3 days
- CSV parser with validation: 1-2 days
- Graph algorithms (cycles, critical path): 2-3 days
- CLI commands (15+ commands): 5-7 days
- Pytest plugin equivalent: N/A (Jest markers would differ)
- **Total: 15-20 days initial, ongoing maintenance**

**Pros**:
- Native performance (<100ms startup)
- Optimal bundle size (<1MB)
- Full npm ecosystem integration
- No runtime dependencies beyond Node.js

**Cons**:
- Significant development effort
- Dual codebase maintenance burden
- Feature drift risk (Python advances, JS lags)
- Pytest plugin has no direct equivalent
- Different dependency ecosystems (no pydantic equivalent)

**Verdict**: Highest quality result but unsustainable maintenance burden for a 0.0.2 project.

#### 5. WebAssembly Compilation (py2wasm)

**Mechanism**: Compile Python to WebAssembly using py2wasm/Nuitka.

**Research Findings** (from [wasmer.io](https://wasmer.io/posts/py2wasm-a-python-to-wasm-compiler)):
- Uses Nuitka to transpile Python to C, then to WASM
- Achieves ~70% of native Python speed
- Requires Python 3.11 environment for compilation
- Early-stage tooling (2024 release)

**Pros**:
- No Python runtime in distribution
- Better performance than interpreter-based approaches
- Single binary output

**Cons**:
- pandas compilation to WASM not well supported
- C extension modules (numpy, pandas) problematic
- Debugging extremely difficult
- Immature tooling, unstable APIs
- Large binary sizes for complex dependencies

**Verdict**: Tooling maturity insufficient. C extension dependencies (pandas) not reliably supported.

#### 6. python-shell / child_process Integration

**Mechanism**: npm package that manages Python venv automatically.

**Research Findings** (from [npm: python-shell](https://www.npmjs.com/package/python-shell)):
- Spawns Python with configurable options
- Supports text, JSON, and binary modes
- Built-in error handling for Python tracebacks

**Implementation Concept**:
```javascript
const { PythonShell } = require('python-shell');

// First run: setup venv and install rtmx
async function ensureRtmx() {
  // Check for .rtmx-venv, create if missing
  // pip install rtmx into venv
}

async function status() {
  await ensureRtmx();
  return PythonShell.run('rtmx', { args: ['status'] });
}
```

**Pros**:
- Full feature parity
- Automatic venv management possible
- Clean separation of concerns

**Cons**:
- Still requires Python installation
- First-run setup adds significant delay
- Platform-specific venv paths
- More complex than simple wrapper

**Verdict**: Improved UX over Approach 1, but still requires Python. Marginal benefit.

## Top 2 Candidates: Proof-of-Concept Notes

### Candidate A: Manual TypeScript Port (Highest Score: 3.20)

**Proof-of-Concept Scope** (if we were to proceed):

1. **Core Models** (`src/models.ts`):
   - Port `Requirement` interface with Zod validation (TypeScript's pydantic equivalent)
   - Port `RTMDatabase` class with CSV loading via `csv-parse`
   - Status/Priority enums

2. **CLI Framework** (`src/cli/`):
   - Use Commander.js (Click equivalent)
   - Implement `status`, `backlog`, `validate` commands only
   - Skip adapter integrations for PoC

3. **Testing**:
   - Jest with custom markers for requirement linking
   - Not a pytest plugin, but similar traceability

**Estimated PoC Effort**: 5 days for minimal viable feature set

**Key Technical Decisions**:
```typescript
// Zod for validation (pydantic equivalent)
import { z } from 'zod';

const RequirementSchema = z.object({
  req_id: z.string().regex(/^REQ-[A-Z]+-\d{3}$/),
  status: z.enum(['COMPLETE', 'PARTIAL', 'MISSING', 'NOT_STARTED']),
  // ...
});

// csv-parse for CSV handling (pandas equivalent)
import { parse } from 'csv-parse/sync';
```

**Why Not Proceed**:
- Maintenance burden disproportionate to user demand
- No evidence of significant JavaScript-only user base
- Python is standard in defense/aerospace compliance tooling
- Team expertise is Python-centric

### Candidate B: py2wasm (Score: 2.70)

**Proof-of-Concept Scope** (if we were to proceed):

1. **Minimal CLI**:
   - Strip pandas dependency (use stdlib csv module)
   - Remove pydantic (use dataclasses only)
   - Compile with py2wasm

2. **Test Execution**:
   ```bash
   pip install py2wasm
   py2wasm rtmx_minimal.py -o rtmx.wasm
   wasmer rtmx.wasm -- status
   ```

3. **Bundle Analysis**:
   - Measure WASM file size
   - Benchmark cold start time

**Estimated PoC Effort**: 3 days

**Why Not Proceed**:
- Requires removing pandas (breaks RTM report generation)
- Tooling still experimental
- No clear path to pydantic support
- Debugging production issues nearly impossible

## Consequences

### Positive

- **No maintenance burden**: Single Python codebase remains the source of truth
- **Focus on core value**: Development effort stays on feature development, not ports
- **Clear distribution story**: `pip install rtmx` is simple and well-understood
- **Preserved architecture**: No compromises to accommodate JavaScript limitations

### Negative

- **Limited reach**: JavaScript-only developers cannot use RTMX without Python
- **CI/CD friction**: Node.js-only CI environments require Python setup step
- **Perception**: May appear less modern without npm presence

### Mitigations

1. **Documentation**: Provide clear Python installation guides for Node.js developers
2. **GitHub Actions**: Publish action that sets up Python and RTMX for Node-based workflows
3. **Docker image**: Provide `ghcr.io/iotactical/rtmx:latest` for containerized usage
4. **MCP Server**: Leverage existing MCP adapter for AI assistant integration (language-agnostic)

## Future Reconsideration Triggers

This decision should be revisited if:

1. **Tooling matures**: py2wasm or similar gains pandas/pydantic support
2. **User demand**: >20% of feature requests mention npm/JavaScript
3. **Project scale**: RTMX reaches 1.0 with stable API worth porting
4. **Team capacity**: Additional contributors with TypeScript expertise join

## Implementation Roadmap

Since the decision is "Rejected," no implementation roadmap is provided.

If the decision were reversed, the recommended path would be:

1. **Phase 1** (v1.0+): TypeScript port of core models and validation
2. **Phase 2**: CLI with Commander.js, subset of commands
3. **Phase 3**: Full command parity, Jest traceability plugin
4. **Phase 4**: npm publish automation in CI/CD

## References

- [Pyodide Documentation](https://pyodide.org/en/stable/)
- [Brython Documentation](https://brython.info/)
- [py2wasm Announcement](https://wasmer.io/posts/py2wasm-a-python-to-wasm-compiler)
- [python-shell npm package](https://www.npmjs.com/package/python-shell)
- [Transcrypt Python-to-JS Compiler](https://www.transcrypt.org/)
- [REQ-DX-006 Specification](../requirements/DEVELOPER_EXPERIENCE/REQ-DX-006.md)
