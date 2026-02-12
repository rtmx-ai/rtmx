package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configValidate bool
	configFormat   string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or validate RTMX configuration",
	Long: `Display the effective configuration after merging defaults with rtmx.yaml.

Examples:
    rtmx config                     # Show current config
    rtmx config --validate          # Check config validity
    rtmx config --format yaml       # Output as YAML`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVar(&configValidate, "validate", false, "validate configuration and check paths")
	configCmd.Flags().StringVar(&configFormat, "format", "terminal", "output format: terminal, yaml, json")

	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if configValidate {
		return validateConfig(cmd, cfg, cwd)
	}

	return displayConfig(cmd, cfg, cwd)
}

func validateConfig(cmd *cobra.Command, cfg *config.Config, cwd string) error {
	width := 80
	cmd.Println(output.Header("Configuration Validation", width))
	cmd.Println()

	errors := []string{}
	warnings := []string{}

	// Check config file exists
	configPath, err := config.FindConfig(cwd)
	if err != nil {
		warnings = append(warnings, "Config file not found (using defaults)")
	} else {
		cmd.Printf("  %s Config file: %s\n", output.Color("[PASS]", output.Green), configPath)
	}

	// Check database path
	dbPath := cfg.DatabasePath(cwd)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("Database not found: %s", dbPath))
	} else {
		cmd.Printf("  %s Database: %s\n", output.Color("[PASS]", output.Green), dbPath)
	}

	// Check requirements directory
	reqDir := cfg.RequirementsPath(cwd)
	if _, err := os.Stat(reqDir); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("Requirements directory not found: %s", reqDir))
	} else {
		cmd.Printf("  %s Requirements dir: %s\n", output.Color("[PASS]", output.Green), reqDir)
	}

	cmd.Println()

	// Display errors
	if len(errors) > 0 {
		for _, e := range errors {
			cmd.Printf("  %s %s\n", output.Color("[FAIL]", output.Red), e)
		}
	}

	// Display warnings
	if len(warnings) > 0 {
		for _, w := range warnings {
			cmd.Printf("  %s %s\n", output.Color("[WARN]", output.Yellow), w)
		}
	}

	cmd.Println()

	if len(errors) > 0 {
		cmd.Printf("Status: %s\n", output.Color("INVALID", output.Red))
		return NewExitError(1, "configuration validation failed")
	} else if len(warnings) > 0 {
		cmd.Printf("Status: %s\n", output.Color("VALID (with warnings)", output.Yellow))
	} else {
		cmd.Printf("Status: %s\n", output.Color("VALID", output.Green))
	}

	return nil
}

func displayConfig(cmd *cobra.Command, cfg *config.Config, cwd string) error {
	switch configFormat {
	case "json":
		return displayConfigJSON(cmd, cfg)
	case "yaml":
		return displayConfigYAML(cmd, cfg)
	default:
		return displayConfigTerminal(cmd, cfg, cwd)
	}
}

func displayConfigJSON(cmd *cobra.Command, cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg.RTMX, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	cmd.Println(string(data))
	return nil
}

func displayConfigYAML(cmd *cobra.Command, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	cmd.Print(string(data))
	return nil
}

func displayConfigTerminal(cmd *cobra.Command, cfg *config.Config, cwd string) error {
	width := 80
	cmd.Println(output.Header("RTMX Configuration", width))
	cmd.Println()

	configPath, _ := config.FindConfig(cwd)
	if configPath == "" {
		configPath = "(defaults)"
	}

	cmd.Println("Paths:")
	cmd.Printf("  Config file:      %s\n", configPath)
	cmd.Printf("  Database:         %s\n", cfg.DatabasePath(cwd))
	cmd.Printf("  Requirements dir: %s\n", cfg.RequirementsPath(cwd))
	cmd.Println()

	cmd.Println("Schema:")
	cmd.Printf("  Type: %s\n", cfg.RTMX.Schema)
	cmd.Println()

	if len(cfg.RTMX.Phases) > 0 {
		cmd.Println("Phases:")
		for i := 1; i <= len(cfg.RTMX.Phases); i++ {
			if name, ok := cfg.RTMX.Phases[i]; ok {
				cmd.Printf("  %d: %s\n", i, name)
			}
		}
		cmd.Println()
	}

	if cfg.RTMX.Pytest.MarkerPrefix != "" {
		cmd.Println("Pytest:")
		cmd.Printf("  Marker prefix: %s\n", cfg.RTMX.Pytest.MarkerPrefix)
		cmd.Printf("  Register markers: %v\n", cfg.RTMX.Pytest.RegisterMarkers)
		cmd.Println()
	}

	return nil
}
