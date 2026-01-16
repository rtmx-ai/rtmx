# REQ-GIT-004: Git attributes configuration

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 15
## Effort: 0.5 weeks

## Description

Provide automated configuration of Git attributes and merge driver settings to enable the RTMX custom merge driver for RTM CSV files. The `rtmx install --git-merge` command shall configure both `.gitattributes` and local Git config.

## Acceptance Criteria

- [ ] `rtmx install --git-merge` configures merge driver in `.git/config`
- [ ] Command adds/updates `.gitattributes` with RTM file pattern
- [ ] Merge driver name is `rtmx` for consistency
- [ ] Configuration supports custom RTM file paths/patterns
- [ ] `--dry-run` flag shows what would be configured
- [ ] `--remove` flag removes RTMX merge driver configuration
- [ ] Existing `.gitattributes` entries are preserved
- [ ] Works with both local and global Git config
- [ ] Provides clear feedback on configuration status

## Test Cases

- `tests/test_git_config.py::test_install_git_merge_driver` - Basic installation
- `tests/test_git_config.py::test_gitattributes_created` - Creates .gitattributes
- `tests/test_git_config.py::test_gitattributes_updated` - Updates existing file
- `tests/test_git_config.py::test_custom_pattern` - Custom file pattern
- `tests/test_git_config.py::test_dry_run` - Dry run output
- `tests/test_git_config.py::test_remove_config` - Clean removal

## Technical Notes

### Git Configuration

The merge driver is configured in `.git/config`:

```ini
[merge "rtmx"]
    name = RTMX Requirements Traceability Matrix merge driver
    driver = rtmx merge-driver %O %A %B %L %P
    recursive = binary
```

### Git Attributes

The `.gitattributes` file maps file patterns to the merge driver:

```gitattributes
# RTMX merge driver for RTM CSV files
docs/rtm_database.csv merge=rtmx
**/rtm_*.csv merge=rtmx
```

### Command Interface

```bash
# Install merge driver (local config + .gitattributes)
rtmx install --git-merge

# Install with custom pattern
rtmx install --git-merge --pattern "requirements/*.csv"

# Install to global Git config
rtmx install --git-merge --global

# Preview changes
rtmx install --git-merge --dry-run

# Remove configuration
rtmx install --git-merge --remove

# Check current configuration
rtmx install --git-merge --status
```

### Default File Patterns

The default patterns to configure:
1. `docs/rtm_database.csv` - Standard RTMX location
2. `**/rtm_*.csv` - Any RTM-prefixed CSV
3. `**/*_requirements.csv` - Common naming convention

### Gitattributes Editing

When updating `.gitattributes`:
1. Read existing file content
2. Remove any existing RTMX merge entries
3. Add new RTMX block with comment header
4. Preserve all other entries

```python
def update_gitattributes(path: Path, patterns: list[str]) -> None:
    content = path.read_text() if path.exists() else ""

    # Remove existing RTMX block
    content = re.sub(r'# RTMX merge driver.*?(?=\n[^#]|\Z)', '', content, flags=re.DOTALL)

    # Add new block
    rtmx_block = "# RTMX merge driver for RTM CSV files\n"
    rtmx_block += "\n".join(f"{p} merge=rtmx" for p in patterns)
    rtmx_block += "\n"

    path.write_text(content.strip() + "\n\n" + rtmx_block)
```

## Files to Create/Modify

- `src/rtmx/cli/install.py` - Add `--git-merge` option
- `src/rtmx/git_config.py` - Git configuration management
- `tests/test_git_config.py` - Configuration tests

## Dependencies

- REQ-GIT-001: Custom merge driver (driver command must exist)

## Blocks

- None
