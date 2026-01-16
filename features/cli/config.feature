@REQ-CLI-008 @REQ-CONFIG-001 @cli @phase-2
Feature: RTM Configuration Display
  As a developer using RTMX
  I want to view and validate my project configuration
  So that I can ensure my RTMX project is properly configured

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display configuration shows current settings
    When I run "rtmx config"
    Then the exit code should be 0
    And I should see "database" in the output

  @scope_system @technique_nominal
  Scenario: Configuration validates successfully
    When I run "rtmx config --validate"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Configuration JSON output
    When I run "rtmx config --format json"
    Then the exit code should be 0
    And the output should be valid JSON

  @scope_system @technique_nominal
  Scenario: Configuration YAML output
    When I run "rtmx config --format yaml"
    Then the exit code should be 0
    And I should see "database" in the output

  @scope_system @technique_stress
  Scenario: Configuration handles project without custom config
    When I run "rtmx config"
    Then the exit code should be 0
