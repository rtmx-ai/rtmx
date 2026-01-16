# REQ-BDD-003: Feature File to Requirement Linking

## Status: MISSING
## Priority: HIGH
## Phase: 18

## Description
System shall establish bidirectional traceability between Gherkin feature files and RTMX requirements through `@REQ-XXX-NNN` tag annotations. The linking system shall support tag inheritance rules, enable integration with `rtmx from-tests --features`, and maintain referential integrity between features and the RTM database.

## Acceptance Criteria
- [ ] Parse `@REQ-XXX-NNN` tags from Feature-level annotations
- [ ] Parse `@REQ-XXX-NNN` tags from Scenario-level annotations
- [ ] Parse `@REQ-XXX-NNN` tags from Scenario Outline-level annotations
- [ ] Implement tag inheritance: Feature tags propagate to all child Scenarios
- [ ] Support multiple requirement tags per Feature/Scenario (e.g., `@REQ-CORE-001 @REQ-API-002`)
- [ ] Validate requirement tag format against regex pattern `@REQ-[A-Z]+-\d{3}`
- [ ] Report orphan tags (tags referencing non-existent requirements in RTM database)
- [ ] Report orphan requirements (requirements with no linked feature files)
- [ ] `rtmx from-tests --features <path>` imports feature->requirement mappings into RTM
- [ ] Update RTM database with feature file paths and scenario names for linked requirements
- [ ] Generate requirement coverage report showing linked vs unlinked requirements
- [ ] Support tag aliasing for legacy tag formats (configurable mapping)

## Technical Notes
- Tag format: `@REQ-{CATEGORY}-{NUMBER}` where CATEGORY is uppercase letters, NUMBER is 3 digits
- Tag inheritance rules:
  - Feature-level tags apply to all Scenarios within the Feature
  - Rule-level tags (Gherkin 6+) apply to all Scenarios within the Rule
  - Scenario tags are additive to inherited tags
- Integration with RTM database: update `test_file`, `test_function` columns
- Orphan detection requires full RTM database scan
- Consider tag namespace prefixes for multi-project repositories
- Store feature->requirement mappings in RTM as JSON in `metadata` column

## Test Cases
1. `tests/test_bdd_linking.py::test_parse_feature_level_req_tags` - Extract feature tags
2. `tests/test_bdd_linking.py::test_parse_scenario_level_req_tags` - Extract scenario tags
3. `tests/test_bdd_linking.py::test_tag_inheritance_feature_to_scenario` - Validate inheritance
4. `tests/test_bdd_linking.py::test_multiple_requirement_tags` - Handle multiple tags
5. `tests/test_bdd_linking.py::test_validate_tag_format` - Regex validation
6. `tests/test_bdd_linking.py::test_detect_orphan_tags` - Find invalid tag references
7. `tests/test_bdd_linking.py::test_detect_orphan_requirements` - Find unlinked requirements
8. `tests/test_bdd_linking.py::test_from_tests_features_import` - CLI import command
9. `tests/test_bdd_linking.py::test_rtm_database_update` - Verify RTM column updates
10. `tests/test_bdd_linking.py::test_requirement_coverage_report` - Generate coverage stats
11. `tests/test_bdd_linking.py::test_tag_aliasing` - Legacy tag format mapping

## Dependencies
- REQ-BDD-001: Gherkin parser for feature files
- REQ-BDD-002: Step definition discovery across languages

## Blocks
- REQ-BDD-004: Living documentation generation

## Effort
2.5 weeks
