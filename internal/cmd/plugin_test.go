package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func newTestPluginCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	plugin := &cobra.Command{
		Use: "plugin",
	}

	var global bool

	list := &cobra.Command{
		Use:  "list",
		RunE: runPluginList,
	}

	install := &cobra.Command{
		Use:  "install",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginGlobal = global
			return runPluginInstall(cmd, args)
		},
	}
	install.Flags().BoolVar(&global, "global", false, "")

	remove := &cobra.Command{
		Use:  "remove",
		Args: cobra.ExactArgs(1),
		RunE: runPluginRemove,
	}

	plugin.AddCommand(list, install, remove)
	root.AddCommand(plugin)
	return root
}

func TestPluginInstall(t *testing.T) {
	rtmx.Req(t, "REQ-PLUGIN-007")

	t.Run("install_from_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(origDir) }()

		// Create project structure
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)
		_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"),
			[]byte("rtmx:\n  database: .rtmx/database.csv\n  schema: core\n"), 0644)

		// Create a valid schema YAML file
		schemaYAML := `name: fedramp
extends: core
columns:
  - name: control_id
    type: string
    description: "NIST control identifier"
  - name: impact_level
    type: enum
    description: "FedRAMP impact level"
    values: [low, moderate, high]
`
		schemaFile := filepath.Join(tmpDir, "fedramp.yaml")
		_ = os.WriteFile(schemaFile, []byte(schemaYAML), 0644)

		_ = os.Chdir(tmpDir)

		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "install", schemaFile})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("plugin install failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Installed") {
			t.Errorf("expected 'Installed' in output, got:\n%s", out)
		}
		if !strings.Contains(out, "fedramp") {
			t.Errorf("expected schema name in output, got:\n%s", out)
		}

		// Verify file was copied to .rtmx/schemas/
		installed := filepath.Join(tmpDir, ".rtmx", "schemas", "fedramp.yaml")
		if _, err := os.Stat(installed); err != nil {
			t.Errorf("schema file should exist at %s: %v", installed, err)
		}
	})

	t.Run("install_invalid_extension", func(t *testing.T) {
		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "install", "notayaml.txt"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for non-YAML source")
		}
		if !strings.Contains(err.Error(), "YAML") {
			t.Errorf("expected YAML error, got: %v", err)
		}
	})

	t.Run("install_invalid_schema", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(origDir) }()
		_ = os.Chdir(tmpDir)

		// Create invalid schema YAML (references nonexistent base)
		bad := filepath.Join(tmpDir, "bad.yaml")
		_ = os.WriteFile(bad, []byte("name: bad\nextends: nonexistent\ncolumns: []\n"), 0644)

		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "install", bad})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid schema")
		}
	})
}

func TestPluginList(t *testing.T) {
	rtmx.Req(t, "REQ-PLUGIN-007")

	t.Run("lists_builtins", func(t *testing.T) {
		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "list"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("plugin list failed: %v", err)
		}

		out := buf.String()
		for _, name := range []string{"core", "phoenix", "do178c", "iso26262"} {
			if !strings.Contains(out, name) {
				t.Errorf("expected %q in list output, got:\n%s", name, out)
			}
		}
		if !strings.Contains(out, "built-in") {
			t.Errorf("expected 'built-in' marker in output, got:\n%s", out)
		}
	})
}

func TestPluginRemove(t *testing.T) {
	rtmx.Req(t, "REQ-PLUGIN-007")

	t.Run("cannot_remove_builtin", func(t *testing.T) {
		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "remove", "core"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error removing built-in schema")
		}
		if !strings.Contains(err.Error(), "built-in") {
			t.Errorf("expected 'built-in' in error, got: %v", err)
		}
	})

	t.Run("removes_installed_plugin", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(origDir) }()
		_ = os.Chdir(tmpDir)

		// Create a fake installed plugin
		schemasDir := filepath.Join(tmpDir, ".rtmx", "schemas")
		_ = os.MkdirAll(schemasDir, 0755)
		_ = os.WriteFile(filepath.Join(schemasDir, "custom.yaml"), []byte("name: custom\n"), 0644)

		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "remove", "custom"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("plugin remove failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Removed") {
			t.Errorf("expected 'Removed' in output, got:\n%s", out)
		}

		// Verify file was deleted
		if _, err := os.Stat(filepath.Join(schemasDir, "custom.yaml")); !os.IsNotExist(err) {
			t.Error("plugin file should have been deleted")
		}
	})

	t.Run("not_found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(origDir) }()
		_ = os.Chdir(tmpDir)

		cmd := newTestPluginCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plugin", "remove", "nonexistent"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for nonexistent plugin")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected 'not found' in error, got: %v", err)
		}
	})
}
