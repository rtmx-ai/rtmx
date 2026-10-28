package cmd

import (
	"net/url"
	"os/exec"
	"strings"
)

// GitRemote is a named git remote URL (no network I/O).
type GitRemote struct {
	Name string
	URL  string
}

// InferProviderHint maps git remotes to an issue-store provider hint for the
// link picker: "github", "gitlab", or "" when unknown / missing.
// Prefers the remote named "origin" when present; otherwise the first match.
// REQ-GO-088 / REQ-MONO-044a.
func InferProviderHint(remotes []GitRemote) string {
	if len(remotes) == 0 {
		return ""
	}
	var origin *GitRemote
	for i := range remotes {
		if remotes[i].Name == "origin" {
			origin = &remotes[i]
			break
		}
	}
	if origin != nil {
		if hint := ProviderHintFromRemoteURL(origin.URL); hint != "" {
			return hint
		}
	}
	for _, r := range remotes {
		if hint := ProviderHintFromRemoteURL(r.URL); hint != "" {
			return hint
		}
	}
	return ""
}

// ProviderHintFromRemoteURL classifies a single remote URL with no network I/O.
func ProviderHintFromRemoteURL(remote string) string {
	host := remoteHost(remote)
	if host == "" {
		return ""
	}
	switch {
	case isGitHubHost(host):
		return "github"
	case isGitLabHost(host):
		return "gitlab"
	default:
		return ""
	}
}

func remoteHost(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}
	// scp-like: git@host:path
	if strings.HasPrefix(remote, "git@") {
		rest := strings.TrimPrefix(remote, "git@")
		host, _, ok := strings.Cut(rest, ":")
		if !ok || host == "" {
			return ""
		}
		return strings.ToLower(host)
	}

	// Ensure url.Parse sees a scheme for ssh:// and https:// forms.
	toParse := remote
	if !strings.Contains(remote, "://") {
		// bare host/path — treat as https for host extraction
		toParse = "https://" + remote
	}
	u, err := url.Parse(toParse)
	if err != nil {
		return ""
	}
	host := u.Hostname()
	if host == "" {
		return ""
	}
	return strings.ToLower(host)
}

func isGitHubHost(host string) bool {
	if host == "github.com" || host == "www.github.com" {
		return true
	}
	// GitHub Enterprise: github.<domain>
	return strings.HasPrefix(host, "github.")
}

func isGitLabHost(host string) bool {
	if host == "gitlab.com" || host == "www.gitlab.com" {
		return true
	}
	// Self-hosted / GitLab Dedicated: host contains "gitlab"
	return strings.Contains(host, "gitlab")
}

// listGitRemotes returns named remotes for dir (empty = cwd). Overridable in tests.
// Missing git / no remotes returns an empty slice without error (hint stays empty).
var listGitRemotes = func(dir string) ([]GitRemote, error) {
	cmd := exec.Command("git", "remote", "-v")
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}
	return parseGitRemoteV(string(out)), nil
}

// parseGitRemoteV parses `git remote -v` output into unique remotes (fetch preferred).
func parseGitRemoteV(out string) []GitRemote {
	byName := map[string]GitRemote{}
	order := []string{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name, url := fields[0], fields[1]
		kind := ""
		if len(fields) >= 3 {
			kind = strings.Trim(fields[2], "()")
		}
		if existing, ok := byName[name]; ok {
			// Prefer fetch over push when both present.
			if kind == "fetch" {
				byName[name] = GitRemote{Name: name, URL: url}
			} else if existing.URL == "" {
				byName[name] = GitRemote{Name: name, URL: url}
			}
			continue
		}
		byName[name] = GitRemote{Name: name, URL: url}
		order = append(order, name)
	}
	remotes := make([]GitRemote, 0, len(order))
	for _, name := range order {
		remotes = append(remotes, byName[name])
	}
	return remotes
}
