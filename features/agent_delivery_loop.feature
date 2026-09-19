Feature: Deployable falling-edge agent delivery loop

  A user deploys `rtmx loop` so an agent keeps working the backlog.
  The loop wakes when a requirement is merged, not on a timer.
  Each wake runs next, then decomposition.
  `rtmx init` installs the delivery rule: one commit per acceptance
  criterion, one pull request per requirement.

  Local RTM is enough. Deploy does not need managed-sync OAuth or Stripe,
  and it does not edit an agent session in this repository.

    rtmx loop
    rtmx loop install

  macOS enable (install does not run this):
    launchctl bootstrap gui/$(id -u) .rtmx/loop/ai.rtmx.loop.plist

  Linux enable (install does not run this):
    systemctl --user link .rtmx/loop/rtmx-loop.service && systemctl --user enable rtmx-loop.service

  Scenario: Init installs the delivery rule (REQ-ORCH-021)
    When I run "rtmx init" in an empty project
    Then ".rtmx/agent/delivery.md" contains "One pull request per requirement"
    And ".rtmx/agent/delivery.md" contains "One commit per acceptance criterion"
    And ".cursor/rules/rtmx-delivery.mdc" contains "One pull request per requirement"
    And the command does not install a loop unit

  Scenario: Merge edge selects next and decomposes (REQ-ORCH-019c)
    Given a backlog with a coarse unblocked requirement "REQ-EX-010"
    And a falling-edge event for a different requirement that just merged
    When I run "rtmx loop --once"
    Then the plan claims "REQ-EX-010" or one of its children
    And the plan records pr_policy "one_pr_per_requirement"
    And the plan records commit_policy "one_commit_per_ac"
    And no git commit was created

  Scenario: Idle when nothing is unblocked (REQ-ORCH-019a)
    Given every requirement is COMPLETE
    When I run "rtmx loop --once"
    Then the command exits 0
    And the plan is idle

  Scenario: Next skips an open pull request (REQ-ORCH-022)
    Given "REQ-EX-001" is incomplete and has an open pull request
    And "REQ-EX-002" is incomplete, unblocked, and has no open pull request
    When I run "rtmx next"
    Then the selected requirement is "REQ-EX-002"
    And "REQ-EX-001" is reported as skipped because of an open pull request
