Feature: Agent Delivery Science
  Warn-first delivery checks, loop tick with re-decompose, scientific MCP, and trade checkpoints.

  @REQ-ORCH-023a
  Scenario: delivery-check warns on unmapped commits by default
    Given a branch with commits that omit AC markers
    When the agent runs "rtmx delivery-check"
    Then the command exits 0
    And the output contains warnings about missing AC markers

  @REQ-ORCH-023a
  Scenario: delivery-check --strict fails on multi-REQ PR body
    Given PR metadata naming two requirement IDs
    When the agent runs "rtmx delivery-check --strict"
    Then the command exits 1

  @REQ-ORCH-023b
  Scenario: loop tick claims and returns a plan
    Given an unblocked atomic requirement and agent_id "agent-1"
    When the agent runs "rtmx loop tick --agent-id agent-1 --json"
    Then the plan includes req_id, claimed, pr_policy, and commit_policy

  @REQ-ORCH-023b
  Scenario: loop tick re-decomposes from delivery notes
    Given a delivery note with a Follow-on decomposition section listing a coarse REQ
    When loop tick selects the next sibling in that web
    Then the plan lists that REQ under redecomposed

  @REQ-ORCH-023c
  Scenario: MCP exposes scientific tools
    When tools/list is called
    Then the list includes loop_tick, decompose, hygiene, cycles, webs, context, delivery_check

  @REQ-ORCH-023d
  Scenario: open trade blocks ready_to_implement
    Given an open trade for the selected requirement
    When loop_tick returns a plan
    Then trade_required is true and ready_to_implement is false
