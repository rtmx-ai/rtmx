@REQ-CLI-004 @REQ-UX-001 @cli @phase-1
Feature: RTM Project Initialization
  As a developer adopting RTMX
  I want to initialize a new RTMX project
  So that I can start tracking requirements immediately

  # Note: These scenarios do NOT use Background because init creates the project.
  # Scenarios requiring a fresh directory need "Given an empty directory" step
  # to be implemented in common_steps.py to properly initialize project_dir.

  @scope_system @technique_nominal
  Scenario: Initialize new RTMX project in empty directory
    Given an empty directory
    When I run "rtmx init"
    Then the exit code should be 0
    And I should see "Initialized"

  @scope_system @technique_nominal
  Scenario: Init refuses to overwrite existing project
    Given an initialized RTMX project
    When I run "rtmx init"
    Then the command should fail
    And I should see "already initialized" in the output

  @scope_system @technique_nominal
  Scenario: Init with --force overwrites existing project
    Given an initialized RTMX project
    When I run "rtmx init --force"
    Then the exit code should be 0
    And I should see "Initialized"

  @scope_system @technique_nominal
  Scenario: Init displays help information
    Given an empty directory
    When I run "rtmx init --help"
    Then the exit code should be 0
    And I should see "init" in the output

  @scope_system @technique_nominal
  Scenario: Init with custom path
    Given an empty directory
    When I run "rtmx init --path custom-rtm"
    Then the exit code should be 0
    And I should see "Initialized"
