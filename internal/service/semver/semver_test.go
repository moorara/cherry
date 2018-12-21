package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestVersion(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedVersion string
	}{
		{
			SemVer{},
			"0.0.0",
		},
		{
			SemVer{0, 1, 0},
			"0.1.0",
		},
		{
			SemVer{1, 2, 0},
			"1.2.0",
		},
	}

	for _, tc := range tests {
		version := tc.semver.Version()
		assert.Equal(t, tc.expectedVersion, version)
	}
}

func TestGitTag(t *testing.T) {
	tests := []struct {
		semver         SemVer
		expectedGitTag string
	}{
		{
			SemVer{},
			"v0.0.0",
		},
		{
			SemVer{0, 1, 0},
			"v0.1.0",
		},
		{
			SemVer{1, 2, 0},
			"v1.2.0",
		},
	}

	for _, tc := range tests {
		gitTag := tc.semver.GitTag()
		assert.Equal(t, tc.expectedGitTag, gitTag)
	}
}

func TestPreRelease(t *testing.T) {
	tests := []struct {
		semver             SemVer
		expectedPreRelease string
	}{
		{
			SemVer{},
			"0.0.0-0",
		},
		{
			SemVer{0, 1, 0},
			"0.1.0-0",
		},
		{
			SemVer{1, 2, 0},
			"1.2.0-0",
		},
	}

	for _, tc := range tests {
		preRelease := tc.semver.PreRelease()
		assert.Equal(t, tc.expectedPreRelease, preRelease)
	}
}

func TestReleasePatch(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedCurrent SemVer
		expectedNext    SemVer
	}{
		{
			SemVer{},
			SemVer{0, 0, 0},
			SemVer{0, 0, 1},
		},
		{
			SemVer{0, 1, 0},
			SemVer{0, 1, 0},
			SemVer{0, 1, 1},
		},
		{
			SemVer{1, 2, 0},
			SemVer{1, 2, 0},
			SemVer{1, 2, 1},
		},
	}

	for _, tc := range tests {
		current, next := tc.semver.ReleasePatch()
		assert.Equal(t, tc.expectedCurrent, current)
		assert.Equal(t, tc.expectedNext, next)
	}
}

func TestReleaseMinor(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedCurrent SemVer
		expectedNext    SemVer
	}{
		{
			SemVer{},
			SemVer{0, 1, 0},
			SemVer{0, 1, 1},
		},
		{
			SemVer{0, 1, 0},
			SemVer{0, 2, 0},
			SemVer{0, 2, 1},
		},
		{
			SemVer{1, 2, 0},
			SemVer{1, 3, 0},
			SemVer{1, 3, 1},
		},
	}

	for _, tc := range tests {
		current, next := tc.semver.ReleaseMinor()
		assert.Equal(t, tc.expectedCurrent, current)
		assert.Equal(t, tc.expectedNext, next)
	}
}

func TestReleaseMajor(t *testing.T) {
	tests := []struct {
		semver          SemVer
		expectedCurrent SemVer
		expectedNext    SemVer
	}{
		{
			SemVer{},
			SemVer{1, 0, 0},
			SemVer{1, 0, 1},
		},
		{
			SemVer{0, 1, 0},
			SemVer{1, 0, 0},
			SemVer{1, 0, 1},
		},
		{
			SemVer{1, 2, 0},
			SemVer{2, 0, 0},
			SemVer{2, 0, 1},
		},
	}

	for _, tc := range tests {
		current, next := tc.semver.ReleaseMajor()
		assert.Equal(t, tc.expectedCurrent, current)
		assert.Equal(t, tc.expectedNext, next)
	}
}
