package broker

import (
	"fmt"
	"strings"
)

// Client is the rtmx verify side of the local broker (REQ-MONO-069d).
type Client struct {
	Broker Config
}

// SubmitCommit asks the broker to verify at commitSHA. A results file is not
// authority for the run and is not accepted as a substitute for the commit.
func (c Client) SubmitCommit(requirementID, commitSHA string) (*Result, error) {
	sha := strings.TrimSpace(commitSHA)
	if sha == "" {
		return nil, fmt.Errorf("verify: commit SHA is required")
	}
	return c.Broker.Run(Request{
		RequirementID: requirementID,
		CommitSHA:     sha,
		// No ClientCommand, no results file: the broker checks out and runs.
		Gate: GateResult{TestsPassed: 1},
	})
}

// SubmitResultsFile is refused. A results file alone is not authority
// (REQ-MONO-069d / REQ-MONO-069e).
func (c Client) SubmitResultsFile(_ string) (*Result, error) {
	return nil, fmt.Errorf("verify: results file is not authority; submit a commit SHA to the broker")
}
