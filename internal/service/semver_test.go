package service

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

func TestNewManager(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		expectedError string
		expectedType  VersionManager
	}{
		{
			name:          "Uknown",
			file:          "version.yaml",
			expectedError: "unknown version file",
		},
		{
			name:         "TextManager",
			file:         "VERSION",
			expectedType: &textVersionManager{},
		},
		{
			name:         "JSONManager",
			file:         "package.json",
			expectedType: &jsonVersionManager{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager, err := NewVersionManager(tc.file)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tc.expectedType, manager)
			}
		})
	}
}

func TestNewTextManager(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "OK",
			file: "VERSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewTextVersionManager(tc.file)
			assert.NotNil(t, manager)
		})
	}
}

func TestTextManagerRead(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		expectedError  string
		expectedSemVer SemVer
	}{
		{
			name:          "NoFile",
			file:          "test/null",
			expectedError: "no such file or directory",
		},
		{
			name:          "EmptyFile",
			file:          "test/VERSION-empty",
			expectedError: "empty version file",
		},
		{
			name:          "InvalidVersion",
			file:          "test/VERSION-invalid",
			expectedError: "invalid semantic version",
		},
		{
			name:           "Success",
			file:           "test/VERSION",
			expectedSemVer: SemVer{Major: 0, Minor: 1, Patch: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &textVersionManager{
				file: tc.file,
			}

			semver, err := manager.Read()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, semver)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, semver, tc.expectedSemVer)
			}
		})
	}
}

func TestTextManagerUpdate(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		version       string
		expectedError string
	}{
		{
			name:          "BadFile",
			file:          "/test/null",
			version:       "0.2.0",
			expectedError: "no such file or directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &textVersionManager{
				file: tc.file,
			}

			err := manager.Update(tc.version)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewJSONManager(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "OK",
			file: "VERSION",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewJSONVersionManager(tc.file)
			assert.NotNil(t, manager)
		})
	}
}

func TestJSONManagerRead(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		expectedError  string
		expectedSemVer SemVer
	}{
		{
			name:          "NoFile",
			file:          "test/null",
			expectedError: "no such file or directory",
		},
		{
			name:          "EmptyFile",
			file:          "test/package-empty.json",
			expectedError: "unexpected end of JSON input",
		},
		{
			name:          "NoVersionKey",
			file:          "test/package-no-version.json",
			expectedError: "no version key",
		},
		{
			name:          "BadVersionKey",
			file:          "test/package-bad-version.json",
			expectedError: "bad version key",
		},
		{
			name:          "InvalidVersionKey",
			file:          "test/package-invalid-version.json",
			expectedError: "invalid semantic version",
		},
		{
			name:           "Success",
			file:           "test/package.json",
			expectedSemVer: SemVer{Major: 0, Minor: 1, Patch: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &jsonVersionManager{
				file: tc.file,
			}

			semver, err := manager.Read()

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Empty(t, semver)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, semver, tc.expectedSemVer)
			}
		})
	}
}

func TestJSONManagerUpdate(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		version       string
		expectedError string
	}{
		{
			name:          "BadFile",
			file:          "/test/null",
			version:       "0.2.0",
			expectedError: "no such file or directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &jsonVersionManager{
				file: tc.file,
			}

			err := manager.Update(tc.version)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
