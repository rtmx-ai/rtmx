package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/spf13/cobra"
)

var docsOutput string

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation from RTMX internals",
	Long: `Auto-generate schema and configuration reference documentation.

Examples:
    rtmx docs schema                # Generate schema.md
    rtmx docs config                # Generate config.md
    rtmx docs schema -o docs/       # Custom output location`,
}

var docsSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate schema documentation",
	RunE:  runDocsSchema,
}

var docsConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate configuration reference",
	RunE:  runDocsConfig,
}

func init() {
	docsCmd.PersistentFlags().StringVarP(&docsOutput, "output", "o", "", "output directory or file")

	docsCmd.AddCommand(docsSchemaCmd)
	docsCmd.AddCommand(docsConfigCmd)
	rootCmd.AddCommand(docsCmd)
}

func runDocsSchema(cmd *cobra.Command, args []string) error {
	content := generateSchemaDoc()

	if docsOutput != "" {
		outPath := docsOutput
		if info, err := os.Stat(docsOutput); err == nil && info.IsDir() {
			outPath = filepath.Join(docsOutput, "schema.md")
		}
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write schema doc: %w", err)
		}
		cmd.Printf("Written to %s\n", outPath)
		return nil
	}

	cmd.Print(content)
	return nil
}

func runDocsConfig(cmd *cobra.Command, args []string) error {
	content := generateConfigDoc()

	if docsOutput != "" {
		outPath := docsOutput
		if info, err := os.Stat(docsOutput); err == nil && info.IsDir() {
			outPath = filepath.Join(docsOutput, "config.md")
		}
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write config doc: %w", err)
		}
		cmd.Printf("Written to %s\n", outPath)
		return nil
	}

	cmd.Print(content)
	return nil
}

func generateSchemaDoc() string {
	var sb strings.Builder

	sb.WriteString("# RTMX Database Schema\n\n")
	sb.WriteString("This document describes the RTM database CSV schema.\n\n")

	sb.WriteString("## Core Schema Fields\n\n")
	sb.WriteString("| Field | Type | Required | Description |\n")
	sb.WriteString("|-------|------|----------|-------------|\n")
	sb.WriteString("| req_id | string | Yes | Unique requirement identifier (e.g., REQ-FEAT-001) |\n")
	sb.WriteString("| category | string | Yes | Requirement category (e.g., CLI, DATA, API) |\n")
	sb.WriteString("| subcategory | string | No | Optional subcategory for grouping |\n")
	sb.WriteString("| requirement_text | string | Yes | Full requirement description |\n")
	sb.WriteString("| target_value | string | No | Measurable target or acceptance criteria |\n")
	sb.WriteString("| test_module | string | No | Test file path |\n")
	sb.WriteString("| test_function | string | No | Test function name |\n")
	sb.WriteString("| validation_method | string | No | Type of test (Unit, Integration, System) |\n")
	sb.WriteString("| status | enum | Yes | COMPLETE, PARTIAL, MISSING, NOT_STARTED |\n")
	sb.WriteString("| priority | enum | Yes | P0, HIGH, MEDIUM, LOW |\n")
	sb.WriteString("| phase | integer | No | Development phase number |\n")
	sb.WriteString("| notes | string | No | Additional notes |\n")
	sb.WriteString("| effort_weeks | float | No | Estimated effort in weeks |\n")
	sb.WriteString("| dependencies | string | No | Pipe-separated list of blocking requirements |\n")
	sb.WriteString("| blocks | string | No | Pipe-separated list of requirements this blocks |\n")
	sb.WriteString("| assignee | string | No | Assigned developer |\n")
	sb.WriteString("| sprint | string | No | Sprint identifier |\n")
	sb.WriteString("| started_date | date | No | Date work started |\n")
	sb.WriteString("| completed_date | date | No | Date work completed |\n")
	sb.WriteString("| requirement_file | string | No | Path to detailed requirement spec |\n")
	sb.WriteString("| external_id | string | No | External tracker ID (GitHub, Jira) |\n")
	sb.WriteString("\n")

	sb.WriteString("## Status Values\n\n")
	sb.WriteString("| Status | Completion % | Description |\n")
	sb.WriteString("|--------|-------------|-------------|\n")
	sb.WriteString("| COMPLETE | 100% | Fully implemented and tested |\n")
	sb.WriteString("| PARTIAL | 50% | Partially implemented |\n")
	sb.WriteString("| MISSING | 0% | Not yet implemented |\n")
	sb.WriteString("| NOT_STARTED | 0% | Not yet started |\n")
	sb.WriteString("\n")

	sb.WriteString("## Priority Values\n\n")
	sb.WriteString("| Priority | Weight | Description |\n")
	sb.WriteString("|----------|--------|-------------|\n")
	sb.WriteString("| P0 | 0 | Critical, must be done first |\n")
	sb.WriteString("| HIGH | 1 | High priority |\n")
	sb.WriteString("| MEDIUM | 2 | Medium priority |\n")
	sb.WriteString("| LOW | 3 | Low priority |\n")
	sb.WriteString("\n")

	return sb.String()
}

func generateConfigDoc() string {
	var sb strings.Builder

	sb.WriteString("# RTMX Configuration Reference\n\n")
	sb.WriteString("This document describes the RTMX configuration file format.\n\n")

	sb.WriteString("## Configuration Files\n\n")
	sb.WriteString("RTMX looks for configuration in the following locations:\n\n")
	sb.WriteString("1. `.rtmx/config.yaml` (recommended)\n")
	sb.WriteString("2. `rtmx.yaml`\n")
	sb.WriteString("3. `.rtmx.yaml`\n")
	sb.WriteString("\n")

	sb.WriteString("## Configuration Structure\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString("rtmx:\n")
	sb.WriteString("  # Path to the RTM database CSV file\n")
	sb.WriteString("  database: .rtmx/database.csv\n")
	sb.WriteString("\n")
	sb.WriteString("  # Directory containing requirement specification files\n")
	sb.WriteString("  requirements_dir: .rtmx/requirements\n")
	sb.WriteString("\n")
	sb.WriteString("  # Schema type: core, phoenix, or custom\n")
	sb.WriteString("  schema: core\n")
	sb.WriteString("\n")
	sb.WriteString("  # Phase definitions with descriptions\n")
	sb.WriteString("  phases:\n")
	sb.WriteString("    1: \"Foundation\"\n")
	sb.WriteString("    2: \"Core Features\"\n")
	sb.WriteString("    3: \"Integration\"\n")
	sb.WriteString("\n")
	sb.WriteString("  # Pytest plugin configuration\n")
	sb.WriteString("  pytest:\n")
	sb.WriteString("    marker_prefix: \"req\"\n")
	sb.WriteString("    register_markers: true\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Fields Reference\n\n")
	sb.WriteString("| Field | Type | Default | Description |\n")
	sb.WriteString("|-------|------|---------|-------------|\n")
	sb.WriteString("| database | string | .rtmx/database.csv | Path to RTM database |\n")
	sb.WriteString("| requirements_dir | string | .rtmx/requirements | Path to requirement specs |\n")
	sb.WriteString("| schema | string | core | Schema type |\n")
	sb.WriteString("| phases | map[int]string | {} | Phase number to name mapping |\n")
	sb.WriteString("| pytest.marker_prefix | string | req | Pytest marker prefix |\n")
	sb.WriteString("| pytest.register_markers | bool | true | Auto-register pytest markers |\n")
	sb.WriteString("\n")

	// Load current config to show example
	if cwd, err := os.Getwd(); err == nil {
		if cfg, err := config.LoadFromDir(cwd); err == nil {
			sb.WriteString("## Current Configuration\n\n")
			sb.WriteString(fmt.Sprintf("- Database: `%s`\n", cfg.RTMX.Database))
			sb.WriteString(fmt.Sprintf("- Requirements Dir: `%s`\n", cfg.RequirementsPath(cwd)))
			sb.WriteString(fmt.Sprintf("- Schema: `%s`\n", cfg.RTMX.Schema))
			sb.WriteString(fmt.Sprintf("- Phases: %d defined\n", len(cfg.RTMX.Phases)))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
