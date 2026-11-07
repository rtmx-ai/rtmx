package broker

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestAirGapVerificationCompletesThroughLocalBroker(t *testing.T) {
	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ag := &AirGapClient{
		Local: Client{Broker: Config{
			WorkDir: dir,
			Command: []string{"./ok.sh"},
			Key:     priv,
		}},
		Dial: func(network, address string) error {
			t.Fatalf("unexpected dial %s %s", network, address)
			return nil
		},
	}
	res, err := ag.SubmitCommit("REQ-MONO-069i", sha)
	if err != nil {
		t.Fatal(err)
	}
	if res.Attestation == nil || res.Attestation.CommitSHA != sha {
		t.Fatalf("attestation = %+v", res.Attestation)
	}
}

func TestAirGapRunDoesNotAttemptOutboundNetworkCall(t *testing.T) {
	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	dialed := false
	ag := &AirGapClient{
		Local: Client{Broker: Config{
			WorkDir: dir,
			Command: []string{"./ok.sh"},
			Key:     priv,
		}},
		Dial: func(network, address string) error {
			dialed = true
			return nil
		},
	}
	if _, err := ag.SubmitCommit("REQ-MONO-069i", sha); err != nil {
		t.Fatal(err)
	}
	if dialed || ag.OutboundCalls() != 0 {
		t.Fatalf("outbound dial attempted: dialed=%v calls=%d", dialed, ag.OutboundCalls())
	}
}
