package broker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxSystemdUnitRunsAsRtmxVerify(t *testing.T) {
	path := filepath.Join("..", "..", "packaging", "systemd", "verify.service")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "User=rtmx-verify") {
		t.Fatalf("unit missing User=rtmx-verify:\n%s", body)
	}
	if !strings.Contains(body, "Group=rtmx-verify") {
		t.Fatalf("unit missing Group=rtmx-verify:\n%s", body)
	}
}
