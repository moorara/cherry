package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version is either patch, minor, or major.
type Version int

const (
	// Patch is the patch number in a semantic version.
	Patch Version = iota
	// Minor is the minor number in a semantic version.
	Minor
	// Major is the major number in a semantic version.
	Major
)

// SemVer represents a semantic versioning
type SemVer struct {
	Major      uint
	Minor      uint
	Patch      uint
	Prerelease []string
	Metadata   []string
}

// Parse reads a semantic version string and returns a SemVer.
// If the second return value is false, it implies that the input semver was incorrect.
func Parse(semver string) (SemVer, bool) {
	var major, minor, patch uint64
	var prerelease, metadata []string

	// Make sure the string is a valid semantic version
	if re := regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(\-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`); !re.MatchString(semver) {
		return SemVer{}, false
	}

	re := regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(\-[\.[0-9A-Za-z-]+)?(\+[\.[0-9A-Za-z-]*)?$`)
	subs := re.FindStringSubmatch(semver)

	major, _ = strconv.ParseUint(subs[1], 10, 64)
	minor, _ = strconv.ParseUint(subs[2], 10, 64)
	patch, _ = strconv.ParseUint(subs[3], 10, 64)

	if subs[4] != "" {
		prerelease = strings.Split(subs[4][1:], ".")
	}

	if subs[5] != "" {
		metadata = strings.Split(subs[5][1:], ".")
	}

	return SemVer{
		Major:      uint(major),
		Minor:      uint(minor),
		Patch:      uint(patch),
		Prerelease: prerelease,
		Metadata:   metadata,
	}, true
}

// AddPrerelease adds a new pre-release identifier to the current semantic version.
// This is a shortcut for v.Prerelease = append(v.Prerelease, s...).
func (v *SemVer) AddPrerelease(s ...string) {
	v.Prerelease = append(v.Prerelease, s...)
}

// AddMetadata adds a new metadata identifier to the current semantic version.
// This is a shortcut for v.Metadata = append(v.Metadata, s...).
func (v *SemVer) AddMetadata(s ...string) {
	v.Metadata = append(v.Metadata, s...)
}

// Next returns the next patch version.
func (v SemVer) Next() SemVer {
	return SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}
}

// Release returns the semantic version for a release
func (v SemVer) Release(version Version) SemVer {
	switch version {
	case Patch:
		return SemVer{
			Major: v.Major,
			Minor: v.Minor,
			Patch: v.Patch,
		}
	case Minor:
		return SemVer{
			Major: v.Major,
			Minor: v.Minor + 1,
			Patch: 0,
		}
	case Major:
		return SemVer{
			Major: v.Major + 1,
			Minor: 0,
			Patch: 0,
		}
	default:
		return SemVer{}
	}
}

// String returns a semantic version string (also implements fmt.Stringer).
func (v SemVer) String() string {
	var tail string

	if len(v.Prerelease) > 0 {
		tail += "-" + strings.Join(v.Prerelease, ".")
	}

	if len(v.Metadata) > 0 {
		tail += "+" + strings.Join(v.Metadata, ".")
	}

	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, tail)
}
