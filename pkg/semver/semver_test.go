package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		semver         string
		expectedSemver SemVer
		expectedOK     bool
	}{
		{
			name:           "Empty",
			semver:         "",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "NoMinor",
			semver:         "1",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "NoPatch",
			semver:         "0.1",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMajor",
			semver:         "X.1.0",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMinor",
			semver:         "0.Y.0",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidPatch",
			semver:         "0.1.Z",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidPrerelease",
			semver:         "0.1.0-",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidPrerelease",
			semver:         "0.1.0-beta.",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMetadata",
			semver:         "0.1.0-beta.1+",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:           "InvalidMetadata",
			semver:         "0.1.0-beta.1+20200818.",
			expectedSemver: SemVer{},
			expectedOK:     false,
		},
		{
			name:   "Release",
			semver: "0.1.0",
			expectedSemver: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
			expectedOK: true,
		},
		{
			name:   "Release",
			semver: "v0.1.0",
			expectedSemver: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 0,
			},
			expectedOK: true,
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
			expectedOK: true,
		},
		{
			name:   "WithPrerelease",
			semver: "v0.1.0-rc.1",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "1"},
			},
			expectedOK: true,
		},
		{
			name:   "WithMetadata",
			semver: "0.1.0+20200820",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"20200820"},
			},
			expectedOK: true,
		},
		{
			name:   "WithMetadata",
			semver: "v0.1.0+sha.abcdeff",
			expectedSemver: SemVer{
				Major:    0,
				Minor:    1,
				Patch:    0,
				Metadata: []string{"sha", "abcdeff"},
			},
			expectedOK: true,
		},
		{
			name:   "WithPrereleaseAndMetadata",
			semver: "0.1.0-beta+20200820",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"beta"},
				Metadata:   []string{"20200820"},
			},
			expectedOK: true,
		},
		{
			name:   "WithPrereleaseAndMetadata",
			semver: "v0.1.0-rc.1+sha.abcdeff.20200820",
			expectedSemver: SemVer{
				Major:      0,
				Minor:      1,
				Patch:      0,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"sha", "abcdeff", "20200820"},
			},
			expectedOK: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			semver, ok := Parse(tc.semver)

			assert.Equal(t, tc.expectedSemver, semver)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestAddPrerelease(t *testing.T) {
	tests := []struct {
		name           string
		semver         SemVer
		identifiers    []string
		expectedSemVer SemVer
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 0,
				Minor: 2,
				Patch: 7,
			},
			identifiers: []string{"rc", "1"},
			expectedSemVer: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
			},
		},
		{
			name: "WithPrerelease",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff"},
			},
			identifiers: []string{"rc", "1"},
			expectedSemVer: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff", "rc", "1"},
			},
		},
		{
			name: "WithMetadata",
			semver: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"20200820"},
			},
			identifiers: []string{"rc", "1"},
			expectedSemVer: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"20200820"},
			},
		},
		{
			name: "WithPrereleaseAndMetadata",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff"},
				Metadata:   []string{"20200820"},
			},
			identifiers: []string{"rc", "1"},
			expectedSemVer: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff", "rc", "1"},
				Metadata:   []string{"20200820"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.semver.AddPrerelease(tc.identifiers...)
			assert.Equal(t, tc.expectedSemVer, tc.semver)
		})
	}
}

func TestAddMetadata(t *testing.T) {
	tests := []struct {
		name           string
		semver         SemVer
		identifiers    []string
		expectedSemVer SemVer
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 0,
				Minor: 2,
				Patch: 7,
			},
			identifiers: []string{"163000"},
			expectedSemVer: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"163000"},
			},
		},
		{
			name: "WithPrerelease",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff"},
			},
			identifiers: []string{"163000"},
			expectedSemVer: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff"},
				Metadata:   []string{"163000"},
			},
		},
		{
			name: "WithMetadata",
			semver: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"20200820"},
			},
			identifiers: []string{"163000"},
			expectedSemVer: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"20200820", "163000"},
			},
		},
		{
			name: "WithPrereleaseAndMetadata",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff"},
				Metadata:   []string{"20200820"},
			},
			identifiers: []string{"163000"},
			expectedSemVer: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"abcdeff"},
				Metadata:   []string{"20200820", "163000"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.semver.AddMetadata(tc.identifiers...)
			assert.Equal(t, tc.expectedSemVer, tc.semver)
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

func TestString(t *testing.T) {
	tests := []struct {
		name            string
		semver          SemVer
		expectedVersion string
	}{
		{
			name: "OK",
			semver: SemVer{
				Major: 0,
				Minor: 2,
				Patch: 7,
			},
			expectedVersion: "0.2.7",
		},
		{
			name: "WithPrerelease",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
			},
			expectedVersion: "0.2.7-rc.1",
		},
		{
			name: "WithMetadata",
			semver: SemVer{
				Major:    0,
				Minor:    2,
				Patch:    7,
				Metadata: []string{"20200820"},
			},
			expectedVersion: "0.2.7+20200820",
		},
		{
			name: "WithPrereleaseAndMetadata",
			semver: SemVer{
				Major:      0,
				Minor:      2,
				Patch:      7,
				Prerelease: []string{"rc", "1"},
				Metadata:   []string{"20200820"},
			},
			expectedVersion: "0.2.7-rc.1+20200820",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedVersion, tc.semver.String())
		})
	}
}
