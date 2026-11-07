// Command verify is the packaged verification broker daemon (REQ-MONO-069).
// It ships in the same rtmx release archive / Homebrew formula as the CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rtmx-ai/rtmx/internal/broker"
)

func main() {
	var (
		socket  = flag.String("socket", broker.DefaultBrokerSocket(), "unix socket path")
		key     = flag.String("key", broker.DefaultPackagePaths().KeyPath, "ed25519 private key path")
		workdir = flag.String("workdir", "/var/lib/rtmx/work", "git workdir the broker checks out into")
	)
	var command multiFlag
	flag.Var(&command, "command", "broker-owned test argv (repeatable; default: go test ./... -count=1)")
	flag.Parse()

	cmd := []string(command)
	if len(cmd) == 0 {
		cmd = []string{"go", "test", "./...", "-count=1"}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := broker.Serve(ctx, broker.ServeConfig{
		SocketPath: *socket,
		KeyPath:    *key,
		WorkDir:    *workdir,
		Command:    cmd,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify: %v\n", err)
		os.Exit(1)
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, " ") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}
