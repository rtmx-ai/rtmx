@REQ-LANG-004 @REQ-VERIFY-001 @phase-14
Feature: Python Extension Dogfood Loop
  As the RTMX development team
  I want to use RTMX markers on Python extension tests
  So that we verify the Python extension using its own capabilities

  Background:
    Given an initialized RTMX project with Go CLI installed
    And the Python rtmx package is installed

  @scope_system @technique_nominal
  Scenario: Python tests produce RTMX results JSON
    Given a Python test file "test_example.py" with:
      """python
      import pytest

      @pytest.mark.req("REQ-LANG-004")
      @pytest.mark.scope_unit
      @pytest.mark.technique_nominal
      @pytest.mark.env_simulation
      def test_marker_registration():
          assert True
      """
    When I run "pytest test_example.py --rtmx-output=results.json"
    Then "results.json" should exist
    And "results.json" should contain "REQ-LANG-004"
    And "results.json" should be valid RTMX results JSON

  @scope_system @technique_nominal
  Scenario: Go CLI verifies Python test results
    Given a Python test file "test_example.py" with:
      """python
      import pytest

      @pytest.mark.req("REQ-LANG-004")
      def test_plugin_works():
          assert True
      """
    And the RTM database has requirement "REQ-LANG-004" with status "MISSING"
    When I run "pytest test_example.py --rtmx-output=results.json"
    And I run "rtmx verify --results results.json --update"
    Then requirement "REQ-LANG-004" should have status "COMPLETE"

  @scope_system @technique_nominal
  Scenario: Mixed Go and Python test results in same project
    Given Go tests with rtmx.Req("REQ-GO-001") markers
    And Python tests with @pytest.mark.req("REQ-PY-001") markers
    And the RTM database has requirements:
      | req_id     | status  |
      | REQ-GO-001 | MISSING |
      | REQ-PY-001 | MISSING |
    When I run Go tests producing "go-results.json"
    And I run Python tests producing "py-results.json"
    And I run "rtmx verify --results go-results.json --update"
    And I run "rtmx verify --results py-results.json --update"
    Then requirement "REQ-GO-001" should have status "COMPLETE"
    And requirement "REQ-PY-001" should have status "COMPLETE"

  @scope_system @technique_nominal
  Scenario: Full dogfood loop on Python extension development
    Given the Python extension test suite with RTMX markers
    When I run the Python extension tests with RTMX output
    And I verify the results with the Go CLI
    Then all REQ-LANG-004 acceptance criteria are verified
    And the RTM database reflects actual test status
