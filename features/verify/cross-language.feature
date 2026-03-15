@REQ-VERIFY-001 @cli @phase-14
Feature: Cross-Language Verification
  As a developer using multiple programming languages
  I want to verify requirements from any language's test results
  So that closed-loop verification works regardless of test framework

  Background:
    Given an initialized RTMX project
    And the RTM database has requirements:
      | req_id       | status  | test_function     |
      | REQ-AUTH-001 | MISSING | test_login        |
      | REQ-AUTH-002 | MISSING | test_logout       |
      | REQ-DATA-001 | MISSING | test_parse_csv    |

  @scope_system @technique_nominal
  Scenario: Verify from results file updates requirement status
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
          "passed": true,
          "timestamp": "2026-02-20T18:45:00Z"
        }
      ]
      """
    When I run "rtmx verify --results results.json --update"
    Then the command should succeed
    And requirement "REQ-AUTH-001" should have status "COMPLETE"
    And requirement "REQ-AUTH-002" should have status "MISSING"

  @scope_system @technique_nominal
  Scenario: Verify from results file with failing tests
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
          "passed": true
        },
        {
          "marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login_edge", "test_file": "test_auth.py"},
          "passed": false,
          "error": "AssertionError: expected 200, got 401"
        }
      ]
      """
    When I run "rtmx verify --results results.json --update"
    Then the command should exit with code 1
    And requirement "REQ-AUTH-001" should have status "MISSING"

  @scope_system @technique_nominal
  Scenario: Verify from results file with multiple requirements
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
          "passed": true
        },
        {
          "marker": {"req_id": "REQ-AUTH-002", "test_name": "test_logout", "test_file": "test_auth.py"},
          "passed": true
        },
        {
          "marker": {"req_id": "REQ-DATA-001", "test_name": "test_parse_csv", "test_file": "test_data.py"},
          "passed": false,
          "error": "FileNotFoundError"
        }
      ]
      """
    When I run "rtmx verify --results results.json --update"
    Then the command should exit with code 1
    And requirement "REQ-AUTH-001" should have status "COMPLETE"
    And requirement "REQ-AUTH-002" should have status "COMPLETE"
    And requirement "REQ-DATA-001" should have status "MISSING"

  @scope_system @technique_nominal
  Scenario: Dry run shows changes without updating
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
          "passed": true
        }
      ]
      """
    When I run "rtmx verify --results results.json --dry-run"
    Then the command should succeed
    And I should see "MISSING" in the output
    And I should see "COMPLETE" in the output
    And I should see "Dry run" in the output
    And requirement "REQ-AUTH-001" should have status "MISSING"

  @scope_system @technique_nominal
  Scenario: Read results from stdin
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
          "passed": true
        }
      ]
      """
    When I pipe "results.json" to "rtmx verify --results - --dry-run"
    Then the command should succeed
    And I should see "COMPLETE" in the output

  @scope_system @technique_boundary
  Scenario: Results flag and default go test are mutually exclusive
    When I run "rtmx verify --results results.json ./..."
    Then the command should fail
    And I should see "mutually exclusive" in the output

  @scope_system @technique_boundary
  Scenario: Missing results file produces clear error
    When I run "rtmx verify --results nonexistent.json"
    Then the command should fail
    And I should see "no such file" in the output

  @scope_system @technique_boundary
  Scenario: Empty results file succeeds with no changes
    Given an RTMX results file "results.json" with:
      """json
      []
      """
    When I run "rtmx verify --results results.json --update"
    Then the command should succeed
    And I should see "No requirements" in the output
