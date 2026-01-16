@REQ-CLI-017 @REQ-TUI-001 @cli @phase-7
Feature: Terminal User Interface
  As a developer using RTMX
  I want to launch a terminal-based user interface
  So that I can interactively navigate and manage requirements

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display TUI command help
    When I run "rtmx tui --help"
    Then the command should succeed

  @scope_system @technique_nominal
  Scenario: TUI help shows available options
    When I run "rtmx tui --help"
    Then the command should succeed
    And I should see "tui" in the output

  @scope_system @technique_stress
  Scenario: TUI command is available
    When I run "rtmx --help"
    Then the command should succeed
    And I should see "tui" in the output
