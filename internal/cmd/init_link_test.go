package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func resetInitFlags() func() {
	of, ol, od, oy := initForce, initLegacy, initDryRun, initYes
	initForce, initLegacy, initDryRun, initYes = false, false, false, false
	return func() {
		initForce, initLegacy, initDryRun, initYes = of, ol, od, oy
	}
}

func TestInitOffersLinkHandoffOnTTY(t *testing.T) {
	rtmx.Req(t, "REQ-GO-090")
	defer resetInitFlags()()

	origConfirm := confirmLinkOffer
	origHandoff := runLinkHandoffAfterInit
	origOut := isTerminalFn
	origIn := stdinIsTerminal
	handoffCalls := 0
	defer func() {
		confirmLinkOffer = origConfirm
		runLinkHandoffAfterInit = origHandoff
		isTerminalFn = origOut
		stdinIsTerminal = origIn
	}()

	isTerminalFn = func() bool { return true }
	stdinIsTerminal = func() bool { return true }
	confirmLinkOffer = func(*cobra.Command, io.Reader) bool { return true }
	runLinkHandoffAfterInit = func(*cobra.Command) error {
		handoffCalls++
		return nil
	}

	dir := t.TempDir()
	t.Chdir(dir)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := runInit(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if handoffCalls != 1 {
		t.Fatalf("handoff calls=%d want 1", handoffCalls)
	}
	if _, err := os.Stat(filepath.Join(dir, ".rtmx", "database.csv")); err != nil {
		t.Fatal("init files missing after handoff")
	}
	if !strings.Contains(buf.String(), "rtmx link") {
		t.Fatalf("next steps should mention link: %q", buf.String())
	}
}

func TestInitYesSkipsLinkHandoff(t *testing.T) {
	rtmx.Req(t, "REQ-GO-090")
	defer resetInitFlags()()

	origHandoff := runLinkHandoffAfterInit
	origOut := isTerminalFn
	origIn := stdinIsTerminal
	handoffCalls := 0
	defer func() {
		runLinkHandoffAfterInit = origHandoff
		isTerminalFn = origOut
		stdinIsTerminal = origIn
	}()

	initYes = true
	isTerminalFn = func() bool { return true }
	stdinIsTerminal = func() bool { return true }
	runLinkHandoffAfterInit = func(*cobra.Command) error {
		handoffCalls++
		return nil
	}

	dir := t.TempDir()
	t.Chdir(dir)
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := runInit(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if handoffCalls != 0 {
		t.Fatalf("--yes should skip handoff, calls=%d", handoffCalls)
	}
}

func TestInitNonTTYSkipsLinkHandoff(t *testing.T) {
	rtmx.Req(t, "REQ-GO-090")
	defer resetInitFlags()()

	origHandoff := runLinkHandoffAfterInit
	origOut := isTerminalFn
	origIn := stdinIsTerminal
	handoffCalls := 0
	defer func() {
		runLinkHandoffAfterInit = origHandoff
		isTerminalFn = origOut
		stdinIsTerminal = origIn
	}()

	isTerminalFn = func() bool { return false }
	stdinIsTerminal = func() bool { return false }
	runLinkHandoffAfterInit = func(*cobra.Command) error {
		handoffCalls++
		return nil
	}

	dir := t.TempDir()
	t.Chdir(dir)
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := runInit(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if handoffCalls != 0 {
		t.Fatalf("non-TTY should skip handoff, calls=%d", handoffCalls)
	}
}

func TestInitKeepsFilesWhenHandoffFails(t *testing.T) {
	rtmx.Req(t, "REQ-GO-090")
	defer resetInitFlags()()

	origConfirm := confirmLinkOffer
	origHandoff := runLinkHandoffAfterInit
	origOut := isTerminalFn
	origIn := stdinIsTerminal
	defer func() {
		confirmLinkOffer = origConfirm
		runLinkHandoffAfterInit = origHandoff
		isTerminalFn = origOut
		stdinIsTerminal = origIn
	}()

	isTerminalFn = func() bool { return true }
	stdinIsTerminal = func() bool { return true }
	confirmLinkOffer = func(*cobra.Command, io.Reader) bool { return true }
	runLinkHandoffAfterInit = func(*cobra.Command) error {
		return errors.New("bootstrap exploded")
	}

	dir := t.TempDir()
	t.Chdir(dir)

	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := runInit(cmd, nil); err != nil {
		t.Fatalf("init should succeed despite handoff failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".rtmx", "database.csv")); err != nil {
		t.Fatal("init files must remain after handoff failure")
	}
	if !strings.Contains(buf.String(), "Link handoff failed") {
		t.Fatalf("want warning, got %q", buf.String())
	}
}

func TestConfirmLinkOfferYesNo(t *testing.T) {
	rtmx.Req(t, "REQ-GO-090")
	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	if !confirmLinkOffer(cmd, strings.NewReader("y\n")) {
		t.Fatal("y should confirm")
	}
	if confirmLinkOffer(cmd, strings.NewReader("n\n")) {
		t.Fatal("n should decline")
	}
}
