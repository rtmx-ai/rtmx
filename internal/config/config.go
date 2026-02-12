// Package config provides configuration management for RTMX.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the RTMX configuration.
type Config struct {
	RTMX RTMXConfig `yaml:"rtmx"`
}

// RTMXConfig contains the main RTMX settings.
type RTMXConfig struct {
	// Database is the path to the RTM database CSV file.
	Database string `yaml:"database"`

	// RequirementsDir is the directory containing requirement spec files.
	RequirementsDir string `yaml:"requirements_dir"`

	// Schema is the schema name (core or custom).
	Schema string `yaml:"schema"`

	// Pytest configuration
	Pytest PytestConfig `yaml:"pytest"`

	// Phases maps phase numbers to descriptions.
	Phases map[int]string `yaml:"phases"`

	// Agents configuration for AI assistants.
	Agents AgentsConfig `yaml:"agents"`

	// Adapters configuration for external integrations.
	Adapters AdaptersConfig `yaml:"adapters"`

	// MCP configuration for Model Context Protocol.
	MCP MCPConfig `yaml:"mcp"`

	// Sync configuration for collaboration.
	Sync SyncConfig `yaml:"sync"`

	// Auth configuration for authentication.
	Auth AuthConfig `yaml:"auth"`

	// Ziti configuration for zero-trust networking.
	Ziti ZitiConfig `yaml:"ziti"`
}

// PytestConfig contains pytest-related settings.
type PytestConfig struct {
	MarkerPrefix    string `yaml:"marker_prefix"`
	RegisterMarkers bool   `yaml:"register_markers"`
}

// AgentsConfig contains AI agent settings.
type AgentsConfig struct {
	Claude      AgentConfig `yaml:"claude"`
	Cursor      AgentConfig `yaml:"cursor"`
	Copilot     AgentConfig `yaml:"copilot"`
	TemplateDir string      `yaml:"template_dir"`
}

// AgentConfig represents configuration for a single AI agent.
type AgentConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ConfigPath string `yaml:"config_path"`
}

// AdaptersConfig contains external integration settings.
type AdaptersConfig struct {
	GitHub GitHubConfig `yaml:"github"`
	Jira   JiraConfig   `yaml:"jira"`
}

// GitHubConfig contains GitHub integration settings.
type GitHubConfig struct {
	Enabled       bool              `yaml:"enabled"`
	Repo          string            `yaml:"repo"`
	TokenEnv      string            `yaml:"token_env"`
	Labels        GitHubLabels      `yaml:"labels"`
	StatusMapping map[string]string `yaml:"status_mapping"`
}

// GitHubLabels contains GitHub label configuration.
type GitHubLabels struct {
	Requirement string `yaml:"requirement"`
}

// GitHubAdapterConfig is an alias for GitHubConfig used by the adapter.
type GitHubAdapterConfig = GitHubConfig

// JiraConfig contains Jira integration settings.
type JiraConfig struct {
	Enabled       bool              `yaml:"enabled"`
	Server        string            `yaml:"server"`
	Project       string            `yaml:"project"`
	TokenEnv      string            `yaml:"token_env"`
	EmailEnv      string            `yaml:"email_env"`
	IssueType     string            `yaml:"issue_type"`
	JQLFilter     string            `yaml:"jql_filter"`
	Labels        []string          `yaml:"labels"`
	StatusMapping map[string]string `yaml:"status_mapping"`
}

// JiraAdapterConfig is an alias for JiraConfig used by the adapter.
type JiraAdapterConfig = JiraConfig

// MCPConfig contains Model Context Protocol settings.
type MCPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
}

// SyncConfig contains collaboration settings.
type SyncConfig struct {
	ConflictResolution string                `yaml:"conflict_resolution"`
	Remotes            map[string]SyncRemote `yaml:"remotes"`
}

// SyncRemote represents a remote RTM repository.
type SyncRemote struct {
	Repo     string `yaml:"repo"`
	Database string `yaml:"database"`
	Path     string `yaml:"path"`
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	Provider     string   `yaml:"provider"`
	Issuer       string   `yaml:"issuer"`
	ClientID     string   `yaml:"client_id"`
	Scopes       []string `yaml:"scopes"`
	CallbackPort int      `yaml:"callback_port"`
}

// ZitiConfig contains OpenZiti zero-trust networking settings.
type ZitiConfig struct {
	Controller  string            `yaml:"controller"`
	IdentityDir string            `yaml:"identity_dir"`
	Services    map[string]string `yaml:"services"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		RTMX: RTMXConfig{
			Database:        ".rtmx/database.csv",
			RequirementsDir: ".rtmx/requirements",
			Schema:          "core",
			Pytest: PytestConfig{
				MarkerPrefix:    "req",
				RegisterMarkers: true,
			},
			Phases: map[int]string{
				1: "Foundation",
				2: "Core Features",
				3: "Integration",
			},
			Agents: AgentsConfig{
				Claude: AgentConfig{
					Enabled:    true,
					ConfigPath: "CLAUDE.md",
				},
				Cursor: AgentConfig{
					Enabled:    false,
					ConfigPath: ".cursorrules",
				},
				Copilot: AgentConfig{
					Enabled:    false,
					ConfigPath: ".github/copilot-instructions.md",
				},
				TemplateDir: ".rtmx/templates/",
			},
			Adapters: AdaptersConfig{
				GitHub: GitHubConfig{
					Enabled:  false,
					TokenEnv: "GITHUB_TOKEN",
					Labels:   GitHubLabels{Requirement: "requirement"},
					StatusMapping: map[string]string{
						"open":   "MISSING",
						"closed": "COMPLETE",
					},
				},
				Jira: JiraConfig{
					Enabled:   false,
					TokenEnv:  "JIRA_API_TOKEN",
					EmailEnv:  "JIRA_EMAIL",
					IssueType: "Requirement",
				},
			},
			MCP: MCPConfig{
				Enabled: false,
				Port:    3000,
				Host:    "localhost",
			},
			Sync: SyncConfig{
				ConflictResolution: "manual",
				Remotes:            make(map[string]SyncRemote),
			},
			Auth: AuthConfig{
				Provider:     "zitadel",
				Scopes:       []string{"openid", "profile", "email"},
				CallbackPort: 8765,
			},
			Ziti: ZitiConfig{
				IdentityDir: "~/.rtmx/ziti",
				Services:    make(map[string]string),
			},
		},
	}
}

// Load loads configuration from a file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to a file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// FindConfig searches for a configuration file starting from the given path.
func FindConfig(startPath string) (string, error) {
	// Check common locations
	candidates := []string{
		".rtmx/config.yaml",
		"rtmx.yaml",
		"rtmx.yml",
	}

	// Search from start path upward
	dir := startPath
	for {
		for _, candidate := range candidates {
			path := filepath.Join(dir, candidate)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no RTMX configuration found")
}

// LoadFromDir loads configuration from the given directory.
func LoadFromDir(dir string) (*Config, error) {
	path, err := FindConfig(dir)
	if err != nil {
		// Return default config if no config file found
		return DefaultConfig(), nil
	}

	return Load(path)
}

// DatabasePath returns the resolved database path.
func (c *Config) DatabasePath(baseDir string) string {
	if filepath.IsAbs(c.RTMX.Database) {
		return c.RTMX.Database
	}
	return filepath.Join(baseDir, c.RTMX.Database)
}

// RequirementsPath returns the resolved requirements directory path.
func (c *Config) RequirementsPath(baseDir string) string {
	if filepath.IsAbs(c.RTMX.RequirementsDir) {
		return c.RTMX.RequirementsDir
	}
	return filepath.Join(baseDir, c.RTMX.RequirementsDir)
}

// PhaseDescription returns the description for a phase number.
func (c *Config) PhaseDescription(phase int) string {
	if desc, ok := c.RTMX.Phases[phase]; ok {
		return desc
	}
	return fmt.Sprintf("Phase %d", phase)
}
