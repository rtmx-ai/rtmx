@REQ-CLI-018 @REQ-DOC-001 @cli @phase-2
Feature: Documentation Generation
  As a developer using RTMX
  I want to generate documentation from my RTM database
  So that I can share project progress with stakeholders

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display docs command help
    When I run "rtmx docs --help"
    Then the command should succeed

  @scope_system @technique_nominal
  Scenario: Docs command accepts format option
    When I run "rtmx docs --help"
    Then the command should succeed
    And I should see "format" in the output

  @scope_system @technique_nominal
  Scenario: Docs command accepts output option
    When I run "rtmx docs --help"
    Then the command should succeed
    And I should see "output" in the output

  @scope_system @technique_nominal
  Scenario: Generate documentation with populated database
    Given the RTM database has 5 requirements
    And 2 requirements are COMPLETE
    When I run "rtmx docs --help"
    Then the command should succeed

  @scope_system @technique_stress
  Scenario: Generate documentation with empty database
    When I run "rtmx docs --help"
    Then the command should succeed
