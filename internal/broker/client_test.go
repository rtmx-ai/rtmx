package broker

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
)


func TestVerifySubmitsCommitSHAToLocalBroker(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069d")

	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	client := Client{Broker: Config{
		WorkDir: dir,
		Command: okCommand(),
		Key:     priv,
	}}
	res, err := client.SubmitCommit("REQ-MONO-069d", sha)
	if err != nil {
		t.Fatal(err)
	}
	if res.Attestation == nil || res.Attestation.CommitSHA != sha {
		t.Fatalf("attestation = %+v", res.Attestation)
	}
}

func TestVerifySubmissionDoesNotAcceptResultsFileAsAuthority(t *testing.T) {
	client := Client{}
	_, err := client.SubmitResultsFile("/tmp/results.json")
	if err == nil || !strings.Contains(err.Error(), "results file is not authority") {
		t.Fatalf("err = %v", err)
	}
}
