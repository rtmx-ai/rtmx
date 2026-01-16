# REQ-LANG-002: JUnit Extension for Java/Kotlin

## Status: MISSING
## Priority: HIGH
## Phase: 14

## Description
System shall provide a JUnit 5 extension for Java and Kotlin projects that enables requirement traceability through custom annotations, with Maven and Gradle plugin support for seamless build integration.

## Acceptance Criteria
- [ ] Maven artifact `io.rtmx:rtmx-junit` published to Maven Central
- [ ] `@Req("REQ-XXX-NNN")` annotation marks tests with requirement IDs
- [ ] `@RtmxScope(Scope.UNIT)` annotation specifies test scope (UNIT, INTEGRATION, SYSTEM, ACCEPTANCE)
- [ ] `@RtmxTechnique(Technique.NOMINAL)` annotation specifies test technique
- [ ] `@RtmxEnv(Environment.SIMULATION)` annotation specifies test environment
- [ ] Composite `@RtmxTest` annotation combines all attributes in single annotation
- [ ] JUnit 5 `TestExecutionListener` captures requirement associations on test completion
- [ ] Maven plugin `rtmx-maven-plugin` with goal `rtmx:extract` generates RTM JSON
- [ ] Gradle plugin `io.rtmx.gradle` with task `rtmxExtract` generates RTM JSON
- [ ] Support for `@ParameterizedTest` and `@RepeatedTest` JUnit features
- [ ] Kotlin DSL support with idiomatic annotation usage
- [ ] `rtmx from-junit <results.json>` imports JUnit test results into RTM database

## Technical Notes
- JUnit 5 extension model via `@ExtendWith(RtmxExtension.class)`
- Annotations defined in separate `rtmx-annotations` artifact for compile-time only dependency
- `TestExecutionListener.executionFinished()` hook for result capture
- Maven plugin uses `maven-plugin-api` 3.9.x and `maven-plugin-annotations`
- Gradle plugin compatible with Gradle 7.x+ and Kotlin DSL
- JSON output follows RTMX marker specification schema (REQ-LANG-007)
- Annotation retention policy: RUNTIME for reflection-based extraction
- Consider annotation inheritance for test class hierarchies

## Test Cases
1. `tests/test_lang_junit.py::test_req_annotation_parsing` - Parse @Req annotations
2. `tests/test_lang_junit.py::test_scope_annotation_parsing` - Parse @RtmxScope
3. `tests/test_lang_junit.py::test_composite_annotation` - Parse @RtmxTest composite
4. `tests/test_lang_junit.py::test_execution_listener_capture` - Verify listener hooks
5. `tests/test_lang_junit.py::test_maven_plugin_extraction` - Maven plugin integration
6. `tests/test_lang_junit.py::test_gradle_plugin_extraction` - Gradle plugin integration
7. `tests/test_lang_junit.py::test_parameterized_test_support` - Parameterized tests
8. `tests/test_lang_junit.py::test_kotlin_annotation_support` - Kotlin compatibility
9. `tests/test_lang_junit.py::test_from_junit_import` - CLI import command

## Dependencies
- REQ-LANG-007: Language-agnostic marker annotation spec

## Blocks
- None

## Effort
4.0 weeks
