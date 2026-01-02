# REQ-DX-006: npm Package Distribution Trade Study

## Status: NOT_STARTED
## Priority: MEDIUM
## Phase: 4
## Effort: 0.5 weeks

## Description

Trade study shall evaluate npm package distribution options for rtmx.

## Background

Currently rtmx is distributed only via PyPI as a Python package. Users in JavaScript/TypeScript ecosystems may benefit from npm distribution, enabling:

- Direct installation via `npm install rtmx`
- Integration with Node.js-based toolchains
- Compatibility with JavaScript-first CI/CD systems
- Broader accessibility for web developers

## Approaches to Evaluate

### 1. Native Python with Node.js Wrapper

Create a thin Node.js wrapper that spawns the Python CLI.

**Pros:**
- Minimal code changes
- Python runtime handles all logic
- Easy to maintain parity

**Cons:**
- Requires Python installed on user system
- Slower startup (process spawn overhead)
- Complex dependency management

### 2. PyScript/Pyodide (Browser-based Python)

Bundle Python interpreter in WebAssembly for browser execution.

**Pros:**
- No Python installation required
- Works in browser environments
- Single distribution artifact

**Cons:**
- Large bundle size (~40MB+)
- Cold start latency
- Limited filesystem access
- Not suitable for CLI usage

### 3. Brython (Python-to-JS Transpilation)

Transpile Python to JavaScript at runtime or build time.

**Pros:**
- Native JavaScript execution
- No runtime dependency on Python
- Smaller bundle than Pyodide

**Cons:**
- Limited Python stdlib support
- Some Python features unsupported
- Runtime overhead for transpilation

### 4. Manual JavaScript Port

Rewrite rtmx core in TypeScript/JavaScript.

**Pros:**
- Native performance
- Full npm ecosystem integration
- No cross-language overhead

**Cons:**
- Significant development effort
- Maintenance burden (two codebases)
- Feature drift risk

### 5. WebAssembly Compilation (py2wasm)

Compile Python to WebAssembly using tools like Pyodide or Nuitka.

**Pros:**
- Near-native performance
- No Python runtime required
- Single binary distribution

**Cons:**
- Immature tooling
- Complex build process
- Debugging challenges

### 6. python-shell / child_process Integration

npm package that manages Python virtual environment automatically.

**Pros:**
- Full Python feature support
- Automatic venv management
- Clean separation of concerns

**Cons:**
- Still requires Python
- More complex installation
- Platform-specific considerations

## Evaluation Criteria

| Criterion | Weight |
|-----------|--------|
| No Python dependency required | 25% |
| Bundle size < 10MB | 15% |
| CLI startup time < 500ms | 20% |
| Feature parity with Python version | 25% |
| Maintenance overhead | 15% |

## Acceptance Criteria

- [ ] All 6 approaches evaluated against criteria
- [ ] Proof-of-concept for top 2 candidates
- [ ] ADR-006 documents final decision with rationale
- [ ] Go/no-go recommendation with justification
- [ ] If go: implementation roadmap included

## Deliverable

`docs/adr/ADR-006-npm-distribution.md` - Architecture Decision Record documenting the trade study findings and decision.

## Notes

The trade study should also consider:
- Target user personas (who would use npm vs pip?)
- CI/CD integration patterns in JS ecosystems
- Competitive analysis (do similar tools offer npm packages?)
- Long-term maintenance implications
