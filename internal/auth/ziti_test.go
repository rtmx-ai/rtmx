package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestZitiIntegration(t *testing.T) {
	rtmx.Req(t, "REQ-GO-041")

	t.Run("new_client_no_config", func(t *testing.T) {
		tmpDir := t.TempDir()
		client, err := NewZitiClient(tmpDir)
		if err != nil {
			t.Fatalf("NewZitiClient failed: %v", err)
		}
		if client.IsEnrolled() {
			t.Error("new client should not be enrolled")
		}
	})

	t.Run("enroll_saves_config", func(t *testing.T) {
		tmpDir := t.TempDir()
		client, err := NewZitiClient(tmpDir)
		if err != nil {
			t.Fatalf("NewZitiClient failed: %v", err)
		}

		identityFile := filepath.Join(tmpDir, "identity.json")
		_ = os.WriteFile(identityFile, []byte(`{"ztAPI":"https://ctrl.example.com"}`), 0600)

		err = client.Enroll("https://ctrl.example.com", identityFile)
		if err != nil {
			t.Fatalf("Enroll failed: %v", err)
		}

		if !client.IsEnrolled() {
			t.Error("client should be enrolled after Enroll()")
		}

		// Verify config file was created
		configPath := filepath.Join(tmpDir, "ziti.json")
		if _, err := os.Stat(configPath); err != nil {
			t.Errorf("config file should exist: %v", err)
		}
	})

	t.Run("load_existing_config", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Write a config file
		configPath := filepath.Join(tmpDir, "ziti.json")
		config := `{"controller_url":"https://ctrl.example.com","identity_file":"/path/to/id.json","service_name":"rtmx-sync","enrolled":true}`
		_ = os.WriteFile(configPath, []byte(config), 0600)

		client, err := NewZitiClient(tmpDir)
		if err != nil {
			t.Fatalf("NewZitiClient failed: %v", err)
		}

		if !client.IsEnrolled() {
			t.Error("should load enrolled state from config")
		}
		if client.Config.ControllerURL != "https://ctrl.example.com" {
			t.Errorf("controller URL = %q, want https://ctrl.example.com", client.Config.ControllerURL)
		}
		if client.Config.ServiceName != "rtmx-sync" {
			t.Errorf("service name = %q, want rtmx-sync", client.Config.ServiceName)
		}
	})

	t.Run("set_service", func(t *testing.T) {
		tmpDir := t.TempDir()
		client, err := NewZitiClient(tmpDir)
		if err != nil {
			t.Fatalf("NewZitiClient failed: %v", err)
		}

		err = client.SetService("rtmx-sync-prod")
		if err != nil {
			t.Fatalf("SetService failed: %v", err)
		}

		if client.Config.ServiceName != "rtmx-sync-prod" {
			t.Errorf("service = %q, want rtmx-sync-prod", client.Config.ServiceName)
		}

		// Reload and verify persistence
		client2, err := NewZitiClient(tmpDir)
		if err != nil {
			t.Fatalf("reload failed: %v", err)
		}
		if client2.Config.ServiceName != "rtmx-sync-prod" {
			t.Errorf("reloaded service = %q, want rtmx-sync-prod", client2.Config.ServiceName)
		}
	})
}
