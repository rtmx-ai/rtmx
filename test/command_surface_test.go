package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/cmd"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestCommandSurface validates that every registered command is reachable
// from the binary's --help and that --help works on each command.
// This test is self-maintaining: it introspects rootCmd rather than
// hardcoding a list that rots.
func TestCommandSurface(t *testing.T) {
	rtmx.Req(t, "REQ-GO-020")

	binaryPath := buildTestBinary(t)
	projectRoot := findTestProjectRoot(t)

	// Get all registered top-level commands from the real rootCmd
	registered := cmd.RegisteredCommands()

	t.Run("all_commands_in_help", func(t *testing.T) {
		helpOut, err := exec.Command(binaryPath, "--help").CombinedOutput()
		if err != nil {
			_ = err // --help may exit non-zero on some setups
		}
		help := string(helpOut)

		for _, name := range registered {
			if name == "help" || name == "completion" {
				continue // built-in Cobra commands
			}
			if !strings.Contains(help, name) {
				t.Errorf("command %q registered but not in --help output", name)
			}
		}
	})

	t.Run("help_flag_works", func(t *testing.T) {
		for _, name := range registered {
			if name == "help" || name == "completion" {
				continue
			}
			t.Run(name, func(t *testing.T) {
				out, err := exec.Command(binaryPath, name, "--help").CombinedOutput()
				if err != nil {
					// Some commands may exit non-zero on --help, check output
					if len(out) == 0 {
						t.Errorf("%s --help produced no output and error: %v", name, err)
					}
				}
				if len(out) == 0 {
					t.Errorf("%s --help produced no output", name)
				}
			})
		}
	})

	// Commands that should work against the real project database
	t.Run("nominal_execution", func(t *testing.T) {
		nominalCmds := []struct {
			name     string
			args     []string
			allowErr bool
		}{
			{"status", []string{"status"}, false},
			{"status_json", []string{"status", "--json"}, false},
			{"backlog", []string{"backlog"}, false},
			{"backlog_json", []string{"backlog", "--json"}, false},
			{"health", []string{"health"}, true}, // may exit 1 with warnings
			{"deps", []string{"deps"}, false},
			{"cycles", []string{"cycles"}, false},
			{"context", []string{"context"}, false},
			{"config", []string{"config"}, false},
			{"markers", []string{"markers"}, false},
			{"next", []string{"next"}, false},
			{"next_one", []string{"next", "--one"}, false},
			{"next_one_json", []string{"next", "--one", "--json"}, false},
			{"status_by_version", []string{"status", "--by-version"}, false},
			{"status_verbose", []string{"status", "-vvv"}, false},
			{"security", []string{"security"}, true}, // may exit 1
			{"version", []string{"version"}, false},
		}

		for _, tc := range nominalCmds {
			t.Run(tc.name, func(t *testing.T) {
				c := exec.Command(binaryPath, tc.args...)
				c.Dir = projectRoot
				out, err := c.CombinedOutput()
				if !tc.allowErr && err != nil {
					t.Errorf("%s failed: %v\n%s", tc.name, err, string(out))
				}
				if len(out) == 0 {
					t.Errorf("%s produced no output", tc.name)
				}
			})
		}
	})

	// Commands that require arguments should fail gracefully without them
	t.Run("missing_args_error", func(t *testing.T) {
		argRequired := []struct {
			name string
			args []string
		}{
			{"release_gate", []string{"release", "gate"}},
			{"release_assign", []string{"release", "assign"}},
			{"release_unassign", []string{"release", "unassign"}},
			{"release_scope", []string{"release", "scope"}},
			{"diff", []string{"diff"}},
			{"move", []string{"move"}},
			{"clone", []string{"clone"}},
		}

		for _, tc := range argRequired {
			t.Run(tc.name, func(t *testing.T) {
				c := exec.Command(binaryPath, tc.args...)
				c.Dir = projectRoot
				out, err := c.CombinedOutput()
				if err == nil {
					t.Errorf("%s should fail without required args, but succeeded:\n%s", tc.name, string(out))
				}
			})
		}
	})

	// Subcommand parents should list their children in help
	t.Run("subcommand_help", func(t *testing.T) {
		parents := []struct {
			name     string
			children []string
		}{
			{"release", []string{"gate", "scope", "assign", "unassign"}},
			{"auth", []string{"login", "logout", "status"}},
			{"grant", []string{"create", "list"}},
			{"remote", []string{"add", "list", "remove"}},
		}

		for _, p := range parents {
			t.Run(p.name, func(t *testing.T) {
				out, _ := exec.Command(binaryPath, p.name, "--help").CombinedOutput()
				help := string(out)
				for _, child := range p.children {
					if !strings.Contains(help, child) {
						t.Errorf("%s --help missing subcommand %q", p.name, child)
					}
				}
			})
		}
	})

	// Invalid flags should produce errors, not panics
	t.Run("invalid_flags", func(t *testing.T) {
		invalidCmds := []struct {
			name string
			args []string
		}{
			{"status_bad_flag", []string{"status", "--nonexistent-flag"}},
			{"backlog_bad_flag", []string{"backlog", "--nonexistent-flag"}},
			{"verify_bad_flag", []string{"verify", "--nonexistent-flag"}},
		}

		for _, tc := range invalidCmds {
			t.Run(tc.name, func(t *testing.T) {
				c := exec.Command(binaryPath, tc.args...)
				c.Dir = projectRoot
				out, err := c.CombinedOutput()
				if err == nil {
					t.Errorf("%s should fail with invalid flag", tc.name)
				}
				// Should produce an error message, not a panic
				if bytes.Contains(out, []byte("panic")) {
					t.Errorf("%s panicked:\n%s", tc.name, string(out))
				}
			})
		}
	})

	// Unknown commands should produce helpful error
	t.Run("unknown_command", func(t *testing.T) {
		c := exec.Command(binaryPath, "nonexistent-command")
		out, err := c.CombinedOutput()
		if err == nil {
			t.Error("unknown command should fail")
		}
		if !bytes.Contains(out, []byte("unknown command")) {
			t.Errorf("expected 'unknown command' in error, got:\n%s", string(out))
		}
	})
}

// buildTestBinary builds the rtmx binary and returns the path.
func buildTestBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	name := "rtmx"
	if runtime.GOOS == "windows" {
		name = "rtmx.exe"
	}
	binaryPath := filepath.Join(tmpDir, name)

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	return binaryPath
}

// findTestProjectRoot locates the project root directory.
func findTestProjectRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	root := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(root, "cmd/rtmx")); err != nil {
		root = wd
	}
	return root
}
