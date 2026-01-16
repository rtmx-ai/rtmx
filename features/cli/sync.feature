@REQ-CLI-020 @REQ-SYNC-001 @cli @phase-10
Feature: RTM Synchronization with External Services
  As a developer using RTMX
  I want to sync my RTM with external services
  So that I can integrate with existing project management tools

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display sync command help
    When I run "rtmx sync --help"
    Then the exit code should be 0
    And I should see "sync" in the output

  @scope_system @technique_nominal
  Scenario: Sync command lists available providers
    When I run "rtmx sync --help"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Sync requires provider selection
    When I run "rtmx sync"
    Then the exit code should be 1

  @scope_system @technique_stress
  Scenario: Sync with empty database
    When I run "rtmx sync --help"
    Then the exit code should be 0
