# REQ-DASH-007: Health Dashboard with Trend Charts

## Metadata
- **Category**: DASH
- **Subcategory**: View
- **Priority**: MEDIUM
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-API-001
- **Blocks**: (none)

## Requirement

The web dashboard shall provide a health dashboard view that displays
project health checks, completion trend charts (burndown and burnup),
and category-level health breakdowns with historical data.

## Rationale

Health monitoring and trend visualization answer "are we on track?" and
"where are we slowing down?". The CLI `rtmx health` command provides a
point-in-time snapshot but no historical trends. Charting completion over
time enables project managers to identify velocity changes and forecast
completion dates.

## Design

### Layout

```
+-- Project Health -------------------------------------------------------+
|                                                                          |
| HEALTH CHECKS                                                            |
| [PASS] Database loads successfully (241 requirements)                    |
| [PASS] No orphaned dependencies                                         |
| [PASS] Test coverage: 92.1%                                             |
| [WARN] 3 requirements stale > 30 days                                   |
|                                                                          |
| COMPLETION TREND                                                         |
| 100% |                                              ****                 |
|  80% |                                   ***********                     |
|  60% |                        ***********                                |
|  40% |             ***********                                           |
|  20% |  ***********                                                      |
|   0% +--+-----+-----+-----+-----+-----+-----+-----+                    |
|      Feb   Mar   Apr   May   Jun   Jul   Aug   Sep                       |
|                                                                          |
| CATEGORY HEALTH                                                          |
| CLI: 100%  MCP: 57.1%  TUI: 0%  DASH: 0%  ADAPT: 100%  ...             |
+--------------------------------------------------------------------------+
```

### Trend Data Collection

Completion history is derived from `completed_date` fields in the database.
For each date, the cumulative count of COMPLETE requirements is computed.
This provides a burnup chart without needing a separate time-series store.

### Charts

Uses Chart.js (lightweight, 70KB gzipped) or inline SVG for simple line
charts. The burnup chart plots cumulative completions over time. A burndown
overlay shows remaining work.

### Category Health Grid

A grid of category cards, each showing completion percentage with a
micro-progress bar. Categories below 50% are highlighted in red.

## Acceptance Criteria

1. Health checks display with PASS/FAIL/WARN indicators.
2. Completion trend chart plots cumulative completions over time.
3. Chart data is derived from completed_date fields (no external store).
4. Category health grid shows per-category completion percentages.
5. Categories below 50% are visually highlighted.
6. Chart renders within 500ms.
7. Tooltip on chart shows exact count at each date.
8. Date range is adjustable (last 30d, 90d, all time).

## Files to Create/Modify

- `dashboard/health.html` -- Health dashboard template
- `dashboard/js/health.js` -- Chart rendering and data transformation
- `dashboard/vendor/chart.min.js` -- Vendored Chart.js (or inline SVG)

## Effort Estimate

1 week

## Test Strategy

- Health checks: verify display matches CLI `rtmx health` output
- Trend calculation: verify cumulative completion counts from test dates
- Category grid: verify per-category percentages match API data
- Date range filter: verify chart updates with different ranges
- Edge case: no completed_date values produces empty chart (not error)
