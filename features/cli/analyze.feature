@REQ-CLI-011 @cli @phase-4
Feature: RTM Analyze Command
  As a developer using RTMX
  I want to analyze the RTM database for insights
  So that I can understand project metrics and health

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Analyze command runs successfully on project with data
    Given the RTM database has 5 requirements
    And 2 requirements are COMPLETE
    When I run "rtmx analyze"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Analyze command with JSON output format
    Given the RTM database has 5 requirements
    And 2 requirements are COMPLETE
    When I run "rtmx analyze --format json"
    Then the exit code should be 0

  @scope_system @technique_stress
  Scenario: Analyze command handles empty database
    When I run "rtmx analyze"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Analyze command on fully complete database
    Given 10 of 10 requirements are COMPLETE
    When I run "rtmx analyze"
    Then the exit code should be 0
