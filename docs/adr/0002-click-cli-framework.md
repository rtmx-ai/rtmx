# ADR-0002: Click Framework for CLI

## Status

Accepted

## Context

RTMX is primarily a command-line tool requiring a robust CLI framework. The options considered were:

1. **argparse** - Python standard library
2. **Click** - Pallets project CLI framework
3. **Typer** - FastAPI author's CLI framework
4. **Fire** - Google's automatic CLI generation

## Decision

We chose **Click** as the CLI framework.

## Rationale

### Better User Experience
- Automatic help generation with rich formatting
- Consistent command structure with groups and subcommands
- Built-in support for environment variables and config files
- Color output and progress bars out of the box

### Composable Commands
- Commands can be organized into groups (`rtmx status`, `rtmx health`)
- Easy to add new commands without modifying existing code
- Supports command aliasing and abbreviation
- Context passing between commands

### Testing Support
- `CliRunner` enables isolated testing without subprocess
- Captures stdout/stderr for assertions
- Supports testing with mock environment variables

### Enterprise Ready
- Used by major projects (Flask, Ansible)
- Excellent documentation
- Active maintenance and community

## Alternatives Considered

### argparse
- **Pro**: No external dependencies
- **Con**: Verbose, poor help formatting
- **Con**: No built-in command grouping

### Typer
- **Pro**: Type hints as CLI definitions
- **Con**: Additional dependency on pydantic
- **Con**: Less mature ecosystem

### Fire
- **Pro**: Automatic CLI from functions
- **Con**: Less control over help text and options
- **Con**: Harder to customize

## Consequences

### Positive
- Rapid CLI development
- Consistent command experience
- Easier testing
- Good discoverability (`--help` everywhere)

### Negative
- External dependency (click package)
- Learning curve for decorators pattern
- Some limitations for very complex argument parsing

## Implementation Details

```python
# Command registration pattern
@main.command()
@click.option("--verbose", "-v", count=True)
def status(verbose: int) -> None:
    """Show RTM status."""
    from rtmx.cli.status import run_status
    run_status(verbose)
```

## References

- [Click Documentation](https://click.palletsprojects.com/)
- [Pallets Projects](https://palletsprojects.com/)
