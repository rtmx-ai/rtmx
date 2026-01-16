# REQ-BDD-005: Scenario Outline Support

## Status: MISSING
## Priority: MEDIUM
## Phase: 18

## Description
System shall fully support Gherkin Scenario Outlines with Examples tables, enabling data-driven BDD testing. The parser shall expand Scenario Outlines into individual Scenarios for traceability, support multiple Examples tables per outline, and handle tagged Examples for selective test execution.

## Acceptance Criteria
- [ ] Parse Scenario Outline keyword and associated step templates
- [ ] Parse Examples tables with header row and data rows
- [ ] Support multiple Examples tables per Scenario Outline
- [ ] Expand Scenario Outline into individual Scenarios using Examples data
- [ ] Preserve placeholder syntax `<placeholder>` in step templates
- [ ] Substitute placeholders with Examples table values during expansion
- [ ] Support tagged Examples tables (e.g., `@smoke @critical Examples:`)
- [ ] Apply Example-level tags to expanded Scenarios
- [ ] Track original Scenario Outline line number for expanded Scenarios
- [ ] `rtmx parse-feature --expand-outlines` CLI flag for expansion output
- [ ] Handle complex data types in Examples: integers, floats, strings, booleans
- [ ] Support Examples table descriptions and comments

## Technical Notes
- Scenario Outline keyword aliases: "Scenario Template" (Gherkin i18n)
- Examples keyword aliases: "Scenarios" (Gherkin i18n)
- Placeholder format: `<column_name>` where column_name matches Examples header
- Expansion algorithm:
  1. Parse Scenario Outline steps with placeholders
  2. For each Examples row, create new Scenario with substituted values
  3. Generated Scenario name: `{outline_name} - {row_index}` or custom format
- Tagged Examples enable test filtering (e.g., run only `@smoke` examples)
- Multiple Examples tables useful for separating test categories (positive/negative)
- Consider lazy expansion for performance with large Examples tables
- Store expansion metadata for traceability back to original outline

## Test Cases
1. `tests/test_bdd_outline.py::test_parse_scenario_outline` - Basic outline parsing
2. `tests/test_bdd_outline.py::test_parse_examples_table` - Examples table extraction
3. `tests/test_bdd_outline.py::test_multiple_examples_tables` - Multiple tables per outline
4. `tests/test_bdd_outline.py::test_expand_outline_to_scenarios` - Expansion algorithm
5. `tests/test_bdd_outline.py::test_placeholder_substitution` - Value substitution
6. `tests/test_bdd_outline.py::test_tagged_examples` - Example-level tags
7. `tests/test_bdd_outline.py::test_tag_inheritance_to_expanded` - Tags on expanded scenarios
8. `tests/test_bdd_outline.py::test_preserve_line_numbers` - Traceability metadata
9. `tests/test_bdd_outline.py::test_expand_outlines_cli_flag` - CLI expansion flag
10. `tests/test_bdd_outline.py::test_complex_data_types` - Type handling in Examples
11. `tests/test_bdd_outline.py::test_examples_description` - Description/comment parsing
12. `tests/test_bdd_outline.py::test_i18n_scenario_template` - Internationalized keywords

## Dependencies
- REQ-BDD-001: Gherkin parser for feature files

## Blocks
- None

## Effort
1.5 weeks
