@REQ-VERIFY-002 @cli @phase-14
Feature: RTMX Results JSON Format
  As a language integration author
  I want a well-defined results JSON schema
  So that my test framework plugin produces compatible output

  @scope_unit @technique_nominal
  Scenario: Valid results file with all fields
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {
            "req_id": "REQ-AUTH-001",
            "scope": "unit",
            "technique": "nominal",
            "env": "simulation",
            "test_name": "test_login_success",
            "test_file": "test_auth.py",
            "line": 42
          },
          "passed": true,
          "duration_ms": 15.5,
          "error": "",
          "timestamp": "2026-02-20T18:45:00Z"
        }
      ]
      """
    When I validate the results file
    Then the validation should succeed

  @scope_unit @technique_nominal
  Scenario: Valid results file with minimal fields
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {
            "req_id": "REQ-AUTH-001",
            "test_name": "test_login",
            "test_file": "test_auth.py"
          },
          "passed": true
        }
      ]
      """
    When I validate the results file
    Then the validation should succeed

  @scope_unit @technique_boundary
  Scenario: Invalid requirement ID format rejected
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {
            "req_id": "INVALID-ID",
            "test_name": "test_foo",
            "test_file": "test.py"
          },
          "passed": true
        }
      ]
      """
    When I validate the results file
    Then the validation should fail
    And the error should mention "req_id"

  @scope_unit @technique_boundary
  Scenario: Missing required marker fields rejected
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {
            "req_id": "REQ-AUTH-001"
          },
          "passed": true
        }
      ]
      """
    When I validate the results file
    Then the validation should fail
    And the error should mention "test_name"

  @scope_unit @technique_boundary
  Scenario: Invalid scope enum rejected
    Given an RTMX results file "results.json" with:
      """json
      [
        {
          "marker": {
            "req_id": "REQ-AUTH-001",
            "scope": "not_a_valid_scope",
            "test_name": "test_foo",
            "test_file": "test.py"
          },
          "passed": true
        }
      ]
      """
    When I validate the results file
    Then the validation should fail
    And the error should mention "scope"

  @scope_unit @technique_nominal
  Scenario: Go WriteResultsJSON output is schema-compatible
    Given a Go test that uses rtmx.Req markers
    When the test produces "rtmx-results.json"
    Then the results file should validate against the schema

  @scope_unit @technique_nominal
  Scenario: Python pytest plugin output is schema-compatible
    Given a Python test with @pytest.mark.req markers
    When the test produces "rtmx-results.json"
    Then the results file should validate against the schema
