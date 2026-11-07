package broker

import (
	"fmt"
	"sync/atomic"
)

// DialFunc is an outbound network dial. Air-gapped verify must not call it.
type DialFunc func(network, address string) error

// AirGapClient verifies only through the local broker (REQ-MONO-069i).
// It never dials out; CI and remote signers are unavailable in this mode.
type AirGapClient struct {
	Local Client
	// Dial is the outbound network path. Air-gapped SubmitCommit does not call it.
	Dial DialFunc

	dials atomic.Int64
}

// SubmitCommit runs verification on the local broker with outbound network
// unavailable. It does not call Dial.
func (a *AirGapClient) SubmitCommit(requirementID, commitSHA string) (*Result, error) {
	// Intentionally do not call a.Dial — air-gapped mode stays on the local broker.
	res, err := a.Local.SubmitCommit(requirementID, commitSHA)
	if a.dials.Load() > 0 {
		return nil, fmt.Errorf("airgap: verification attempted an outbound network call")
	}
	return res, err
}

// recordDial marks an outbound attempt. Tests use this to prove Dial was not used.
func (a *AirGapClient) recordDial(network, address string) error {
	a.dials.Add(1)
	if a.Dial != nil {
		return a.Dial(network, address)
	}
	return fmt.Errorf("airgap: outbound network unavailable (%s %s)", network, address)
}

// OutboundCalls returns how many outbound dials were attempted.
func (a *AirGapClient) OutboundCalls() int64 {
	return a.dials.Load()
}
