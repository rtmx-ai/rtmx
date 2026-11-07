# First-session success metric (REQ-MONO-063b)

North Star after marketplace / deeplink / emit install:

> **First-session success** = within **30 minutes** of agent plugin (or MCP)
> install, the agent invokes MCP tools **`next`** and **`claim`** at least once
> each against a real project database.

This is the product pull for distribution BML (063): discovery → install →
useful agent loop, not merely “plugin listed.”

## Success event (definition)

| Field | Value |
|-------|--------|
| Event | Ordered pair: `tools/call` `next` **and** `tools/call` `claim` (any `agent_id`) |
| Window | 30 minutes from first successful MCP `initialize` after install |
| Unit | Per developer machine / project (not a global SaaS counter) |
| Privacy | No personal data required — counts of tool names only |

Related funnel: [DISCOVERY_FUNNEL.md](./DISCOVERY_FUNNEL.md) (063a).

## How to observe (offline-first)

Default remains **offline-first**. Prefer these observation modes in order:

1. **Local dogfood / CI smoke** — run the marketplace smoke or a short MCP
   stdio script that calls `next` then `claim` after install
   (`scripts/marketplace_smoke.sh`, tapes under `docs/tapes/`).
2. **Managed Sync (optional)** — when the project uses Sync, room claim
   activity already appears in Sync presence / audit without new CLI
   telemetry.
3. **Opt-in local log** — if enabled (see below), parse a local file for
   tool names. Never upload unless the user explicitly configures Sync or
   another remote.

## Opt-in measurement (off by default)

| Knob | Default | Behavior |
|------|---------|----------|
| CLI phone-home / install beacon | **Off / absent** | Forbidden as a default (063a non-goal) |
| Future `RTMX_SESSION_METRICS=1` (or config equivalent) | **Unset = off** | If implemented later, MAY append tool-name events to a **local** file under `.rtmx/` only; no network |

Until an opt-in knob ships, observation is manual dogfood, Sync (if used),
or aggregating anonymous site funnel clicks (063a) as a leading indicator —
not a substitute for the North Star.

## Usable without personal data

- Success is boolean (did next+claim occur in-window?), not who did it.
- Do not require email, GitHub login, or device fingerprinting for the metric
  definition.
- Aggregated site UTMs (063a) remain separate leading indicators.

## Linking to bootstrap (063f)

`rtmx bootstrap agents` (or equivalent) SHOULD optimize for this North Star:
one invocation that gets marketplace/emit + PATH reminder so an agent can
reach `next`/`claim` in minutes.
