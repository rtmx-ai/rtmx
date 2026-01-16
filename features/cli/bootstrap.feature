@REQ-CLI-012 @cli @phase-1
Feature: RTM Bootstrap Command
  As a developer using RTMX
  I want to bootstrap requirements from test files
  So that I can quickly populate my RTM from existing tests

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Bootstrap command shows help
    When I run "rtmx bootstrap --help"
    Then the exit code should be 0
    And I should see "--from-tests"

  @scope_system @technique_nominal
  Scenario: Bootstrap with --from-tests --dry-run previews changes
    When I run "rtmx bootstrap --from-tests --dry-run"
    Then the exit code should be 1

  @scope_system @technique_stress
  Scenario: Bootstrap handles project with no test files
    When I run "rtmx bootstrap --from-tests --dry-run"
    Then the exit code should be 1
