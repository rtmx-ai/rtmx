package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestInitInstallsDeliveryRule(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-021")
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	initForce = false
	initLegacy = false
	initDryRun = false
	defer func() {
		initForce = false
		initDryRun = false
	}()

	buf := new(bytes.Buffer)
	root := newTestRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	delivery, err := os.ReadFile(filepath.Join(dir, ".rtmx", "agent", "delivery.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(delivery)
	if !strings.Contains(text, "One pull request per requirement") || !strings.Contains(text, "One commit per acceptance criterion") {
		t.Fatalf("delivery.md missing rule:\n%s", text)
	}
	mdc, err := os.ReadFile(filepath.Join(dir, ".cursor", "rules", "rtmx-delivery.mdc"))
	if err != nil || !strings.Contains(string(mdc), "One pull request per requirement") {
		t.Fatalf("cursor rule: %v\n%s", err, mdc)
	}
	claude, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil || !strings.Contains(string(claude), deliveryRuleStart) || !strings.Contains(string(claude), deliveryRuleEnd) {
		t.Fatalf("CLAUDE.md block: %v\n%s", err, claude)
	}
	if _, err := os.Stat(filepath.Join(dir, "Library", "LaunchAgents")); err == nil {
		t.Fatal("init must not install a LaunchAgent")
	}
	if strings.Contains(buf.String(), "systemd") {
		t.Fatal("init must not mention installing a loop unit")
	}

	// Force path replaces the block and keeps surrounding text.
	preamble := "# Project notes\nKeep this line.\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(preamble+"\n"+string(claude)), 0644); err != nil {
		t.Fatal(err)
	}
	initForce = true
	buf.Reset()
	root = newTestRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--force"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "Keep this line.") {
		t.Fatalf("force init clobbered CLAUDE.md:\n%s", after)
	}
	if strings.Count(string(after), deliveryRuleStart) != 1 {
		t.Fatalf("expected one delivery block:\n%s", after)
	}
}

func TestInitDryRunWritesNothing(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-021")
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	initDryRun = true
	initForce = false
	initLegacy = false
	defer func() { initDryRun = false }()

	buf := new(bytes.Buffer)
	root := newTestRootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "delivery.md") || !strings.Contains(out, "Dry run") {
		t.Fatalf("dry-run output: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".rtmx")); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote .rtmx")
	}
}
