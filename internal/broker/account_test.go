package broker

import (
	"os/user"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestEnsureVerifierAccountInvokesPlatformProvisioners(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069b")
	if runtime.GOOS == "windows" {
		t.Skip("windows uses service contract account")
	}

	orig := RunCommand
	defer func() { RunCommand = orig }()

	var saw []string
	RunCommand = func(name string, args ...string) error {
		saw = append(saw, name+" "+strings.Join(args, " "))
		return nil
	}

	account := "rtmx-verify-missing-for-test"
	if _, err := user.Lookup(account); err == nil {
		t.Skip("test account unexpectedly exists")
	}
	if _, err := user.LookupGroup(account); err == nil {
		t.Skip("test group unexpectedly exists")
	}

	if err := EnsureVerifierAccount(account); err != nil {
		t.Fatalf("EnsureVerifierAccount: %v", err)
	}
	joined := strings.Join(saw, "\n")
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(joined, "dscl . -create /Groups/"+account) {
			t.Fatalf("expected dscl group create, got:\n%s", joined)
		}
		if !strings.Contains(joined, "dscl . -create /Users/"+account) {
			t.Fatalf("expected dscl user create, got:\n%s", joined)
		}
		if !strings.Contains(joined, "PrimaryGroupID") {
			t.Fatalf("expected PrimaryGroupID, got:\n%s", joined)
		}
	default:
		if !strings.Contains(joined, "groupadd --system "+account) {
			t.Fatalf("expected groupadd, got:\n%s", joined)
		}
		if !strings.Contains(joined, "useradd --system") || !strings.Contains(joined, account) {
			t.Fatalf("expected useradd, got:\n%s", joined)
		}
	}
}

func TestEnsureVerifierAccountNoopWhenPresent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows")
	}
	account := "nobody"
	if _, err := user.Lookup(account); err != nil {
		t.Skip("nobody user missing")
	}
	if _, err := user.LookupGroup(account); err != nil {
		t.Skip("nobody group missing")
	}
	orig := RunCommand
	defer func() { RunCommand = orig }()
	RunCommand = func(string, ...string) error {
		t.Fatal("must not provision when account exists")
		return nil
	}
	if err := EnsureVerifierAccount(account); err != nil {
		t.Fatal(err)
	}
}
