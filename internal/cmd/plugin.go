package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/internal/schema"
	"github.com/spf13/cobra"
)

var (
	pluginGlobal bool
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage schema plugins",
	Long: `Install, list, and remove schema plugins.

Plugins add domain-specific columns to the RTMX database. Built-in schemas
(core, phoenix, do178c, iso26262) are always available. Additional schemas
can be installed from files or Git repositories.

Examples:
    rtmx plugin list                          # list installed schemas
    rtmx plugin install ./myschema.yaml       # install from file
    rtmx plugin install --global ./s.yaml     # install globally
    rtmx plugin remove myschema               # remove a plugin`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available schema plugins",
	RunE:  runPluginList,
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a schema plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginInstall,
}

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed schema plugin",
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginRemove,
}

func init() {
	pluginInstallCmd.Flags().BoolVar(&pluginGlobal, "global", false, "install to user-global schemas directory")
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	rootCmd.AddCommand(pluginCmd)
}

func runPluginList(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	names := schema.Names()
	builtins := map[string]bool{"core": true, "phoenix": true, "do178c": true, "iso26262": true}

	width := output.TerminalWidth()
	cmd.Println(output.Header("Schema Plugins", width))
	cmd.Println()

	for _, name := range names {
		s := schema.Get(name)
		if s == nil {
			continue
		}
		source := "built-in"
		if !builtins[name] {
			source = "installed"
		}
		cmd.Printf("  %-15s  %-10s  %d columns  %s\n",
			output.Color(name, output.Cyan),
			source,
			len(s.Columns),
			s.Description)
	}

	cmd.Println()

	// Check for locally installed plugins
	cwd, err := os.Getwd()
	if err == nil {
		localDir := filepath.Join(cwd, ".rtmx", "schemas")
		if entries, err := os.ReadDir(localDir); err == nil {
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
					name := strings.TrimSuffix(strings.TrimSuffix(e.Name(), ".yaml"), ".yml")
					if schema.Get(name) == nil {
						cmd.Printf("  %-15s  %-10s  (not loaded)\n",
							output.Color(name, output.Dim),
							"local")
					}
				}
			}
		}
	}

	return nil
}

func runPluginInstall(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Determine the target directory
	var targetDir string
	if pluginGlobal {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		targetDir = filepath.Join(home, ".rtmx", "schemas")
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}
		targetDir = filepath.Join(cwd, ".rtmx", "schemas")
	}

	// Check source is a YAML file
	if !strings.HasSuffix(source, ".yaml") && !strings.HasSuffix(source, ".yml") {
		return fmt.Errorf("source must be a YAML file (got %q)", source)
	}

	// Validate the schema file can be loaded
	s, err := schema.LoadCustomSchema(source)
	if err != nil {
		return fmt.Errorf("invalid schema file: %w", err)
	}

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("cannot create schemas directory: %w", err)
	}

	// Read source and write to target
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("cannot read source: %w", err)
	}

	targetFile := filepath.Join(targetDir, filepath.Base(source))
	if err := os.WriteFile(targetFile, data, 0644); err != nil {
		return fmt.Errorf("cannot write plugin: %w", err)
	}

	scope := "project"
	if pluginGlobal {
		scope = "global"
	}
	cmd.Printf("Installed schema %q (%d columns) to %s schemas\n", s.Name, len(s.Columns), scope)
	cmd.Printf("  Path: %s\n", targetFile)

	return nil
}

func runPluginRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	builtins := map[string]bool{"core": true, "phoenix": true, "do178c": true, "iso26262": true}
	if builtins[name] {
		return fmt.Errorf("cannot remove built-in schema %q", name)
	}

	// Search in project and global directories
	cwd, _ := os.Getwd()
	searchDirs := []string{filepath.Join(cwd, ".rtmx", "schemas")}

	home, err := os.UserHomeDir()
	if err == nil {
		searchDirs = append(searchDirs, filepath.Join(home, ".rtmx", "schemas"))
	}

	var removed []string
	for _, dir := range searchDirs {
		for _, ext := range []string{".yaml", ".yml"} {
			path := filepath.Join(dir, name+ext)
			if _, err := os.Stat(path); err == nil {
				if err := os.Remove(path); err != nil {
					return fmt.Errorf("failed to remove %s: %w", path, err)
				}
				removed = append(removed, path)
			}
		}
	}

	sort.Strings(removed)
	if len(removed) == 0 {
		return fmt.Errorf("schema plugin %q not found", name)
	}

	for _, path := range removed {
		cmd.Printf("Removed: %s\n", path)
	}

	return nil
}
