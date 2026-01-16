@REQ-CLI-002 @REQ-UX-001 @cli @phase-5
Feature: RTM Backlog Display
  As a developer using RTMX
  I want to view incomplete requirements in the backlog
  So that I can prioritize and plan my work effectively

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display backlog with incomplete requirements
    Given the RTM database has 10 requirements
    And 5 requirements are COMPLETE
    When I run "rtmx backlog"
    Then the exit code should be 0
    And I should see "REQ-" in the output

  @scope_system @technique_nominal
  Scenario: Backlog shows only incomplete items by default
    Given 3 of 10 requirements are COMPLETE
    When I run "rtmx backlog"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Filter backlog by phase
    Given the RTM database has 5 requirements
    When I run "rtmx backlog --phase 1"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: View all backlog items
    Given the RTM database has 10 requirements
    And 5 requirements are COMPLETE
    When I run "rtmx backlog --view all"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: View critical backlog items
    Given the RTM database has 10 requirements
    When I run "rtmx backlog --view critical"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: View blockers in backlog
    Given the RTM database has 10 requirements
    When I run "rtmx backlog --view blockers"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Limit backlog output
    Given the RTM database has 10 requirements
    When I run "rtmx backlog --limit 2"
    Then the exit code should be 0

  @scope_system @technique_stress
  Scenario: Backlog handles empty database gracefully
    When I run "rtmx backlog"
    Then the exit code should be 0

  @scope_system @technique_parametric
  Scenario Outline: Filter backlog by different views
    Given the RTM database has 10 requirements
    And 5 requirements are COMPLETE
    When I run "rtmx backlog --view <view>"
    Then the exit code should be 0

    Examples:
      | view     |
      | all      |
      | critical |
      | blockers |

  @scope_system @technique_parametric
  Scenario Outline: Filter backlog by different phases
    Given the RTM database has 10 requirements
    When I run "rtmx backlog --phase <phase>"
    Then the exit code should be 0

    Examples:
      | phase |
      | 1     |
      | 2     |
      | 3     |

  @scope_system @technique_parametric
  Scenario Outline: Combine view and limit options
    Given the RTM database has 10 requirements
    When I run "rtmx backlog --view <view> --limit <limit>"
    Then the exit code should be 0

    Examples:
      | view     | limit |
      | all      | 5     |
      | critical | 3     |
      | blockers | 2     |
