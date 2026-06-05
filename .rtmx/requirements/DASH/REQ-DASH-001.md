# REQ-DASH-001: Embedded SPA with Client-Side Routing

## Metadata
- **Category**: DASH
- **Subcategory**: Framework
- **Priority**: P0
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-DASH-002, REQ-DASH-003, REQ-DASH-004, REQ-DASH-005, REQ-DASH-006, REQ-DASH-007, REQ-DASH-008, REQ-DASH-009

## Requirement

The `rtmx serve` command shall serve an embedded single-page application
(SPA) with client-side routing, replacing the current inline HTML string
with a modern web framework. The SPA and all its assets shall be compiled
into the Go binary via `embed.FS`, requiring no external files at runtime.

## Rationale

The current serve command renders a static HTML page with 4 status count
cards. A project management dashboard requires multiple views, navigation,
data tables, and interactive components that are not feasible with inline
HTML string templates. An embedded SPA maintains the single-binary
distribution advantage while providing a rich web experience.

## Design

### Technology Choice: htmx + Alpine.js + Tailwind CSS

Instead of a heavy Node.js build pipeline (React/Vue), the dashboard uses
a lightweight stack that embeds naturally into Go:

- **htmx** (14KB gzipped): Server-driven dynamic content via HTML attributes
- **Alpine.js** (15KB gzipped): Lightweight reactivity for client-side state
- **Tailwind CSS** (standalone CLI, tree-shaken): Utility-first styling

This stack avoids Node.js as a build dependency, keeps the binary small,
and produces HTML that Go's template engine can serve directly.

### Embedded Assets

```go
//go:embed dashboard/dist/*
var dashboardFS embed.FS

func NewDashboardMux(db *database.Database, cfg *config.Config) http.Handler {
    mux := http.NewServeMux()
    // Serve static assets
    mux.Handle("/assets/", http.FileServer(http.FS(dashboardFS)))
    // SPA fallback: serve index.html for all non-API routes
    mux.HandleFunc("/", serveSPA)
    // API routes
    mux.HandleFunc("/api/", serveAPI)
    return mux
}
```

### Client-Side Routing

The SPA uses hash-based routing (`/#/requirements`, `/#/backlog`, etc.)
to avoid server-side route configuration. The Go server always returns
index.html for non-API, non-asset paths.

### Page Structure

```
/#/                 -- Dashboard overview (status summary + charts)
/#/requirements     -- Requirements table (REQ-DASH-002)
/#/requirements/:id -- Requirement detail (REQ-DASH-003)
/#/graph            -- Dependency graph (REQ-DASH-004)
/#/backlog          -- Backlog board (REQ-DASH-005)
/#/releases         -- Release planning (REQ-DASH-006)
/#/health           -- Health dashboard (REQ-DASH-007)
/#/agents           -- Agent activity (REQ-DASH-008)
```

### Build Process

```makefile
# dashboard/Makefile
build:
    npx @tailwindcss/cli -i input.css -o dist/style.css --minify
    cp vendor/{htmx,alpine}.min.js dist/
    cp *.html dist/
```

The dashboard build runs as part of `make build` and produces static files
in `dashboard/dist/` that Go's embed directive includes.

## Acceptance Criteria

1. `rtmx serve` serves a full SPA with navigation between views.
2. All assets are embedded in the Go binary (no external files needed).
3. Client-side routing works for all defined routes.
4. Direct URL access to any route works (SPA fallback to index.html).
5. Assets are served with correct Content-Type headers and cache headers.
6. Total asset size < 100KB gzipped (htmx + Alpine + Tailwind + templates).
7. Dashboard loads in < 500ms on localhost.
8. No Node.js runtime required (build-time only for Tailwind).
9. Existing `/api/*` endpoints continue to function.

## Files to Create/Modify

- `dashboard/index.html` -- SPA shell with navigation
- `dashboard/input.css` -- Tailwind input file
- `dashboard/Makefile` -- Asset build pipeline
- `dashboard/vendor/` -- Vendored htmx and Alpine.js
- `internal/cmd/serve.go` -- Replace inline HTML with embedded SPA
- `internal/cmd/serve_embed.go` -- embed.FS declaration and SPA handler
- `Makefile` -- Add dashboard build step to `build` target

## Effort Estimate

1.5 weeks

## Test Strategy

- Embed test: verify all expected files present in embed.FS
- SPA fallback: non-API paths return index.html with 200
- API paths: return JSON (not index.html)
- Asset serving: correct Content-Type for .js, .css, .html
- Build test: `make build` produces working binary with embedded dashboard
- Binary size: verify total < 20MB (current ~12MB + 100KB dashboard)
