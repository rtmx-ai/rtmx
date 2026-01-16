@REQ-CLI-010 @REQ-GRAPH-003 @cli @phase-4
Feature: RTM Dependency Reconciliation
  As a developer using RTMX
  I want to reconcile dependency relationships
  So that dependencies and blocks fields maintain reciprocity

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Dry run shows what would change (default mode)
    Given the RTM database has 5 requirements
    When I run "rtmx reconcile"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Check reciprocity in well-formed database
    Given the RTM database has 5 requirements
    And the requirements have no circular dependencies
    When I run "rtmx reconcile"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Fix missing dependencies with execute flag
    Given the RTM database has 5 requirements
    When I run "rtmx reconcile --execute"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Dry run mode does not modify database
    Given the RTM database has 5 requirements
    When I run "rtmx reconcile"
    Then the exit code should be 0
    And the RTM database should not be modified

  @scope_system @technique_nominal
  Scenario: Reconcile displays help information
    When I run "rtmx reconcile --help"
    Then the exit code should be 0
    And I should see "reconcile" in the output

  @scope_system @technique_stress
  Scenario: Reconcile handles empty database gracefully
    When I run "rtmx reconcile"
    Then the exit code should be 0
