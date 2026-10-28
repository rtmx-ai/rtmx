package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestApplyLinkChoice(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	tests := []struct {
		id     string
		wantOK bool
		check  func(t *testing.T)
	}{
		{"github", true, func(t *testing.T) {
			if !bootstrapFromGitHub || bootstrapFromJira {
				t.Fatal("expected github only")
			}
		}},
		{"jira", true, func(t *testing.T) {
			if !bootstrapFromJira {
				t.Fatal("expected jira")
			}
		}},
		{"asana", true, func(t *testing.T) {
			if !bootstrapFromAsana {
				t.Fatal("expected asana")
			}
		}},
		{"monday", true, func(t *testing.T) {
			if !bootstrapFromMonday {
				t.Fatal("expected monday")
			}
		}},
		{"gitlab", true, func(t *testing.T) {
			if !bootstrapFromGitLab {
				t.Fatal("expected gitlab")
			}
		}},
		{"tests", true, func(t *testing.T) {
			if !bootstrapFromTests {
				t.Fatal("expected tests")
			}
		}},
		{"skip", false, func(t *testing.T) {
			if bootstrapSourceSelected() {
				t.Fatal("skip should clear sources")
			}
		}},
		{"", false, nil},
		{"unknown", false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			defer resetBootstrapFlags()()
			ok := applyLinkChoice(tt.id)
			if ok != tt.wantOK {
				t.Fatalf("ok=%v want %v", ok, tt.wantOK)
			}
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestLinkPickerInvokesBootstrapSource(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	defer resetBootstrapFlags()()

	origPicker := runProviderPicker
	origRemotes := listGitRemotes
	defer func() {
		runProviderPicker = origPicker
		listGitRemotes = origRemotes
	}()

	listGitRemotes = func(dir string) ([]GitRemote, error) {
		return []GitRemote{{Name: "origin", URL: "https://github.com/acme/app.git"}}, nil
	}
	runProviderPicker = func(hint string) (string, error) {
		if hint != "github" {
			t.Fatalf("hint=%q want github", hint)
		}
		return "github", nil
	}

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := selectBootstrapSourceInteractive(cmd); err != nil {
		t.Fatal(err)
	}
	if !bootstrapFromGitHub {
		t.Fatal("expected github flag set from picker")
	}
	if !strings.Contains(out.String(), "github") {
		t.Fatalf("want inferred hint message, got %q", out.String())
	}
}

func TestLinkSkipExitsClean(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	defer resetBootstrapFlags()()

	origPicker := runProviderPicker
	origRemotes := listGitRemotes
	origTerm := isTerminalFn
	defer func() {
		runProviderPicker = origPicker
		listGitRemotes = origRemotes
		isTerminalFn = origTerm
	}()

	listGitRemotes = func(string) ([]GitRemote, error) { return nil, nil }
	runProviderPicker = func(string) (string, error) { return "skip", nil }
	isTerminalFn = func() bool { return true }

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := runLink(cmd, nil); err != nil {
		t.Fatalf("skip should exit 0: %v", err)
	}
	if !strings.Contains(out.String(), "Skipped") {
		t.Fatalf("want Skipped message, got %q", out.String())
	}
	if bootstrapSourceSelected() {
		t.Fatal("sources should remain clear after skip")
	}
}

func TestLinkRequiresTTY(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	defer resetBootstrapFlags()()

	origTerm := isTerminalFn
	defer func() { isTerminalFn = origTerm }()
	isTerminalFn = func() bool { return false }

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := runLink(cmd, nil)
	if err == nil {
		t.Fatal("expected error on non-TTY")
	}
	if !strings.Contains(out.String(), "TTY") {
		t.Fatalf("want TTY message, got %q", out.String())
	}
}

func TestBootstrapFlaglessNonTTYStillErrors(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	defer resetBootstrapFlags()()

	origTerm := isTerminalFn
	defer func() { isTerminalFn = origTerm }()
	isTerminalFn = func() bool { return false }

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := runBootstrap(cmd, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(buf.String(), "No source specified") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestBootstrapFlaglessTTYPickerDryRun(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	defer resetBootstrapFlags()()

	origPicker := runProviderPicker
	origRemotes := listGitRemotes
	origTerm := isTerminalFn
	defer func() {
		runProviderPicker = origPicker
		listGitRemotes = origRemotes
		isTerminalFn = origTerm
	}()

	isTerminalFn = func() bool { return true }
	listGitRemotes = func(string) ([]GitRemote, error) {
		return []GitRemote{{Name: "origin", URL: "git@gitlab.com:acme/x.git"}}, nil
	}
	runProviderPicker = func(hint string) (string, error) {
		if hint != "gitlab" {
			t.Fatalf("hint=%q want gitlab", hint)
		}
		return "tests", nil
	}
	bootstrapDryRun = true

	dir := t.TempDir()
	t.Chdir(dir)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := runBootstrap(cmd, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRY RUN") {
		t.Fatalf("want DRY RUN, got %q", out)
	}
	if !bootstrapFromTests {
		t.Fatal("expected tests source from picker")
	}
}

func TestParseGitRemoteV(t *testing.T) {
	out := parseGitRemoteV(`
origin	https://github.com/acme/app.git (fetch)
origin	https://github.com/acme/app.git (push)
upstream	https://gitlab.com/acme/x.git (fetch)
`)
	if len(out) != 2 {
		t.Fatalf("len=%d want 2: %+v", len(out), out)
	}
	if out[0].Name != "origin" || !strings.Contains(out[0].URL, "github.com") {
		t.Fatalf("origin=%+v", out[0])
	}
}

func TestLinkCommandRegistered(t *testing.T) {
	rtmx.Req(t, "REQ-GO-089")
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "link" {
			found = true
			if !strings.Contains(c.Long, "issue store") {
				t.Fatalf("link help should document providers: %q", c.Long)
			}
			break
		}
	}
	if !found {
		t.Fatal("rtmx link not registered")
	}
}
