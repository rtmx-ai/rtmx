# ADR-0005: Adapter Pattern for External Integrations

## Status

Accepted

## Context

RTMX needs to integrate with multiple external systems:

1. **GitHub** - Issue tracking, PRs, repository sync
2. **Jira** - Enterprise issue tracking
3. **MCP** - Model Context Protocol for AI assistants

Each integration has different:
- Authentication methods
- API structures
- Rate limiting
- Error handling requirements

## Decision

We use the **Adapter Pattern** with a common base class to standardize external integrations.

## Rationale

### Uniform Interface
- All adapters implement `sync()` method
- Consistent error handling across integrations
- Predictable behavior for CLI commands

### Isolation of Dependencies
- Each adapter encapsulates its SDK/library
- Optional dependencies don't affect core functionality
- Clear separation of concerns

### Testability
- Mock individual adapters without affecting others
- Integration tests can use fake implementations
- Base class provides testing utilities

## Implementation

### Base Adapter Class
```python
# rtmx/adapters/base.py
class BaseAdapter:
    """Base class for external service adapters."""

    def __init__(self, config: RTMXConfig):
        self.config = config

    def sync(self, db: RTMDatabase) -> SyncResult:
        """Sync requirements with external system."""
        raise NotImplementedError

    def validate_connection(self) -> bool:
        """Test connectivity to external system."""
        raise NotImplementedError
```

### Concrete Implementations
```python
# rtmx/adapters/github.py
class GitHubAdapter(BaseAdapter):
    def sync(self, db: RTMDatabase) -> SyncResult:
        from github import Github  # Lazy import
        # Implementation...

# rtmx/adapters/jira.py
class JiraAdapter(BaseAdapter):
    def sync(self, db: RTMDatabase) -> SyncResult:
        from jira import JIRA  # Lazy import
        # Implementation...
```

### Factory Pattern for Adapter Selection
```python
def get_adapter(adapter_type: str, config: RTMXConfig) -> BaseAdapter:
    adapters = {
        "github": GitHubAdapter,
        "jira": JiraAdapter,
        "mcp": MCPAdapter,
    }
    return adapters[adapter_type](config)
```

## Consequences

### Positive
- Easy to add new integrations (implement BaseAdapter)
- Consistent CLI commands across all integrations
- Each adapter independently testable
- Clear extension points for enterprise adapters

### Negative
- Some code duplication across adapters
- Base class may not fit all integration patterns
- Version compatibility across multiple SDKs

### Mitigations
- Shared utilities in base class for common operations
- Adapter-specific configuration sections
- CI tests for each supported SDK version

## References

- [Adapter Pattern (Gang of Four)](https://refactoring.guru/design-patterns/adapter)
- [PyGithub Documentation](https://pygithub.readthedocs.io/)
- [Python-Jira Documentation](https://jira.readthedocs.io/)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
