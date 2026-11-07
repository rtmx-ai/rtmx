# `rtmx bootstrap agents` (REQ-MONO-063f)

One-shot path toward first-session success (`next` + `claim` within minutes).
See [FIRST_SESSION_SUCCESS.md](./FIRST_SESSION_SUCCESS.md).

```bash
rtmx bootstrap agents                 # marketplace text + offline emit
rtmx bootstrap agents --dry-run       # preview only
rtmx bootstrap agents --emit=false    # guidance only (no writes)
rtmx bootstrap agents --agents-md     # optional AGENTS.md nudge
rtmx bootstrap agents --force         # overwrite emit outputs / AGENTS.md
```

## Behavior

1. Reminds that `rtmx` must be on `PATH` (`brew install rtmx-ai/tap/rtmx`).
2. Prints marketplace / deeplink / Codex install commands (user runs these online).
3. By default runs offline `plugin emit` for `claude,cursor,codex` (air-gapped OK).
4. Does not overwrite existing `AGENTS.md` unless `--force`.

Marketplace remains preferred when the network is available; emit is the
air-gapped / dogfood path (REQ-MONO-062p).

After agents can `next`/`claim`, multiplayer rooms use Managed Sync with the
same MCP pack — see [SYNC_MARKETPLACE_COHESION.md](./SYNC_MARKETPLACE_COHESION.md).
