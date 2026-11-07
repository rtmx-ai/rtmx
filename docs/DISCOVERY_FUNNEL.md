# Discovery funnel instrumentation (REQ-MONO-063a)

Measure which discovery surfaces convert (marketplace, deeplink, directory,
docs) **without** default CLI phone-home.

## Privacy posture

| Surface | Allowed | Forbidden (v1) |
|---------|---------|----------------|
| rtmx.ai | Privacy-safe web analytics (Plausible or self-hosted equivalent); standard server logs | Ad pixels / third-party retargeting required for product |
| README / docs outbound links | Documented UTM (or campaign) query params | |
| Go CLI (`rtmx`) | — | **Default install telemetry / phone-home** (explicit non-goal) |

## UTM dictionary

Append these query params to marketplace / deeplink CTAs so Plausible (or
server logs on intermediate pages) can attribute clicks:

| Param | Values |
|-------|--------|
| `utm_source` | `rtmx_readme` · `rtmx_docs` · `rtmx_site` |
| `utm_medium` | `referral` · `docs` |
| `utm_campaign` | `agent_discovery` |
| `utm_content` | `cursor_deeplink` · `cursor_directory` · `github_fab` |

Example (README Cursor deeplink):

```
https://cursor.com/link/mcp?name=rtmx&config=…&utm_source=rtmx_readme&utm_medium=referral&utm_campaign=agent_discovery&utm_content=cursor_deeplink
```

Example (docs directory):

```
https://cursor.directory/plugins/rtmx?utm_source=rtmx_docs&utm_medium=docs&utm_campaign=agent_discovery&utm_content=cursor_directory
```

Keep launch `command`/`args` identical to the shared pack; UTMs are
attribution only.

## Privacy-safe analytics (measure path)

1. Create a Plausible (or self-hosted Plausible / Umami) site for `rtmx.ai`.
2. Set build env `PUBLIC_PLAUSIBLE_DOMAIN=rtmx.ai` (Pages / CI). When set,
   Starlight injects the deferred Plausible script from `astro.config.mjs`.
3. Enable outbound-link goals for Cursor / cursor.directory hosts if available.
4. Review funnel weekly alongside first-session success
   ([FIRST_SESSION_SUCCESS.md](./FIRST_SESSION_SUCCESS.md), REQ-MONO-063b).

Until the env var is set, this document is the measure path of record — CTAs
still carry UTMs for when analytics is enabled.

## Non-goal

The open-source CLI MUST NOT phone home by default. No install-id beacons,
no anonymous usage pings in `rtmx` without an explicit, documented opt-in
(out of scope for 063a).
