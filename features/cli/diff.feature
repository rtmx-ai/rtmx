@REQ-CLI-013 @cli @phase-4
Feature: RTM Diff Command
  As a developer using RTMX
  I want to compare RTM database versions
  So that I can see what requirements have changed

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Diff command shows no changes for same file
    Given the RTM database has 5 requirements
    When I run "rtmx diff docs/rtm_database.csv docs/rtm_database.csv"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Diff command with JSON output format
    Given the RTM database has 5 requirements
    When I run "rtmx diff docs/rtm_database.csv docs/rtm_database.csv --format json"
    Then the exit code should be 0
    And the output should be valid JSON

  @scope_system @technique_nominal
  Scenario: Diff command with Markdown output format
    Given the RTM database has 5 requirements
    When I run "rtmx diff docs/rtm_database.csv docs/rtm_database.csv --format markdown"
    Then the exit code should be 0
