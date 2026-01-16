@REQ-CLI-009 @REQ-PYTEST-001 @cli @phase-3
Feature: RTM From-Tests Command
  As a developer using RTMX
  I want to scan test files for requirement markers
  So that I can discover test coverage and sync tests with my RTM database

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Scan tests directory and discover markers
    Given a test file with requirement markers
    When I run "rtmx from-tests"
    Then the exit code should be 0
    And I should see "Scanning" in the output
    And I should see "requirement" in the output

  @scope_system @technique_nominal
  Scenario: Display summary of discovered markers
    Given a test file with requirement markers
    When I run "rtmx from-tests"
    Then the exit code should be 0
    And I should see "Summary" in the output

  @scope_system @technique_nominal
  Scenario: Show all discovered requirements with --all flag
    Given a test file with requirement markers
    When I run "rtmx from-tests --all"
    Then the exit code should be 0
    And I should see "All Requirements" in the output

  @scope_system @technique_nominal
  Scenario: Show requirements not in database with --missing flag
    Given a test file with requirement markers
    When I run "rtmx from-tests --missing"
    Then the exit code should be 0

  @scope_system @technique_nominal
  Scenario: Update RTM database with test information
    Given a test file with requirement markers
    And the RTM database has 5 requirements
    When I run "rtmx from-tests --update"
    Then the exit code should be 0
    And I should see "Updated" in the output

  @scope_system @technique_nominal
  Scenario: Scan specific test path
    Given a test file with requirement markers
    When I run "rtmx from-tests --path tests"
    Then the exit code should be 0
    And I should see "Scanning" in the output

  @scope_system @technique_nominal
  Scenario: Display help information
    When I run "rtmx from-tests --help"
    Then the exit code should be 0
    And I should see "from-tests" in the output

  @scope_system @technique_stress
  Scenario: Handle missing tests directory gracefully
    When I run "rtmx from-tests --path nonexistent"
    Then the command should fail
    And I should see "Error" in the output

  @scope_system @technique_stress
  Scenario: Handle empty tests directory gracefully
    When I run "rtmx from-tests"
    Then the exit code should be 0
    And I should see "No requirement markers found" in the output

  @scope_system @technique_nominal
  Scenario: Discover class-level requirement markers
    Given a test file with class-level requirement markers
    When I run "rtmx from-tests --all"
    Then the exit code should be 0
    And I should see "REQ-" in the output

  @scope_system @technique_nominal
  Scenario: Report requirements in database without tests
    Given the RTM database has 5 requirements
    When I run "rtmx from-tests"
    Then the exit code should be 0
    And I should see "without tests" in the output
