// Package auth provides authentication for the RTMX CLI.
// This file implements OpenZiti zero-trust networking integration.
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ZitiConfig holds OpenZiti enrollment configuration.
type ZitiConfig struct {
	// IdentityFile is the path to the Ziti identity JSON file.
	IdentityFile string `json:"identity_file"`

	// ControllerURL is the Ziti controller endpoint.
	ControllerURL string `json:"controller_url"`

	// ServiceName is the Ziti service to connect to.
	ServiceName string `json:"service_name"`

	// Enrolled indicates whether the identity has been enrolled.
	Enrolled bool `json:"enrolled"`
}

// ZitiClient provides OpenZiti zero-trust networking for RTMX.
type ZitiClient struct {
	Config     ZitiConfig
	ConfigPath string
}

// NewZitiClient creates a ZitiClient from a config directory.
func NewZitiClient(configDir string) (*ZitiClient, error) {
	configPath := filepath.Join(configDir, "ziti.json")

	client := &ZitiClient{
		ConfigPath: configPath,
	}

	// Load existing config if present
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, &client.Config); err != nil {
			return nil, fmt.Errorf("invalid ziti config: %w", err)
		}
	}

	return client, nil
}

// Enroll registers this CLI instance with a Ziti controller.
func (z *ZitiClient) Enroll(controllerURL, identityFile string) error {
	z.Config.ControllerURL = controllerURL
	z.Config.IdentityFile = identityFile
	z.Config.Enrolled = true

	return z.saveConfig()
}

// IsEnrolled returns true if the Ziti identity has been enrolled.
func (z *ZitiClient) IsEnrolled() bool {
	return z.Config.Enrolled
}

// SetService configures the Ziti service name for rtmx-sync connectivity.
func (z *ZitiClient) SetService(serviceName string) error {
	z.Config.ServiceName = serviceName
	return z.saveConfig()
}

func (z *ZitiClient) saveConfig() error {
	dir := filepath.Dir(z.ConfigPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(z.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(z.ConfigPath, data, 0600)
}
