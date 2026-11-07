package markers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SpecIDPattern is the requirement ID grammar from requirement-markers-v1
// (REQ-MONO-068a / REQ-MONO-068b).
var SpecIDPattern = reqIDPattern

// LooseREQ finds candidate requirement tokens in source, including malformed ones.
var LooseREQ = regexp.MustCompile(`REQ-[A-Za-z0-9_-]*`)

// Issue is one dangling or malformed marker finding.
type Issue struct {
	Kind    string // "dangling" or "malformed"
	Token   string
	Path    string
	Message string
}

func (i Issue) Error() string {
	return fmt.Sprintf("%s: %s (%s)", i.Path, i.Message, i.Token)
}

// CheckSources scans files for requirement markers and reports dangling or
// malformed IDs (REQ-MONO-068b). knownIDs is the set of requirement IDs in the RTM.
func CheckSources(paths []string, knownIDs map[string]struct{}) []Issue {
	var issues []Issue
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			issues = append(issues, Issue{
				Kind:    "malformed",
				Token:   path,
				Path:    path,
				Message: fmt.Sprintf("read failed: %v", err),
			})
			continue
		}
		for _, tok := range LooseREQ.FindAllString(string(raw), -1) {
			if !SpecIDPattern.MatchString(tok) {
				issues = append(issues, Issue{
					Kind:    "malformed",
					Token:   tok,
					Path:    path,
					Message: "marker does not match the requirement-markers-v1 grammar",
				})
				continue
			}
			if _, ok := knownIDs[tok]; !ok {
				issues = append(issues, Issue{
					Kind:    "dangling",
					Token:   tok,
					Path:    path,
					Message: "marker names a requirement ID that is not in the RTM",
				})
			}
		}
	}
	return issues
}

// ExitCode returns 1 when any issue is present, otherwise 0.
func ExitCode(issues []Issue) int {
	if len(issues) == 0 {
		return 0
	}
	return 1
}

// CollectSourceFiles walks root for common source extensions.
func CollectSourceFiles(root string) ([]string, error) {
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == "node_modules" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go", ".py", ".rs", ".ts", ".js", ".java", ".cs", ".md":
			out = append(out, path)
		}
		return nil
	})
	return out, err
}
