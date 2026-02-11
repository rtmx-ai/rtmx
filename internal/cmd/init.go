package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	initForce  bool
	initLegacy bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize RTM structure in a new project",
	Long: `Initialize the RTM (Requirements Traceability Matrix) structure.

Creates the necessary directory structure and sample files to start
tracking requirements with RTMX.

By default, creates the modern .rtmx/ directory structure:
  .rtmx/
  ├── database.csv        # Requirements database
  ├── config.yaml         # Configuration
  ├── requirements/       # Requirement specification files
  │   └── EXAMPLE/
  │       └── REQ-EX-001.md
  └── cache/              # Cache directory (gitignored)

Use --legacy to create the older docs/ structure instead.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite existing files")
	initCmd.Flags().BoolVar(&initLegacy, "legacy", false, "use legacy docs/ directory structure")

	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if initLegacy {
		return initLegacyStructure(cmd, cwd)
	}
	return initRtmxStructure(cmd, cwd)
}

func initRtmxStructure(cmd *cobra.Command, cwd string) error {
	rtmxDir := filepath.Join(cwd, ".rtmx")
	rtmCSV := filepath.Join(rtmxDir, "database.csv")
	requirementsDir := filepath.Join(rtmxDir, "requirements")
	configFile := filepath.Join(rtmxDir, "config.yaml")
	cacheDir := filepath.Join(rtmxDir, "cache")
	gitignore := filepath.Join(rtmxDir, ".gitignore")

	// Check for existing files
	if !initForce {
		if _, err := os.Stat(rtmxDir); err == nil {
			cmd.Printf("%s The following already exist:\n", output.Color("Warning:", output.Yellow))
			cmd.Printf("  %s\n\n", rtmxDir)
			cmd.Printf("%s\n", output.Color("Use --force to overwrite", output.Dim))
			os.Exit(1)
		}
	}

	// Create directories
	cmd.Printf("Creating RTM structure in %s/.rtmx/\n\n", cwd)

	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		return fmt.Errorf("failed to create .rtmx directory: %w", err)
	}
	if err := os.MkdirAll(requirementsDir, 0755); err != nil {
		return fmt.Errorf("failed to create requirements directory: %w", err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create .gitignore
	gitignoreContent := `# RTMX cache and generated files
cache/
`
	if err := os.WriteFile(gitignore, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), gitignore)

	// Create sample RTM database
	sampleRTM := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-EX-001,EXAMPLE,SAMPLE,Sample requirement for demonstration,Target value here,tests/test_example.py,test_sample,Unit Test,MISSING,MEDIUM,1,This is a sample requirement,1.0,,,developer,v0.1,,,.rtmx/requirements/EXAMPLE/REQ-EX-001.md
`
	if err := os.WriteFile(rtmCSV, []byte(sampleRTM), 0644); err != nil {
		return fmt.Errorf("failed to create database.csv: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), rtmCSV)

	// Create sample requirement file
	sampleReqDir := filepath.Join(requirementsDir, "EXAMPLE")
	if err := os.MkdirAll(sampleReqDir, 0755); err != nil {
		return fmt.Errorf("failed to create EXAMPLE directory: %w", err)
	}

	sampleReqFile := filepath.Join(sampleReqDir, "REQ-EX-001.md")
	sampleReqContent := `# REQ-EX-001: Sample Requirement

## Description
This is a sample requirement demonstrating the RTMX requirement file format.

## Target
**Metric**: Target value here

## Acceptance Criteria
- [ ] Achieves target value
- [ ] Test implemented and passing
- [ ] Documentation complete

## Implementation
- **Status**: MISSING
- **Phase**: 1
- **Priority**: MEDIUM

## Validation
- **Test**: tests/test_example.py::test_sample
- **Method**: Unit Test

## Dependencies
None

## Notes
This is a sample requirement. Replace with your actual requirements.
`
	if err := os.WriteFile(sampleReqFile, []byte(sampleReqContent), 0644); err != nil {
		return fmt.Errorf("failed to create sample requirement: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), sampleReqFile)

	// Create config file
	configContent := `# RTMX Configuration
# See https://rtmx.ai for documentation

rtmx:
  database: .rtmx/database.csv
  requirements_dir: .rtmx/requirements
  schema: core
  pytest:
    marker_prefix: "req"
    register_markers: true
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config.yaml: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), configFile)

	cmd.Println()
	cmd.Printf("%s\n", output.Color("✓ RTM initialized successfully!", output.Green))
	cmd.Println()
	cmd.Println("Next steps:")
	cmd.Printf("  1. Edit %s to add your requirements\n", rtmCSV)
	cmd.Printf("  2. Create requirement spec files in %s\n", requirementsDir)
	cmd.Println("  3. Run 'rtmx status' to see progress")

	return nil
}

func initLegacyStructure(cmd *cobra.Command, cwd string) error {
	rtmCSV := filepath.Join(cwd, "docs", "rtm_database.csv")
	requirementsDir := filepath.Join(cwd, "docs", "requirements")
	configFile := filepath.Join(cwd, "rtmx.yaml")

	// Check for existing files
	if !initForce {
		var existing []string
		if _, err := os.Stat(rtmCSV); err == nil {
			existing = append(existing, rtmCSV)
		}
		if _, err := os.Stat(configFile); err == nil {
			existing = append(existing, configFile)
		}

		if len(existing) > 0 {
			cmd.Printf("%s The following files already exist:\n", output.Color("Warning:", output.Yellow))
			for _, f := range existing {
				cmd.Printf("  %s\n", f)
			}
			cmd.Println()
			cmd.Printf("%s\n", output.Color("Use --force to overwrite", output.Dim))
			os.Exit(1)
		}
	}

	// Create directories
	cmd.Printf("Creating RTM structure in %s\n\n", cwd)

	if err := os.MkdirAll(filepath.Dir(rtmCSV), 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}
	if err := os.MkdirAll(requirementsDir, 0755); err != nil {
		return fmt.Errorf("failed to create requirements directory: %w", err)
	}

	// Create sample RTM database
	sampleRTM := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-EX-001,EXAMPLE,SAMPLE,Sample requirement for demonstration,Target value here,tests/test_example.py,test_sample,Unit Test,MISSING,MEDIUM,1,This is a sample requirement,1.0,,,developer,v0.1,,,docs/requirements/EXAMPLE/REQ-EX-001.md
`
	if err := os.WriteFile(rtmCSV, []byte(sampleRTM), 0644); err != nil {
		return fmt.Errorf("failed to create rtm_database.csv: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), rtmCSV)

	// Create sample requirement file
	sampleReqDir := filepath.Join(requirementsDir, "EXAMPLE")
	if err := os.MkdirAll(sampleReqDir, 0755); err != nil {
		return fmt.Errorf("failed to create EXAMPLE directory: %w", err)
	}

	sampleReqFile := filepath.Join(sampleReqDir, "REQ-EX-001.md")
	sampleReqContent := `# REQ-EX-001: Sample Requirement

## Description
This is a sample requirement demonstrating the RTMX requirement file format.

## Target
**Metric**: Target value here

## Acceptance Criteria
- [ ] Achieves target value
- [ ] Test implemented and passing
- [ ] Documentation complete

## Implementation
- **Status**: MISSING
- **Phase**: 1
- **Priority**: MEDIUM

## Validation
- **Test**: tests/test_example.py::test_sample
- **Method**: Unit Test

## Dependencies
None

## Notes
This is a sample requirement. Replace with your actual requirements.
`
	if err := os.WriteFile(sampleReqFile, []byte(sampleReqContent), 0644); err != nil {
		return fmt.Errorf("failed to create sample requirement: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), sampleReqFile)

	// Create config file
	configContent := `# RTMX Configuration
# See https://rtmx.ai for documentation

rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core
  pytest:
    marker_prefix: "req"
    register_markers: true
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create rtmx.yaml: %w", err)
	}
	cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), configFile)

	cmd.Println()
	cmd.Printf("%s\n", output.Color("✓ RTM initialized successfully!", output.Green))
	cmd.Println()
	cmd.Println("Next steps:")
	cmd.Printf("  1. Edit %s to add your requirements\n", rtmCSV)
	cmd.Printf("  2. Create requirement spec files in %s\n", requirementsDir)
	cmd.Println("  3. Run 'rtmx status' to see progress")

	return nil
}
