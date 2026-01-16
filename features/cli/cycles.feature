@REQ-CLI-007 @REQ-GRAPH-002 @cli @phase-4
Feature: RTM Dependency Cycle Detection
  As a developer using RTMX
  I want to detect circular dependencies in my requirements
  So that I can resolve them before they block project progress

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: No cycles found in acyclic dependency graph
    Given the RTM database has 5 requirements
    And the requirements have no circular dependencies
    When I run "rtmx cycles"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Cycles detected with simple two-node cycle
    Given the RTM database has 2 requirements with a cycle
    When I run "rtmx cycles"
    Then I should see "cycle" in the output

  @scope_system @technique_nominal
  Scenario: Cycles output shows detection status
    Given the RTM database has 5 requirements
    When I run "rtmx cycles"
    Then the exit code should be 0

  @scope_system @technique_stress
  Scenario: Cycles handles empty database gracefully
    When I run "rtmx cycles"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Cycles displays help information
    When I run "rtmx cycles --help"
    Then the exit code should be 0
    And I should see "cycles" in the output
