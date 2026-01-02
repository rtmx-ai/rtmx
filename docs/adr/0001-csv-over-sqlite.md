# ADR-0001: CSV Database Format Over SQLite

## Status

Accepted

## Context

RTMX needs a persistent storage format for the Requirements Traceability Matrix (RTM) database. The options considered were:

1. **SQLite** - Embedded relational database
2. **CSV** - Comma-separated values flat file
3. **JSON** - JavaScript Object Notation file
4. **YAML** - Human-friendly data serialization

The primary use case is tracking software requirements across development teams, with heavy integration with version control systems (Git) and AI code assistants.

## Decision

We chose **CSV format** for the RTM database (`rtm_database.csv`).

## Rationale

### Human-Readable
- Developers can view and edit requirements directly in any text editor
- CSV renders as a table in GitHub/GitLab web interfaces
- No special tooling required for basic operations

### Git-Friendly
- Line-based format enables meaningful diffs
- Merge conflicts are easier to understand and resolve
- Each requirement is a single line, making blame/annotate useful
- History shows exactly which requirements changed in each commit

### AI-Parseable
- LLMs can read and understand CSV without specialized parsers
- Enables AI agents to directly interact with requirements
- Schema is self-documenting via header row
- Compatible with Model Context Protocol (MCP) tools

### Simplicity
- No database connections or drivers needed
- Works on all platforms without additional dependencies
- Backup is trivial (it's just a text file)
- Import/export with spreadsheet tools (Excel, Google Sheets)

## Alternatives Considered

### SQLite
- **Pro**: ACID transactions, complex queries
- **Con**: Binary format creates opaque diffs, merge conflicts are cryptic
- **Con**: Requires sqlite3 driver

### JSON/YAML
- **Pro**: Hierarchical data support
- **Con**: Multi-line records make diffs harder to read
- **Con**: Merge conflicts across nested structures are complex

## Consequences

### Positive
- Zero-configuration setup
- Works seamlessly with Git workflows
- Easy integration with AI/LLM tools
- Low barrier to adoption

### Negative
- No built-in query language (must parse in code)
- Limited data types (everything is strings)
- No referential integrity enforcement
- Large databases may have performance issues

### Mitigations
- Use Python dataclasses for type conversion
- Implement validation layer for referential integrity
- For very large RTMs (10,000+ requirements), consider SQLite export

## References

- [RFC 4180: Common Format and MIME Type for CSV Files](https://tools.ietf.org/html/rfc4180)
- [Git - Merge Strategy Options](https://git-scm.com/docs/merge-strategies)
