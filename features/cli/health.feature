@REQ-CLI-003 @REQ-UX-001 @cli @phase-5
Feature: RTM Health Check
  As a developer using RTMX
  I want to run health checks on my RTM database
  So that I can identify and fix issues before they cause problems

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Basic health check shows check results
    When I run "rtmx health"
    Then the exit code should be 0
    And I should see "check" in the output

  @scope_system @technique_nominal
  Scenario: Health check with verbose output
    When I run "rtmx health -v"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Health check JSON output
    When I run "rtmx health --format json"
    Then the exit code should be 0
    And the output should be valid JSON

  @scope_system @technique_nominal
  Scenario: Health check CI format output
    When I run "rtmx health --format ci"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Health check specific check only
    When I run "rtmx health --check rtm_exists"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Health check strict mode with healthy project
    When I run "rtmx health --strict"
    Then the exit code should be 0

  @scope_system @technique_stress
  Scenario: Health check detects issues in project with problems
    Given the RTM database has 5 requirements
    And 2 requirements are COMPLETE
    When I run "rtmx health"
    Then the exit code should be 0

  @scope_system @technique_stress
  Scenario: Strict mode fails on warnings
    Given the RTM database has 3 requirements
    When I run "rtmx health --strict"
    Then the exit code should be 1
