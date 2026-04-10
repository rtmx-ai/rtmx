Feature: rtmx verify --results parses cross-language results
  Cross-language test runners produce an RTMX results JSON file that
  rtmx verify --results consumes to update requirement status. The
  parser must accept the canonical nested form, the convenience flat
  form, and a status string, while rejecting unknown fields.

  Background:
    Given a database containing requirement "REQ-INGEST-030"

  Scenario: Canonical nested marker form (REQ-VERIFY-002)
    Given a results file:
      """
      [{"marker":{"req_id":"REQ-INGEST-030","test_name":"t","test_file":"t.py"},"passed":true}]
      """
    When I run "rtmx verify --results results.json --dry-run"
    Then the command exits 0
    And REQ-INGEST-030 is reported as passing

  Scenario: Flat-form compatibility (REQ-VERIFY-004)
    Given a results file:
      """
      [{"req_id":"REQ-INGEST-030","test_name":"test_foo","test_file":"tests/unit/test_foo.py","status":"pass"}]
      """
    When I run "rtmx verify --results results.json --dry-run"
    Then the command exits 0
    And REQ-INGEST-030 is reported as passing

  Scenario: Status string maps to passed=false (REQ-VERIFY-004)
    Given a results file with status "fail" for REQ-INGEST-030
    When I run "rtmx verify --results results.json --dry-run"
    Then REQ-INGEST-030 is reported as failing

  Scenario: Unknown top-level field is rejected (REQ-VERIFY-004)
    Given a results file containing an unknown key "reqid"
    When I run "rtmx verify --results results.json"
    Then the command exits non-zero
    And the error mentions the unknown field

  Scenario: Mixed nested and flat: nested wins (REQ-VERIFY-004)
    Given a results file with both a "marker" object and a top-level "req_id"
    When I run "rtmx verify --results results.json --dry-run"
    Then the requirement matched is the one inside the marker object
