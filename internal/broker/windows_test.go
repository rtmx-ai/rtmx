package broker

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsServiceRunsAsNTServiceRtmx(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069b")

	path := filepath.Join("..", "..", "packaging", "windows", "verify-service.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		ServiceAccount string `json:"serviceAccount"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.ServiceAccount != `NT SERVICE\rtmx` {
		t.Fatalf("serviceAccount = %q, want NT SERVICE\\rtmx", cfg.ServiceAccount)
	}
}

func TestWindowsKeyACLGrantsServiceDeniesDeveloper(t *testing.T) {
	developer := `DESKTOP\developer`
	acl := DefaultWindowsKeyACL(developer)
	if !acl.GrantsVerifierDeniesDeveloper(developer) {
		t.Fatalf("ACL = %+v does not grant NT SERVICE\\rtmx and deny %s", acl, developer)
	}
	if acl.GrantsVerifierDeniesDeveloper(`DESKTOP\other`) {
		t.Fatal("ACL should not report deny for an unlisted developer account")
	}
}
