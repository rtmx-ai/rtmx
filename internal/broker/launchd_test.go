package broker

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacOSLaunchdDaemonRunsAsDedicatedVerifierUser(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069b")

	path := filepath.Join("..", "..", "packaging", "launchd", "com.rtmx.verify.plist")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "<key>UserName</key>") || !strings.Contains(body, "<string>rtmx-verify</string>") {
		t.Fatalf("plist missing UserName rtmx-verify:\n%s", body)
	}
	if strings.Contains(body, "LaunchAgents") {
		t.Fatal("must be a LaunchDaemon, not a LaunchAgent")
	}
}

func TestMacOSVerifierKeyNotReadableByDeveloper(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "verify.key")
	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePrivateKeyMode600(path); err != nil {
		t.Fatal(err)
	}
	readable, err := PrivateKeyReadableByOthers(path)
	if err != nil {
		t.Fatal(err)
	}
	if readable {
		t.Fatal("developer must not read the macOS verifier private key")
	}
	plist, err := os.ReadFile(filepath.Join("..", "..", "packaging", "launchd", "com.rtmx.verify.plist"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plist), "/var/lib/rtmx/verify.key") {
		t.Fatal("plist must name the private key path under the verifier account")
	}
}
