@REQ-CLI-006 @REQ-GRAPH-001 @cli @phase-4
Feature: Dependency Analysis
  As a developer using RTMX
  I want to analyze requirement dependencies
  So that I can understand the project structure and identify blockers

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display dependency tree for project
    Given the RTM database has 5 requirements
    When I run "rtmx deps"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Show dependencies for specific requirement
    Given the RTM database has 5 requirements
    When I run "rtmx deps --req REQ-TEST-001"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Filter dependencies by category
    Given the RTM database has 5 requirements
    When I run "rtmx deps --category TEST"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Filter dependencies by phase
    Given the RTM database has 5 requirements
    When I run "rtmx deps --phase 1"
    Then the exit code should be 0

  @scope_system @technique_stress
  Scenario: Handle empty database gracefully
    When I run "rtmx deps"
    Then the exit code should be 0
