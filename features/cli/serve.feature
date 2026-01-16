@REQ-CLI-016 @REQ-WEB-001 @cli @phase-6
Feature: Web Dashboard Server
  As a developer using RTMX
  I want to start a web dashboard server
  So that I can view RTM status in a browser

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display serve command help
    When I run "rtmx serve --help"
    Then the command should succeed
    And I should see "port" in the output

  @scope_system @technique_nominal
  Scenario: Serve command accepts port option
    When I run "rtmx serve --help"
    Then the command should succeed
    And I should see "--port" in the output

  @scope_system @technique_nominal
  Scenario: Serve command accepts host option
    When I run "rtmx serve --help"
    Then the command should succeed
    And I should see "--host" in the output

  @scope_system @technique_stress
  Scenario: Serve with empty database shows help
    When I run "rtmx serve --help"
    Then the command should succeed
