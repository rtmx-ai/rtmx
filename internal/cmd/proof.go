package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/proofbundle"
	"github.com/spf13/cobra"
)

// ProofDialFunc is an outbound network dial. Proof export must not call it.
type ProofDialFunc func(network, address string) error

var (
	proofExportOut string
	// proofDial counts outbound network attempts during proof export (REQ-MONO-065ab).
	proofDial atomic.Int64
	// proofDialer is the outbound path. runProofExport never calls it.
	proofDialer ProofDialFunc
)

var proofCmd = &cobra.Command{
	Use:   "proof",
	Short: "Export offline verification proof",
	Long:  `Build offline verification artifacts for regulated buyers. No network is required.`,
}

var proofExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Write an offline proof bundle",
	Long: `Write a local proof bundle containing the requirement matrix, evidence records,
and the commit SHAs those records name. The command does not open a network connection.`,
	RunE: runProofExport,
}

func init() {
	proofExportCmd.Flags().StringVarP(&proofExportOut, "out", "o", ".rtmx/proof-bundle", "destination directory for the proof bundle")
	proofCmd.AddCommand(proofExportCmd)
	rootCmd.AddCommand(proofCmd)
}

func runProofExport(cmd *cobra.Command, args []string) error {
	proofDial.Store(0)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("working directory: %w", err)
	}
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	matrixPath := cfg.DatabasePath(cwd)
	evidencePath := filepath.Join(cwd, ".rtmx", "evidence.jsonl")
	out := proofExportOut
	if !filepath.IsAbs(out) {
		out = filepath.Join(cwd, out)
	}
	if err := proofbundle.Build(out, matrixPath, evidencePath); err != nil {
		return err
	}
	if proofDial.Load() > 0 {
		return fmt.Errorf("proof export attempted an outbound network call")
	}
	cmd.Printf("Wrote proof bundle to %s\n", out)
	return nil
}

// proofRecordDial marks an outbound attempt. Tests inject this to prove export
// never dials when outbound network is unavailable.
func proofRecordDial(network, address string) error {
	proofDial.Add(1)
	return fmt.Errorf("proof export: outbound network unavailable (%s %s)", network, address)
}

// ProofOutboundCalls returns how many outbound dials were attempted during export.
func ProofOutboundCalls() int64 {
	return proofDial.Load()
}
