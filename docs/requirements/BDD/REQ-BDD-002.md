# REQ-BDD-002: Step Definition Discovery Across Languages

## Status: PARTIAL
## Priority: HIGH
## Phase: 18

## Description
System shall discover and index step definitions across multiple programming languages and BDD frameworks, enabling requirement traceability from Gherkin steps to their implementing code. The system shall support Python (behave, pytest-bdd), JavaScript (Cucumber.js), and Java (Cucumber-JVM) step definition patterns.

## Acceptance Criteria
- [ ] Discover Python step definitions using `@given`, `@when`, `@then` decorators from behave
- [ ] Discover Python step definitions using `@given`, `@when`, `@then` decorators from pytest-bdd
- [ ] Discover JavaScript step definitions using Cucumber.js `Given()`, `When()`, `Then()` functions
- [ ] Discover Java step definitions using Cucumber-JVM `@Given`, `@When`, `@Then` annotations
- [ ] Extract step definition regex/expression patterns for matching
- [ ] Match Gherkin step text to discovered step definitions using regex evaluation
- [ ] Report unimplemented steps (steps without matching definitions)
- [ ] Report step definition ambiguity (multiple definitions matching same step)
- [ ] Index step definition file locations and line numbers
- [ ] `rtmx discover-steps <path>` CLI command outputs discovered definitions as JSON
- [ ] Support for cucumber expressions in addition to regex patterns
- [ ] Handle parameterized step definitions with capture groups

## Technical Notes
- Python behave: Parse `@given(r"...")`, `@when(r"...")`, `@then(r"...")` decorators
- Python pytest-bdd: Parse `@given("...")`, `@when("...")`, `@then("...")` decorators
- JavaScript: Parse `Given(/.../)`, `When(/.../)`, `Then(/.../)` function calls
- Java: Parse `@Given("...")`, `@When("...")`, `@Then("...")` annotations
- Use tree-sitter parsers for robust AST-based extraction (leverage REQ-LANG-001)
- Cucumber expression format: `{int}`, `{float}`, `{string}`, `{word}`, custom types
- Step matching algorithm: compile regex, test against step text, extract parameters
- Consider step definition caching with file modification time invalidation
- Handle async step definitions in JavaScript (async/await patterns)

## Test Cases
1. `tests/test_bdd_steps.py::test_discover_behave_python_steps` - Parse behave decorators
2. `tests/test_bdd_steps.py::test_discover_pytest_bdd_steps` - Parse pytest-bdd decorators
3. `tests/test_bdd_steps.py::test_discover_cucumberjs_steps` - Parse Cucumber.js functions
4. `tests/test_bdd_steps.py::test_discover_cucumber_jvm_steps` - Parse Java annotations
5. `tests/test_bdd_steps.py::test_match_step_to_definition` - Regex matching algorithm
6. `tests/test_bdd_steps.py::test_report_unimplemented_steps` - Identify missing implementations
7. `tests/test_bdd_steps.py::test_report_ambiguous_steps` - Detect multiple matches
8. `tests/test_bdd_steps.py::test_extract_step_parameters` - Capture group extraction
9. `tests/test_bdd_steps.py::test_cucumber_expressions` - Parse cucumber expression syntax
10. `tests/test_bdd_steps.py::test_discover_steps_cli_output` - CLI JSON output format
11. `tests/test_bdd_steps.py::test_async_step_definitions` - Handle async JavaScript steps

## Dependencies
- REQ-BDD-001: Gherkin parser for feature files
- REQ-LANG-001: Language-specific plugin architecture (tree-sitter integration)

## Blocks
- REQ-BDD-003: Feature file to requirement linking
- REQ-BDD-006: Background/hooks inheritance

## Effort
3.0 weeks
