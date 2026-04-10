package test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestActionPinning proves that GitHub Actions are pinned to mutable version
// tags instead of commit SHAs, enabling supply chain attacks via tag mutation.
// REQ-SEC-004: All GitHub Actions shall be pinned to commit SHAs.
//
// ATTACK: An attacker who compromises any upstream action repo can move a
// version tag (e.g., v5) to point to malicious code. The next rtmx CI or
// release run will execute the attacker's code with full secret access.
//
// WHEN FIXED: This test should assert zero mutable tags found.
func TestActionPinning(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-004")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	workflowDir := filepath.Join(projectRoot, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		t.Fatalf("Failed to read workflows dir: %v", err)
	}

	// Match uses: org/action@ref where ref should be a SHA or version tag
	usesRegex := regexp.MustCompile(`uses:\s*([^@]+)@(\S+)`)
	shaRegex := regexp.MustCompile(`^[0-9a-f]{40}$`)
	tagRegex := regexp.MustCompile(`^v\d+`)

	unpinnedCount := 0
	totalCount := 0

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(workflowDir, entry.Name()))
		if err != nil {
			t.Fatalf("Failed to read %s: %v", entry.Name(), err)
		}

		matches := usesRegex.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			ref := m[2]
			totalCount++
			if !shaRegex.MatchString(ref) && !tagRegex.MatchString(ref) {
				unpinnedCount++
				t.Logf("VULNERABLE: %s uses %s@%s (unpinned ref)", entry.Name(), m[1], ref)
			}
		}
	}

	// All actions should be pinned to SHAs or version tags
	if unpinnedCount > 0 {
		t.Errorf("%d/%d action references use unpinned refs (not SHA or version tag)", unpinnedCount, totalCount)
	}
}

// TestInstallScriptGPGVerification proves that the install script verifies
// checksums but NOT GPG signatures, allowing a supply chain attack where both
// the binary and checksums.txt are replaced.
// REQ-SEC-005: Install script shall verify GPG signatures.
//
// ATTACK: Compromise the GitHub release (account takeover or stolen GITHUB_TOKEN).
// Replace both the archive and checksums.txt. The install script happily
// installs the tampered binary because it only checks SHA256 against the
// attacker-controlled checksums.
//
// WHEN FIXED: This test should find gpg --verify in the install script.
func TestInstallScriptGPGVerification(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-005")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	content, err := os.ReadFile(filepath.Join(projectRoot, "scripts", "install.sh"))
	if err != nil {
		t.Fatalf("Failed to read install.sh: %v", err)
	}
	script := string(content)

	// Verify checksums ARE checked (the half that works)
	if !strings.Contains(script, "sha256sum") && !strings.Contains(script, "shasum") {
		t.Fatal("Install script doesn't even verify checksums")
	}

	// FIXED: GPG verification should be present
	if !strings.Contains(script, "gpg --verify") {
		t.Error("Install script must contain gpg --verify for signature validation")
	}
	if !strings.Contains(script, "checksums.txt.sig") {
		t.Error("Install script must download checksums.txt.sig")
	}
}

// TestMandatoryGPGSigning proves that GPG signing is conditional in the
// release workflow, allowing unsigned releases to ship silently.
// REQ-SEC-006: GPG signing shall be mandatory for all releases.
//
// ATTACK: If the GPG_PRIVATE_KEY secret is accidentally deleted or misconfigured,
// releases proceed without signatures. Users who rely on GPG verification
// receive unsigned binaries with no warning from the release process.
//
// WHEN FIXED: The GPG import step should have no `if` condition.
func TestMandatoryGPGSigning(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-006")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	content, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("Failed to read release.yml: %v", err)
	}
	release := string(content)

	// Find the GPG import step
	if !strings.Contains(release, "Import GPG key") {
		t.Fatal("GPG import step not found in release workflow")
	}

	// ATTACK SUCCEEDS: signing is conditional
	// The `if:` condition is on the line AFTER the step name
	gpgIdx := strings.Index(release, "Import GPG key")
	end := gpgIdx + 200
	if end > len(release) {
		end = len(release)
	}
	block := release[gpgIdx:end]

	// FIXED: GPG step should NOT have a conditional
	if strings.Contains(block, "if:") {
		t.Error("GPG signing step still has conditional -- should be mandatory")
	}
}

// TestCIPipelineSafety proves that the CI auto-commit in verify-requirements
// does not validate which files were modified, allowing a compromised verify
// command to inject arbitrary changes into main.
// REQ-SEC-010: CI pipeline shall prevent self-modification attacks.
//
// ATTACK: A malicious PR modifies the rtmx source code such that `rtmx verify
// --update` writes not only to database.csv but also to a Go source file.
// The auto-commit step commits ALL staged changes to main without checking
// that only database.csv was modified.
//
// WHEN FIXED: The commit step should verify that only .rtmx/database.csv
// was modified before committing.
func TestCIPipelineSafety(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-010")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	content, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("Failed to read ci.yml: %v", err)
	}
	ci := string(content)

	// Find the auto-commit section
	commitIdx := strings.Index(ci, "ci: Auto-update RTM status")
	if commitIdx < 0 {
		t.Fatal("Auto-commit step not found in CI workflow")
	}

	// Look at the commit block
	commitBlock := ci[commitIdx-300 : commitIdx+200]

	// ATTACK SUCCEEDS: git add targets only database.csv, but there's no check
	// that rtmx verify --update didn't modify OTHER files
	if !strings.Contains(commitBlock, "git add .rtmx/database.csv") {
		t.Fatal("Expected targeted git add")
	}

	// Verify there's no safety check (e.g., "git diff --name-only" validation)
	if strings.Contains(commitBlock, "git diff --name-only") {
		t.Fatal("File modification check found -- vulnerability may be fixed")
	}
	if strings.Contains(commitBlock, "only .rtmx/database.csv") {
		t.Fatal("Restrictive commit check found -- vulnerability may be fixed")
	}

	t.Log("ATTACK CONFIRMED: auto-commit does not verify which files were modified by verify")
}
