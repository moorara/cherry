package semver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConext(t *testing.T) {
	tests := []struct {
		name            string
		segment         Segment
		expectedSegment Segment
	}{
		{
			name:            "Patch",
			segment:         Patch,
			expectedSegment: Patch,
		},
		{
			name:            "Minor",
			segment:         Minor,
			expectedSegment: Minor,
		},
		{
			name:            "Major",
			segment:         Major,
			expectedSegment: Major,
		},
		{
			name:            "Default",
			segment:         Segment(-1),
			expectedSegment: Patch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			if tc.segment >= 0 {
				ctx = ContextWithSegment(ctx, tc.segment)
			}

			segment := SegmentFromContext(ctx)

			assert.Equal(t, tc.expectedSegment, segment)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		expectedError  string
		expectedSemver SemVer
	}{
		{
			name:          "Empty",
			version:       "",
			expectedError: "invalid semantic version",
		},
		{
			name:          "NoPatch",
			version:       "0.1",
			expectedError: "invalid semantic version",
		},
		{
			name:          "NoMinor",
			version:       "1",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidPatch",
			version:       "0.1.Z",
			expectedError: "invalid patch version",
		},
		{
			name:          "InvalidMinor",
			version:       "0.Y.0",
			expectedError: "invalid minor version",
		},
		{
			name:          "InvalidMajor",
			version:       "X.1.0",
			expectedError: "invalid major version",
		},
		{
			name:           "Valid",
			version:        "0.1.0",
			expectedSemver: SemVer{0, 1, 0},
		},
		{
			name:           "WithPrerelease",
			version:        "0.1.0-0",
			expectedSemver: SemVer{0, 1, 0},
		},
		{
			name:           "WithMetadata",
			version:        "0.1.0+2018",
			expectedSemver: SemVer{0, 1, 0},
		},
		{
			name:           "WithPrereleaseAndMetadata",
			version:        "0.1.0-0+2019",
			expectedSemver: SemVer{0, 1, 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			semver, err := Parse(tc.version)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSemver, semver)
			}
		})
	}
}

func TestSemVer(t *testing.T) {
	tests := []struct {
		name               string
		semver             SemVer
		expectedVersion    string
		expectedGitTag     string
		expectedPreRelease string
	}{
		{
			name:               "OK",
			semver:             SemVer{1, 2, 4},
			expectedVersion:    "1.2.4",
			expectedGitTag:     "v1.2.4",
			expectedPreRelease: "1.2.4-0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVersion, tc.semver.Version())
			assert.Equal(t, tc.expectedGitTag, tc.semver.GitTag())
			assert.Equal(t, tc.expectedPreRelease, tc.semver.PreRelease())
		})
	}
}

func TestRelease(t *testing.T) {
	tests := []struct {
		semver          SemVer
		segment         Segment
		expectedCurrent SemVer
		expectedNext    SemVer
	}{
		{
			SemVer{},
			Patch,
			SemVer{0, 0, 0},
			SemVer{0, 0, 1},
		},
		{
			SemVer{0, 1, 0},
			Patch,
			SemVer{0, 1, 0},
			SemVer{0, 1, 1},
		},
		{
			SemVer{1, 2, 0},
			Patch,
			SemVer{1, 2, 0},
			SemVer{1, 2, 1},
		},
		{
			SemVer{},
			Minor,
			SemVer{0, 1, 0},
			SemVer{0, 1, 1},
		},
		{
			SemVer{0, 1, 0},
			Minor,
			SemVer{0, 2, 0},
			SemVer{0, 2, 1},
		},
		{
			SemVer{1, 2, 0},
			Minor,
			SemVer{1, 3, 0},
			SemVer{1, 3, 1},
		},
		{
			SemVer{},
			Major,
			SemVer{1, 0, 0},
			SemVer{1, 0, 1},
		},
		{
			SemVer{0, 1, 0},
			Major,
			SemVer{1, 0, 0},
			SemVer{1, 0, 1},
		},
		{
			SemVer{1, 2, 0},
			Major,
			SemVer{2, 0, 0},
			SemVer{2, 0, 1},
		},
		{
			SemVer{},
			Segment(-1),
			SemVer{},
			SemVer{},
		},
	}

	for _, tc := range tests {
		current, next := tc.semver.Release(tc.segment)

		assert.Equal(t, tc.expectedCurrent, current)
		assert.Equal(t, tc.expectedNext, next)
	}
}
