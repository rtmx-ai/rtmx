# Claude Code marketplace (first-party)

RTMX ships a first-party Claude marketplace so agents can install the plugin
**without waiting** on Anthropic’s official `external_plugins` review
(REQ-MONO-062e). Official listing remains amplification
([REQ-MONO-062f](./ANTHROPIC_OFFICIAL_SUBMIT.md)).

Marketplace manifest: [`.claude-plugin/marketplace.json`](../.claude-plugin/marketplace.json)  
Plugin package: [`plugins/claude-rtmx/`](../plugins/claude-rtmx/)

## Install (one-liner path)

```bash
# Add marketplace (GitHub — after rtmx repo publishes this tree)
claude plugin marketplace add rtmx-ai/rtmx

# Install plugin for the next Claude Code session
claude plugin install rtmx@rtmx
```

In a session: `/plugin marketplace add rtmx-ai/rtmx` then `/plugin install rtmx@rtmx`.

### Local / monorepo dogfood

```bash
claude plugin marketplace add /absolute/path/to/rtmx
claude plugin install rtmx@rtmx
claude plugin validate --strict /absolute/path/to/rtmx/plugins/claude-rtmx
```

## Prerequisite

`rtmx` on `PATH` (`brew install rtmx-ai/tap/rtmx`, scoop, or go install).
See `plugins/rtmx/skills/rtmx-prerequisite/`.

## Smoke

```bash
claude plugin marketplace add "$PWD"   # from rtmx/ checkout
claude plugin install rtmx@rtmx --scope user
claude plugin list | grep -i rtmx
```
