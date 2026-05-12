# Contributing to RTMX

Thank you for your interest in contributing to RTMX.

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Git
- golangci-lint v2 (`curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin`)

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/rtmx.git
   cd rtmx
   ```

2. **Install pre-commit hooks**
   ```bash
   make hooks
   ```
   This installs `.githooks/pre-commit` which runs build, test, lint, and
   vet before each commit. Pre-commit hooks must pass -- do not use
   `--no-verify`.

3. **Build and test**
   ```bash
   make build
   make test
   make lint
   ```

## Code Style

- `go fmt` and `goimports` for formatting (`make fmt`)
- golangci-lint v2 for static analysis (`make lint`)
- No CGO -- pure Go for static binary distribution
- Minimize dependencies -- 2 external deps (Cobra + YAML parser)
- All packages under `internal/` are unexported; public API lives in `pkg/rtmx/`

## Testing

### Running Tests

```bash
make test          # Full test suite with race detector and coverage
make test-short    # Short tests only
make ci            # Full local CI (build, test, coverage threshold, vet, lint, markers)
```

### Writing Tests

Every test must be linked to a requirement using `rtmx.Req()`:

```go
func TestFeature(t *testing.T) {
    rtmx.Req(t, "REQ-XX-NNN")
    // test body
}
```

Use table-driven tests. Use golden files for output formatting
(`testdata/`). Use interfaces for external dependencies (HTTP, filesystem,
environment) -- never call `os.Getenv` or `http.DefaultClient` directly.

### Coverage

CI enforces a minimum 80% coverage threshold. Individual packages in
`internal/adapters` and `internal/cmd` target 100%.

## Pull Request Process

### Branch Protection

The `main` branch is protected:
- Direct pushes are restricted to maintainers
- All external contributions must come via pull request
- CI (`build-and-test`) must pass before merge
- At least 1 approving review is required
- Stale reviews are dismissed on new pushes
- Force pushes and branch deletion are blocked

### Before Submitting

1. **Fork the repo and create a feature branch from `main`**
   ```bash
   git checkout -b feature/your-feature-name main
   ```

2. **Rebase on `main` before opening your PR**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```
   We require a linear history on feature branches. Merge commits in PRs
   will not be accepted -- rebase to resolve conflicts.

3. **Run the full CI suite locally**
   ```bash
   make ci
   ```
   This runs build, test with coverage threshold, vet, lint, and marker
   compliance. Your PR will not be reviewed until `make ci` passes locally.

4. **Link your work to a requirement**
   - If your change addresses an existing requirement, reference it in the
     commit message (`REQ-XX-NNN`)
   - If your change adds new functionality, open an issue first to discuss
     the requirement before writing code

### Submitting

1. Push your branch and create a Pull Request against `main`
2. Fill out the PR description: what changed, why, and how to test
3. Link any related issues
4. Wait for CI to pass
5. Address review feedback by pushing new commits (do not force-push during review)

### PR Requirements

- All CI checks pass (build, test, lint, vet)
- New code has test coverage with `rtmx.Req()` markers
- Tests are table-driven where applicable
- No `--no-verify` commits
- Rebased on current `main` (no merge commits)
- Documentation updated if applicable

## Commit Messages

Use conventional commit prefixes:

```
feat: add support for Jira integration

- Implement JiraAdapter for ticket sync
- Add rtmx sync jira command
- Update configuration schema

Closes #42
```

Prefixes: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`

## Reporting Issues

### Bug Reports

Include:
- Go version and OS (`go version`, `uname -a`)
- rtmx version (`rtmx --version`)
- Steps to reproduce
- Expected vs actual behavior
- Error messages

### Feature Requests

Describe:
- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Questions?

- [GitHub Discussions](https://github.com/rtmx-ai/rtmx/discussions)
- [Issues](https://github.com/rtmx-ai/rtmx/issues)
- Email: dev@rtmx.ai

## License

By contributing, you agree that your contributions will be licensed under
the Apache License 2.0.
