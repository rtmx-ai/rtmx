package broker

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestDefaultPackagePathsShipWithRtmxLayout(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069b")

	p := DefaultPackagePaths()
	if p.VerifierAccount == "" {
		t.Fatal("verifier account required")
	}
	if p.KeyPath == "" || p.VerifyBinary == "" {
		t.Fatalf("incomplete paths: %+v", p)
	}
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(p.VerifyBinary, "libexec/rtmx/verify") {
			t.Fatalf("darwin binary = %q", p.VerifyBinary)
		}
		if p.LaunchdPlist == "" {
			t.Fatal("launchd plist path required on darwin")
		}
	case "windows":
		if !strings.Contains(p.VerifyBinary, "verify.exe") {
			t.Fatalf("windows binary = %q", p.VerifyBinary)
		}
	default:
		if p.VerifyBinary != "/usr/bin/verify" {
			t.Fatalf("linux binary = %q", p.VerifyBinary)
		}
		if p.SystemdUnit == "" {
			t.Fatal("systemd unit path required on linux")
		}
	}
}

func TestStagedPathsLiveUnderPackagePrefix(t *testing.T) {
	p := StagedPaths("/opt/homebrew")
	if p.VerifyBinary != filepath.Join("/opt/homebrew", "libexec", "rtmx", "verify") {
		t.Fatalf("staged binary = %q", p.VerifyBinary)
	}
	if !strings.Contains(p.LaunchdPlist, "share/rtmx") && !strings.Contains(p.SystemdUnit, "share/rtmx") {
		// At least one unit template is staged under share/rtmx.
		if p.LaunchdPlist == "" && p.SystemdUnit == "" && p.WindowsServiceJSON == "" {
			t.Fatal("expected staged unit templates under share/rtmx")
		}
	}
}
