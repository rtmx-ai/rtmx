# Public rtmx-ai/rtmx sync (REQ-MONO-063e)

Marketplace listings, cursor.directory, and the MCP Registry crawl
**https://github.com/rtmx-ai/rtmx**. The monorepo vendors that tree under
`rtmx/`. Keeping them aligned is part of the distribution flywheel.

## Why not `make sync-push` / `git subtree push`?

Public `rtmx` **main** requires verified commit signatures. Subtree-split
commits are unsigned and GitHub rejects them (GH013).

## Canonical path: CI rsync → PR

Workflow: [`.github/workflows/publish-rtmx.yml`](../../.github/workflows/publish-rtmx.yml)  
Script: [`system/scripts/ci-sync-rtmx.sh`](../../system/scripts/ci-sync-rtmx.sh)

Triggers:

- Push to monorepo `main` touching `rtmx/**`
- `workflow_dispatch` (optional dry-run)

Behavior:

1. Shallow-clone `rtmx-ai/rtmx`
2. `rsync -a --delete` from monorepo `rtmx/` (excludes claims, workdir, bins)
3. Commit on branch `sync/mono-rtmx-<sha>`
4. Open/update a PR (human or admin squash-merge → verified merge on main)

### Secret

```bash
# Fine-grained PAT: Contents R/W + Pull requests R/W on rtmx-ai/rtmx
gh secret set RTMX_PUSH_TOKEN -R rtmx-ai/system
```

### Local dry-run

```bash
export RTMX_PUSH_TOKEN=…   # or GH_TOKEN with access
RTMX_SYNC_DRY_RUN=1 system/scripts/ci-sync-rtmx.sh
```

Break-glass direct main (needs signature bypass; prefer PR):

```bash
RTMX_SYNC_PUSH_MAIN=1 system/scripts/ci-sync-rtmx.sh
```

## After marketplace / plugin changes

1. Land changes on monorepo main.
2. Wait for **Publish rtmx** (or dispatch it).
3. Merge the opened `rtmx-ai/rtmx` PR when CI is green.
4. Confirm `plugins/cursor-rtmx` on public main before Marketplace refreshes.
