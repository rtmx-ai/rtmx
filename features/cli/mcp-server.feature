@REQ-CLI-019 @REQ-MCP-001 @cli @phase-7
Feature: MCP Server for AI Agent Integration
  As an AI agent developer
  I want to start the MCP protocol server
  So that AI agents can query and update the RTM

  Background:
    Given an initialized RTMX project

  @scope_system @technique_nominal
  Scenario: Display MCP server help
    When I run "rtmx mcp-server --help"
    Then the exit code should be 0
    And I should see "MCP" in the output

  @scope_system @technique_nominal
  Scenario: MCP server accepts port option
    When I run "rtmx mcp-server --help"
    Then the exit code should be 0
    And I should see "port" in the output

  @scope_system @technique_nominal
  Scenario: MCP server accepts stdio mode
    When I run "rtmx mcp-server --help"
    Then the exit code should be 0
    And I should see "stdio" in the output

  @scope_system @technique_stress
  Scenario: MCP server with empty database shows help
    When I run "rtmx mcp-server --help"
    Then the exit code should be 0
