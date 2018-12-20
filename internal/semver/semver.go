package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// SemVer represents a semantic versioning
type SemVer struct {
	Major int64
	Minor int64
	Patch int64
}

// Parse reads a semantic version string and returns a SemVer
func Parse(version string) (semver SemVer, err error) {
	re := regexp.MustCompile("[.+-]")
	comps := re.Split(version, -1)
	if len(comps) < 3 {
		err = errors.New("invalid semantic version")
		return
	}

	// Major version number
	semver.Major, err = strconv.ParseInt(comps[0], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid major version: %v", err)
		return
	}

	// Minor version number
	semver.Minor, err = strconv.ParseInt(comps[1], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid minor version: %v", err)
		return
	}

	// Patch version number
	semver.Patch, err = strconv.ParseInt(comps[2], 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid patch version: %v", err)
		return
	}

	return
}

// Version returns a string representation of semantic version
func (v SemVer) Version() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// GitTag returns a string representation of semantic version to be used as a git tag
func (v SemVer) GitTag() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// PreRelease returns a string representation of semantic version to be used as a prelease version
func (v SemVer) PreRelease() string {
	return fmt.Sprintf("%d.%d.%d-0", v.Major, v.Minor, v.Patch)
}

// ReleasePatch returns the current and next semantic versions for a patch release
func (v SemVer) ReleasePatch() (current SemVer, next SemVer) {
	current = SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}

	next = SemVer{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch + 1,
	}

	return
}

// ReleaseMinor returns the current and next semantic versions for a minor release
func (v SemVer) ReleaseMinor() (current SemVer, next SemVer) {
	current = SemVer{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 0,
	}

	next = SemVer{
		Major: v.Major,
		Minor: v.Minor + 1,
		Patch: 1,
	}

	return
}

// ReleaseMajor returns the current and next semantic versions for a major release
func (v SemVer) ReleaseMajor() (current SemVer, next SemVer) {
	current = SemVer{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 0,
	}

	next = SemVer{
		Major: v.Major + 1,
		Minor: 0,
		Patch: 1,
	}

	return
}
