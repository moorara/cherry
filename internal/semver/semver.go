package semver

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

// contextKey is the type for the keys added to context.
type contextKey string

const segmentKey = contextKey("ReleaseSegment")

// ContextWithSegment adds a segment to a context.
func ContextWithSegment(ctx context.Context, segment Segment) context.Context {
	return context.WithValue(ctx, segmentKey, segment)
}

// SegmentFromContext retrieves a segment from a context.
// If no segment found on context, a default segment (patch) will be returned.
func SegmentFromContext(ctx context.Context) Segment {
	segment, ok := ctx.Value(segmentKey).(Segment)
	if ok {
		return segment
	}

	return Patch
}

// SemVer represents a semantic versioning
type SemVer struct {
	Major uint
	Minor uint
	Patch uint
}

// Parse reads a semantic version string and returns a SemVer.
func Parse(version string) (SemVer, error) {
	re := regexp.MustCompile("[.+-]")
	comps := re.Split(version, -1)
	if len(comps) < 3 {
		return SemVer{}, errors.New("invalid semantic version")
	}

	// Major version number
	major, err := strconv.ParseUint(comps[0], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid major version: %v", err)
	}

	// Minor version number
	minor, err := strconv.ParseUint(comps[1], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid minor version: %v", err)
	}

	// Patch version number
	patch, err := strconv.ParseUint(comps[2], 10, 64)
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid patch version: %v", err)
	}

	return SemVer{
		Major: uint(major),
		Minor: uint(minor),
		Patch: uint(patch),
	}, nil
}

// Version returns a string representation of semantic version.
func (v SemVer) Version() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// GitTag returns a string representation of semantic version to be used as a git tag.
func (v SemVer) GitTag() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// PreRelease returns a string representation of semantic version to be used as a prelease version.
func (v SemVer) PreRelease() string {
	return fmt.Sprintf("%d.%d.%d-0", v.Major, v.Minor, v.Patch)
}

// Release returns the current and next semantic versions for a release
func (v SemVer) Release(segment Segment) (SemVer, SemVer) {
	switch segment {
	case Patch:
		return SemVer{v.Major, v.Minor, v.Patch}, SemVer{v.Major, v.Minor, v.Patch + 1}
	case Minor:
		return SemVer{v.Major, v.Minor + 1, 0}, SemVer{v.Major, v.Minor + 1, 1}
	case Major:
		return SemVer{v.Major + 1, 0, 0}, SemVer{v.Major + 1, 0, 1}
	default:
		return SemVer{}, SemVer{}
	}
}
