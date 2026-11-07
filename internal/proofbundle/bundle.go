// Package proofbundle builds an offline verification proof for regulated buyers
// (REQ-MONO-065aa). The bundle carries the requirement matrix, evidence records,
// and the commit SHAs those records name. It never carries git objects.
package proofbundle

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	MatrixName    = "matrix.csv"
	EvidenceName  = "evidence.jsonl"
	SHAsName      = "commit_shas.json"
	ManifestName  = "manifest.sha256"
)

// bundledFiles are hashed into the manifest (REQ-MONO-065b).
var bundledFiles = []string{MatrixName, EvidenceName, SHAsName}

var forbiddenKeys = map[string]struct{}{
	"blob": {}, "tree": {}, "patch": {}, "diff": {}, "content": {}, "contents": {},
}

// evidenceRecord is one verification-evidence-v1 line.
type evidenceRecord struct {
	RequirementID string `json:"requirement_id"`
	TestID        string `json:"test_id"`
	CommitSHA     string `json:"commit_sha"`
	Result        string `json:"result"`
	Timestamp     string `json:"timestamp"`
}

// Build writes a proof bundle under destDir from a matrix CSV and evidence JSONL.
// Git object fields on evidence lines are dropped. The bundle never contains
// blobs, trees, or patches.
func Build(destDir, matrixPath, evidencePath string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create bundle dir: %w", err)
	}
	if err := copyFile(matrixPath, filepath.Join(destDir, MatrixName)); err != nil {
		return fmt.Errorf("copy matrix: %w", err)
	}
	recs, err := readEvidence(evidencePath)
	if err != nil {
		return err
	}
	if err := writeEvidence(filepath.Join(destDir, EvidenceName), recs); err != nil {
		return err
	}
	shas := uniqueSHAs(recs)
	raw, err := json.MarshalIndent(shas, "", "  ")
	if err != nil {
		return fmt.Errorf("encode commit SHAs: %w", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(filepath.Join(destDir, SHAsName), raw, 0o644); err != nil {
		return fmt.Errorf("write commit SHAs: %w", err)
	}
	if err := writeManifest(destDir); err != nil {
		return err
	}
	return nil
}

// writeManifest records SHA-256 digests of every bundled artifact (REQ-MONO-065b).
func writeManifest(destDir string) error {
	var b strings.Builder
	for _, name := range bundledFiles {
		sum, err := fileSHA256(filepath.Join(destDir, name))
		if err != nil {
			return fmt.Errorf("hash %s: %w", name, err)
		}
		fmt.Fprintf(&b, "%s  %s\n", sum, name)
	}
	if err := os.WriteFile(filepath.Join(destDir, ManifestName), []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// Verify checks the hash manifest against the bundled files. It does not open
// a network connection (REQ-MONO-065b).
func Verify(destDir string) error {
	raw, err := os.ReadFile(filepath.Join(destDir, ManifestName))
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	expected := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return fmt.Errorf("malformed manifest line: %q", line)
		}
		expected[parts[1]] = parts[0]
	}
	for _, name := range bundledFiles {
		want, ok := expected[name]
		if !ok {
			return fmt.Errorf("manifest missing %s", name)
		}
		got, err := fileSHA256(filepath.Join(destDir, name))
		if err != nil {
			return fmt.Errorf("hash %s: %w", name, err)
		}
		if got != want {
			return fmt.Errorf("%s: digest mismatch", name)
		}
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func readEvidence(path string) ([]evidenceRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open evidence: %w", err)
	}
	defer func() { _ = f.Close() }()
	var out []evidenceRecord
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, fmt.Errorf("parse evidence: %w", err)
		}
		for k := range forbiddenKeys {
			delete(raw, k)
		}
		rec := evidenceRecord{
			RequirementID: asString(raw["requirement_id"]),
			TestID:        asString(raw["test_id"]),
			CommitSHA:     strings.ToLower(asString(raw["commit_sha"])),
			Result:        asString(raw["result"]),
			Timestamp:     asString(raw["timestamp"]),
		}
		out = append(out, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan evidence: %w", err)
	}
	return out, nil
}

func writeEvidence(path string, recs []evidenceRecord) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create evidence: %w", err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, rec := range recs {
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("write evidence: %w", err)
		}
	}
	return nil
}

func uniqueSHAs(recs []evidenceRecord) []string {
	seen := map[string]struct{}{}
	for _, r := range recs {
		if r.CommitSHA != "" {
			seen[r.CommitSHA] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for sha := range seen {
		out = append(out, sha)
	}
	sort.Strings(out)
	return out
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprint(v)
	}
	return s
}
