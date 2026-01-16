@REQ-CLI-014 @cli @phase-4
Feature: RTM Install Command
  As a developer using RTMX
  I want to install RTMX integrations and plugins
  So that I can extend RTMX functionality in my project

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Install command shows help with available plugins
    When I run "rtmx install --help"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Install command shows available options
    When I run "rtmx install --help"
    Then the exit code should be 0
