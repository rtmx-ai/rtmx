// Package version provides semantic versioning parsing and comparison.
package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a parsed semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
}

// Parse parses a version string like "v1.2.3" or "1.2.3-rc1".
func Parse(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")

	// Split off prerelease
	var pre string
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		pre = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid semver: expected MAJOR.MINOR.PATCH, got %q", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}

	return Version{Major: major, Minor: minor, Patch: patch, Prerelease: pre}, nil
}

// String returns the version as "vMAJOR.MINOR.PATCH".
func (v Version) String() string {
	s := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	return s
}

// BumpMajor returns the next major version (v+1.0.0).
func (v Version) BumpMajor() Version {
	return Version{Major: v.Major + 1}
}

// BumpMinor returns the next minor version (v.MAJOR.minor+1.0).
func (v Version) BumpMinor() Version {
	return Version{Major: v.Major, Minor: v.Minor + 1}
}

// BumpPatch returns the next patch version (v.MAJOR.MINOR.patch+1).
func (v Version) BumpPatch() Version {
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

// BumpLevel represents a semver increment level.
type BumpLevel int

const (
	BumpNone  BumpLevel = iota
	BumpPatch
	BumpMinor
	BumpMajor
)

// String returns the human-readable name.
func (b BumpLevel) String() string {
	switch b {
	case BumpMajor:
		return "major"
	case BumpMinor:
		return "minor"
	case BumpPatch:
		return "patch"
	default:
		return "none"
	}
}

// ParseBumpLevel parses a string into a BumpLevel.
func ParseBumpLevel(s string) BumpLevel {
	switch strings.ToLower(s) {
	case "major":
		return BumpMajor
	case "minor":
		return BumpMinor
	case "patch":
		return BumpPatch
	default:
		return BumpNone
	}
}

// ActualBump determines the bump level between two versions.
func ActualBump(from, to Version) BumpLevel {
	if to.Major > from.Major {
		return BumpMajor
	}
	if to.Minor > from.Minor {
		return BumpMinor
	}
	if to.Patch > from.Patch {
		return BumpPatch
	}
	return BumpNone
}
