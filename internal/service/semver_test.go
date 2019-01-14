package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSemVer(t *testing.T) {
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
			semver, err := ParseSemVer(tc.version)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSemver, semver)
			}
		})
	}
}

func TestSemVerVersion(t *testing.T) {
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

func TestSemVerGitTag(t *testing.T) {
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

func TestSemVerPreRelease(t *testing.T) {
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

func TestSemVerReleasePatch(t *testing.T) {
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

func TestSemVerReleaseMinor(t *testing.T) {
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

func TestSemVerReleaseMajor(t *testing.T) {
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

func TestNewVersionManager(t *testing.T) {
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

func TestNewTextVersionManager(t *testing.T) {
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

func TestTextVersionManagerRead(t *testing.T) {
	tests := []struct {
		name               string
		mockVersionFile    bool
		versionFileContent string
		expectedError      string
		expectedSemVer     SemVer
	}{
		{
			name:          "NoFile",
			expectedError: "no such file or directory",
		},
		{
			name:               "EmptyFile",
			mockVersionFile:    true,
			versionFileContent: "",
			expectedError:      "empty version file",
		},
		{
			name:               "InvalidVersion",
			mockVersionFile:    true,
			versionFileContent: "1.0",
			expectedError:      "invalid semantic version",
		},
		{
			name:               "Success",
			mockVersionFile:    true,
			versionFileContent: "0.1.0-0",
			expectedSemVer:     SemVer{Major: 0, Minor: 1, Patch: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &textVersionManager{}

			if tc.mockVersionFile {
				file, remove, err := createTempFile(tc.versionFileContent)
				assert.NoError(t, err)
				defer remove()
				manager.file = file
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

func TestTextVersionManagerUpdate(t *testing.T) {
	tests := []struct {
		name               string
		mockVersionFile    bool
		versionFileContent string
		newVersion         string
		expectedError      string
	}{
		{
			name:            "NoFile",
			mockVersionFile: false,
			newVersion:      "0.1.0",
			expectedError:   "no such file or directory",
		},
		{
			name:               "Success",
			mockVersionFile:    true,
			versionFileContent: "0.1.0-0",
			newVersion:         "0.1.0",
			expectedError:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &textVersionManager{}

			if tc.mockVersionFile {
				file, remove, err := createTempFile(tc.versionFileContent)
				assert.NoError(t, err)
				defer remove()
				manager.file = file
			}

			err := manager.Update(tc.newVersion)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewJSONVersionManager(t *testing.T) {
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

func TestJSONVersionManagerRead(t *testing.T) {
	tests := []struct {
		name               string
		mockVersionFile    bool
		versionFileContent string
		expectedError      string
		expectedSemVer     SemVer
	}{
		{
			name:            "NoFile",
			mockVersionFile: false,
			expectedError:   "no such file or directory",
		},
		{
			name:               "EmptyFile",
			mockVersionFile:    true,
			versionFileContent: ``,
			expectedError:      "unexpected end of JSON input",
		},
		{
			name:               "NoVersionKey",
			mockVersionFile:    true,
			versionFileContent: `{ "name": "node-service" }`,
			expectedError:      "no version key",
		},
		{
			name:               "BadVersionKey",
			mockVersionFile:    true,
			versionFileContent: `{ "name": "node-service", "version": 1 }`,
			expectedError:      "bad version key",
		},
		{
			name:               "InvalidVersionKey",
			mockVersionFile:    true,
			versionFileContent: `{ "name": "node-service", "version": "1.0" }`,
			expectedError:      "invalid semantic version",
		},
		{
			name:               "Success",
			mockVersionFile:    true,
			versionFileContent: `{ "name": "node-service", "version": "0.1.0-0" }`,
			expectedSemVer:     SemVer{Major: 0, Minor: 1, Patch: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &jsonVersionManager{}

			if tc.mockVersionFile {
				file, remove, err := createTempFile(tc.versionFileContent)
				assert.NoError(t, err)
				defer remove()
				manager.file = file
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

func TestJSONVersionManagerUpdate(t *testing.T) {
	tests := []struct {
		name               string
		mockVersionFile    bool
		versionFileContent string
		newVersion         string
		expectedError      string
	}{
		{
			name:            "NoFile",
			mockVersionFile: false,
			newVersion:      "0.1.0",
			expectedError:   "no such file or directory",
		},
		{
			name:               "Success",
			mockVersionFile:    true,
			versionFileContent: `{ "name": "node-service", "version": "0.1.0-0" }`,
			newVersion:         "0.1.0",
			expectedError:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := &jsonVersionManager{}

			if tc.mockVersionFile {
				file, remove, err := createTempFile(tc.versionFileContent)
				assert.NoError(t, err)
				defer remove()
				manager.file = file
			}

			err := manager.Update(tc.newVersion)

			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
