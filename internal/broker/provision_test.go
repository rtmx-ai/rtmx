package broker

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestInstallPlanApplyStagesBrokerFromPackage(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069j")

	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "verify")
	if err := os.WriteFile(src, []byte("#!/bin/sh\necho broker\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	paths := DefaultPackagePaths()
	plan := InstallPlan{Layout: InstallLayout{
		Root:         root,
		SourceBinary: src,
		Paths:        paths,
		Command:      []string{"go", "test", "./..."},
	}}
	if err := plan.Apply(); err != nil {
		t.Fatal(err)
	}
	bin := resolvePath(root, paths.VerifyBinary)
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("binary missing: %v", err)
	}
	key := resolvePath(root, paths.KeyPath)
	if _, err := os.Stat(key); err != nil {
		t.Fatalf("key missing: %v", err)
	}
	switch runtime.GOOS {
	case "darwin":
		body, err := os.ReadFile(resolvePath(root, paths.LaunchdPlist))
		if err != nil {
			t.Fatal(err)
		}
		s := string(body)
		if !strings.Contains(s, "rtmx-verify") || !strings.Contains(s, "RunAtLoad") {
			t.Fatalf("plist = %s", s)
		}
		if !strings.Contains(s, paths.VerifyBinary) {
			t.Fatalf("plist missing binary path")
		}
	case "windows":
		// contract written
	default:
		body, err := os.ReadFile(resolvePath(root, paths.SystemdUnit))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "WantedBy=multi-user.target") {
			t.Fatalf("unit = %s", body)
		}
	}
}

func TestRenderLaunchdIncludesKeyPath(t *testing.T) {
	p := DefaultPackagePaths()
	body := RenderLaunchdPlist(p, nil)
	if !strings.Contains(body, "/var/lib/rtmx/verify.key") && runtime.GOOS == "darwin" {
		// On non-darwin DefaultPackagePaths still uses unix key path except windows.
		if runtime.GOOS != "windows" && !strings.Contains(body, p.KeyPath) {
			t.Fatalf("plist missing key: %s", body)
		}
	}
}
