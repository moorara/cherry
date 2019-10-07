package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Segment represents either patch, minor, major numbers is a semantic version.
type Segment int

const (
	// Patch is the patch number in a semantic version.
	Patch Segment = iota
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
func Parse(version string) (SemVer, error) {
	var zero SemVer
	var major, minor, patch uint64
	var prerelease, metadata []string

	// Make sure the string is a valid semantic version
	svRE := regexp.MustCompile(`^\d+\.\d+\.\d+(\-\w+(\.\w+)*)?(\+\w+(\.\w+)*)?$`)
	if !svRE.MatchString(version) {
		return zero, errors.New("invalid semantic version")
	}

	subsRE := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(\-[\.\w]+)?(\+[\.\w]*)?$`)
	subs := subsRE.FindStringSubmatch(version)

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
	}, nil
}

// Version returns a semantic version string.
func (v SemVer) Version() string {
	var tail string

	if len(v.Prerelease) > 0 {
		tail += "-" + strings.Join(v.Prerelease, ".")
	}

	if len(v.Metadata) > 0 {
		tail += "+" + strings.Join(v.Metadata, ".")
	}

	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, tail)
}

// GitTag returns a semantic version string to be used as a git tag.
func (v SemVer) GitTag() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Release returns the current and next semantic versions for a release
func (v SemVer) Release(segment Segment) (SemVer, SemVer) {
	switch segment {
	case Patch:
		return SemVer{
				Major: v.Major,
				Minor: v.Minor,
				Patch: v.Patch,
			}, SemVer{
				Major: v.Major,
				Minor: v.Minor,
				Patch: v.Patch + 1,
			}
	case Minor:
		return SemVer{
				Major: v.Major,
				Minor: v.Minor + 1,
				Patch: 0,
			}, SemVer{
				Major: v.Major,
				Minor: v.Minor + 1,
				Patch: 1,
			}
	case Major:
		return SemVer{
				Major: v.Major + 1,
				Minor: 0,
				Patch: 0,
			}, SemVer{
				Major: v.Major + 1,
				Minor: 0,
				Patch: 1,
			}
	default:
		return SemVer{}, SemVer{}
	}
}
