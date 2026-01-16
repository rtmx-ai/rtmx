@REQ-CLI-005 @REQ-UX-001 @cli @phase-1
Feature: RTM Project Setup
  As a developer adopting RTMX
  I want to set up RTMX in my existing project
  So that I can integrate requirements tracking with my workflow

  # Note: These scenarios do NOT use Background because setup creates the project.
  # The setup command is designed to integrate into existing projects,
  # configuring agent instructions, Makefile targets, and other integrations.

  @scope_system @technique_nominal
  Scenario: Minimal setup in empty directory
    Given an empty directory
    When I run "rtmx setup --minimal"
    Then the exit code should be 0
    And I should see "Setup complete" in the output

  @scope_system @technique_nominal
  Scenario: Full setup creates all configuration files
    Given an empty directory
    When I run "rtmx setup"
    Then the exit code should be 0
    And I should see "rtmx.yaml" in the output
    And I should see "CLAUDE.md" in the output

  @scope_system @technique_nominal
  Scenario: Setup displays dry run preview
    Given an empty directory
    When I run "rtmx setup --dry-run"
    Then the exit code should be 0
    And I should see "dry run" in the output

  @scope_system @technique_nominal
  Scenario: Setup skips existing configuration
    Given an initialized RTMX project
    When I run "rtmx setup"
    Then the exit code should be 0
    And I should see "skipped" in the output

  @scope_system @technique_nominal
  Scenario: Setup with force overwrites existing configuration
    Given an initialized RTMX project
    When I run "rtmx setup --force"
    Then the exit code should be 0
    And I should see "Setup complete" in the output

  @scope_system @technique_nominal
  Scenario: Setup displays help information
    Given an empty directory
    When I run "rtmx setup --help"
    Then the exit code should be 0
    And I should see "setup" in the output

  @scope_system @technique_nominal
  Scenario: Setup with skip-agents flag
    Given an empty directory
    When I run "rtmx setup --skip-agents"
    Then the exit code should be 0
    And I should see "Setup complete" in the output

  @scope_system @technique_nominal
  Scenario: Setup with skip-makefile flag
    Given an empty directory
    When I run "rtmx setup --skip-makefile"
    Then the exit code should be 0
    And I should see "Setup complete" in the output
