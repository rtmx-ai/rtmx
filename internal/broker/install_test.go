package broker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyConnectsWithoutSudoPrompt(t *testing.T) {
	if err := ConnectWithoutAdmin(VerifyConnectOptions{
		SocketPath:  DefaultBrokerSocket(),
		AdminPrompt: false,
	}); err != nil {
		t.Fatal(err)
	}
	err := ConnectWithoutAdmin(VerifyConnectOptions{AdminPrompt: true})
	if err == nil || !strings.Contains(err.Error(), "must not prompt") {
		t.Fatalf("err = %v", err)
	}
	unit, err := os.ReadFile(filepath.Join("..", "..", "packaging", "systemd", "verify.service"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unit), "WantedBy=multi-user.target") {
		t.Fatal("systemd unit must start at boot")
	}
	plist, err := os.ReadFile(filepath.Join("..", "..", "packaging", "launchd", "com.rtmx.verify.plist"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plist), "<key>RunAtLoad</key>") {
		t.Fatal("launchd daemon must start at boot")
	}
}

func TestFleetAdminProvisionsBrokerWithoutEndUser(t *testing.T) {
	ok := FleetProvision{
		VerifierAccount: "rtmx-verify",
		EnableAtBoot:    true,
		RequiresEndUser: false,
	}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := FleetProvision{
		VerifierAccount: "rtmx-verify",
		EnableAtBoot:    true,
		RequiresEndUser: true,
	}
	err := bad.Validate()
	if err == nil || !strings.Contains(err.Error(), "end user must not run") {
		t.Fatalf("err = %v", err)
	}
}
