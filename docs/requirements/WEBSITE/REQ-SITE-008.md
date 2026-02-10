# REQ-SITE-008: Website Patterns Page

## Requirement

Website shall have /patterns page generated from docs/patterns.md.

## Description

Create a /patterns page on rtmx.ai that displays the RTMX Patterns and Anti-Patterns content. This page should be auto-generated or synchronized from the canonical `docs/patterns.md` source in the rtmx repository to ensure consistency.

## Acceptance Criteria

- [ ] Page accessible at rtmx.ai/patterns
- [ ] Content matches docs/patterns.md from rtmx repo
- [ ] Styled consistently with rest of rtmx.ai
- [ ] Responsive design for mobile
- [ ] Navigation links in site sidebar
- [ ] SEO metadata for discoverability
- [ ] Anchor links for each section

## Implementation Options

1. **Submodule sync**: rtmx.ai already has rtmx as submodule; copy patterns.md during build
2. **Astro content collection**: Import markdown directly from submodule path
3. **Manual sync**: Copy content when it changes (not recommended)

## Recommended Approach

Use Astro's content collections to import directly from the rtmx submodule:

```javascript
// astro.config.mjs
import { defineConfig } from 'astro/config';

export default defineConfig({
  // ...
  vite: {
    resolve: {
      alias: {
        '@rtmx': './rtmx/docs'
      }
    }
  }
});
```

## Dependencies

- REQ-SITE-001: Website framework (Astro)
