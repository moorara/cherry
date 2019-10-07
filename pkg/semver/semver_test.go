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
			name:          "NoMinor",
			version:       "1",
			expectedError: "invalid semantic version",
		},
		{
			name:          "NoPatch",
			version:       "0.1",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidMajor",
			version:       "X.1.0",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidMinor",
			version:       "0.Y.0",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidPatch",
			version:       "0.1.Z",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidPrerelease",
			version:       "0.1.0-",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidPrerelease",
			version:       "0.1.0-beta.",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidMetadata",
			version:       "0.1.0-beta.1+",
			expectedError: "invalid semantic version",
		},
		{
			name:          "InvalidMetadata",
			version:       "0.1.0-beta.1+20191006.",
			expectedError: "invalid semantic version",
		},
		{
			name:    "OK",
			version: "0.1.0",
			expectedSemver: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
		},
		{
			name:    "WithPrerelease",
			version: "0.1.0-beta",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"beta"},
			},
		},
		{
			name:    "WithPrerelease",
			version: "0.1.0-rc.1",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "1"},
			},
		},
		{
			name:    "WithMetadata",
			version: "0.1.0+20191006",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"20191006"},
			},
		},
		{
			name:    "WithMetadata",
			version: "0.1.0+sha.aabbccd",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"sha", "aabbccd"},
			},
		},
		{
			name:    "WithPrereleaseAndMetadata",
			version: "0.1.0-rc.2+sha.aabbccd.20191006",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "2"},
				Metadata:   []string{"sha", "aabbccd", "20191006"},
			},
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
		name            string
		semver          SemVer
		expectedVersion string
		expectedGitTag  string
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 4,
			},
			expectedVersion: "1.2.4",
			expectedGitTag:  "v1.2.4",
		},
		{
			name: "WithPrerelease",
			semver: SemVer{
				Major:      1,
				Minor:      2,
				Patch:      4,
				Prerelease: []string{"rc", "1"},
			},
			expectedVersion: "1.2.4-rc.1",
			expectedGitTag:  "v1.2.4",
		},
		{
			name: "WithMetadata",
			semver: SemVer{
				Major:    1,
				Minor:    2,
				Patch:    4,
				Metadata: []string{"sha", "aabbccd"},
			},
			expectedVersion: "1.2.4+sha.aabbccd",
			expectedGitTag:  "v1.2.4",
		},
		{
			name: "WithPrereleaseAndMetadata",
			semver: SemVer{
				Major:      1,
				Minor:      2,
				Patch:      4,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"sha", "aabbccd"},
			},
			expectedVersion: "1.2.4-rc.1+sha.aabbccd",
			expectedGitTag:  "v1.2.4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVersion, tc.semver.Version())
			assert.Equal(t, tc.expectedGitTag, tc.semver.GitTag())
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
			SemVer{Major: 0, Minor: 0, Patch: 0},
			SemVer{Major: 0, Minor: 0, Patch: 1},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			Patch,
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			Patch,
			SemVer{Major: 1, Minor: 2, Patch: 0},
			SemVer{Major: 1, Minor: 2, Patch: 1},
		},
		{
			SemVer{},
			Minor,
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			Minor,
			SemVer{Major: 0, Minor: 2, Patch: 0},
			SemVer{Major: 0, Minor: 2, Patch: 1},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			Minor,
			SemVer{Major: 1, Minor: 3, Patch: 0},
			SemVer{Major: 1, Minor: 3, Patch: 1},
		},
		{
			SemVer{},
			Major,
			SemVer{Major: 1, Minor: 0, Patch: 0},
			SemVer{Major: 1, Minor: 0, Patch: 1},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			Major,
			SemVer{Major: 1, Minor: 0, Patch: 0},
			SemVer{Major: 1, Minor: 0, Patch: 1},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			Major,
			SemVer{Major: 2, Minor: 0, Patch: 0},
			SemVer{Major: 2, Minor: 0, Patch: 1},
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
