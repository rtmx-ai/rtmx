# REQ-BDD-004: Living Documentation Generation

## Status: MISSING
## Priority: MEDIUM
## Phase: 18

## Description
System shall generate living documentation from Gherkin feature files, incorporating real-time test execution status to produce human-readable requirement and acceptance documentation. The documentation generator shall use Jinja2 templates for customization and support watch mode for continuous regeneration during development.

## Acceptance Criteria
- [ ] Generate Markdown documentation from parsed feature files
- [ ] Include Feature name, description, and tags in generated output
- [ ] Include Scenario name, steps, and requirement links in generated output
- [ ] Embed test execution status (PASS/FAIL/PENDING) per Scenario
- [ ] Support Jinja2 template customization for output formatting
- [ ] Provide default templates: markdown, HTML, and reStructuredText
- [ ] `rtmx docs generate --features <path> --output <dir>` CLI command
- [ ] Watch mode: `rtmx docs generate --watch` regenerates on file changes
- [ ] Include data tables and doc strings in generated documentation
- [ ] Generate table of contents with anchor links
- [ ] Aggregate documentation across multiple feature file directories
- [ ] Include requirement traceability matrix in generated output

## Technical Notes
- Use Jinja2 for template rendering with custom filters for Gherkin structures
- Template variables: `feature`, `scenarios`, `steps`, `tags`, `test_status`, `requirements`
- Watch mode implementation: use `watchdog` library for file system monitoring
- Test status integration: read from `rtmx status` or CI test result artifacts
- Default template locations: `~/.rtmx/templates/` and `<project>/.rtmx/templates/`
- Consider generating per-feature files and an aggregated index file
- HTML template can include collapsible sections for large feature files
- reStructuredText output enables integration with Sphinx documentation

## Test Cases
1. `tests/test_bdd_docs.py::test_generate_markdown_from_feature` - Basic markdown generation
2. `tests/test_bdd_docs.py::test_include_feature_metadata` - Feature name/description/tags
3. `tests/test_bdd_docs.py::test_include_scenario_details` - Scenario steps and links
4. `tests/test_bdd_docs.py::test_embed_test_status` - PASS/FAIL/PENDING status
5. `tests/test_bdd_docs.py::test_jinja2_custom_template` - Custom template rendering
6. `tests/test_bdd_docs.py::test_default_html_template` - HTML output format
7. `tests/test_bdd_docs.py::test_default_rst_template` - reStructuredText output
8. `tests/test_bdd_docs.py::test_docs_generate_cli` - CLI command execution
9. `tests/test_bdd_docs.py::test_watch_mode_regeneration` - File change detection
10. `tests/test_bdd_docs.py::test_data_tables_in_docs` - Render step data tables
11. `tests/test_bdd_docs.py::test_aggregate_multiple_directories` - Multi-directory docs
12. `tests/test_bdd_docs.py::test_traceability_matrix_output` - RTM integration

## Dependencies
- REQ-BDD-001: Gherkin parser for feature files
- REQ-BDD-003: Feature file to requirement linking

## Blocks
- None

## Effort
2.0 weeks
