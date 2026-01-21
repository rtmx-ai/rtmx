# Contributing to RTMX

Thank you for your interest in contributing to RTMX! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Python 3.10 or higher
- Git

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/rtm.git
   cd rtm
   ```

2. **Create a virtual environment**
   ```bash
   python -m venv .venv
   source .venv/bin/activate  # On Windows: .venv\Scripts\activate
   ```

3. **Install development dependencies**
   ```bash
   make dev
   # Or manually: pip install -e ".[dev]"
   ```

4. **Install pre-commit hooks**
   ```bash
   make pre-commit-install
   # Or manually: pre-commit install
   ```

5. **Verify setup**
   ```bash
   make check  # Runs lint, typecheck, and tests
   ```

## Code Style

We use automated tools to maintain consistent code style:

### Linting with Ruff

```bash
make lint       # Check for issues
make lint-fix   # Auto-fix issues
```

Ruff is configured with:
- pycodestyle (E, W)
- Pyflakes (F)
- isort imports (I)
- flake8-bugbear (B)
- flake8-comprehensions (C4)
- pyupgrade (UP)
- flake8-unused-arguments (ARG)
- flake8-simplify (SIM)

### Type Checking with MyPy

```bash
make typecheck
```

We use strict mode. All code must be fully typed.

### Formatting

```bash
make format        # Format code
make format-check  # Check formatting
```

## Testing

### Running Tests

```bash
make test          # Run all tests
make test-fast     # Run without coverage
make test-cov      # Run with detailed coverage report
```

### Writing Tests

All tests should be linked to requirements using the `@pytest.mark.req()` marker:

```python
import pytest

@pytest.mark.req("REQ-XX-NNN")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
def test_feature_works():
    """Test that feature works correctly."""
    assert feature() == expected
```

**Available markers:**

| Category | Markers |
|----------|---------|
| **Requirement** | `@pytest.mark.req("REQ-XX-NNN")` |
| **Scope** | `scope_unit`, `scope_integration`, `scope_system` |
| **Technique** | `technique_nominal`, `technique_parametric`, `technique_monte_carlo`, `technique_stress` |
| **Environment** | `env_simulation`, `env_hil`, `env_anechoic`, `env_static_field`, `env_dynamic_field` |

## Pull Request Process

### Before Submitting

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clear, concise commit messages
   - Add tests for new functionality
   - Update documentation if needed

3. **Run the full check suite**
   ```bash
   make check
   ```

4. **Ensure all tests pass**
   ```bash
   make test
   ```

### Submitting

1. Push your branch and create a Pull Request
2. Fill out the PR template completely
3. Link any related issues
4. Wait for CI to pass
5. Request review from maintainers

### PR Requirements

- [ ] All CI checks pass (tests, lint, typecheck)
- [ ] New code has test coverage
- [ ] Documentation updated if applicable
- [ ] Changelog entry added for user-facing changes

## Commit Messages

Use clear, descriptive commit messages:

```
feat: Add support for Jira integration

- Implement JiraAdapter for ticket sync
- Add rtmx sync jira command
- Update configuration schema

Closes #42
```

**Prefixes:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `chore:` - Build/tooling changes

## Reporting Issues

### Bug Reports

Please include:
- Python version and OS
- rtmx version (`rtmx --version`)
- Steps to reproduce
- Expected vs actual behavior
- Error messages/tracebacks

### Feature Requests

Please describe:
- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

## Questions?

- Open a [GitHub Discussion](https://github.com/rtmx-ai/rtmx/discussions)
- Check existing [Issues](https://github.com/rtmx-ai/rtmx/issues)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
