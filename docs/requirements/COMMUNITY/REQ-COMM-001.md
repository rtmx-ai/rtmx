# REQ-COMM-001: HackerNews Community Feedback Post

## Status: MISSING
## Priority: MEDIUM
## Phase: 12

## Description
RTMX team shall publish a Show HN post on Hacker News to introduce the tool to the developer community and solicit feedback on the requirements traceability approach, CLI design, and roadmap priorities.

## Acceptance Criteria
- [ ] Post title follows Show HN format: "Show HN: RTMX - Requirements Traceability for AI-Assisted Development"
- [ ] Post includes concise description of problem solved and unique value proposition
- [ ] Links to GitHub repo, documentation site (rtmx.ai), and PyPI package
- [ ] Post timing considers HN traffic patterns (weekday morning US time)
- [ ] Team monitors and responds to comments within 24 hours
- [ ] Feedback is captured and triaged into GitHub issues or backlog items
- [ ] Post-mortem documents lessons learned and community sentiment

## Post Content Guidelines

### Title Options
1. "Show HN: RTMX - Git-native requirements traceability for TDD teams"
2. "Show HN: RTMX - Track requirements to tests in your Python projects"
3. "Show HN: RTMX - Requirements management that lives in your repo"

### Key Points to Cover
- Problem: Requirements tracking is disconnected from code, especially in AI-assisted workflows
- Solution: CSV-based RTM database in git, pytest markers link tests to requirements
- Differentiators: Git-native, AI-friendly, no external service dependency
- Current state: v0.0.5, Python CLI, pytest plugin, MCP server for Claude Code
- Asking for: Feedback on approach, feature priorities, use cases we haven't considered

### Links to Include
- GitHub: https://github.com/rtmx-ai/rtmx
- Docs: https://rtmx.ai
- PyPI: https://pypi.org/project/rtmx/

## Feedback Triage Process
1. Create GitHub issue for each actionable suggestion
2. Tag issues with `community-feedback` label
3. Link back to HN comment in issue description
4. Prioritize based on upvotes and alignment with roadmap
5. Respond on HN when issue is addressed

## Success Metrics
- [ ] Post reaches HN front page (>50 points)
- [ ] Minimum 10 substantive comments with feedback
- [ ] At least 3 new GitHub issues created from feedback
- [ ] No major negative sentiment about core approach

## Test Cases
- Manual: Post published and visible on HN
- Manual: All links in post are functional
- Manual: Feedback captured in GitHub issues

## Dependencies
- REQ-SITE-001 (Website live) - COMPLETE
- PyPI package published - COMPLETE

## Blocks
- None

## Effort
0.5 weeks

## Notes
- Consider cross-posting to Reddit r/programming, r/Python after HN
- Prepare for common questions: "Why not Jira?", "How does this scale?", "What about non-Python?"
- Have demo GIF or video ready for quick understanding
- Be transparent about early stage and seeking feedback
