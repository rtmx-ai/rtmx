@REQ-CLI-001 @REQ-UX-001 @cli @phase-5
Feature: RTM Status Display
  As a developer using RTMX
  I want to see the current RTM completion status
  So that I can track project progress at a glance

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display status summary with partial completion
    Given the RTM database has 10 requirements
    And 5 requirements are COMPLETE
    When I run "rtmx status"
    Then the exit code should be 1
    And I should see the completion percentage

  @scope_system @technique_nominal
  Scenario: Status shows 100% for fully complete database
    Given 10 of 10 requirements are COMPLETE
    When I run "rtmx status"
    Then the exit code should be 0
    And I should see "100.0%"

  @scope_system @technique_nominal
  Scenario: Status shows progress for incomplete database
    Given 5 of 10 requirements are COMPLETE
    When I run "rtmx status"
    Then the exit code should be 1
    And I should see "50.0%"

  @scope_system @technique_stress
  Scenario: Status handles empty database gracefully
    When I run "rtmx status"
    Then the exit code should be 1

  @scope_system @technique_nominal
  Scenario: Verbose status shows category breakdown
    Given the RTM database has 5 requirements
    And 2 requirements are COMPLETE
    When I run "rtmx status -v"
    Then the exit code should be 1
    And I should see the completion percentage

  @scope_system @technique_nominal
  Scenario Outline: Exit codes reflect completion status
    Given <complete> of <total> requirements are COMPLETE
    When I run "rtmx status"
    Then the exit code should be <code>

    Examples:
      | complete | total | code |
      | 10       | 10    | 0    |
      | 5        | 10    | 1    |
      | 0        | 10    | 1    |
      | 0        | 0     | 1    |
