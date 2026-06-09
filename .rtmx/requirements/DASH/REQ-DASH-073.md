# REQ-DASH-073: Treemap Drill-Down

## Summary

Clicking a category group rectangle in the treemap zooms into that category, expanding it to fill the entire treemap viewport and showing only its child requirements at full detail. A breadcrumb navigation bar appears at the top showing "All > Category Name", allowing the user to return to the full treemap view. The zoom transition is animated using D3 transitions to provide spatial continuity.

## Acceptance Criteria

1. Clicking a category group rectangle (the outer rectangle, not an individual requirement) triggers a drill-down into that category.
2. The drill-down re-renders the treemap using only the requirements within the selected category, filling the full SVG viewport.
3. A breadcrumb bar renders above the treemap SVG showing the navigation path: "All > {Category Name}".
4. Clicking "All" in the breadcrumb returns to the full treemap view showing all categories.
5. The zoom-in transition animates over 500ms: the selected category rectangle expands to fill the viewport while other rectangles fade out.
6. The zoom-out transition animates over 500ms: the drilled-in view shrinks back to the category position while other categories fade in.
7. During drill-down, the requirement rectangles within the category are re-laid out by D3 treemap to use the full available space, not just scaled up from their previous positions.
8. All treemap interactions (hover tooltip, click to detail, color modes) continue to work within the drilled-down view.
9. Drill-down state is stored in an Alpine.js reactive property. It is not persisted in the URL.
10. If the active category has only one requirement, drill-down still works and shows that single requirement filling the viewport.

## Dependencies

- REQ-DASH-071 (treemap view provides the base layout and rectangle structure)

## Blocks

(none)

## Files to Modify

- `internal/dashboard/templates/partials/treemap.html` (add breadcrumb bar, drill-down click handler)
- `internal/dashboard/static/app.js` (drill-down state management, D3 zoom transitions, re-layout logic)

## Test Strategy

- Unit test: drill-down filter correctly isolates requirements for a single category
- Unit test: breadcrumb state updates to show category name on drill-down and resets on zoom-out
- Integration test: click a category rectangle, verify treemap re-renders showing only that category's requirements
- Integration test: verify breadcrumb shows "All > {Category}" during drill-down
- Integration test: click "All" in breadcrumb, verify full treemap restores
- Integration test: verify zoom-in animation executes (rectangle transforms change over time)
- Integration test: color mode changes work within drilled-down view

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32
