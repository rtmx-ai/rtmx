@REQ-CLI-015 @cli @phase-4
Feature: Makefile Generation
  As a developer using RTMX
  I want to generate a Makefile with RTMX targets
  So that I can integrate RTMX commands into my build workflow

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Generate Makefile to stdout by default
    When I run "rtmx makefile"
    Then the command should succeed
    And I should see ".PHONY" in the output

  @scope_system @technique_nominal
  Scenario: Generated Makefile contains rtmx targets
    When I run "rtmx makefile"
    Then the command should succeed
    And I should see "rtmx" in the output

  @scope_system @technique_nominal
  Scenario: Output Makefile to file with -o option
    When I run "rtmx makefile -o rtmx.mk"
    Then the command should succeed

  @scope_system @technique_nominal
  Scenario: Output Makefile to file with --output option
    When I run "rtmx makefile --output rtmx.mk"
    Then the command should succeed

  @scope_system @technique_stress
  Scenario: Generate Makefile with empty database
    When I run "rtmx makefile"
    Then the command should succeed
    And I should see ".PHONY" in the output
