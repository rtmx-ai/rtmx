# REQ-DX-005: Git Hook Integration

## Status: NOT_STARTED
## Priority: LOW
## Phase: 4
## Effort: 0.5 weeks

## Description

Add optional git hooks for automated validation on commit and push.

## Acceptance Criteria

- [ ] `rtmx install --hooks` installs pre-commit hook
- [ ] Pre-commit hook runs `rtmx health --strict`
- [ ] Hook fails commit if health check fails
- [ ] `rtmx install --hooks --pre-push` adds marker compliance check
- [ ] Hooks are shell scripts, not Python (faster)
- [ ] Uninstall via `rtmx install --hooks --remove`

## Usage

```bash
# Install pre-commit hook
rtmx install --hooks

# Install both pre-commit and pre-push hooks
rtmx install --hooks --pre-push

# Remove all rtmx hooks
rtmx install --hooks --remove

# Preview hook content without installing
rtmx install --hooks --dry-run
```

## Pre-commit Hook Content

```bash
#!/bin/sh
# RTMX pre-commit hook
# Installed by: rtmx install --hooks

echo "Running RTMX health check..."
rtmx health --strict
if [ $? -ne 0 ]; then
    echo "RTMX health check failed. Commit aborted."
    echo "Run 'rtmx health' for details, or commit with --no-verify to skip."
    exit 1
fi
```

## Pre-push Hook Content

```bash
#!/bin/sh
# RTMX pre-push hook
# Installed by: rtmx install --hooks --pre-push

echo "Checking test marker compliance..."
pytest tests/ --collect-only -m req 2>/dev/null | grep -c "::test_" > /tmp/rtmx_with_req
pytest tests/ --collect-only 2>/dev/null | grep -c "::test_" > /tmp/rtmx_total

WITH_REQ=$(cat /tmp/rtmx_with_req)
TOTAL=$(cat /tmp/rtmx_total)

if [ "$TOTAL" -gt 0 ]; then
    PCT=$((WITH_REQ * 100 / TOTAL))
    if [ "$PCT" -lt 80 ]; then
        echo "Test marker compliance is ${PCT}% (requires 80%)."
        echo "Push aborted. Add @pytest.mark.req() markers to tests."
        exit 1
    fi
fi
```

## Files to Modify

- `src/rtmx/cli/install.py` - Add --hooks option

## Dependencies

- REQ-DX-001 (assumes .rtmx/ structure)

## Notes

Hooks should be fast and have minimal dependencies. Shell scripts avoid Python startup overhead.
