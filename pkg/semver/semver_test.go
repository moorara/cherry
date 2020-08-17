package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		semver         string
		expectedError  string
		expectedSemver SemVer
	}{
		{
			name:          "Empty",
			semver:        "",
			expectedError: "invalid semantic version: ",
		},
		{
			name:          "NoMinor",
			semver:        "1",
			expectedError: "invalid semantic version: 1",
		},
		{
			name:          "NoPatch",
			semver:        "0.1",
			expectedError: "invalid semantic version: 0.1",
		},
		{
			name:          "InvalidMajor",
			semver:        "X.1.0",
			expectedError: "invalid semantic version: X.1.0",
		},
		{
			name:          "InvalidMinor",
			semver:        "0.Y.0",
			expectedError: "invalid semantic version: 0.Y.0",
		},
		{
			name:          "InvalidPatch",
			semver:        "0.1.Z",
			expectedError: "invalid semantic version: 0.1.Z",
		},
		{
			name:          "InvalidPrerelease",
			semver:        "0.1.0-",
			expectedError: "invalid semantic version: 0.1.0-",
		},
		{
			name:          "InvalidPrerelease",
			semver:        "0.1.0-beta.",
			expectedError: "invalid semantic version: 0.1.0-beta.",
		},
		{
			name:          "InvalidMetadata",
			semver:        "0.1.0-beta.1+",
			expectedError: "invalid semantic version: 0.1.0-beta.1+",
		},
		{
			name:          "InvalidMetadata",
			semver:        "0.1.0-beta.1+20200818.",
			expectedError: "invalid semantic version: 0.1.0-beta.1+20200818.",
		},
		{
			name:   "OK",
			semver: "0.1.0",
			expectedSemver: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
		},
		{
			name:   "WithPrerelease",
			semver: "0.1.0-beta",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"beta"},
			},
		},
		{
			name:   "WithPrerelease",
			semver: "0.1.0-rc.1",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "1"},
			},
		},
		{
			name:   "WithMetadata",
			semver: "0.1.0+20191006",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"20191006"},
			},
		},
		{
			name:   "WithMetadata",
			semver: "0.1.0+sha.aabbccd",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"sha", "aabbccd"},
			},
		},
		{
			name:   "WithPrereleaseAndMetadata",
			semver: "0.1.0-rc.2+sha.aabbccd.20191006",
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
			semver, err := Parse(tc.semver)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSemver, semver)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name            string
		semver          SemVer
		expectedVersion string
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 4,
			},
			expectedVersion: "1.2.4",
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVersion, tc.semver.String())
		})
	}
}

func TestNext(t *testing.T) {
	tests := []struct {
		semver       SemVer
		expectedNext SemVer
	}{
		{
			SemVer{},
			SemVer{Major: 0, Minor: 0, Patch: 1},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			SemVer{Major: 0, Minor: 1, Patch: 1},
		},
		{
			SemVer{Major: 1, Minor: 0, Patch: 0},
			SemVer{Major: 1, Minor: 0, Patch: 1},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedNext, tc.semver.Next())
	}
}

func TestRelease(t *testing.T) {
	tests := []struct {
		semver          SemVer
		version         Version
		expectedRelease SemVer
	}{
		{
			SemVer{},
			Patch,
			SemVer{Major: 0, Minor: 0, Patch: 0},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			Patch,
			SemVer{Major: 0, Minor: 1, Patch: 0},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			Patch,
			SemVer{Major: 1, Minor: 2, Patch: 0},
		},
		{
			SemVer{},
			Minor,
			SemVer{Major: 0, Minor: 1, Patch: 0},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			Minor,
			SemVer{Major: 0, Minor: 2, Patch: 0},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			Minor,
			SemVer{Major: 1, Minor: 3, Patch: 0},
		},
		{
			SemVer{},
			Major,
			SemVer{Major: 1, Minor: 0, Patch: 0},
		},
		{
			SemVer{Major: 0, Minor: 1, Patch: 0},
			Major,
			SemVer{Major: 1, Minor: 0, Patch: 0},
		},
		{
			SemVer{Major: 1, Minor: 2, Patch: 0},
			Major,
			SemVer{Major: 2, Minor: 0, Patch: 0},
		},
		{
			SemVer{},
			Version(-1),
			SemVer{},
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedRelease, tc.semver.Release(tc.version))
	}
}
