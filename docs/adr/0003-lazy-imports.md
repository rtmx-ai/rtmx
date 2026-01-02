# ADR-0003: Lazy Imports for CLI Startup Performance

## Status

Accepted

## Context

CLI tools should start quickly - users expect sub-second response for help text and simple commands. RTMX has several optional heavy dependencies:

- `PyGithub` for GitHub integration
- `jira` for Jira integration
- `mcp` for Model Context Protocol
- Various analysis libraries

Importing all modules at startup would significantly slow down the CLI.

## Decision

We use **lazy imports** - modules are only imported when their functionality is actually needed.

## Rationale

### Fast Startup
- `rtmx --help` responds in <100ms
- Simple commands don't pay for unused features
- Users notice and appreciate responsiveness

### Optional Dependencies
- Core features work without optional packages
- Clear error messages when features require missing packages
- Graceful degradation of functionality

### Reduced Memory
- Only load what's needed
- Better for systems with limited resources
- Important for CI/CD environments

## Implementation Pattern

```python
# In cli/main.py - the click command wrapper
@main.command()
def status() -> None:
    """Show RTM status."""
    # Import happens only when command is invoked
    from rtmx.cli.status import run_status
    run_status()

# In adapters - check for optional dependencies
def _get_client(self) -> Github:
    try:
        from github import Github
    except ImportError as e:
        raise ImportError(
            "PyGithub is required. Install with: pip install rtmx[github]"
        ) from e
```

## Consequences

### Positive
- Sub-100ms startup time for simple commands
- Optional features truly optional
- Users can install minimal package

### Negative
- Import errors appear at runtime, not startup
- Slightly more complex import structure
- Must test each command's imports

### Mitigations
- Clear error messages with installation instructions
- CI tests verify all imports work
- Documentation lists optional dependencies

## Measurements

| Command | With Lazy | Without Lazy |
|---------|-----------|--------------|
| `rtmx --help` | 80ms | 400ms |
| `rtmx status` | 150ms | 450ms |
| `rtmx sync --github` | 600ms | 700ms |

## References

- [Python Lazy Import Patterns](https://docs.python.org/3/library/importlib.html)
- [Click CLI Performance](https://click.palletsprojects.com/en/8.1.x/quickstart/)
