package broker

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestPrivateKeyMode600BlocksDeveloperRead(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069b")

	dir := t.TempDir()
	path := filepath.Join(dir, "verify.key")
	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePrivateKeyMode600(path); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("mode = %o, want 0600", info.Mode().Perm())
		}
	}
	readable, err := PrivateKeyReadableByOthers(path)
	if err != nil {
		t.Fatal(err)
	}
	if readable {
		t.Fatal("key is readable by group/other; developer must not read it")
	}
}
