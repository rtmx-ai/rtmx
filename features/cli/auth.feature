@REQ-ZT-001 @cli @phase-10
Feature: Authentication Commands
  As a developer using RTMX
  I want to authenticate with RTMX sync services
  So that I can access cross-repo requirements across trust boundaries

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Check authentication status when not logged in
    When I run "rtmx auth status"
    Then I should see "Not authenticated" in the output
    And I should see "Run 'rtmx auth login'" in the output

  @scope_system @technique_nominal
  Scenario: Logout clears credentials
    When I run "rtmx auth logout"
    Then the command should succeed
    And I should see "Logged out successfully" in the output

  @scope_system @technique_nominal
  Scenario: Login shows authentication provider info
    When I run "rtmx auth status"
    Then I should see "Provider:" in the output
    And I should see "Issuer:" in the output

  @scope_system @technique_nominal
  Scenario: Login without browser prints URL
    When I run "rtmx auth login --no-browser"
    Then I should see "Starting authentication" in the output
    # Note: Full login test requires mock OIDC server

  @scope_system @technique_nominal
  Scenario: Auth help shows available commands
    When I run "rtmx auth --help"
    Then the command should succeed
    And I should see "login" in the output
    And I should see "logout" in the output
    And I should see "status" in the output
